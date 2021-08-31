package util

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"

	"github.com/AVENTER-UG/rexray/core"
)

var (

	// BinFileName is the name of the executing binary.
	BinFileName string

	// BinFilePath is the full path of the executing binary.
	BinFilePath string

	// BinFileDirPath is the full path of the executing binary's parent
	// directory.
	BinFileDirPath string

	// UnitFileName is the name of the SystemD service's unit file.
	UnitFileName string

	// UnitFilePath is the path to the SystemD service's unit file.
	UnitFilePath string

	// InitFileName is the name of the SystemV service's unit file.
	InitFileName string

	// InitFilePath is the path to the SystemV service's init script.
	InitFilePath string

	// PIDFileName is the name of the PID file.
	PIDFileName string

	// DotDirName is the name of the hidden app directory.
	DotDirName string
)

func init() {
	BinFileDirPath, BinFileName, BinFilePath = gotil.GetThisPathParts()
	UnitFileName = fmt.Sprintf("%s.service", BinFileName)
	UnitFilePath = path.Join("/etc/systemd/system", UnitFileName)
	InitFileName = BinFileName
	InitFilePath = path.Join("/etc/init.d", InitFileName)
	PIDFileName = fmt.Sprintf("%s.pid", BinFileName)

	DotDirName = fmt.Sprintf(".rexray")
}

// IsPrefixed determines whether REX-Ray is prefixed.
func IsPrefixed(ctx apitypes.Context) bool {
	pc := context.MustPathConfig(ctx)
	return !(pc.Home == "" || pc.Home == "/")
}

// GetPrefix gets the root path to the REX-Ray data.
func GetPrefix(ctx apitypes.Context) string {
	return context.MustPathConfig(ctx).Home
}

// ScriptDirPath returns the path to the REX-Ray script directory.
func ScriptDirPath(ctx apitypes.Context) string {
	pathConfig := context.MustPathConfig(ctx)
	v := path.Join(pathConfig.Lib, "scripts")
	if !gotil.FileExists(v) {
		os.MkdirAll(v, 0755)
	}
	return v
}

// ScriptFilePath returns the path to a file inside the REX-Ray script directory
// with the provided file name.
func ScriptFilePath(ctx apitypes.Context, fileName string) string {
	return path.Join(ScriptDirPath(ctx), fileName)
}

// EtcFilePath returns the path to a file inside the REX-Ray etc directory
// with the provided file name.
func EtcFilePath(ctx apitypes.Context, fileName string) string {
	return path.Join(context.MustPathConfig(ctx).Etc, fileName)
}

// RunFilePath returns the path to a file inside the REX-Ray run directory
// with the provided file name.
func RunFilePath(ctx apitypes.Context, fileName string) string {
	return path.Join(context.MustPathConfig(ctx).Run, fileName)
}

// PidFilePath returns the path to the REX-Ray PID file.
func PidFilePath(ctx apitypes.Context) string {
	return path.Join(
		context.MustPathConfig(ctx).Run,
		fmt.Sprintf("%s.pid", BinFileName))
}

// EnvFilePath returns the path to the REX-Ray env file.
func EnvFilePath(ctx apitypes.Context) string {
	return path.Join(
		context.MustPathConfig(ctx).Etc,
		fmt.Sprintf("%s.env", BinFileName))
}

