package volume

import (
	"errors"
	"fmt"
	"os"
	"strings"

	volumedriver "github.com/emccode/rexray/drivers/volume"
)

var (
	debug         string
	volumeDrivers string
)

func init() {
	debug = strings.ToUpper(os.Getenv("REXRAY_DEBUG"))
	initVolumeDrivers()
}

func initVolumeDrivers() {
	volumeDrivers = strings.ToLower(os.Getenv("REXRAY_VOLUMEDRIVERS"))
	var err error
	volumedriver.Adapters, err = volumedriver.GetDrivers(volumeDrivers)
	if err != nil && debug == "TRUE" {
		fmt.Println(err)
	}
	if len(volumedriver.Adapters) == 0 {
		if debug == "true" {
			fmt.Println("Rexray: No volume manager adapters initialized")
		}
	}

}

func MountVolume(volumeName, volumeID string, overwriteFs bool, newFsType string) (string, error) {
	for _, driver := range volumedriver.Adapters {
		return driver.MountVolume(volumeName, volumeID, overwriteFs, newFsType)
	}
	return "", errors.New("No Volume Manager specified")
}

func UnmountVolume(volumeName, volumeID string) error {
	for _, driver := range volumedriver.Adapters {
		return driver.UnmountVolume(volumeName, volumeID)
	}
	return errors.New("No Volume Manager specified")
}
