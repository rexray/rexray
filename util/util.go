package util

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"
	apiversion "github.com/codedellemc/libstorage/api"
	"github.com/codedellemc/libstorage/api/context"
	apiserver "github.com/codedellemc/libstorage/api/server"
	apitypes "github.com/codedellemc/libstorage/api/types"

	"github.com/codedellemc/rexray/core"
)

const (
	logDirPathSuffix = "/var/log/rexray"
	etcDirPathSuffix = "/etc/rexray"
	binDirPathSuffix = "/usr/bin"
	runDirPathSuffix = "/var/run/rexray"
	libDirPathSuffix = "/var/lib/rexray"

	// UnitFilePath is the path to the SystemD service's unit file.
	UnitFilePath = "/etc/systemd/system/rexray.service"

	// InitFilePath is the path to the SystemV Service's init script.
	InitFilePath = "/etc/init.d/rexray"

	// EnvFileName is the name of the environment file used by the SystemD
	// service.
	EnvFileName = "rexray.env"
)

var (
	thisExeDir     string
	thisExeName    string
	thisExeAbsPath string

	prefix string

	binDirPath  string
	binFilePath string
	logDirPath  string
	libDirPath  string
	runDirPath  string
	etcDirPath  string
	pidFilePath string
	spcFilePath string
)

func init() {
	prefix = os.Getenv("REXRAY_HOME")

	thisExeDir, thisExeName, thisExeAbsPath = gotil.GetThisPathParts()
}

// GetPrefix gets the root path to the REX-Ray data.
func GetPrefix() string {
	return prefix
}

// Prefix sets the root path to the REX-Ray data.
func Prefix(p string) {
	if p == "" || p == "/" {
		return
	}

	binDirPath = ""
	binFilePath = ""
	logDirPath = ""
	libDirPath = ""
	runDirPath = ""
	etcDirPath = ""
	pidFilePath = ""
	spcFilePath = ""

	prefix = p
}

// IsPrefixed returns a flag indicating whether or not a prefix value is set.
func IsPrefixed() bool {
	return !(prefix == "" || prefix == "/")
}

// Install executes the system install command.
func Install(args ...string) {
	exec.Command("install", args...).Run()
}

// InstallChownRoot executes the system install command and chowns the target
// to the root user and group.
func InstallChownRoot(args ...string) {
	a := []string{"-o", "0", "-g", "0"}
	for _, i := range args {
		a = append(a, i)
	}
	exec.Command("install", a...).Run()
}

// InstallDirChownRoot executes the system install command with a -d flag and
// chowns the target to the root user and group.
func InstallDirChownRoot(dirPath string) {
	InstallChownRoot("-d", dirPath)
}

// EtcDirPath returns the path to the REX-Ray etc directory.
func EtcDirPath() string {
	if etcDirPath == "" {
		etcDirPath = fmt.Sprintf("%s%s", prefix, etcDirPathSuffix)
		os.MkdirAll(etcDirPath, 0755)
	}
	return etcDirPath
}

// RunDirPath returns the path to the REX-Ray run directory.
func RunDirPath() string {
	if runDirPath == "" {
		runDirPath = fmt.Sprintf("%s%s", prefix, runDirPathSuffix)
		os.MkdirAll(runDirPath, 0755)
	}
	return runDirPath
}

// LogDirPath returns the path to the REX-Ray log directory.
func LogDirPath() string {
	if logDirPath == "" {
		logDirPath = fmt.Sprintf("%s%s", prefix, logDirPathSuffix)
		os.MkdirAll(logDirPath, 0755)
	}
	return logDirPath
}

// LibDirPath returns the path to the REX-Ray bin directory.
func LibDirPath() string {
	if libDirPath == "" {
		libDirPath = fmt.Sprintf("%s%s", prefix, libDirPathSuffix)
		os.MkdirAll(libDirPath, 0755)
	}
	return libDirPath
}