// SpecFilePath returns the path to the REX-Ray spec file.
func SpecFilePath(ctx apitypes.Context) string {
	return RunFilePath(ctx, "rexray.spec")
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

// LogFilePath returns the path to a file inside the REX-Ray log directory
// with the provided file name.
func LogFilePath(ctx apitypes.Context, fileName string) string {
	return path.Join(context.MustPathConfig(ctx).Log, fileName)
}

// LogFile returns a writer to a file inside the REX-Ray log directory
// with the provided file name.
func LogFile(ctx apitypes.Context, fileName string) (io.Writer, error) {
	return os.OpenFile(
		LogFilePath(ctx, fileName),
		os.O_CREATE|os.O_APPEND|os.O_RDWR,
		0644)
}

// StdOutAndLogFile returns a mutltiplexed writer for the current process's
// stdout descriptor and a REX-Ray log file with the provided name.
func StdOutAndLogFile(ctx apitypes.Context, fileName string) (io.Writer, error) {
	lf, lfErr := LogFile(ctx, fileName)
	if lfErr != nil {
		return nil, lfErr
	}
	return io.MultiWriter(os.Stdout, lf), nil
}

// WritePidFile writes the current process ID to the REX-Ray PID file.
func WritePidFile(ctx apitypes.Context, pid int) error {
	if pid < 0 {
		pid = os.Getpid()
	}
	return gotil.WriteStringToFile(fmt.Sprintf("%d", pid), PidFilePath(ctx))
}

// ReadPidFile reads the REX-Ray PID from the PID file.
func ReadPidFile(ctx apitypes.Context) (int, error) {
	pidStr, pidStrErr := gotil.ReadFileToString(PidFilePath(ctx))
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
func WriteSpecFile(ctx apitypes.Context, host string) error {
	return gotil.WriteStringToFile(host, SpecFilePath(ctx))
}

// ReadSpecFile reads the REX-Ray host address from the spec file.
func ReadSpecFile(ctx apitypes.Context) (string, error) {
	host, err := gotil.ReadFileToString(SpecFilePath(ctx))
	if err != nil {
		return "", err
	}
	return gotil.Trim(host), nil
}

// PrintVersion prints the current version information to the provided writer.
func PrintVersion(out io.Writer) {
	fmt.Fprintln(out, "REX-Ray")
	fmt.Fprintln(out, "-------")
	fmt.Fprintf(out, "Binary: %s\n", BinFilePath)
	fmt.Fprintf(out, "Flavor: %s\n", core.BuildType)
	fmt.Fprintf(out, "SemVer: %s\n", core.SemVer)
	fmt.Fprintf(out, "OsArch: %s\n", core.Arch)
	fmt.Fprintf(out, "Commit: %s\n", core.CommitSha32)
	fmt.Fprintf(out, "Formed: %s\n",
		core.CommitTime.Format(time.RFC1123))
}

var localHostRX = regexp.MustCompile(
	`(?i)^(localhost|(?:127\.0\.0\.1))(?::(\d+))?$`)

func logHostSpec(ctx apitypes.Context, h, m string) {
	ctx.WithFields(log.Fields{
		"path": SpecFilePath(ctx),
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
		specFile = SpecFilePath(ctx)
	)

	if gotil.FileExists(specFile) {
		if h, _ := ReadSpecFile(ctx); h != "" {
			host = h
			host = parseSafeHost(ctx, host)
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

// ErrHostDetectionFailed is returned when ActivateLibStorage is invoked
// and the there is no host specified in the configuration file and
// no existing host was detected locally.
var ErrHostDetectionFailed = errors.New("host detection failed")

// ActivateLibStorage activates libStorage and returns a possibly mutated
// context.
func ActivateLibStorage(
	ctx apitypes.Context,
	config gofig.Config) (apitypes.Context, gofig.Config, <-chan error, error) {

	if config.GetBool("libstorage.disabled") {
		return ctx, config, nil, nil
	}

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
		config.Set(
			apitypes.ConfigIgVolOpsMountPath,
			path.Join(context.MustPathConfig(ctx).Lib, "volumes"))
	}

	var (
		err  error
		errs <-chan error
	)

	ctx, config, errs, err = activateLibStorage(ctx, config)
	if err != nil {
		return ctx, config, errs, err
	}

	return ctx, config, errs, nil
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

// WaitUntilLibStorageStopped blocks until libStorage is stopped.
func WaitUntilLibStorageStopped(ctx apitypes.Context, errs <-chan error) {
	waitUntilLibStorageStopped(ctx, errs)
}

// ErrMissingService occurs when the client configuration is
// missing the property "libstorage.service" either at the root
// or as part of a module definition.
var ErrMissingService = goof.New("client must specify service")

// NewClient returns a new libStorage client.
func NewClient(
	ctx apitypes.Context, config gofig.Config) (apitypes.Client, error) {

	if config.GetBool("libstorage.disabled") {
		return nil, nil
	}

	if v := config.Get(apitypes.ConfigService); v == "" {
		return nil, ErrMissingService
	}
	return newClient(ctx, config)
}

// NewConfig returns a new config object.
func NewConfig(ctx apitypes.Context) (config gofig.Config) {

	pathConfig := context.MustPathConfig(ctx)

	defer func() {
		if config != nil {
			config = config.Scope("rexray")
		}
	}()

	const cfgFileExt = "yml"

	loadConfig := func(
		allExists, usrExists, ignoreExists bool,
		allPath, usrPath, name string) (gofig.Config, bool) {

		fields := log.Fields{
			"buildType":              core.BuildType,
			"ignoreExists":           ignoreExists,
			"configFileName":         name,
			"globalConfigFilePath":   allPath,
			"userConfigFilePath":     usrPath,
			"globalConfigFileExists": allExists,
			"userConfigFileExists":   usrExists,
		}
		ctx.WithFields(fields).Debug("loading config")
		if ignoreExists {
			ctx.WithFields(fields).Debug("disabled config file exist check")
		} else if !allExists && !usrExists {
			ctx.WithFields(fields).Debug("cannot find global or user file")
			return nil, false
		}
		if allExists {
			ctx.WithFields(fields).Debug("validating global config")
			ValidateConfig(allPath)
		}
		if usrExists {
			ctx.WithFields(fields).Debug("validating user config")
			ValidateConfig(usrPath)
		}
		ctx.WithFields(fields).Debug("created new config")
		return gofigCore.NewConfig(
			allExists || ignoreExists,
			usrExists || ignoreExists,
			name, cfgFileExt), true
	}

	// load build-type specific config
	switch core.BuildType {
	case "client", "agent", "controller":
		var (
			fileName    = BinFileName
			fileNameExt = fileName + "." + cfgFileExt
			allFilePath = EtcFilePath(ctx, fileNameExt)
			usrFilePath = path.Join(pathConfig.UserHome, fileNameExt)
		)
		if config, ok := loadConfig(
			gotil.FileExists(allFilePath),
			gotil.FileExists(usrFilePath),
			false,
			allFilePath, usrFilePath,
			fileName); ok {
			return config
		}
	}

	// load config from rexray.yml?
	{
		var (
			fileName    = "rexray"
			fileNameExt = fileName + "." + cfgFileExt
			allFilePath = EtcFilePath(ctx, fileNameExt)
			usrFilePath = path.Join(pathConfig.UserHome, fileNameExt)
		)
		if config, ok := loadConfig(
			gotil.FileExists(allFilePath),
			gotil.FileExists(usrFilePath),
			false,
			allFilePath, usrFilePath,
			fileName); ok {
			return config
		}
	}

	// load default config
	{
		var (
			fileName    = "config"
			fileNameExt = fileName + "." + cfgFileExt
			allFilePath = EtcFilePath(ctx, fileNameExt)
			usrFilePath = path.Join(pathConfig.UserHome, fileNameExt)
		)
		config, _ := loadConfig(
			gotil.FileExists(allFilePath),
			gotil.FileExists(usrFilePath),
			true,
			allFilePath, usrFilePath,
			fileName)
		return config
	}
}

// ValidateConfig validates a provided configuration file.
func ValidateConfig(path string) {
	if !gotil.FileExists(path) {
		return
	}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "rexray: error reading config: %s\n%v\n", path, err)
		os.Exit(1)
	}
	s := string(buf)
	if _, err := gofigCore.ValidateYAMLString(s); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"rexray: invalid config: %s\n\n  %v\n\n", path, err)
		fmt.Fprint(
			os.Stderr,
			"paste the contents between ---BEGIN--- and ---END---\n")
		fmt.Fprint(
			os.Stderr,
			"into http://www.yamllint.com/ to discover the issue\n\n")
		fmt.Fprintln(os.Stderr, "---BEGIN---")
		fmt.Fprintln(os.Stderr, s)
		fmt.Fprintln(os.Stderr, "---END---")
		os.Exit(1)
	}
}

const tcp127 = "tcp://127.0.0.1"

var rxParseSafeHost = regexp.MustCompile(`^tcp://(?:(?:0\.0\.0\.0)|\*)?:(\d+)$`)

func parseSafeHost(ctx apitypes.Context, h string) string {
	if h == "" || h == "0.0.0.0" || h == "*" {
		ctx.WithFields(map[string]interface{}{
			"preParse":  h,
			"postParse": tcp127,
		}).Debug(`parseSafeHost - h == "" || h == "0.0.0.0" || h == "*"`)
		return tcp127
	}
	if m := rxParseSafeHost.FindStringSubmatch(h); len(m) > 0 {
		sh := fmt.Sprintf("tcp://127.0.0.1:%s", m[1])
		ctx.WithFields(map[string]interface{}{
			"preParse":  h,
			"postParse": sh,
		}).Debug(`parseSafeHost - rxParseSafeHost.FindStringSubmatch(h)`)
		return sh
	}

	ctx.WithFields(map[string]interface{}{
		"preParse":  h,
		"postParse": h,
	}).Debug(`parseSafeHost - no change`)
	return h
}

// FindFlagVal looks at the provided list of arguments and attempts
// to find the value for the specified flag. The indicies of the
// arguments that are the flag and its value are also returned. If
// the flag is specified as --flag=val then only a single index will
// be returned.
func FindFlagVal(name string, args ...string) (string, []int) {
	var format string
	if strings.HasPrefix(name, "--") {
		format = `(?i)^%s(?:=(.*))?$`
	} else {
		format = `(?i)^%s$`
	}
	rx := regexp.MustCompile(fmt.Sprintf(format, name))
	for i, v := range args {
		m := rx.FindStringSubmatch(v)
		switch len(m) {
		case 0:
			continue
		case 1:
			if i+1 < len(args) {
				return args[i+1], []int{i, i + 1}
			}
			return "", []int{i}
		case 2:
			if m[1] == "" {
				if i+1 < len(args) {
					return args[i+1], []int{i, i + 1}
				}
				return "", []int{i}
			}
			return m[1], []int{i}
		}
	}

	return "", nil
}

// IsTerminal returns a flag indicating whether or not the provided
// writer is connected to a terminal device.
func IsTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		return terminal.IsTerminal(int(v.Fd()))
	default:
		return false
	}
}
