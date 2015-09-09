package storage

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/emccode/rexray/config"
	"github.com/emccode/rexray/errors"
	"github.com/emccode/rexray/util"
)

var driverInitFuncs map[string]InitFunc

func init() {
	driverInitFuncs = make(map[string]InitFunc)
}

type StorageDriverManager struct {
	Drivers map[string]Driver
	Config  *config.Config
}

type InitFunc func(conf *config.Config) (Driver, error)

func Register(name string, initFunc InitFunc) {
	driverInitFuncs[name] = initFunc
}

func (sdm *StorageDriverManager) GetDriverNames() []string {
	return GetDriverNames()
}

func GetDriverNames() []string {
	names := make([]string, 0, len(driverInitFuncs))
	for n := range driverInitFuncs {
		names = append(names, n)
	}
	return names
}

func getDrivers(conf *config.Config) (map[string]Driver, error) {

	driverNames := conf.StorageDrivers

	log.WithFields(log.Fields{
		"driverInitFuncs": driverInitFuncs,
		"driverNames":     driverNames}).Debug("getting driver instances")

	drivers := map[string]Driver{}

	for name, initFunc := range driverInitFuncs {
		if len(driverNames) > 0 && !util.StringInSlice(name, driverNames) {
			continue
		}

		var initErr error
		drivers[name], initErr = initFunc(conf)
		if initErr != nil {
			log.WithFields(log.Fields{
				"driverName": name,
				"error":      initErr}).Debug("error initializing driver")
			delete(drivers, name)
			continue
		}

		log.WithField("driverName", name).Debug("initialized driver")
	}

	return drivers, nil
}

var (
	ErrDriverBlockDeviceDiscovery = errors.New("Driver Block Device discovery failed")
	ErrDriverInstanceDiscovery    = errors.New("Driver Instance discovery failed")
	ErrDriverVolumeDiscovery      = errors.New("Driver Volume discovery failed")
	ErrDriverSnapshotDiscovery    = errors.New("Driver Snapshot discovery failed")
	ErrMultipleDriversDetected    = errors.New("Multiple drivers detected, must declare with driver with env of REXRAY_STORAGEDRIVER=")
)

func NewStorageDriverManager(conf *config.Config) (*StorageDriverManager, error) {

	sd, sdErr := getDrivers(conf)
	if sdErr != nil {
		return nil, sdErr
	}

	if len(sd) == 0 {
		return nil, errors.New("no storage drivers initialized")
	}

	return &StorageDriverManager{
		Drivers: sd,
		Config:  conf,
	}, nil
}

func (sdm *StorageDriverManager) IsDrivers() bool {
	return len(sdm.Drivers) > 0
}

// GetVolumeMapping performs storage introspection and
// returns a listing of block devices from the guest
func (sdm *StorageDriverManager) GetVolumeMapping() ([]*BlockDevice, error) {
	var allBlockDevices []*BlockDevice
	for _, driver := range sdm.Drivers {
		blockDevices, err := driver.GetVolumeMapping()
		if err != nil {
			return []*BlockDevice{}, fmt.Errorf("Error: %s: %s", ErrDriverBlockDeviceDiscovery, err)
		}

		if len(blockDevices) > 0 {
			for _, blockDevice := range blockDevices {
				allBlockDevices = append(allBlockDevices, blockDevice)
			}
		}
	}

	return allBlockDevices, nil

}

func (sdm *StorageDriverManager) GetInstance() ([]*Instance, error) {
	var allInstances []*Instance
	for _, driver := range sdm.Drivers {
		instance, err := driver.GetInstance()
		if err != nil {
			return nil, fmt.Errorf("Error: %s: %s", ErrDriverInstanceDiscovery, err)
		}

		allInstances = append(allInstances, instance)

	}

	return allInstances, nil
}

func (sdm *StorageDriverManager) GetVolume(volumeID, volumeName string) ([]*Volume, error) {
	var allVolumes []*Volume

	for _, driver := range sdm.Drivers {
		volumes, err := driver.GetVolume(volumeID, volumeName)
		if err != nil {
			return []*Volume{}, fmt.Errorf("Error: %s: %s", ErrDriverVolumeDiscovery, err)
		}

		if len(volumes) > 0 {
			for _, volume := range volumes {
				allVolumes = append(allVolumes, volume)
			}
		}
	}
	return allVolumes, nil
}

