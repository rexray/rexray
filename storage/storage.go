package storage

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	storagedriver "github.com/emccode/rexray/drivers/storage"
)

var (
	debug          string
	storageDrivers string
)

var (
	ErrDriverBlockDeviceDiscovery = errors.New("Driver Block Device discovery failed")
	ErrDriverInstanceDiscovery    = errors.New("Driver Instance discovery failed")
	ErrDriverVolumeDiscovery      = errors.New("Driver Volume discovery failed")
	ErrDriverSnapshotDiscovery    = errors.New("Driver Snapshot discovery failed")
	ErrMultipleDriversDetected    = errors.New("Multiple drivers detected, must declare with driver with env of REXRAY_STORAGEDRIVER=")
)

func init() {
	debug = strings.ToUpper(os.Getenv("REXRAY_DEBUG"))
	initStorageDrivers()
}

func initStorageDrivers() {
	storageDrivers = strings.ToLower(os.Getenv("REXRAY_STORAGEDRIVERS"))
	var err error
	storagedriver.Adapters, err = storagedriver.GetDrivers(storageDrivers)
	if err != nil && debug == "TRUE" {
		fmt.Println(err)
	}

	if len(storagedriver.Adapters) == 0 {
		if debug == "true" {
			fmt.Println("Rexray: No storage adapters initialized")
		}
	}
}

// GetVolumeMapping performs storage introspection and
// returns a listing of block devices from the guest
func GetVolumeMapping() ([]*storagedriver.BlockDevice, error) {
	var allBlockDevices []*storagedriver.BlockDevice
	for _, driver := range storagedriver.Adapters {
		blockDevices, err := driver.GetVolumeMapping()
		if err != nil {
			return []*storagedriver.BlockDevice{}, fmt.Errorf("Error: %s: %s", ErrDriverBlockDeviceDiscovery, err)
		}

		if len(blockDevices) > 0 {
			for _, blockDevice := range blockDevices {
				allBlockDevices = append(allBlockDevices, blockDevice)
			}
		}
	}

	return allBlockDevices, nil

}

func GetInstance() ([]*storagedriver.Instance, error) {
	var allInstances []*storagedriver.Instance
	for _, driver := range storagedriver.Adapters {
		instance, err := driver.GetInstance()
		if err != nil {
			return nil, fmt.Errorf("Error: %s: %s", ErrDriverInstanceDiscovery, err)
		}

		allInstances = append(allInstances, instance)

	}

	return allInstances, nil
}

func GetVolume(volumeID, volumeName string) ([]*storagedriver.Volume, error) {
	var allVolumes []*storagedriver.Volume

	for _, driver := range storagedriver.Adapters {
		volumes, err := driver.GetVolume(volumeID, volumeName)
		if err != nil {
			return []*storagedriver.Volume{}, fmt.Errorf("Error: %s: %s", ErrDriverVolumeDiscovery, err)
		}

		if len(volumes) > 0 {
			for _, volume := range volumes {
				allVolumes = append(allVolumes, volume)
			}
		}
	}
	return allVolumes, nil
}

func GetSnapshot(volumeID, snapshotID, snapshotName string) ([]*storagedriver.Snapshot, error) {
	var allSnapshots []*storagedriver.Snapshot

	for _, driver := range storagedriver.Adapters {
		snapshots, err := driver.GetSnapshot(volumeID, snapshotID, snapshotName)
		if err != nil {
			return nil, fmt.Errorf("Error: %s: %s", ErrDriverSnapshotDiscovery, err)
		}

		if len(snapshots) > 0 {
			for _, snapshot := range snapshots {
				allSnapshots = append(allSnapshots, snapshot)
			}
		}
	}
	return allSnapshots, nil
}

func CreateSnapshot(runAsync bool, snapshotName, volumeID, description string) ([]*storagedriver.Snapshot, error) {
	if len(storagedriver.Adapters) > 1 {
		return nil, ErrMultipleDriversDetected
	}
	for _, driver := range storagedriver.Adapters {
		snapshot, err := driver.CreateSnapshot(runAsync, snapshotName, volumeID, description)
		if err != nil {
			return nil, err
		}
		return snapshot, nil
	}
	return nil, nil
}

