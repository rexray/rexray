package cli

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/context"
	apitypes "github.com/emccode/libstorage/api/types"
	"github.com/emccode/rexray/core"
	rrdaemon "github.com/emccode/rexray/daemon"
	"github.com/emccode/rexray/util"
)

var (
	useSystemDForSCMCmds = gotil.FileExists(util.UnitFilePath) &&
		getInitSystemType() == SystemD

	serverSockFile = util.RunFilePath("server.sock")
	clientSockFile = util.RunFilePath("client.sock")
)

func (c *CLI) start() {
	if !c.fg && useSystemDForSCMCmds {
		startViaSystemD()
		return
	}

	checkOpPerms("started")

	c.ctx.WithField("os.Args", os.Args).Debug("invoking service start")

	pidFile := util.PidFilePath()

	if gotil.FileExists(pidFile) {
		pid, pidErr := util.ReadPidFile()
		if pidErr != nil {
			fmt.Printf("Error reading REX-Ray PID file at %s\n", pidFile)
			panic(1)
		}

		rrproc, err := findProcess(pid)
		if err != nil {
			fmt.Printf("Error finding process for PID %d", pid)
			panic(1)
		}

		if rrproc != nil {
			fmt.Printf("REX-Ray already running at PID %d\n", pid)
			panic(1)
		}

		if err := os.RemoveAll(pidFile); err != nil {
			fmt.Println("Error removing REX-Ray PID file")
			panic(1)
		}
	}

	if c.fg || c.fork {
		c.ctx.Debug("starting in foreground")
		c.startDaemon()
	} else {
		c.ctx.Debug("starting in background")
		c.tryToStartDaemon()
	}
}

func failOnError(err error) {
	if err != nil {
		fmt.Printf("FAILED!\n  %v\n", err)
		panic(err)
	}
}

func startViaSystemD() {
	execSystemDCmd("start")
	statusViaSystemD()
}

func stopViaSystemD() {
	execSystemDCmd("stop")
	statusViaSystemD()
}

func statusViaSystemD() {
	execSystemDCmd("status")
}

func execSystemDCmd(cmdType string) {
	cmd := exec.Command("systemctl", cmdType, "-l", "rexray")
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

func (c *CLI) startDaemon() {

	var out io.Writer = os.Stdout
	if !log.IsTerminal() {
		logFile, logFileErr := util.LogFile("rexray.log")
		failOnError(logFileErr)
		out = io.MultiWriter(os.Stdout, logFile)
	}
	log.SetOutput(out)

	c.ctx = c.ctx.WithValue(context.LoggerKey,
		&log.Logger{
			Formatter: log.StandardLogger().Formatter,
			Out:       out,
			Hooks:     log.StandardLogger().Hooks,
			Level:     log.StandardLogger().Level,
		})

	fmt.Fprintf(out, "%s\n", rexRayLogoASCII)
	util.PrintVersion(out)
	fmt.Fprintln(out)

	var success []byte
	var failure []byte
	var conn net.Conn

	if !c.fg {

		success = []byte{0}
		failure = []byte{1}

		var dialErr error

		c.ctx.WithField("addr", clientSockFile).Debug("dialing rex-ray client")
		conn, dialErr = net.Dial("unix", clientSockFile)
		if dialErr != nil {
			panic(dialErr)
		}
	}

	writePidErr := util.WritePidFile(-1)
	if writePidErr != nil {
		if conn != nil {
			conn.Write(failure)
		}
		panic(writePidErr)
	}

	defer func() {
		r := recover()
		os.Remove(util.PidFilePath())
		if r != nil {
			panic(r)
		}
	}()

	c.ctx.WithField("pid", os.Getpid()).Info("created pid file")

	stop := make(chan os.Signal)

	os.Remove(serverSockFile)
	host := fmt.Sprintf("unix://%s", serverSockFile)
	errs, err := rrdaemon.Start(c.ctx, c.config, host, stop)
	if err != nil {
		c.ctx.WithError(err).Error("error starting rex-ray")
		if conn != nil {
			conn.Write(failure)
			conn.Close()
		}
		return
	}

	if conn != nil {
		conn.Write(success)
		conn.Close()
	}

	wait := make(chan os.Signal)
	core.RegisterSignalHandler(func(ctx apitypes.Context, s os.Signal) {

		if ok, _ := core.IsExitSignal(s); !ok {
			return
		}

		ctx = ctx.Join(c.ctx)

		ctx.Info("signal received; shutting down services")

		stop <- s
		close(stop)

		os.Remove(serverSockFile)

		// wait until the daemon stops
		for range errs {
		}

		ctx.Info("service shutdown complete")

		wait <- s
		close(wait)
	})

	<-wait
}

func (c *CLI) tryToStartDaemon() {
	_, _, thisAbsPath := gotil.GetThisPathParts()

	fmt.Print("Starting REX-Ray...")

	signal := make(chan byte)
	os.Remove(clientSockFile)
	c.ctx.WithField("client", clientSockFile).Debug("trying to start service")

	l, lErr := net.Listen("unix", clientSockFile)
	failOnError(lErr)

	go func() {
		conn, acceptErr := l.Accept()
		if acceptErr != nil {
			fmt.Printf("FAILED!\n  %v\n", acceptErr)
			panic(acceptErr)
		}
		defer conn.Close()
		defer os.Remove(clientSockFile)

		c.ctx.Debug("accepted connection")

		buff := make([]byte, 1)
		conn.Read(buff)

		c.ctx.Debug("received data")

		signal <- buff[0]
	}()

	cmdArgs := []string{
		"start", "--fork",
		fmt.Sprintf("--logLevel=%v", c.logLevel())}

	if c.cfgFile != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--config=%s", c.cfgFile))
	}

	cmd := exec.Command(thisAbsPath, cmdArgs...)

	cmdErr := cmd.Start()
	failOnError(cmdErr)

	sigVal := <-signal
	if sigVal != 0 {
		fmt.Println("FAILED!")
		panic(1)
	}

	pid, _ := util.ReadPidFile()
	fmt.Printf("SUCCESS!\n\n")
	fmt.Printf("  The REX-Ray daemon is now running at PID %d. To\n", pid)
	fmt.Printf("  shutdown the daemon execute the following command:\n\n")
	fmt.Printf("    sudo %s stop\n\n", thisAbsPath)
}

func stop() {
	if useSystemDForSCMCmds {
		stopViaSystemD()
		return
	}

	checkOpPerms("stopped")

	if !gotil.FileExists(util.PidFilePath()) {
		fmt.Println("REX-Ray is already stopped")
		panic(1)
	}

	fmt.Print("Shutting down REX-Ray...")

	pid, pidErr := util.ReadPidFile()
	failOnError(pidErr)

	proc, procErr := os.FindProcess(pid)
	failOnError(procErr)

	killErr := proc.Signal(syscall.SIGHUP)
	failOnError(killErr)

	fmt.Println("SUCCESS!")
}

func (c *CLI) status() {
	if useSystemDForSCMCmds {
		statusViaSystemD()
		return
	}

	pidFile := util.PidFilePath()

	if !gotil.FileExists(pidFile) {
		fmt.Println("REX-Ray is stopped")
		return
	}

	pid, pidErr := util.ReadPidFile()
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

func (c *CLI) restart() {
	checkOpPerms("restarted")

	if gotil.FileExists(util.PidFilePath()) {
		stop()
	}

	c.start()
}

func checkOpPerms(op string) error {
	//if os.Geteuid() != 0 {
	//	return goof.Newf("REX-Ray can only be %s by root", op)
	//}

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
