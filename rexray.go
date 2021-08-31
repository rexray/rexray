//go:generate go generate ./core
//go:generate go run core/semver/semver.go -f mk -o semver.mk
//go:generate go run core/semver/semver.go -f env -o semver.env -x

package main

import (
	"fmt"
	golog "log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"strings"
	"sync"

	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"
	log "github.com/sirupsen/logrus"

	"github.com/AVENTER-UG/rexray/cli"
	"github.com/AVENTER-UG/rexray/core"
	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	rrutils "github.com/AVENTER-UG/rexray/util"

	// import the libstorage config package
	_ "github.com/AVENTER-UG/rexray/libstorage/imports/config"

	// load the profiler
	_ "net/http/pprof"
)

func main() {
	// If X_CSI_NATIVE is set to a truthy value then disable libStorage.
	if v := os.Getenv("X_CSI_NATIVE"); v != "" {
		if ok, _ := strconv.ParseBool(v); ok {
			os.Setenv("LIBSTORAGE", "false")
		}
	}

	// Brand libStorage's path structure with "rexray"
	if v := os.Getenv("LIBSTORAGE_APPTOKEN"); v == "" {
		os.Setenv("LIBSTORAGE_APPTOKEN", "rexray")
	}

	// Update REXRAY_HOME and LIBSTORAGE_HOME from the other if
	// one is set and the other is not.
	rrHome := os.Getenv("REXRAY_HOME")
	lsHome := os.Getenv("LIBSTORAGE_HOME")
	if rrHome != "" && lsHome == "" {
		os.Setenv("LIBSTORAGE_HOME", rrHome)
	} else if rrHome == "" && lsHome != "" {
		os.Setenv("REXRAY_HOME", lsHome)
	}

	// Since flags are not parsed yet, manually check to see if a
	// -l or --logLevel were provided via the command line's arguments.
	if v, _ := rrutils.FindFlagVal(
		"-l", os.Args...); v != "" {
		os.Setenv("REXRAY_LOGLEVEL", v)
		os.Setenv("LIBSTORAGE_LOGGING_LEVEL", v)
	} else if v, _ := rrutils.FindFlagVal(
		"--loglevel", os.Args...); v != "" {
		os.Setenv("REXRAY_LOGLEVEL", v)
		os.Setenv("LIBSTORAGE_LOGGING_LEVEL", v)
	}

	// Since flags are not parsed yet, manually check to see if a
	// -c or --config were provided via the command line's arguments.
	var configFile string
	if v, _ := rrutils.FindFlagVal(
		"-c", os.Args...); v != "" {
		configFile = v
	} else if v, _ := rrutils.FindFlagVal(
		"--config", os.Args...); v != "" {
		configFile = v
	}

	// Register REX-Ray's global config options.
	registerConfig()

	// Create a new context and process registration configs.
	ctx := context.Background()
	pathConfig := utils.NewPathConfig()
	ctx = ctx.WithValue(
		context.PathConfigKey, pathConfig)
	registry.ProcessRegisteredConfigs(ctx)

	// If the configFile value is empty then configure Gofig's global
	// search locations.
	var config gofig.Config
	if configFile == "" {
		gofigCore.SetGlobalConfigPath(pathConfig.Etc)
		gofigCore.SetUserConfigPath(pathConfig.Home)
	} else if !gotil.FileExists(configFile) {
		fmt.Fprintf(os.Stderr,
			"error: invalid config file: %s\n", configFile)
		os.Exit(1)
	} else {
		rrutils.ValidateConfig(configFile)
		config = gofigCore.New()
		if err := config.ReadConfigFile(configFile); err != nil {
			fmt.Fprintf(os.Stderr,
				"error: invalid config file: %s: %v\n", configFile, err)
			os.Exit(1)
		}
		config = config.Scope("rexray")
	}

	// Update the log level after it's been parsed from every possible
	// location.
	context.SetLogLevel(ctx, updateLogLevel(config))

	// Get the context logger and update Go's standard log facility so that
	// it emits logs using the context logger.
	golog.SetFlags(0)
	golog.SetOutput(rrutils.NewWriterFor(ctx.Infof))

	// Print the status of DockerLegacyMode.
	ctx.WithField("enabled", core.DockerLegacyMode).Info("DockerLegacyMode")

	if config != nil {
		ctx.WithField("path", configFile).Info("loaded custom config")
	}

	var (
		err          error
		traceProfile *os.File
		cpuProfile   *os.File
		exit         sync.Once
	)

	createUserKnownHostsFile(ctx, pathConfig)

	onExit := func() {
		if traceProfile != nil {
			ctx.Info("stopping trace profile")
			trace.Stop()
			traceProfile.Close()
			ctx.Debug("stopped trace profile")
		}

		if cpuProfile != nil {
			ctx.Info("stopping cpu profile")
			pprof.StopCPUProfile()
			cpuProfile.Close()
			ctx.Debug("stopped cpu profile")
		}

		ctx.Info("exiting process")
	}

	var waitForExit chan int

	core.RegisterSignalHandler(func(ctx apitypes.Context, s os.Signal) {
		if ok, _ := core.IsExitSignal(s); ok {
			waitForExit = make(chan int)
			ctx.Info("received exit signal")
			exit.Do(onExit)
		}
	})

	if p := os.Getenv("REXRAY_TRACE_PROFILE"); p != "" {
		if traceProfile, err = os.Create(p); err != nil {
			panic(err)
		}
		if err = trace.Start(traceProfile); err != nil {
			panic(err)
		}
		ctx.WithField("path", traceProfile.Name()).Info("trace profile enabled")
	}

	if p := os.Getenv("REXRAY_CPU_PROFILE"); p != "" {
		if cpuProfile, err = os.Create(p); err != nil {
			panic(err)
		}
		if err = pprof.StartCPUProfile(cpuProfile); err != nil {
			panic(err)
		}
		ctx.WithField("path", cpuProfile.Name()).Info("cpu profile enabled")
	}

	if p := os.Getenv("REXRAY_PROFILE_ADDR"); p != "" {
		go http.ListenAndServe(p, http.DefaultServeMux)
		ctx.WithField("address", p).Info("http pprof enabled")
	}

	core.TrapSignals(ctx)
	ctx.Debug("trapped signals")

	cli.Execute(ctx, config)
	ctx.Debug("completed cli execution")

	exit.Do(onExit)
	ctx.Debug("completed onExit at end of program")

	// If an exit signal was received then just block until the
	// handler exits the process.
	if waitForExit != nil {
		<-waitForExit
	}
}

