// +build !client

package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"text/template"

	"github.com/akutz/gotil"
	log "github.com/sirupsen/logrus"

	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/util"
)

func init() {
	installFunc = install
	uninstallFunc = uninstall
}

// init system types
const (
	Unknown = iota
	SystemD
	UpdateRcD
	ChkConfig
)

func install(ctx apitypes.Context) {
	checkOpPerms("installed")
	switch runtime.GOOS {
	case "linux":
		switch getInitSystemType() {
		case SystemD:
			installSystemD(ctx)
		case UpdateRcD:
			installUpdateRcd()
		case ChkConfig:
			installChkConfig()
		}
	}
}

func isRpmInstall(pkgName *string) bool {
	exePath := util.BinFilePath
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
	*pkgName = gotil.Trim(soutput)

	log.WithFields(log.Fields{
		"exePath": exePath,
		"pkgName": *pkgName,
	}).Debug("is rpm install success")
	return true
}

func isDebInstall(pkgName *string) bool {
	exePath := util.BinFilePath
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
	*pkgName = strings.Split(gotil.Trim(soutput), ":")[0]

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

func uninstall(ctx apitypes.Context, pkgManager bool) {
	checkOpPerms("uninstalled")

	binFilePath := util.BinFilePath

	// if the uninstall command was executed manually we should check to see
	// if this file is owned by a package manager and remove it that way if so
	if !pkgManager {
		log.WithField("binFilePath", binFilePath).Debug(
			"is this a managed file?")
		var pkgName string
		if isRpmInstall(&pkgName) {
			uninstallRpm(pkgName)
			return
		} else if isDebInstall(&pkgName) {
			uninstallDeb(pkgName)
			return
		}
	}

	func() {
		defer func() {
			recover()
		}()
		serviceStop(ctx)
	}()

	switch getInitSystemType() {
	case SystemD:
		uninstallSystemD()
	case UpdateRcD:
		uninstallUpdateRcd()
	case ChkConfig:
		uninstallChkConfig()
	}

	if !pkgManager {
		os.Remove(binFilePath)
		if util.IsPrefixed(ctx) {
			os.RemoveAll(util.GetPrefix(ctx))
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
	if gotil.FileExistsInPath("systemctl") {
		return SystemD
	}

	if gotil.FileExistsInPath("update-rc.d") {
		return UpdateRcD
	}

	if gotil.FileExistsInPath("chkconfig") {
		return ChkConfig
	}

	return Unknown
}

func installSystemD(ctx apitypes.Context) {
	if err := createUnitFile(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: install failed: %v\n", err)
		os.Exit(1)
	}
	if err := createEnvFile(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: install failed: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("systemctl", "enable", "-q", util.UnitFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: install failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("REX-Ray is now installed. Before starting it please check ")
	fmt.Print("http://github.com/AVENTER-UG/rexray for instructions on how to ")
	fmt.Print("configure it.\n\n Once configured the REX-Ray service can be ")
	fmt.Print("started with the command 'sudo systemctl start rexray'.\n\n")
}

func uninstallSystemD() {

	// a link created by systemd as docker should "want" rexray as a service.
	// the uninstaller will fail
	os.Remove(path.Join("/etc/systemd/system/docker.service.wants",
		util.UnitFileName))

	cmd := exec.Command("systemctl", "disable", "-q", util.UnitFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: uninstall failed: %v\n", err)
		os.Exit(1)
	}

	os.Remove(util.UnitFilePath)
}

func installUpdateRcd() {
	if err := createInitFile(); err != nil {
		fmt.Fprintf(os.Stderr, "error: install failed: %v\n", err)
		os.Exit(1)
	}
	cmd := exec.Command("update-rc.d", util.InitFileName, "defaults")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: install failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Print("REX-Ray is now installed. Before starting it please check ")
	fmt.Print("http://github.com/AVENTER-UG/rexray for instructions on how to ")
	fmt.Print("configure it.\n\n Once configured the REX-Ray service can be ")
	fmt.Print("started with the command ")
	fmt.Printf("'sudo %s start'.\n\n", util.InitFilePath)
}

func uninstallUpdateRcd() {
	os.Remove(util.InitFilePath)
	cmd := exec.Command("update-rc.d", util.InitFileName, "remove")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: uninstall failed: %v\n", err)
		os.Exit(1)
	}
}

func installChkConfig() {
	if err := createInitFile(); err != nil {
		fmt.Fprintf(os.Stderr, "error: install failed: %v\n", err)
		os.Exit(1)
	}
	cmd := exec.Command("chkconfig", util.InitFileName, "on")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: install failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Print("REX-Ray is now installed. Before starting it please check ")
	fmt.Print("http://github.com/AVENTER-UG/rexray for instructions on how to ")
	fmt.Print("configure it.\n\n Once configured the REX-Ray service can be ")
	fmt.Print("started with the command ")
	fmt.Printf("'sudo %s start'.\n\n", util.InitFilePath)
}

func uninstallChkConfig() {
	cmd := exec.Command("chkconfig", "--del", util.InitFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: uninstall failed: %v\n", err)
		os.Exit(1)
	}
	os.Remove(util.InitFilePath)
}

func createEnvFile(ctx apitypes.Context) error {
	envFilePath := util.EnvFilePath(ctx)
	f, err := os.OpenFile(envFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open env file failed: %s: %v", envFilePath, err)
	}
	defer f.Close()
	if util.IsPrefixed(ctx) {
		f.WriteString("REXRAY_HOME=")
		f.WriteString(util.GetPrefix(ctx))
	}
	return nil
}

func createUnitFile(ctx apitypes.Context) error {
	data := struct {
		BinFileName string
		BinFilePath string
		EnvFilePath string
	}{
		util.BinFileName,
		util.BinFilePath,
		util.EnvFilePath(ctx),
	}
	tmpl, err := template.New("UnitFile").Parse(unitFileTemplate)
	if err != nil {
		return fmt.Errorf("create unit file template failed: %v", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("exec unit file template failed: %v", err)
	}
	text := buf.String()
	f, err := os.OpenFile(util.UnitFilePath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf(
			"create unit file failed: %s: %v", util.UnitFilePath, err)
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		return fmt.Errorf(
			"write unit file failed: %s: %v", util.UnitFilePath, err)
	}

	return nil
}

const unitFileTemplate = `[Unit]
Description={{.BinFileName}}
Wants=scini.service
Before=docker.service
After=scini.service

[Service]
EnvironmentFile={{.EnvFilePath}}
ExecStart={{.BinFilePath}} start
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process

[Install]
WantedBy=docker.service
`

func createInitFile() error {
	data := struct {
		BinFileName string
		BinFilePath string
	}{
		util.BinFileName,
		util.BinFilePath,
	}
	tmpl, err := template.New("InitScript").Parse(initScriptTemplate)
	if err != nil {
		return fmt.Errorf("create init file template failed: %v", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("exec unit file template failed: %v", err)
	}
	text := buf.String()
	// wrapped in a function to defer the close to ensure file is written to
	// disk before subsequent chmod below
	if err := func() error {
		f, err := os.OpenFile(util.InitFilePath,
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf(
				"create init file failed: %s: %v", util.InitFilePath, err)
		}
		defer f.Close()
		if _, err := f.WriteString(text); err != nil {
			return fmt.Errorf(
				"write init file failed: %s: %v", util.InitFilePath, err)
		}
		return nil
	}(); err != nil {
		return err
	}
	if err := os.Chmod(util.InitFilePath, 0755); err != nil {
		return fmt.Errorf(
			"chmod init file failed: %s: %v", util.InitFilePath, err)
	}
	return nil
}

const initScriptTemplate = `### BEGIN INIT INFO
# Provides:          {{.BinFileName}}
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Should-Start:      scini
# X-Start-Before:    docker
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Start daemon at boot time
# Description:       Enable service provided by daemon.
### END INIT INFO

case "$1" in
  start)
    {{.BinFilePath}} start
    ;;
  stop)
    {{.BinFilePath}} stop
    ;;
  status)
    {{.BinFilePath}} status
    ;;
  restart)
    {{.BinFilePath}} restart
    ;;
  reload)
    {{.BinFilePath}} reload
    ;;
  force-reload)
    {{.BinFilePath}} force-reload
    ;;
  *)
    echo "Usage: $0 {start|stop|status|restart|reload|force-reload}"
esac
`