// LibFilePath returns the path to a file inside the REX-Ray lib directory
// with the provided file name.
func LibFilePath(fileName string) string {
	return path.Join(LibDirPath(), fileName)
}

// RunFilePath returns the path to a file inside the REX-Ray run directory
// with the provided file name.
func RunFilePath(fileName string) string {
	return path.Join(RunDirPath(), fileName)
}

// BinDirPath returns the path to the REX-Ray bin directory.
func BinDirPath() string {
	if binDirPath == "" {
		binDirPath = fmt.Sprintf("%s%s", prefix, binDirPathSuffix)
		os.MkdirAll(binDirPath, 0755)
	}
	return binDirPath
}

// PidFilePath returns the path to the REX-Ray PID file.
func PidFilePath() string {
	if pidFilePath == "" {
		pidFilePath = RunFilePath("rexray.pid")
	}
	return pidFilePath
}

// SpecFilePath returns the path to the REX-Ray spec file.
func SpecFilePath() string {
	if spcFilePath == "" {
		spcFilePath = RunFilePath("rexray.spec")
	}
	return spcFilePath
}

// BinFilePath returns the path to the REX-Ray executable.
func BinFilePath() string {
	if binFilePath == "" {
		binFilePath = path.Join(BinDirPath(), "rexray")
	}
	return binFilePath
}

// EtcFilePath returns the path to a file inside the REX-Ray etc directory
// with the provided file name.
func EtcFilePath(fileName string) string {
	return path.Join(EtcDirPath(), fileName)
}

// LogFilePath returns the path to a file inside the REX-Ray log directory
// with the provided file name.
func LogFilePath(fileName string) string {
	return path.Join(LogDirPath(), fileName)
}

