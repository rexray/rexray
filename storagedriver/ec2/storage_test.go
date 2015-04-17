package ec2

import (
	"fmt"
	"log"
)

import "testing"
import (
	"github.com/emccode/rexray/storagedriver"
	"github.com/goamz/goamz/ec2"
)

var driver storagedriver.Driver

func init() {
	var err error
	driver, err = Init()
	if err != nil {
		panic(err)
	}
}

func TestGetInstanceIdentityDocument(*testing.T) {
	instanceIdentityDocument, err := getInstanceIdendityDocument()
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", instanceIdentityDocument))

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

func TestCreateSnapshot(*testing.T) {
	// (ec2 *EC2) CreateSnapshot(volumeId, description string)
	blockDeviceMapping, err := driver.GetBlockDeviceMapping()
	if err != nil {
		panic(err)
	}

	snapshot, err := driver.CreateSnapshot(false, "", blockDeviceMapping.([]*storagedriver.BlockDevice)[0].VolumeID, "test")
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", snapshot))
}

func TestGetSnapshot(*testing.T) {
	blockDeviceMapping, err := driver.GetBlockDeviceMapping()
	if err != nil {
		panic(err)
	}

	snapshots, err := driver.GetSnapshot(blockDeviceMapping.([]*storagedriver.BlockDevice)[0].VolumeID, "", "")
	if err != nil {
		panic(err)
	}

	for _, snapshot := range snapshots.([]*storagedriver.Snapshot) {
		fmt.Println(fmt.Sprintf("%+v", snapshot))
	}
}

func TestRemoveSnapshot(*testing.T) {
	blockDeviceMapping, err := driver.GetBlockDeviceMapping()
	if err != nil {
		panic(err)
	}
	snapshots, err := driver.GetSnapshot(blockDeviceMapping.([]*storagedriver.BlockDevice)[0].VolumeID, "", "")
	if err != nil {
		panic(err)
	}

	for _, snapshot := range snapshots.([]*storagedriver.Snapshot) {
		fmt.Println(fmt.Sprintf("%+v", snapshot))
	}

	snapshot, err := driver.CreateSnapshot(false, "", blockDeviceMapping.([]*storagedriver.BlockDevice)[0].VolumeID, "test")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", snapshot))

	err = driver.RemoveSnapshot(snapshot.(ec2.Snapshot).Id)
	if err != nil {
		panic(err)
	}

	snapshots, err = driver.GetSnapshot(blockDeviceMapping.([]*storagedriver.BlockDevice)[0].VolumeID, "", "")
	if err != nil {
		panic(err)
	}

	for _, snapshot := range snapshots.([]*storagedriver.Snapshot) {
		fmt.Println(fmt.Sprintf("%+v", snapshot))
	}
}

func TestGetDeviceNextAvailable(*testing.T) {

	deviceName, err := driver.GetDeviceNextAvailable()
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf(deviceName))

}

func TestCreateSnapshotVolume(*testing.T) {
	blockDeviceMapping, err := driver.GetBlockDeviceMapping()
	if err != nil {
		panic(err)
	}

	snapshot, err := driver.CreateSnapshot(false, "", blockDeviceMapping.([]*storagedriver.BlockDevice)[0].VolumeID, "test")
	if err != nil {
		panic(err)
	}

	volumeID, err := driver.CreateSnapshotVolume(false, "testing", snapshot.([]*storagedriver.Snapshot)[0].SnapshotID)
	if err != nil {
		panic(err)
	}

	err = driver.RemoveVolume(volumeID)
	if err != nil {
		panic(err)
	}

	err = driver.RemoveSnapshot(snapshot.([]*storagedriver.Snapshot)[0].SnapshotID)
	if err != nil {
		panic(err)
	}
}

func TestAttachVolume(*testing.T) {
	instance, err := driver.GetInstance()
	if err != nil {
		panic(err)
	}

	volume, err := driver.CreateVolume(false, "testing", "", "", 0, 2)
	if err != nil {
		panic(err)
	}

	volumeAttachment, err := driver.GetVolumeAttach(volume.(storagedriver.Volume).VolumeID, instance.(*storagedriver.Instance).InstanceID)
	if err != nil {
		panic(err)
	}

	log.Println(fmt.Sprintf("Volumes attached: %+v", volumeAttachment))

	volumeAttachment, err = driver.AttachVolume(false, volume.(storagedriver.Volume).VolumeID, instance.(*storagedriver.Instance).InstanceID)
	if err != nil {
		panic(err)
	}

	log.Println(fmt.Sprintf("Volumes attached: %+v", volumeAttachment))

	err = driver.DetachVolume(false, volumeAttachment.(storagedriver.VolumeAttachment).VolumeID, "")
	if err != nil {
		panic(err)
	}

	log.Println(fmt.Sprintf("Volume detached: %+v", volumeAttachment.(storagedriver.VolumeAttachment).VolumeID))

	err = driver.RemoveVolume(volume.(storagedriver.Volume).VolumeID)
	if err != nil {
		panic(err)
	}

	log.Println(fmt.Sprintf("Volume removed: %+v", volumeAttachment.(storagedriver.VolumeAttachment).VolumeID))
}

func TestGetVolume(*testing.T) {
	volume, err := driver.GetVolume("", "testing")
	if err != nil {
		panic(err)
	}
	for _, volume := range volume.([]*storagedriver.Volume) {
		fmt.Println(fmt.Sprintf("%+v", volume))
	}
}

func TestCreateVolume(*testing.T) {
	volume, err := driver.CreateVolume(true, "testing", "", "", 0, 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", volume.(*storagedriver.Volume)))
}

func TestCreateSnapshot2(*testing.T) {
	snapshots, err := driver.CreateSnapshot(false, "testing", "vol-8295eb9f", "test")
	if err != nil {
		panic(err)
	}
	for _, snapshot := range snapshots.([]*storagedriver.Snapshot) {
		fmt.Println(fmt.Sprintf("%+v", snapshot))
	}
}

func TestGetSnapshotByName(*testing.T) {
	volume, err := driver.GetSnapshot("", "", "testing")
	if err != nil {
		panic(err)
	}
	for _, snapshot := range volume.([]*storagedriver.Snapshot) {
		fmt.Println(fmt.Sprintf("%+v", snapshot))
	}
}

// func TestListTables(*testing.T) {
// 	ListTables()
// }
//
// func TestGetDDValue(*testing.T) {
// 	value, err := driver.GetDDValue("name1")
// 	if err != nil {
// 		panic(err)
// 	}
// }
