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
	// Mount will return a mount point path when specifying either a volumeName or volumeID.  If a overwriteFs boolean
	// is specified it will overwrite the FS based on newFsType if it is detected that there is no FS present.
	Mount(volumeName, volumeID string, overwriteFs bool, newFsType string) (string, error)

	// Unmount will unmount the specified volume by volumeName or volumeID.
	Unmount(volumeName, volumeID string) error

	// Path will return the mounted path of the volumeName or volumeID.
	Path(volumeName, volumeID string) (string, error)

	// Create will create a new volume with the volumeName.
	Create(volumeName string) error

	// Remove will remove a volume of volumeName.
	Remove(volumeName string) error

	// Attach will attach a volume based on volumeName to the instance of instanceID.
	Attach(volumeName, instanceID string) (string, error)

	// Detach will detach a volume based on volumeName to the instance of instanceID.
	Detach(volumeName, instanceID string) error

	// NetworkName will return an identifier of a volume that is relevant when corelating a
	// local device to a device that is the volumeName to the local instanceID.
	NetworkName(volumeName, instanceID string) (string, error)
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
