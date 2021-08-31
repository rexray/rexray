//go:generate go run semver/semver.go -f semver.tpl -o core_generated.go

package core

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
)

var (

	// DockerLegacyMode is true if Docker legacy mode is enabled.
	DockerLegacyMode bool

	// Arch is the uname OS-Arch string.
	Arch = fmt.Sprintf(
		"%s-%s",
		goosToUname[runtime.GOOS],
		goarchToUname[runtime.GOARCH])

	// SemVer is the semantic version.
	SemVer string

	// CommitSha7 is the short version of the commit hash from which
	// this program was built.
	CommitSha7 string

	// CommitSha32 is the long version of the commit hash from which
	// this program was built.
	CommitSha32 string

	// CommitTime is the commit timestamp of the commit from which
	// this program was built.
	CommitTime time.Time

	// BuildType is the build type of this binary.
	BuildType = "client+agent+controller"

	// Debug is whether or not the REXRAY_DEBUG environment variable is set
	// to a truthy value.
	Debug, _ = strconv.ParseBool(os.Getenv("REXRAY_DEBUG"))
)

func init() {
	if v, ok := os.LookupEnv("DOCKER_LEGACY"); ok {
		DockerLegacyMode, _ = strconv.ParseBool(v)
	} else {
		DockerLegacyMode = true
	}
}

// SignalHandlerFunc is a function that can be registered with
// `core.RegisterSignalHandler` to receive a callback when the process receives
// a signal.
type SignalHandlerFunc func(ctx apitypes.Context, s os.Signal)

var (
	sigHandlers    []SignalHandlerFunc
	sigHandlersRWL = &sync.RWMutex{}
)

// RegisterSignalHandler registers a SignalHandlerFunc.
func RegisterSignalHandler(f SignalHandlerFunc) {
	sigHandlersRWL.Lock()
	defer sigHandlersRWL.Unlock()
	sigHandlers = append(sigHandlers, f)
}

type signalContextKeyType int

const signalContextKey signalContextKeyType = 0

func (k signalContextKeyType) String() string {
	return "signal"
}

// TrapSignals tells the process to trap incoming process signals.
func TrapSignals(ctx apitypes.Context) {

	context.RegisterCustomKey(signalContextKey, context.CustomLoggerKey)

	sigc := make(chan os.Signal, 1)
	signal.Notify(
		sigc,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT)

	go func() {
		for s := range sigc {

			ctx := ctx.WithValue(signalContextKey, s.String())
			isExitSignal, isGraceful := IsExitSignal(s)

			if isExitSignal {
				if isGraceful {
					ctx.Info("received signal; shutting down")
				} else {
					ctx.Error("received signal; aborting")
					os.Exit(1)
				}
			}

			func() {
				sigHandlersRWL.RLock()
				defer sigHandlersRWL.RUnlock()

				// execute the signal handlers in reverse order. the first
				// one registered should be executed last as it was registered
				// the earliest
				for i := len(sigHandlers) - 1; i >= 0; i-- {
					sigHandlers[i](ctx, s)
				}
			}()

			if isExitSignal {
				os.Exit(0)
			}
		}
	}()
}

// IsExitSignal returns a flag indicating whether a signal is SIGKILL, SIGHUP,
// SIGINT, SIGTERM, or SIGQUIT. The second return value is whether it is a
// graceful exit. This flag is true for SIGTERM, SIGHUP, SIGINT, and SIGQUIT.
func IsExitSignal(s os.Signal) (bool, bool) {
	switch s {
	case syscall.SIGKILL:
		return true, false
	case syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT:
		return true, true
	default:
		return false, false
	}
}

var goosToUname = map[string]string{
	"android":   "Android",
	"darwin":    "Darwin",
	"dragonfly": "DragonFly",
	"freebsd":   "kFreeBSD",
	"linux":     "Linux",
	"nacl":      "NaCl",
	"netbsd":    "NetBSD",
	"openbsd":   "OpenBSD",
	"plan9":     "Plan9",
	"solaris":   "Solaris",
	"windows":   "Windows",
}

var goarchToUname = map[string]string{
	"386":      "i386",
	"amd64":    "x86_64",
	"amd64p32": "x86_64_P32",
	"arm":      "ARMv7",
	"arm64":    "ARMv8",
	"mips":     "MIPS32",
	"mips64":   "MIPS64",
	"mips64le": "MIPS64LE",
	"mipsle":   "MIPS32LE",
	"ppc64":    "PPC64",
	"ppc64le":  "PPC64LE",
	"s390x":    "s390x",
}
