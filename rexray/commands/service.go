package commands

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	log "github.com/Sirupsen/logrus"

	_ "github.com/emccode/rexray/config"
	rrdaemon "github.com/emccode/rexray/daemon"
	"github.com/emccode/rexray/errors"
	"github.com/emccode/rexray/util"
)

func GetInitSystemCmd() string {
	switch GetInitSystemType() {
	case SYSTEMD:
		return "systemd"
	case UPDATERCD:
		return "update-rc.d"
	case CHKCONFIG:
		return "chkconfig"
	default:
		return "unknown"
	}
}

func GetInitSystemType() int {
	if util.FileExistsInPath("systemctl") {
		return SYSTEMD
	}

	if util.FileExistsInPath("update-rc.d") {
		return UPDATERCD
	}

	if util.FileExistsInPath("chkconfig") {
		return CHKCONFIG
	}

	return UNKNOWN
}

func Install() {
	checkOpPerms("installed")

	_, _, exeFile := util.GetThisPathParts()

	os.Chown(exeFile, 0, 0)
	exec.Command("chmod", "4755", exeFile).Run()

	switch GetInitSystemType() {
	case SYSTEMD:
		installSystemD(exeFile)
	case UPDATERCD:
		installUpdateRcd(exeFile)
	case CHKCONFIG:
		installChkConfig(exeFile)
	}
}

func Start() {
	checkOpPerms("started")

	log.WithField("os.Args", os.Args).Debug("invoking service start")

	pidFile := util.PidFile()

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

	fmt.Printf("%s\n", RexRayLogoAscii)

	var success []byte
	var failure []byte
	var conn net.Conn

	if !fg {

		success = []byte{0}
		failure = []byte{1}

		var dialErr error

		log.Printf("Dialing %s\n", client)
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
		os.Remove(util.PidFile())
		if r != nil {
			panic(r)
		}
	}()

	log.Printf("Created pid file, pid=%d\n", os.Getpid())

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
		rrdaemon.Start(c.Host, init, stop)
	}()

	initErrors := make([]error, 0)

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
	log.Printf("Received shutdown signal %v\n", sigv)
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
		fmt.Sprintf("--logLevel=%v", c.LogLevel)}

	if c.Host != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--host=%s", c.Host))
	}

	cmd := exec.Command(thisAbsPath, cmdArgs...)

	logFile, logFileErr :=
		os.OpenFile(util.LogFile(), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	failOnError(logFileErr)

	cmd.Stdout = logFile
	cmd.Stderr = logFile

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

func Stop() {
	checkOpPerms("stopped")

	if !util.FileExists(util.PidFile()) {
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

func Status() {
	if !util.FileExists(util.PidFile()) {
		fmt.Println("REX-Ray is stopped")
		return
	}
	pid, _ := util.ReadPidFile()
	fmt.Printf("REX-Ray is running at pid %d\n", pid)
}

func Restart() {
	checkOpPerms("restarted")

	if util.FileExists(util.PidFile()) {
		Stop()
	}

	Start()
}

func checkOpPerms(op string) error {
	if os.Geteuid() != 0 {
		return errors.Newf("REX-Ray can only be %s by root", op)
	}

	return nil
}

func installSystemD(exeFile string) {
	createUnitFile(exeFile)
	createEnvFile()

	cmd := exec.Command("systemctl", "enable", "rexray.service")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Fatalf("installation error %v", err)
	}

	fmt.Print("REX-Ray is now installed. Before starting it please check ")
	fmt.Print("http://github.com/emccode/rexray for instructions on how to ")
	fmt.Print("configure it.\n\n Once configured the REX-Ray service can be ")
	fmt.Print("started with the command 'sudo systemctl start rexray'.\n\n")
}

func installUpdateRcd(exeFile string) {
	createInitFile(exeFile)
	cmd := exec.Command("update-rc.d", "rexray", "defaults")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Fatalf("installation error %v", err)
	}

	fmt.Print("REX-Ray is now installed. Before starting it please check ")
	fmt.Print("http://github.com/emccode/rexray for instructions on how to ")
	fmt.Print("configure it.\n\n Once configured the REX-Ray service can be ")
	fmt.Print("started with the command 'sudo /etc/init.d/rexray start'.\n\n")
}

func installChkConfig(exeFile string) {
	createInitFile(exeFile)
	cmd := exec.Command("chkconfig", "rexray", "on")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Fatalf("installation error %v", err)
	}

	fmt.Print("REX-Ray is now installed. Before starting it please check ")
	fmt.Print("http://github.com/emccode/rexray for instructions on how to ")
	fmt.Print("configure it.\n\n Once configured the REX-Ray service can be ")
	fmt.Print("started with the command 'sudo /etc/init.d/rexray start'.\n\n")
}

func createEnvFile() {

	envdir := filepath.Dir(ENVFILE)
	os.MkdirAll(envdir, 0755)

	f, err := os.OpenFile(ENVFILE, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString("REXRAY_HOME=/opt/rexray")
}

func createUnitFile(exeFile string) {
	f, err := os.OpenFile(UNTFILE, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString("[Unit]\n")
	f.WriteString("Description=rexray\n")
	f.WriteString("Before=docker.service\n")
	f.WriteString("\n")
	f.WriteString("[Service]\n")
	f.WriteString(fmt.Sprintf("EnvironmentFile=%s\n", ENVFILE))
	f.WriteString(fmt.Sprintf("ExecStart=%s start -f\n", exeFile))
	f.WriteString(fmt.Sprintf("ExecReload=/bin/kill -HUP $MAINPID\n"))
	f.WriteString("KillMode=process\n")
	f.WriteString("\n")
	f.WriteString("[Install]\n")
	f.WriteString("WantedBy=docker.service\n")
	f.WriteString("\n")
}

func createInitFile(exeFile string) {
	os.Symlink(exeFile, INTFILE)
}

const RexRayLogoAscii = `
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