const (
	configRR   = "rexray"
	configRRLL = configRR + ".loglevel"
)

func setConfigLogLevel(config gofig.Config, k1, k2 string, level string) {
	v, ok := config.Get(k1).(map[string]interface{})
	if ok {
		v[strings.Replace(k2, k1+".", "", 1)] = level
	} else {
		config.Set(k2, level)
	}
}

func updateLogLevel(config gofig.Config) (level log.Level) {
	defer func() {
		if config == nil {
			return
		}
		szl := level.String()
		setConfigLogLevel(
			config, configRR, configRRLL, szl)
		setConfigLogLevel(
			config, apitypes.ConfigLogging, apitypes.ConfigLogLevel, szl)
	}()

	if ok, _ := strconv.ParseBool(os.Getenv("REXRAY_DEBUG")); ok {
		enableDebugMode()
		return log.DebugLevel
	}

	if ok, _ := strconv.ParseBool(os.Getenv("LIBSTORAGE_DEBUG")); ok {
		enableDebugMode()
		return log.DebugLevel
	}

	if ll := os.Getenv("REXRAY_LOGLEVEL"); ll != "" {
		if lvl, err := log.ParseLevel(ll); err == nil {
			if lvl == log.DebugLevel {
				enableDebugMode()
			} else {
				setLogLevels(lvl)
			}
			return lvl
		}
	}

	if ll := os.Getenv("LIBSTORAGE_LOGGING_LEVEL"); ll != "" {
		if lvl, err := log.ParseLevel(ll); err == nil {
			if lvl == log.DebugLevel {
				enableDebugMode()
			} else {
				setLogLevels(lvl)
			}
			return lvl
		}
	}

	if config != nil {
		if ll := config.GetString(configRRLL); ll != "" {
			if lvl, err := log.ParseLevel(ll); err == nil {
				if lvl == log.DebugLevel {
					enableDebugMode()
				} else {
					setLogLevels(lvl)
				}
				return lvl
			}
		}
		if ll := config.GetString(apitypes.ConfigLogLevel); ll != "" {
			if lvl, err := log.ParseLevel(ll); err == nil {
				if lvl == log.DebugLevel {
					enableDebugMode()
				} else {
					setLogLevels(lvl)
				}
				return lvl
			}
		}
	}

	return log.WarnLevel
}

func enableDebugMode() {
	core.Debug = true
	apitypes.Debug = true
	os.Setenv("REXRAY_DEBUG", "true")
	os.Setenv("LIBSTORAGE_DEBUG", "true")
	setLogLevels(log.DebugLevel)
	log.SetLevel(log.DebugLevel)
}

func setLogLevels(lvl log.Level) {
	os.Setenv("REXRAY_LOGLEVEL", lvl.String())
	os.Setenv("LIBSTORAGE_LOGGING_LEVEL", lvl.String())
}

func createUserKnownHostsFile(
	ctx apitypes.Context,
	pathConfig *apitypes.PathConfig) {

	khPath := pathConfig.UserDefaultTLSKnownHosts

	if gotil.FileExists(khPath) {
		return
	}

	khDirPath := filepath.Dir(khPath)
	os.MkdirAll(khDirPath, 0755)
	khFile, err := os.Create(khPath)
	if err != nil {
		ctx.WithField("path", khPath).Fatal(
			"failed to create known_hosts")
	}
	defer khFile.Close()
}

func registerConfig() {
	r := gofigCore.NewRegistration("Global")
	r.SetYAML(`
rexray:
    logLevel: warn
`)
	r.Key(gofig.String, "h", "",
		"The libStorage host.", "rexray.host",
		"host")
	r.Key(gofig.String, "s", "",
		"The libStorage service.", "rexray.service",
		"service")
	r.Key(gofig.String, "l", "warn",
		"The log level (error, warn, info, debug)", "rexray.logLevel",
		"logLevel")
	gofigCore.Register(r)
}
