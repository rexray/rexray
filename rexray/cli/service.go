package cli

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/emccode/rexray/core/errors"
	rrdaemon "github.com/emccode/rexray/daemon"
	"github.com/emccode/rexray/util"
)

func start() {
	checkOpPerms("started")

	log.WithField("os.Args", os.Args).Debug("invoking service start")

	pidFile := util.PidFilePath()

	if util.FileExists(pidFile) {
		pid, pidErr := util.ReadPidFile()
		if pidErr != nil {
			fmt.Printf("Error reading REX-Ray PID file at %s\n", pidFile)
		} else {
			fmt.Printf("REX-Ray already running at PID %d\n", pid)
		}
		panic(1)
	}

	if fg || client != "" {
		startDaemon()
	} else {
		tryToStartDaemon()
	}
}

func failOnError(err error) {
	if err != nil {
		fmt.Printf("FAILED!\n  %v\n", err)
		panic(err)
	}
}

func startDaemon() {

	var out io.Writer = os.Stdout
	if !log.IsTerminal() {
		logFile, logFileErr := util.LogFile("rexray.log")
		failOnError(logFileErr)
		out = io.MultiWriter(os.Stdout, logFile)
	}
	log.SetOutput(out)

	fmt.Fprintf(out, "%s\n", rexRayLogoASCII)
	util.PrintVersion(out)
	fmt.Fprintln(out)

	var success []byte
	var failure []byte
	var conn net.Conn

	if !fg {

		success = []byte{0}
		failure = []byte{1}

		var dialErr error

		log.Printf("dialing %s", client)
		conn, dialErr = net.Dial("unix", client)
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

	log.Printf("created pid file, pid=%d", os.Getpid())

	init := make(chan error)
	sigc := make(chan os.Signal, 1)
	stop := make(chan os.Signal)

	signal.Notify(sigc,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		rrdaemon.Start(r.Config.Host, init, stop)
	}()

	var initErrors []error

	for initErr := range init {
		initErrors = append(initErrors, initErr)
		log.Println(initErr)
	}

	if conn != nil {
		if len(initErrors) == 0 {
			conn.Write(success)
		} else {
			conn.Write(failure)
		}

		conn.Close()
	}

	if len(initErrors) > 0 {
		return
	}

	sigv := <-sigc
	log.Printf("received shutdown signal %v", sigv)
	stop <- sigv
}

func tryToStartDaemon() {
	_, _, thisAbsPath := util.GetThisPathParts()

	fmt.Print("Starting REX-Ray...")

	signal := make(chan byte)
	client := fmt.Sprintf("%s/%s.sock", os.TempDir(), util.RandomString(32))
	log.WithField("client", client).Debug("trying to start service")

	l, lErr := net.Listen("unix", client)
	failOnError(lErr)

	go func() {
		conn, acceptErr := l.Accept()
		if acceptErr != nil {
			fmt.Printf("FAILED!\n  %v\n", acceptErr)
			panic(acceptErr)
		}
		defer conn.Close()
		defer os.Remove(client)

		log.Debug("accepted connection")

		buff := make([]byte, 1)
		conn.Read(buff)

		log.Debug("received data")

		signal <- buff[0]
	}()

	cmdArgs := []string{
		"start",
		fmt.Sprintf("--client=%s", client),
		fmt.Sprintf("--logLevel=%v", r.Config.LogLevel)}

	if r.Config.Host != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--host=%s", r.Config.Host))
	}

	cmd := exec.Command(thisAbsPath, cmdArgs...)
	cmd.Stderr = os.Stderr

	cmdErr := cmd.Start()
	failOnError(cmdErr)

	sigVal := <-signal
	if sigVal != 0 {
		fmt.Println("FAILED!")
		panic(1)
	}

	pid, _ := util.ReadPidFile()
	fmt.Printf("SUCESS!\n\n")
	fmt.Printf("  The REX-Ray daemon is now running at PID %d. To\n", pid)
	fmt.Printf("  shutdown the daemon execute the following command:\n\n")
	fmt.Printf("    sudo %s stop\n\n", thisAbsPath)
}

func stop() {
	checkOpPerms("stopped")

	if !util.FileExists(util.PidFilePath()) {
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

func status() {
	if !util.FileExists(util.PidFilePath()) {
		fmt.Println("REX-Ray is stopped")
		return
	}
	pid, _ := util.ReadPidFile()
	fmt.Printf("REX-Ray is running at pid %d\n", pid)
}

func restart() {
	checkOpPerms("restarted")

	if util.FileExists(util.PidFilePath()) {
		stop()
	}

	start()
}

func checkOpPerms(op string) error {
	if os.Geteuid() != 0 {
		return errors.Newf("REX-Ray can only be %s by root", op)
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
