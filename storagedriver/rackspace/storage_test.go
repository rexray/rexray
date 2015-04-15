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

func TestGetVolume(*testing.T) {
	volume, err := driver.GetVolume("43de157d-3dfb-441f-b832-4d2d8cf457cc")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", volume))
}

func TestGetVolumeAttach(*testing.T) {
	volume, err := driver.GetVolumeAttach("43de157d-3dfb-441f-b832-4d2d8cf457cc", "5ad7727c-aa5a-43e4-8ab7-a499295032d7")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", volume))
}

func TestGetSnapshotFromVolumeID(*testing.T) {
	snapshots, err := driver.GetSnapshot("738ea6b9-8c49-416c-97b7-a5264a799eb6", "")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", snapshots.([]*storagedriver.Snapshot)))
}

func TestGetSnapshotFromSnapshotID(*testing.T) {
	snapshots, err := driver.GetSnapshot("", "83743ccc-200f-45bb-8144-e802ceb4b555")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", snapshots.([]*storagedriver.Snapshot)[0]))
}
