package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/emccode/libstorage/api/context"
	apitypes "github.com/emccode/libstorage/api/types"
	"github.com/emccode/rexray/cli/cli"
	"github.com/emccode/rexray/core"

	// load REX-Ray
	_ "github.com/emccode/rexray"
)

func main() {

	updateLogLevel()

	var (
		err          error
		traceProfile *os.File
		cpuProfile   *os.File
		ctx          = context.Background()
		exit         sync.Once
	)

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