// LogFile returns a writer to a file inside the REX-Ray log directory
// with the provided file name.
func LogFile(fileName string) (io.Writer, error) {
	return os.OpenFile(
		LogFilePath(fileName), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
}

// StdOutAndLogFile returns a mutltiplexed writer for the current process's
// stdout descriptor and a REX-Ray log file with the provided name.
func StdOutAndLogFile(fileName string) (io.Writer, error) {
	lf, lfErr := LogFile(fileName)
	if lfErr != nil {
		return nil, lfErr
	}
	return io.MultiWriter(os.Stdout, lf), nil
}

// WritePidFile writes the current process ID to the REX-Ray PID file.
func WritePidFile(pid int) error {
	if pid < 0 {
		pid = os.Getpid()
	}
	return gotil.WriteStringToFile(fmt.Sprintf("%d", pid), PidFilePath())
}

// ReadPidFile reads the REX-Ray PID from the PID file.
func ReadPidFile() (int, error) {
	pidStr, pidStrErr := gotil.ReadFileToString(PidFilePath())
	if pidStrErr != nil {
		return -1, pidStrErr
	}
	pid, atoiErr := strconv.Atoi(pidStr)
	if atoiErr != nil {
		return -1, atoiErr
	}
	return pid, nil
}

// WriteSpecFile writes the current host address to the REX-Ray spec file.
func WriteSpecFile(host string) error {
	return gotil.WriteStringToFile(host, SpecFilePath())
}

// ReadSpecFile reads the REX-Ray host address from the spec file.
func ReadSpecFile() (string, error) {
	host, err := gotil.ReadFileToString(SpecFilePath())
	if err != nil {
		return "", err
	}
	return gotil.Trim(host), nil
}

// PrintVersion prints the current version information to the provided writer.
func PrintVersion(out io.Writer) {
	fmt.Fprintln(out, "REX-Ray")
	fmt.Fprintln(out, "-------")
	fmt.Fprintf(out, "Binary: %s\n", thisExeAbsPath)
	fmt.Fprintf(out, "SemVer: %s\n", core.Version.SemVer)
	fmt.Fprintf(out, "OsArch: %s\n", core.Version.Arch)
	fmt.Fprintf(out, "Branch: %s\n", core.Version.Branch)
	fmt.Fprintf(out, "Commit: %s\n", core.Version.ShaLong)
	fmt.Fprintf(out, "Formed: %s\n\n",
		core.Version.BuildTimestamp.Format(time.RFC1123))

	fmt.Fprintln(out, "libStorage")
	fmt.Fprintln(out, "----------")
	fmt.Fprintf(out, "SemVer: %s\n", apiversion.Version.SemVer)
	fmt.Fprintf(out, "OsArch: %s\n", apiversion.Version.Arch)
	fmt.Fprintf(out, "Branch: %s\n", apiversion.Version.Branch)
	fmt.Fprintf(out, "Commit: %s\n", apiversion.Version.ShaLong)

	timestamp := apiversion.Version.BuildTimestamp.Format(time.RFC1123)
	fmt.Fprintf(out, "Formed: %s\n", timestamp)
}

// WaitUntilLibStorageStopped blocks until libStorage is stopped.
func WaitUntilLibStorageStopped(ctx apitypes.Context, errs <-chan error) {
	ctx.Debug("waiting until libStorage is stopped")

	// if there is no err channel then do not wait until libStorage is stopped
	// as the absence of the err channel means libStorage was not started in
	// embedded mode
	if errs == nil {
		ctx.Debug("done waiting on err chan; err chan is nil")
		return
	}

	// in a goroutine, range over the apiserver.Close channel until it's closed
	for range apiserver.Close() {
	}
	ctx.Debug("done sending close signals to libStorage")

	// block until the err channel is closed
	for err := range errs {
		if err == nil {
			continue
		}
		ctx.WithError(err).Error("error on closing libStorage server")
	}
	ctx.Debug("done waiting on err chan")
}

var localHostRX = regexp.MustCompile(
	`(?i)^(localhost|(?:127\.0\.0\.1))(?::(\d+))?$`)

func logHostSpec(ctx apitypes.Context, h, m string) {
	ctx.WithFields(log.Fields{
		"path": SpecFilePath(),
		"host": h,
	}).Debug(m)
}

// IsAddressActive returns a flag indicating whether or not a an address is
// responding to connection attempts. This does not validate whether the
// address is using TLS or such a connection is possible.
func IsAddressActive(proto, addr string) bool {
	dialer := &net.Dialer{Timeout: time.Second * 1}
	if _, err := dialer.Dial(proto, addr); err != nil {
		return false
	}
	return true
}

// IsLocalServerActive returns a flag indicating whether or not a local
// libStorage is already running.
func IsLocalServerActive(
	ctx apitypes.Context, config gofig.Config) (host string, running bool) {

	var (
		isLocal  bool
		specFile = SpecFilePath()
	)

	if gotil.FileExists(specFile) {
		if h, _ := ReadSpecFile(); h != "" {
			host = h
			logHostSpec(ctx, host, "read spec file")
			defer func() {
				if running || !isLocal {
					return
				}
				host = ""
				os.RemoveAll(specFile)
				ctx.WithField("specFile", specFile).Info(
					"removed invalid spec file")
			}()
		}
	}
	if host == "" {
		host = config.GetString(apitypes.ConfigHost)
	}
	if host == "" {
		return "", false
	}

	proto, addr, err := gotil.ParseAddress(host)
	if err != nil {
		return "", false
	}

	switch proto {
	case "unix":
		isLocal = true
		ctx.WithField("sock", addr).Debug("is local unix server active")
		var sockExists, isActive bool
		if sockExists = gotil.FileExists(addr); sockExists {
			if isActive = IsAddressActive(proto, addr); !isActive {
				os.RemoveAll(addr)
				ctx.WithField("sockFile", addr).Info(
					"removed invalid sock file")
			}
		}
		return host, isActive
	case "tcp":
		m := localHostRX.FindStringSubmatch(addr)
		if len(m) < 3 {
			return "", false
		}
		isLocal = true
		port, err := strconv.Atoi(m[2])
		if err != nil {
			return "", false
		}
		ctx.WithField("port", port).Debug("is local tcp server active")
		return host, IsAddressActive(proto, addr)
	}
	return "", false
}

func setHost(
	ctx apitypes.Context,
	config gofig.Config,
	host string) apitypes.Context {
	ctx = ctx.WithValue(context.HostKey, host)
	ctx.WithField("host", host).Debug("set host in context")
	config.Set(apitypes.ConfigHost, host)
	ctx.WithField("host", host).Debug("set host in config")
	return ctx
}

// ActivateLibStorage activates a libStorage server if conditions are met and
// returns a possibly mutated context.
func ActivateLibStorage(
	ctx apitypes.Context,
	config gofig.Config) (apitypes.Context, gofig.Config, <-chan error, error) {

	config = config.Scope("rexray")

	// set the `libstorage.service` property to the value of
	// `rexray.storageDrivers` if the former is not defined and the
	// latter is
	if !config.IsSet(apitypes.ConfigService) &&
		config.IsSet("rexray.storageDrivers") {
		if sd := config.GetStringSlice("rexray.storageDrivers"); len(sd) > 0 {
			config.Set(apitypes.ConfigService, sd[0])
		} else if sd := config.GetString("rexray.storageDrivers"); sd != "" {
			config.Set(apitypes.ConfigService, sd)
		}
	}

	if !config.IsSet(apitypes.ConfigIgVolOpsMountPath) {
		config.Set(apitypes.ConfigIgVolOpsMountPath, LibFilePath("volumes"))
	}

	var (
		host          string
		err           error
		isRunning     bool
		errs          chan error
		serverErrChan <-chan error
		server        apitypes.Server
	)

	if host = config.GetString(apitypes.ConfigHost); host != "" {
		if !config.GetBool(apitypes.ConfigEmbedded) {
			ctx.WithField(
				"host", host,
			).Debug("not starting embedded server; embedded mode disabled")
			return ctx, config, nil, nil
		}
	}

	if host, isRunning = IsLocalServerActive(ctx, config); isRunning {
		ctx = setHost(ctx, config, host)
		ctx.WithField("host", host).Debug(
			"not starting embedded server; already running")
		return ctx, config, nil, nil
	}

	// if no host was specified then see if a set of default services need to
	// be initialized
	if host == "" {
		if err = initDefaultLibStorageServices(ctx, config); err != nil {
			return ctx, config, nil, err
		}
	}

	ctx.Debug("starting embedded libStorage server")

	if server, serverErrChan, err = apiserver.Serve(ctx, config); err != nil {
		return ctx, config, nil, err
	}

	if host == "" {
		host = server.Addrs()[0]
		ctx.WithField("host", host).Debug("got host from new server address")
	}
	ctx = setHost(ctx, config, host)

	errs = make(chan error)
	go func() {
		for err := range serverErrChan {
			if err != nil {
				errs <- err
			}
		}
		if err := os.RemoveAll(SpecFilePath()); err == nil {
			logHostSpec(ctx, host, "removed spec file")
		}
		close(errs)
	}()

	// write the host to the spec file so that other rex-ray invocations can
	// find it, even if running as an embedded libStorage server
	if err := WriteSpecFile(host); err != nil {
		specFile := SpecFilePath()
		if os.IsPermission(err) {
			ctx.WithError(err).Errorf(
				"user does not have write permissions for %s", specFile)
		} else {
			ctx.WithError(err).Errorf(
				"error writing spec file at %s", specFile)
		}
		WaitUntilLibStorageStopped(ctx, serverErrChan)
		return ctx, config, errs, err
	}
	logHostSpec(ctx, host, "created spec file")

	return ctx, config, errs, nil
}
