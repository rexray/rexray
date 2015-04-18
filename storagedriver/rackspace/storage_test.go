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
	volume, err := driver.GetVolume("ccde08e3-d21b-467a-a7d3-bc92ffe0a14f", "")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", volume))
}

func TestGetVolumeByName(*testing.T) {
	volumes, err := driver.GetVolume("", "Volume-1")
	if err != nil {
		panic(err)
	}
	for _, volume := range volumes.([]*storagedriver.Volume) {
		fmt.Println(fmt.Sprintf("%+v", volume))
	}
}

func TestGetVolumeAttach(*testing.T) {
	volume, err := driver.GetVolumeAttach("12b64bd3-2c34-4fe1-b389-5cf8df668ef5", "5ad7727c-aa5a-43e4-8ab7-a499295032d7")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", volume))
}

func TestGetSnapshotFromVolumeID(*testing.T) {
	snapshots, err := driver.GetSnapshot("738ea6b9-8c49-416c-97b7-a5264a799eb6", "", "")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", snapshots.([]*storagedriver.Snapshot)))
}

func TestGetSnapshotBySnapshotName(*testing.T) {
	snapshots, err := driver.GetSnapshot("", "", "Volume-1-1")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", snapshots.([]*storagedriver.Snapshot)))
}

func TestGetSnapshotFromSnapshotID(*testing.T) {
	snapshots, err := driver.GetSnapshot("", "83743ccc-200f-45bb-8144-e802ceb4b555", "")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", snapshots.([]*storagedriver.Snapshot)))
}

func TestCreateSnapshot(*testing.T) {
	snapshot, err := driver.CreateSnapshot(false, "testing", "87ef25ed-9c5f-4030-ada7-eeaf4cba0814", "")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", snapshot))
}

func TestRemoveSnapshot(*testing.T) {
	err := driver.RemoveSnapshot("ea14a2f0-16b2-47e9-b7ba-01d812f65205")
	if err != nil {
		panic(err)
	}
}

func TestCreateVolume(*testing.T) {
	volume, err := driver.CreateVolume(false, "testing", "", "", "", 0, 75)
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", volume))
}

func TestRemoveVolume(*testing.T) {
	err := driver.RemoveVolume("743e9de5-8de4-4f09-8249-0238849a3a29")
	if err != nil {
		panic(err)
	}
}

func TestGetDeviceNextAvailable(*testing.T) {
	deviceName, err := driver.GetDeviceNextAvailable()
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf(deviceName))
}

func TestAttachVolume(*testing.T) {
	volumeAttachment, err := driver.AttachVolume(false, "94e02a4a-71dc-4026-b561-1cd0cad37bce", "5ad7727c-aa5a-43e4-8ab7-a499295032d7")
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", volumeAttachment.(storagedriver.VolumeAttachment)))
}

func TestDetachVolume(*testing.T) {
	err := driver.DetachVolume(false, "94e02a4a-71dc-4026-b561-1cd0cad37bce", "5ad7727c-aa5a-43e4-8ab7-a499295032d7")
	if err != nil {
		panic(err)
	}

}
