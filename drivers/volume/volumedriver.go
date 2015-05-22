package volumedriver

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
	//MountVolume will attach a Volume, prepare for mounting, and mount
	Mount(string, string, bool, string) (string, error)
	//UnmountVolume will unmount and detach a Volume
	Unmount(string, string) error
	//Path will return the mountpoint of a volume
	Path(string, string) (string, error)
	//Create will create a remote volume
	Create(string) error
	//Remove will remove a remote volume
	Remove(string) error
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

func GetDrivers(volumeDrivers string) (map[string]Driver, error) {
	var err error
	var volumeDriversArr []string
	if volumeDrivers != "" {
		volumeDriversArr = strings.Split(volumeDrivers, ",")
	}

	if debug == "TRUE" {
		fmt.Println(driverInitFuncs)
	}

	for name, initFunc := range driverInitFuncs {
		if len(volumeDriversArr) > 0 && !util.StringInSlice(name, volumeDriversArr) {
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
