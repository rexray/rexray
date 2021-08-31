// +build !client

package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"
	log "github.com/sirupsen/logrus"

	"github.com/AVENTER-UG/rexray/core"
	apictx "github.com/AVENTER-UG/rexray/libstorage/api/context"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/util"
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

	var (
		exit            sync.Once
		wg              sync.WaitGroup
		serviceStartErr error
		cancel          context.CancelFunc
	)

	wg.Add(1)
	defer wg.Done()

	ctx, cancel = apictx.WithCancel(ctx)

	onExit := func() {
		if serviceStartErr == nil {
			return
		}
		fmt.Fprintf(
			os.Stderr,
			"error: %v\n",
			serviceStartErr)
		os.Exit(1)
	}
	defer exit.Do(onExit)

	core.RegisterSignalHandler(func(ctx apitypes.Context, s os.Signal) {
		if ok, _ := core.IsExitSignal(s); ok {
			ctx.Info("received exit signal")
			cancel()
			wg.Wait()
			exit.Do(onExit)
		}
	})

	if err := serviceStartAndWait(ctx, config, nopid); err != nil {
		serviceStartErr = fmt.Errorf("service startup failed: %v", err)
	}
}

func serviceStartAndWait(
	ctx apitypes.Context,
	config gofig.Config,
	nopid bool) error {

	checkOpPerms("started")

	if !nopid {
		if err := handleStalePIDFile(ctx); err != nil {
			return err
		}
	}

	var out io.Writer = os.Stdout
	if !util.IsTerminal(out) {
		logFile, err := util.LogFile(ctx, logFileName)
		if err != nil {
			return fmt.Errorf(
				"create log file failed: %s: %v", logFileName, err)
		}
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

	pidFile := util.PidFilePath(ctx)

	if !nopid {
		if err := util.WritePidFile(ctx, -1); err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf(
					"write access denied to pid file: %s", pidFile)
			}
			return fmt.Errorf(
				"write pid file failed: %s: %v", pidFile, err)
		}
		ctx.WithFields(map[string]interface{}{
			"path": pidFile,
			"pid":  os.Getpid(),
		}).Info("created pid file")
	}

	var (
		wg              sync.WaitGroup
		cancel          context.CancelFunc
		serviceStartErr error
	)

	// Create a context that can be cancelled.
	ctx, cancel = apictx.WithCancel(ctx)

	// Start the registered services via their start functions.
	for _, f := range startFuncs {
		fctx, errs, err := f(ctx, config)

		// If the service failed to start then jump out of this loop
		if err != nil {
			serviceStartErr = err
			break
		}

		// Skip this service if it did not return an error channel.
		if errs == nil {
			continue
		}

		wg.Add(1)

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
					ctx.WithError(err).Error("service failure detected")
				}
			}
			wg.Done()
		}(errs)
	}

	// If any service startup errors occurred then make sure any
	// services that *were* started receive a cancel signal.
	//
	// If no errors occurred then it's safe to print the logo and version.
	if serviceStartErr != nil {
		cancel()
	} else {
		fmt.Fprintf(out, "%s\n", rexRayLogoASCII)
		util.PrintVersion(out)
		fmt.Fprintln(out)
	}

	// Wait until this context has been cancelled.
	<-ctx.Done()

	// Wait until all the service start functions have returned.
	wg.Wait()
	ctx.Info("service shutdown complete")

	if !nopid {
		os.RemoveAll(pidFile)
		ctx.WithField("path", pidFile).Info("removed pid file")
	}

	return serviceStartErr
}

var errAlreadyStopped = errors.New("already stopped")

func serviceStop(ctx apitypes.Context) error {
	if useSystemDForSCMCmds {
		stopViaSystemD()
		return nil
	}

	checkOpPerms("stopped")

	if !gotil.FileExists(util.PidFilePath(ctx)) {
		return errAlreadyStopped
	}

	pid, err := util.ReadPidFile(ctx)
	if err != nil {
		return fmt.Errorf(
			"read pid file failed: %s: %v",
			util.PidFilePath(ctx),
			err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process for pid failed: %d: %v", pid, err)
	}

	if err := proc.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("kill process failed: %d: %v", pid, err)
	}

	return nil
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
		os.Exit(1)
	}

	rrproc, err := findProcess(pid)

	if err != nil || rrproc == nil {
		if err := os.RemoveAll(pidFile); err != nil {
			fmt.Println("Error removing stale REX-Ray PID file")
			os.Exit(1)
		}
		fmt.Println("REX-Ray is stopped")
		return
	}

	fmt.Printf("REX-Ray is running at PID %d\n", pid)
	return
}

func stopViaSystemD() {
	execSystemDCmd("stop")
	//statusViaSystemD()
}

func statusViaSystemD() {
	execSystemDCmd("status")
}

func execSystemDCmd(cmdType string) {
	cmdAndArgs := []string{"systemctl", cmdType, "-l", util.BinFileName}
	cmd := exec.Command(cmdAndArgs[0], cmdAndArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
		fmt.Fprintf(
			os.Stderr,
			"error: systemd cmd failed: %s: %v\n",
			strings.Join(cmdAndArgs, " "),
			err)
		os.Exit(exitCode)
	}
}

func handleStalePIDFile(ctx apitypes.Context) error {
	pidFile := util.PidFilePath(ctx)
	if !gotil.FileExists(pidFile) {
		return nil
	}

	pid, err := util.ReadPidFile(ctx)
	if err != nil {
		return fmt.Errorf("read pid file failed: %s: %v", pidFile, err)
	}

	proc, err := findProcess(pid)
	if err != nil {
		return fmt.Errorf("find process for pid failed: %d: %v", pid, err)
	}

	if proc != nil {
		return fmt.Errorf("already running at pid %d", pid)
	}

	if err := os.RemoveAll(pidFile); err != nil {
		return fmt.Errorf("remove pid file failed: %s: %v", pidFile, err)
	}

	return nil
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
