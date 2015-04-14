package rexray

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/emccode/rexray/storagedriver"
)

var (
	debug          string
	storageDrivers string
)
var (
	ErrDriverBlockDeviceDiscovery = errors.New("Driver Block Device discovery failed")
)

func init() {
	debug = strings.ToUpper(os.Getenv("REXRAY_DEBUG"))
	storageDrivers = strings.ToLower(os.Getenv("REXRAY_STORAGEDRIVERS"))
}

// GetBlockDeviceMapping performs storage introspection and
// returns a listing of block devices from the guest
func GetBlockDeviceMapping() ([]*storagedriver.BlockDevice, error) {
	drivers, err := storagedriver.GetDrivers(storageDrivers)
	if err != nil && debug == "TRUE" {
		fmt.Println(err)
	}

	var allBlockDevices []*storagedriver.BlockDevice
	for _, driver := range drivers {
		blockDevices, err := driver.GetBlockDeviceMapping()
		if err != nil {
			return []*storagedriver.BlockDevice{}, fmt.Errorf("Error: %s: %s", ErrDriverBlockDeviceDiscovery, err)
		}

		if len(blockDevices.([]*storagedriver.BlockDevice)) > 0 {
			for _, blockDevice := range blockDevices.([]*storagedriver.BlockDevice) {
				allBlockDevices = append(allBlockDevices, blockDevice)
			}
		}
	}

	return allBlockDevices, nil

}
