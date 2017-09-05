// +build !client

package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"
	log "github.com/sirupsen/logrus"

	"github.com/codedellemc/rexray/core"
	apictx "github.com/codedellemc/rexray/libstorage/api/context"
	apitypes "github.com/codedellemc/rexray/libstorage/api/types"
	"github.com/codedellemc/rexray/util"
)

type startFunc func(
	ctx apitypes.Context,
	config gofig.Config) (apitypes.Context, <-chan error, error)

var (
	startFuncs           []startFunc
	logFileName          = fmt.Sprintf("%s.log", util.BinFileName)
	useSystemDForSCMCmds = gotil.FileExists(util.UnitFilePath) &&
		getInitSystemType() == SystemD
)

func serviceStart(
	ctx apitypes.Context,
	config gofig.Config,
	nopid bool) {

	var cancel context.CancelFunc
	ctx, cancel = apictx.WithCancel(ctx)

	var wg sync.WaitGroup
	wg.Add(1)

	core.RegisterSignalHandler(func(ctx apitypes.Context, s os.Signal) {
		if ok, _ := core.IsExitSignal(s); ok {
			ctx.Info("received exit signal")
			cancel()
			wg.Wait()
		}
	})

	serviceStartAndWait(ctx, config, nopid)
	wg.Done()
}

func serviceStartAndWait(
	ctx apitypes.Context,
	config gofig.Config,
	nopid bool) {

	checkOpPerms("started")

	if !nopid {
		handleStalePIDFile(ctx)
	}

	var out io.Writer = os.Stdout
	if !util.IsTerminal(out) {
		logFile, logFileErr := util.LogFile(ctx, logFileName)
		failOnError(logFileErr)
		out = io.MultiWriter(os.Stdout, logFile)
	}
	log.SetOutput(out)

	ctx = ctx.WithValue(apictx.LoggerKey,
		&log.Logger{
			Formatter: log.StandardLogger().Formatter,
			Out:       out,
			Hooks:     log.StandardLogger().Hooks,
			Level:     log.StandardLogger().Level,
		})

	fmt.Fprintf(out, "%s\n", rexRayLogoASCII)
	util.PrintVersion(out)
	fmt.Fprintln(out)

	pidFile := util.PidFilePath(ctx)

	if !nopid {
		if err := util.WritePidFile(ctx, -1); err != nil {
			if os.IsPermission(err) {
				ctx.WithError(err).Errorf(
					"user does not have write permissions for %s",
					pidFile)
			} else {
				ctx.WithError(err).Errorf(
					"error writing PID file at %s",
					pidFile)
			}
			os.Exit(1)
		}
		ctx.WithFields(map[string]interface{}{
			"path": pidFile,
			"pid":  os.Getpid(),
		}).Info("created pid file")
	}

	var wg sync.WaitGroup
	wg.Add(len(startFuncs))

	// Start the registered services via their start functions.
	for _, f := range startFuncs {
		fctx, errs, err := f(ctx, config)
		if err != nil {
			panic(err)
		}

		if errs == nil {
			wg.Done()
			continue
		}

		// Update the outer context with the context returned from
		// the service start function. This is important as a start
		// function may have injected some data into the context used
		// by subsequent services.
		ctx = fctx

		// Inside a goroutine wait until the service's error channel
		// is closed.
		go func(errs <-chan error) {
			for err := range errs {
				if err != nil {
					panic(err)
				}
			}
			wg.Done()
		}(errs)
	}

	// Wait until this context has been cancelled.
	<-ctx.Done()
	ctx.Info("cancelled; shutting down services")

	// Wait until all the service start functions have returned.
	wg.Wait()
	ctx.Info("service shutdown complete")

	if !nopid {
		os.RemoveAll(pidFile)
		ctx.WithField("path", pidFile).Info("removed pid file")
	}
}

func serviceStop(ctx apitypes.Context) {
	if useSystemDForSCMCmds {
		stopViaSystemD()
		return
	}

	checkOpPerms("stopped")

	if !gotil.FileExists(util.PidFilePath(ctx)) {
		fmt.Println("REX-Ray is already stopped")
		panic(1)
	}

	fmt.Print("Shutting down REX-Ray...")

	pid, pidErr := util.ReadPidFile(ctx)
	failOnError(pidErr)

	proc, procErr := os.FindProcess(pid)
	failOnError(procErr)

	killErr := proc.Signal(os.Interrupt)
	failOnError(killErr)

	fmt.Println("SUCCESS!")
}

func serviceRestart(
	ctx apitypes.Context,
	config gofig.Config,
	nopid bool) {

	checkOpPerms("restarted")
	if gotil.FileExists(util.PidFilePath(ctx)) {
		serviceStop(ctx)
	}
	serviceStart(ctx, config, nopid)
}

func serviceStatus(ctx apitypes.Context) {
	if useSystemDForSCMCmds {
		statusViaSystemD()
		return
	}

	pidFile := util.PidFilePath(ctx)

	if !gotil.FileExists(pidFile) {
		fmt.Println("REX-Ray is stopped")
		return
	}

	pid, pidErr := util.ReadPidFile(ctx)
	if pidErr != nil {
		fmt.Printf("Error reading REX-Ray PID file at %s\n", pidFile)
		panic(1)
	}

	rrproc, err := findProcess(pid)

	if err != nil || rrproc == nil {
		if err := os.RemoveAll(pidFile); err != nil {
			fmt.Println("Error removing stale REX-Ray PID file")
			panic(1)
		}
		fmt.Println("REX-Ray is stopped")
		return
	}

	fmt.Printf("REX-Ray is running at PID %d\n", pid)
	return
}

