package os

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/pkg/mount"
	osdriver "github.com/emccode/rexray/drivers/os"
)

var (
	debug     string
	osDrivers string
)

func init() {
	debug = strings.ToUpper(os.Getenv("REXRAY_DEBUG"))
	initOSDrivers()
}

func initOSDrivers() {
	osDrivers = strings.ToLower(os.Getenv("REXRAY_OSDRIVERS"))
	var err error
	osdriver.Adapters, err = osdriver.GetDrivers(osDrivers)
	if err != nil && debug == "TRUE" {
		fmt.Println(err)
	}
	if len(osdriver.Adapters) == 0 {
		if debug == "true" {
			fmt.Println("Rexray: No OS adapters initialized")
		}
	}
}

func GetMounts(deviceName, mountPoint string) ([]*mount.Info, error) {

	for _, driver := range osdriver.Adapters {
		mounts, err := driver.GetMounts(deviceName, mountPoint)
		if err != nil {
			return nil, err
		}
		return mounts, nil
	}

	return nil, errors.New("No OS detected")
}

func Mounted(mountPoint string) (bool, error) {
	for _, driver := range osdriver.Adapters {
		return driver.Mounted(mountPoint)
	}
	return false, errors.New("No OS detected")
}

func Unmount(mountPoint string) error {
	for _, driver := range osdriver.Adapters {
		return driver.Unmount(mountPoint)
	}
	return errors.New("No OS detected")
}

func Mount(device, target, mountOptions, mountLabel string) error {
	for _, driver := range osdriver.Adapters {
		return driver.Mount(device, target, mountOptions, mountLabel)
	}
	return errors.New("No OS detected")
}

func Format(deviceName, fsType string, overwriteFs bool) error {
	for _, driver := range osdriver.Adapters {
		return driver.Format(deviceName, fsType, overwriteFs)
	}
	return errors.New("No OS detected")
}
