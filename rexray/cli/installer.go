package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"

	"github.com/emccode/rexray/util"
)

// init system types
const (
	Unknown = iota
	SystemD
	UpdateRcD
	ChkConfig
)

func install() {
	checkOpPerms("installed")

	_, _, exeFile := util.GetThisPathParts()

	if runtime.GOOS == "linux" {
		switch getInitSystemType() {
		case SystemD:
			installSystemD(exeFile)
		case UpdateRcD:
			installUpdateRcd(exeFile)
		case ChkConfig:
			installChkConfig(exeFile)
		}
	}
}

func isRpmInstall(exePath string, pkgName *string) bool {
	cmd := exec.Command("rpm", "-qf", exePath)
	output, err := cmd.CombinedOutput()
	soutput := string(output)
	if err != nil {
		log.WithFields(log.Fields{
			"exePath": exePath,
			"output":  soutput,
			"error":   err,
		}).Debug("error checking if rpm install")
		return false
	}
	log.WithField("output", soutput).Debug("rpm install query result")
	*pkgName = util.Trim(soutput)

	log.WithFields(log.Fields{
		"exePath": exePath,
		"pkgName": *pkgName,
	}).Debug("is rpm install success")
	return true
}

func isDebInstall(exePath string, pkgName *string) bool {
	cmd := exec.Command("dpkg-query", "-S", exePath)
	output, err := cmd.CombinedOutput()
	soutput := string(output)
	if err != nil {
		log.WithFields(log.Fields{
			"exePath": exePath,
			"output":  soutput,
			"error":   err,
		}).Debug("error checking if deb install")
		return false
	}
	log.WithField("output", soutput).Debug("deb install query result")
	*pkgName = strings.Split(util.Trim(soutput), ":")[0]

	log.WithFields(log.Fields{
		"exePath": exePath,
		"pkgName": *pkgName,
	}).Debug("is deb install success")
	return true
}

func uninstallRpm(pkgName string) bool {
	output, err := exec.Command("rpm", "-e", pkgName).CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{
			"pkgName": pkgName,
			"output":  string(output),
			"error":   err,
		}).Error("error uninstalling rpm")
	}
	return true
}

func uninstallDeb(pkgName string) bool {
	output, err := exec.Command("dpkg", "-r", pkgName).CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{
			"pkgName": pkgName,
			"output":  string(output),
			"error":   err,
		}).Error("error uninstalling deb")
	}
	return true
}

func uninstall(pkgManager bool) {
	checkOpPerms("uninstalled")

	_, _, binFile := util.GetThisPathParts()

	// if the uninstall command was executed manually we should check to see
	// if this file is owned by a package manager and remove it that way if so
	if !pkgManager {
		log.WithField("binFile", binFile).Debug("is this a managed file?")
		var pkgName string
		if isRpmInstall(binFile, &pkgName) {
			uninstallRpm(pkgName)
			return
		} else if isDebInstall(binFile, &pkgName) {
			uninstallDeb(pkgName)
			return
		}
	}

	func() {
		defer func() {
			recover()
		}()
		stop()
	}()

	switch getInitSystemType() {
	case SystemD:
		uninstallSystemD()
	case UpdateRcD:
		uninstallUpdateRcd()
	case ChkConfig:
		uninstallChkConfig()
	}

	os.RemoveAll(util.EtcDirPath())
	os.RemoveAll(util.RunDirPath())
	os.RemoveAll(util.LibDirPath())
	os.RemoveAll(util.LogDirPath())

	if !pkgManager {
		os.Remove(binFile)
		if util.IsPrefixed() {
			os.RemoveAll(util.GetPrefix())
		}
	}
}

func getInitSystemCmd() string {
	switch getInitSystemType() {
	case SystemD:
		return "systemd"
	case UpdateRcD:
		return "update-rc.d"
	case ChkConfig:
		return "chkconfig"
	default:
		return "unknown"
	}
}

func getInitSystemType() int {
	if util.FileExistsInPath("systemctl") {
		return SystemD
	}

	if util.FileExistsInPath("update-rc.d") {
		return UpdateRcD
	}

	if util.FileExistsInPath("chkconfig") {
		return ChkConfig
	}

	return Unknown
}

func installSystemD(exeFile string) {
	createUnitFile(exeFile)
	createEnvFile()

	cmd := exec.Command("systemctl", "enable", "-q", "rexray.service")
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

func uninstallSystemD() {

	// a link created by systemd as docker should "want" rexray as a service.
	// the uninstaller will fail
	os.Remove("/etc/systemd/system/docker.service.wants/rexray.service")

	cmd := exec.Command("systemctl", "disable", "-q", "rexray.service")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Fatalf("uninstallation error %v", err)
	}

	os.Remove(util.UnitFilePath)
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

func uninstallUpdateRcd() {

	os.Remove(util.InitFilePath)

	cmd := exec.Command("update-rc.d", "rexray", "remove")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Fatalf("uninstallation error %v", err)
	}
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

func uninstallChkConfig() {
	cmd := exec.Command("chkconfig", "--del", "rexray")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Fatalf("uninstallation error %v", err)
	}

	os.Remove(util.InitFilePath)
}

func createEnvFile() {
	f, err := os.OpenFile(
		util.EtcFilePath(util.EnvFileName), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if util.IsPrefixed() {
		f.WriteString("REXRAY_HOME=")
		f.WriteString(util.GetPrefix())
	}
}

func createUnitFile(exeFile string) {

	data := struct {
		RexrayBin string
		EnvFile   string
	}{
		exeFile,
		util.EtcFilePath(util.EnvFileName),
	}

	tmpl, err := template.New("UnitFile").Parse(unitFileTemplate)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	text := buf.String()

	f, err := os.OpenFile(util.UnitFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString(text)
}

const unitFileTemplate = `[Unit]
Description=rexray
Before=docker.service

[Service]
EnvironmentFile={{.EnvFile}}
ExecStart={{.RexrayBin}} start -f
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process

[Install]
WantedBy=docker.service
`

func createInitFile(exeFile string) {

	data := struct {
		RexrayBin string
	}{
		exeFile,
	}

	tmpl, err := template.New("InitScript").Parse(initScriptTemplate)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	text := buf.String()

	// wrapped in a function to defer the close to ensure file is written to
	// disk before subsequent chmod below
	func() {
		f, err := os.OpenFile(util.InitFilePath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		f.WriteString(text)
	}()

	os.Chmod(util.InitFilePath, 0755)
}

const initScriptTemplate = `### BEGIN INIT INFO
# Provides:          rexray
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Start daemon at boot time
# Description:       Enable service provided by daemon.
### END INIT INFO

case "$1" in
  start)
    {{.RexrayBin}} start
    ;;
  stop)
    {{.RexrayBin}} stop
    ;;
  status)
    {{.RexrayBin}} status
    ;;
  retart)
    {{.RexrayBin}} restart
    ;;
  reload)
    {{.RexrayBin}} reload
    ;;
  force-reload)
    {{.RexrayBin}} force-reload
    ;;
  *)
    echo "Usage: $0 {start|stop|status|restart|reload|force-reload}"
esac
`
