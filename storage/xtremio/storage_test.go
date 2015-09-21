package xtremio

import "fmt"
import "testing"
import "github.com/emccode/godogged/storagedriver"

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

func TestGetVolumeMapping(*testing.T) {
	blockDeviceMapping, err := driver.GetVolumeMapping()
	if err != nil {
		panic(err)
	}

	for _, blockDevice := range blockDeviceMapping {
		fmt.Println(fmt.Sprintf("%+v", blockDevice))
	}
}

func TestGetVolume(*testing.T) {
	volumes, err := driver.GetVolume("", "")
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

func TestGetVolumeByID(*testing.T) {
	volumes, err := driver.GetVolume("30", "")
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

func TestGetVolumeByName(*testing.T) {
	volumes, err := driver.GetVolume("", "ubuntu-vol4")
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

func TestCreateVolume(*testing.T) {
	volume, err := driver.CreateVolume(false, "testing12", "", "", "", 0, 1, "")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", volume))
}

func TestCreateVolumeFromVolID(*testing.T) {
	volume, err := driver.CreateVolume(false, "testing12", "24", "", "", 0, 1, "")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", volume))
}

func TestCreateVolumeFromSnapshotID(*testing.T) {
	volume, err := driver.CreateVolume(false, "testing12", "", "25", "", 0, 1, "")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", volume))
}

func TestRemoveVolume(*testing.T) {
	err := driver.RemoveVolume("30")
	if err != nil {
		panic(err)
	}
}

func TestGetSnapshot(*testing.T) {
	volumes, err := driver.GetSnapshot("", "", "")
	if err != nil {
		panic(err)
	}
	for _, snapshot := range volumes {
		fmt.Println(fmt.Sprintf("%+v", snapshot))
	}
}

func TestGetSnapshotByID(*testing.T) {
	volumes, err := driver.GetSnapshot("", "25", "")
	if err != nil {
		panic(err)
	}
	for _, snapshot := range volumes {
		fmt.Println(fmt.Sprintf("%+v", snapshot))
	}
}

func TestGetSnapshotByName(*testing.T) {
	volumes, err := driver.GetSnapshot("", "", "ubuntu-vol4.snap.06022015-09:58")
	if err != nil {
		panic(err)
	}
	for _, snapshot := range volumes {
		fmt.Println(fmt.Sprintf("%+v", snapshot))
	}
}

func TestGetSnapshotByVolID(*testing.T) {
	volumes, err := driver.GetSnapshot("24", "", "")
	if err != nil {
		panic(err)
	}
	for _, snapshot := range volumes {
		fmt.Println(fmt.Sprintf("%+v", snapshot))
	}
}

func TestCreateSnapshot(*testing.T) {
	snapshots, err := driver.CreateSnapshot(false, "testing6", "24", "")
	if err != nil {
		panic(err)
	}
	for _, snapshot := range snapshots {
		fmt.Println(fmt.Sprintf("%+v", snapshot))
	}
}

func TestRemoveSnapshot(*testing.T) {
	err := driver.RemoveSnapshot("26")
	if err != nil {
		panic(err)
	}
}

func TestGetVolumeAttach(*testing.T) {
	volume, err := driver.GetVolumeAttach("24", "")
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", volume[0]))
}

func TestAttachVolume(*testing.T) {
	volumeAttachments, err := driver.AttachVolume(false, "24", "")
	if err != nil {
		panic(err)
	}

	for volumeAttachment := range volumeAttachments {
		fmt.Println(fmt.Sprintf("%+v", volumeAttachment))
	}
}

func TestGetLunMaps(*testing.T) {
	instanceLunMaps, err := getLunMaps("iqn.1993-08.org.debian:01:dca5bccb64", "24")
	if err != nil {
		panic(err)
	}

	for _, lunMap := range instanceLunMaps {
		fmt.Println(fmt.Sprintf("%+v", lunMap))
	}

}

func TestDetachVolume(*testing.T) {
	err := driver.DetachVolume(false, "24", "")
	if err != nil {
		panic(err)
	}

}