func (sdm *StorageDriverManager) GetSnapshot(volumeID, snapshotID, snapshotName string) ([]*Snapshot, error) {
	var allSnapshots []*Snapshot

	for _, driver := range sdm.Drivers {
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

func (sdm *StorageDriverManager) CreateSnapshot(runAsync bool, snapshotName, volumeID, description string) ([]*Snapshot, error) {
	if len(sdm.Drivers) > 1 {
		return nil, ErrMultipleDriversDetected
	}
	for _, driver := range sdm.Drivers {
		snapshot, err := driver.CreateSnapshot(runAsync, snapshotName, volumeID, description)
		if err != nil {
			return nil, err
		}
		return snapshot, nil
	}
	return nil, nil
}

func (sdm *StorageDriverManager) RemoveSnapshot(snapshotID string) error {
	if len(sdm.Drivers) > 1 {
		return ErrMultipleDriversDetected
	}
	for _, driver := range sdm.Drivers {
		err := driver.RemoveSnapshot(snapshotID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sdm *StorageDriverManager) CreateVolume(runAsync bool, volumeName string, volumeID, snapshotID string, volumeType string, IOPS int64, size int64, availabilityZone string) (*Volume, error) {
	if len(sdm.Drivers) > 1 {
		return &Volume{}, ErrMultipleDriversDetected
	}
	for _, driver := range sdm.Drivers {
		var minSize int
		var err error
		minVolSize := os.Getenv("REXRAY_MINVOLSIZE")
		if size != 0 && minVolSize != "" {
			minSize, err = strconv.Atoi(os.Getenv("REXRAY_MINVOLSIZE"))
			if err != nil {
				return &Volume{}, err
			}
		}
		if minSize > 0 && int64(minSize) > size {
			size = int64(minSize)
		}
		volume, err := driver.CreateVolume(runAsync, volumeName, volumeID, snapshotID, volumeType, IOPS, size, availabilityZone)
		if err != nil {
			return &Volume{}, err
		}
		return volume, nil
	}
	return &Volume{}, nil
}

func (sdm *StorageDriverManager) RemoveVolume(volumeID string) error {
	if len(sdm.Drivers) > 1 {
		return ErrMultipleDriversDetected
	}
	for _, driver := range sdm.Drivers {
		err := driver.RemoveVolume(volumeID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sdm *StorageDriverManager) AttachVolume(runAsync bool, volumeID string, instanceID string) ([]*VolumeAttachment, error) {
	if len(sdm.Drivers) > 1 {
		return []*VolumeAttachment{}, ErrMultipleDriversDetected
	}
	for _, driver := range sdm.Drivers {
		if instanceID == "" {
			instance, err := driver.GetInstance()
			if err != nil {
				return []*VolumeAttachment{}, fmt.Errorf("Error: %s: %s", ErrDriverInstanceDiscovery, err)
			}
			instanceID = instance.InstanceID
		}

		volumeAttachment, err := driver.AttachVolume(runAsync, volumeID, instanceID)
		if err != nil {
			return []*VolumeAttachment{}, err
		}
		return volumeAttachment, nil
	}
	return []*VolumeAttachment{}, nil
}

func (sdm *StorageDriverManager) DetachVolume(runAsync bool, volumeID string, instanceID string) error {
	if len(sdm.Drivers) > 1 {
		return ErrMultipleDriversDetected
	}
	for _, driver := range sdm.Drivers {
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

func (sdm *StorageDriverManager) GetVolumeAttach(volumeID string, instanceID string) ([]*VolumeAttachment, error) {
	if len(sdm.Drivers) > 1 {
		return []*VolumeAttachment{}, ErrMultipleDriversDetected
	}
	for _, driver := range sdm.Drivers {
		volumeAttachments, err := driver.GetVolumeAttach(volumeID, instanceID)
		if err != nil {
			return []*VolumeAttachment{}, err
		}
		return volumeAttachments, nil
	}

	return []*VolumeAttachment{}, nil
}

func (sdm *StorageDriverManager) CopySnapshot(runAsync bool, volumeID, snapshotID, snapshotName, targetSnapshotName, targetRegion string) (*Snapshot, error) {
	if len(sdm.Drivers) > 1 {
		return nil, ErrMultipleDriversDetected
	}
	for _, driver := range sdm.Drivers {
		snapshot, err := driver.CopySnapshot(runAsync, volumeID, snapshotID, snapshotName, targetSnapshotName, targetRegion)
		if err != nil {
			return nil, err
		}
		return snapshot, nil
	}
	return nil, nil
}