func RemoveSnapshot(snapshotID string) error {
	if len(storagedriver.Adapters) > 1 {
		return ErrMultipleDriversDetected
	}
	for _, driver := range storagedriver.Adapters {
		err := driver.RemoveSnapshot(snapshotID)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateVolume(runAsync bool, volumeName string, volumeID, snapshotID string, volumeType string, IOPS int64, size int64, availabilityZone string) (*storagedriver.Volume, error) {
	if len(storagedriver.Adapters) > 1 {
		return &storagedriver.Volume{}, ErrMultipleDriversDetected
	}
	for _, driver := range storagedriver.Adapters {
		var minSize int
		var err error
		minVolSize := os.Getenv("REXRAY_MINVOLSIZE")
		if size != 0 && minVolSize != "" {
			minSize, err = strconv.Atoi(os.Getenv("REXRAY_MINVOLSIZE"))
			if err != nil {
				return &storagedriver.Volume{}, err
			}
		}
		if minSize > 0 && int64(minSize) > size {
			size = int64(minSize)
		}
		volume, err := driver.CreateVolume(runAsync, volumeName, volumeID, snapshotID, volumeType, IOPS, size, availabilityZone)
		if err != nil {
			return &storagedriver.Volume{}, err
		}
		return volume, nil
	}
	return &storagedriver.Volume{}, nil
}

func RemoveVolume(volumeID string) error {
	if len(storagedriver.Adapters) > 1 {
		return ErrMultipleDriversDetected
	}
	for _, driver := range storagedriver.Adapters {
		err := driver.RemoveVolume(volumeID)
		if err != nil {
			return err
		}
	}
	return nil
}

func AttachVolume(runAsync bool, volumeID string, instanceID string) ([]*storagedriver.VolumeAttachment, error) {
	if len(storagedriver.Adapters) > 1 {
		return []*storagedriver.VolumeAttachment{}, ErrMultipleDriversDetected
	}
	for _, driver := range storagedriver.Adapters {
		if instanceID == "" {
			instance, err := driver.GetInstance()
			if err != nil {
				return []*storagedriver.VolumeAttachment{}, fmt.Errorf("Error: %s: %s", ErrDriverInstanceDiscovery, err)
			}
			instanceID = instance.InstanceID
		}

		volumeAttachment, err := driver.AttachVolume(runAsync, volumeID, instanceID)
		if err != nil {
			return []*storagedriver.VolumeAttachment{}, err
		}
		return volumeAttachment, nil
	}
	return []*storagedriver.VolumeAttachment{}, nil
}

func DetachVolume(runAsync bool, volumeID string, instanceID string) error {
	if len(storagedriver.Adapters) > 1 {
		return ErrMultipleDriversDetected
	}
	for _, driver := range storagedriver.Adapters {
		if instanceID == "" {
			instance, err := driver.GetInstance()
			if err != nil {
				fmt.Errorf("Error: %s: %s", ErrDriverInstanceDiscovery, err)
			}
			instanceID = instance.InstanceID
		}

		err := driver.DetachVolume(runAsync, volumeID, instanceID)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func GetVolumeAttach(volumeID string, instanceID string) ([]*storagedriver.VolumeAttachment, error) {
	if len(storagedriver.Adapters) > 1 {
		return []*storagedriver.VolumeAttachment{}, ErrMultipleDriversDetected
	}
	for _, driver := range storagedriver.Adapters {
		volumeAttachments, err := driver.GetVolumeAttach(volumeID, instanceID)
		if err != nil {
			return []*storagedriver.VolumeAttachment{}, err
		}
		return volumeAttachments, nil
	}

	return []*storagedriver.VolumeAttachment{}, nil
}

func CopySnapshot(runAsync bool, volumeID, snapshotID, snapshotName, targetSnapshotName, targetRegion string) (*storagedriver.Snapshot, error) {
	if len(storagedriver.Adapters) > 1 {
		return nil, ErrMultipleDriversDetected
	}
	for _, driver := range storagedriver.Adapters {
		snapshot, err := driver.CopySnapshot(runAsync, volumeID, snapshotID, snapshotName, targetSnapshotName, targetRegion)
		if err != nil {
			return nil, err
		}
		return snapshot, nil
	}
	return nil, nil
}

func GetDriverNames() []string {
	return storagedriver.GetDriverNames()
}
