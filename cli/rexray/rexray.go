package rexray

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gotil"

	"github.com/codedellemc/rexray/libstorage/api/context"
	"github.com/codedellemc/rexray/libstorage/api/registry"
	apitypes "github.com/codedellemc/rexray/libstorage/api/types"
	"github.com/codedellemc/rexray/libstorage/api/utils"
	"github.com/codedellemc/rexray/cli/cli"
	"github.com/codedellemc/rexray/core"

	// load REX-Ray
	_ "github.com/codedellemc/rexray"

	// load the profiler
	_ "net/http/pprof"
)

// Run the CLI.
func Run() {

	updateLogLevel()

	var (
		err          error
		traceProfile *os.File
		cpuProfile   *os.File
		ctx          = context.Background()
		exit         sync.Once
	)

	pathConfig := utils.NewPathConfig(ctx, "", "")
	ctx = ctx.WithValue(
		context.PathConfigKey, pathConfig)
	registry.ProcessRegisteredConfigs(ctx)

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

	core.RegisterSignalHandler(func(ctx apitypes.Context, s os.Signal) {
		if ok, _ := core.IsExitSignal(s); ok {
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

	cli.Execute(ctx)
	ctx.Debug("completed cli execution")

	exit.Do(onExit)
	ctx.Debug("completed onExit at end of program")
}

func updateLogLevel() {
	if ok, _ := strconv.ParseBool(os.Getenv("REXRAY_DEBUG")); ok {
		enableDebugMode()
		return
	}

	if ok, _ := strconv.ParseBool(os.Getenv("LIBSTORAGE_DEBUG")); ok {
		enableDebugMode()
		return
	}

	if ll := os.Getenv("REXRAY_LOG_LEVEL"); ll != "" {
		if lvl, err := log.ParseLevel(ll); err != nil {
			setLogLevels(lvl)
			return
		}
	}

	if ll := os.Getenv("LIBSTORAGE_LOGGING_LEVEL"); ll != "" {
		if lvl, err := log.ParseLevel(ll); err != nil {
			setLogLevels(lvl)
		}
	}
}

func enableDebugMode() {
	log.SetLevel(log.DebugLevel)
	apitypes.Debug = true
	setLogLevels(log.DebugLevel)
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
