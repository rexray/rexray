package scaleio

import "fmt"
import "testing"
import "github.com/emccode/rexray/drivers/storage"

var driver storagedriver.Driver

func init() {
	var err error
	driver, err = Init()
	if err != nil {
		panic(err)
	}
}

func TestGetInstance(*testing.T) {
	instance, err := driver.GetInstance()
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", instance))
}

func TestGetBlockDeviceMapping(*testing.T) {
	blockDeviceMapping, err := driver.GetBlockDeviceMapping()
	if err != nil {
		panic(err)
	}

	for _, blockDevice := range blockDeviceMapping {
		fmt.Println(fmt.Sprintf("%+v", blockDevice))
	}
}

func TestGetVolume(*testing.T) {
	instance, err := driver.GetInstance()
	if err != nil {
		panic(err)
	}

	volumes, err := driver.GetVolume("", instance.InstanceID)
	if err != nil {
		panic(err)
	}
	for _, volume := range volumes {
		fmt.Println(fmt.Sprintf("%+v", volume))
		for _, attachment := range volume.Attachments {
			fmt.Println(fmt.Sprintf("%+v", attachment))
		}
	}
}

func TestGetVolumeAttach(*testing.T) {
	volume, err := driver.GetVolumeAttach("e55b0ead00000000", "d28bed1900000000")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", volume[0]))
}

func TestCreateSnapshot(*testing.T) {
	snapshots, err := driver.CreateSnapshot(false, "testing6", "e55b0ead00000000", "")
	if err != nil {
		panic(err)
	}
	for _, snapshot := range snapshots {
		fmt.Println(fmt.Sprintf("%+v", snapshot))
	}
}

func TestGetSnapshotFromVolumeID(*testing.T) {
	snapshots, err := driver.GetSnapshot("e55b0eb000000003", "", "")
	if err != nil {
		panic(err)
	}
	for _, snapshot := range snapshots {
		fmt.Println(fmt.Sprintf("%+v", snapshot))
	}
}

func TestCreateVolume(*testing.T) {
	volume, err := driver.CreateVolume(false, "testing12", "", "", "", 0, 1, "")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", volume))
}

func TestRemoveVolume(*testing.T) {
	err := driver.RemoveVolume("e55b0eb000000003")
	if err != nil {
		panic(err)
	}
}

func TestRemoveSnapshot(*testing.T) {
	err := driver.RemoveVolume("e55b0ec500000003")
	if err != nil {
		panic(err)
	}
}

func TestAttachVolume(*testing.T) {
	volumeAttachments, err := driver.AttachVolume(false, "e55b0ec600000003", "d28bed1900000000")
	if err != nil {
		panic(err)
	}

	for volumeAttachment := range volumeAttachments {
		fmt.Println(fmt.Sprintf("%+v", volumeAttachment))
	}
}

func TestDetachVolume(*testing.T) {
	err := driver.DetachVolume(false, "e55b0ec600000003", "d28bed1900000000")
	if err != nil {
		panic(err)
	}

}
