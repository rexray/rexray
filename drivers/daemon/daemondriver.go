package daemondriver

import (
	"fmt"
	"os"
	"strings"

	"github.com/emccode/rexray/util"
)

var (
	driverInitFuncs map[string]InitFunc
	drivers         map[string]Driver
	debug           string
)

var Adapters map[string]Driver

type Driver interface {
	// Starts the daemon
	Start(string) error
}

type InitFunc func() (Driver, error)

func Register(name string, initFunc InitFunc) error {
	driverInitFuncs[name] = initFunc
	return nil
}

func init() {
	driverInitFuncs = make(map[string]InitFunc)
	drivers = make(map[string]Driver)
	debug = strings.ToUpper(os.Getenv("REXRAY_DEBUG"))
}

func GetDrivers(daemonDrivers string) (map[string]Driver, error) {
	var err error
	var daemonDriversArr []string
	if daemonDrivers != "" {
		daemonDriversArr = strings.Split(daemonDrivers, ",")
	}

	if debug == "TRUE" {
		fmt.Println(driverInitFuncs)
	}

	driverPriority := []string{
		"dockervolumedriver",
		"dockerremotevolumedriver",
	}

	for _, name := range driverPriority {

		if len(daemonDriversArr) > 0 && !util.StringInSlice(name, daemonDriversArr) {
			continue
		}

		var initFunc InitFunc
		var ok bool
		if initFunc, ok = driverInitFuncs[name]; !ok {
			continue
		}

		drivers[name], err = initFunc()
		if err != nil {
			if debug == "TRUE" {
				fmt.Println(fmt.Sprintf("Info (%s): %s", name, err))
			}
			delete(drivers, name)
		}
	}

	return drivers, nil
}