func stopViaSystemD() {
	execSystemDCmd("stop")
	statusViaSystemD()
}

func statusViaSystemD() {
	execSystemDCmd("status")
}

func execSystemDCmd(cmdType string) {
	cmd := exec.Command("systemctl", cmdType, "-l", util.BinFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				panic(status.ExitStatus())
			}
		}
	}
}

func handleStalePIDFile(ctx apitypes.Context) {
	pidFile := util.PidFilePath(ctx)
	if !gotil.FileExists(pidFile) {
		return
	}

	pid, err := util.ReadPidFile(ctx)
	if err != nil {
		fmt.Printf("Error reading REX-Ray PID file at %s\n", pidFile)
		panic(1)
	}

	proc, err := findProcess(pid)
	if err != nil {
		fmt.Printf("Error finding process for PID %d", pid)
		panic(1)
	}

	if proc != nil {
		fmt.Printf("REX-Ray already running at PID %d\n", pid)
		panic(1)
	}

	if err := os.RemoveAll(pidFile); err != nil {
		fmt.Println("Error removing REX-Ray PID file")
		panic(1)
	}
}

func failOnError(err error) {
	if err != nil {
		fmt.Printf("FAILED!\n  %v\n", err)
		panic(err)
	}
}

const rexRayLogoASCII = `
                          ⌐▄Q▓▄Ç▓▄,▄_
                         Σ▄▓▓▓▓▓▓▓▓▓▓▄π
                       ╒▓▓▌▓▓▓▓▓▓▓▓▓▓▀▓▄▄.
                    ,_▄▀▓▓ ▓▓ ▓▓▓▓▓▓▓▓▓▓▓█
                   │▄▓▓ _▓▓▓▓▓▓▓▓▓┌▓▓▓▓▓█
                  _J┤▓▓▓▓▓▓▓▓▓▓▓▓▓├█▓█▓▀Γ
            ,▄▓▓▓▓▓▓^██▓▓▓▓▓▓▓▓▓▓▓▓▄▀▄▄▓▓Ω▄
            F▌▓▌█ⁿⁿⁿ  ⁿ└▀ⁿ██▓▀▀▀▀▀▀▀▀▀▀▌▓▓▓▌
             'ⁿ_  ,▄▄▄▄▄▄▄▄▄█_▄▄▄▄▄▄▄▄▄ⁿ▀~██
               Γ  ├▓▓▓▓▓█▀ⁿ█▌▓Ω]█▓▓▓▓▓▓ ├▓
               │  ├▓▓▓▓▓▌≡,__▄▓▓▓█▓▓▓▓▓ ╞█~   Y,┐
               ╞  ├▓▓▓▓▓▄▄__^^▓▓▓▌▓▓▓▓▓  ▓   /▓▓▓
                  ├▓▓▓▓▓▓▓▄▄═▄▓▓▓▓▓▓▓▓▓  π ⌐▄▓▓█║n
                _ ├▓▓▓▓▓▓▓▓▓~▓▓▓▓▓▓▓▓▓▓  ▄4▄▓▓▓██
                µ ├▓▓▓▓█▀█▓▓_▓▓███▓▓▓▓▓  ▓▓▓▓▓Ω4
                µ ├▓▀▀L   └ⁿ  ▀   ▀ ▓▓█w ▓▓▓▀ìⁿ
                ⌐ ├_                τ▀▓  Σ⌐└
                ~ ├▓▓  ▄  _     ╒  ┌▄▓▓  Γ
                  ├▓▓▓▌█═┴▓▄╒▀▄_▄▌═¢▓▓▓  ╚
               ⌠  ├▓▓▓▓▓ⁿ▄▓▓▓▓▓▓▓┐▄▓▓▓▓  └
               Ω_.└██▓▀ⁿÇⁿ▀▀▀▀▀▀█≡▀▀▀▀▀   µ
               ⁿ  .▄▄▓▓▓▓▄▄┌ ╖__▓_▄▄▄▄▄*Oⁿ
                 û▌├▓█▓▓▓██ⁿ ¡▓▓▓▓▓▓▓▓█▓╪
                 ╙Ω▀█ ▓██ⁿ    └█▀██▀▓█├█Å
                     ⁿⁿ             ⁿ ⁿ^
:::::::..  .,::::::    .,::      .::::::::..    :::.  .-:.     ::-.
;;;;'';;;; ;;;;''''    ';;;,  .,;; ;;;;'';;;;   ;;';;  ';;.   ;;;;'
 [[[,/[[['  [[cccc       '[[,,[['   [[[,/[[['  ,[[ '[[,  '[[,[[['
 $$$$$$c    $$""""        Y$$$Pcccc $$$$$$c   c$$$cc$$$c   c$$"
 888b "88bo,888oo,__    oP"''"Yo,   888b "88bo,888   888,,8P"'
 MMMM   "W" """"YUMMM,m"       "Mm, MMMM   "W" YMM   ""'mM"
`
