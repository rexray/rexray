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

func Mount(volumeName, volumeID string, overwriteFs bool, newFsType string) (string, error) {
	for _, driver := range volumedriver.Adapters {
		return driver.Mount(volumeName, volumeID, overwriteFs, newFsType)
	}
	return "", errors.New("No Volume Manager specified")
}

func Unmount(volumeName, volumeID string) error {
	for _, driver := range volumedriver.Adapters {
		return driver.Unmount(volumeName, volumeID)
	}
	return errors.New("No Volume Manager specified")
}

func Path(volumeName, volumeID string) (string, error) {
	for _, driver := range volumedriver.Adapters {
		return driver.Path(volumeName, volumeID)
	}
	return "", errors.New("No Volume Manager specified")
}

func Create(volumeName string) error {
	for _, driver := range volumedriver.Adapters {
		return driver.Create(volumeName)
	}
	return errors.New("No Volume Manager specified")
}

func Remove(volumeName string) error {
	for _, driver := range volumedriver.Adapters {
		return driver.Remove(volumeName)
	}
	return errors.New("No Volume Manager specified")
}

func Attach(volumeName, instanceID string) (string, error) {
	for _, driver := range volumedriver.Adapters {
		return driver.Attach(volumeName, instanceID)
	}
	return "", errors.New("No Volume Manager specified")
}

func Detach(volumeName, instanceID string) error {
	for _, driver := range volumedriver.Adapters {
		return driver.Detach(volumeName, instanceID)
	}
	return errors.New("No Volume Manager specified")
}

func NetworkName(volumeName, instanceID string) (string, error) {
	for _, driver := range volumedriver.Adapters {
		return driver.NetworkName(volumeName, instanceID)
	}
	return "", errors.New("No Volume Manager specified")
}
