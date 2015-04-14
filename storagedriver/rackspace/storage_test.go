package rackspace

import "fmt"
import "testing"
import "github.com/emccode/rexray/storagedriver"

var driver storagedriver.Driver

func init() {
	var err error
	driver, err = Init()
	if err != nil {
		panic(err)
	}
}

func TestGetInstanceID(*testing.T) {
	instanceID, err := getInstanceID()
	if err != nil {
		panic(err)
	}

	fmt.Println(instanceID)
}

func TestGetBlockDeviceMapping(*testing.T) {
	blockDeviceMapping, err := driver.GetBlockDeviceMapping()
	if err != nil {
		panic(err)
	}

	for _, blockDevice := range blockDeviceMapping.([]*storagedriver.BlockDevice) {
		fmt.Println(fmt.Sprintf("%+v", blockDevice))
	}
}

func TestGetInstance(*testing.T) {
	instance, err := driver.GetInstance()
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", instance))
}
