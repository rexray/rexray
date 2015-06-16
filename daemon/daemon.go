package daemon

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/emccode/rexray/drivers/daemon"
)

var (
	debug         string
	daemonDrivers string
)

const (
	defaultDaemon string = "dockervolumedriver"
)

func init() {
	debug = strings.ToUpper(os.Getenv("REXRAY_DEBUG"))
	initDaemonDrivers()
}

func initDaemonDrivers() {
	daemonDrivers = strings.ToLower(os.Getenv("REXRAY_DAEMONDRIVERS"))
	var err error
	daemondriver.Adapters, err = daemondriver.GetDrivers(daemonDrivers)
	if err != nil && debug == "TRUE" {
		fmt.Println(err)
	}
	if len(daemondriver.Adapters) == 0 {
		if debug == "true" {
			fmt.Println("Rexray: No daemon adapters initialized")
		}
	}

}

func Start(hostname string) error {
	if len(daemondriver.Adapters) > 0 {
		for _, driver := range daemondriver.Adapters {
			return driver.Start(hostname)
		}
	}

	return errors.New("No daemon driver initialized")
}
