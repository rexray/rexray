package core

import (
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/codedellemc/rexray/libstorage/api/context"
	apitypes "github.com/codedellemc/rexray/libstorage/api/types"
)

var (
	// Version of REX-Ray.
	Version *apitypes.VersionInfo

	// BuildType is the build type of this binary.
	BuildType = "client+agent+controller"

	// Debug is whether or not the REXRAY_DEBUG environment variable is set
	// to a truthy value.
	Debug, _ = strconv.ParseBool(os.Getenv("REXRAY_DEBUG"))
)

type osString string

func (o osString) String() string {
	switch o {
	case "linux":
		return "Linux"
	case "darwin":
		return "Darwin"
	case "windows":
		return "Windows"
	default:
		return string(o)
	}
}

type archString string

func (a archString) String() string {
	switch a {
	case "386":
		return "i386"
	case "amd64":
		return "x86_64"
	default:
		return string(a)
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
	signal.Notify(sigc)

	go func() {
		for s := range sigc {

			ctx := ctx.WithValue(signalContextKey, s.String())
			if ok, graceful := IsExitSignal(s); ok && !graceful {
				ctx.Error("received signal; aborting")
				os.Exit(1)
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

			if ok, graceful := IsExitSignal(s); ok && graceful {
				ctx.Error("received signal; shutting down")
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
