package storagedriver

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/emccode/rexray/util"
)

var (
	driverInitFuncs map[string]InitFunc
	drivers         map[string]Driver
	debug           string
)

var (
	ErrDriverInstanceDiscovery = errors.New("Driver Instance discovery failed")
)

var Adapters map[string]Driver

type BlockDevice struct {
	ProviderName string
	InstanceID   string
	VolumeID     string
	DeviceName   string
	Region       string
	Status       string
	NetworkName  string
}

type Instance struct {
	ProviderName string
	InstanceID   string
	Region       string
	Name         string
}

type Snapshot struct {
	Name        string
	VolumeID    string
	SnapshotID  string
	VolumeSize  string
	StartTime   string
	Description string
	Status      string
}

type Volume struct {
	Name             string
	VolumeID         string
	AvailabilityZone string
	Status           string
	VolumeType       string
	IOPS             int64
	Size             string
	NetworkName      string
	Attachments      []*VolumeAttachment
}

type VolumeAttachment struct {
	VolumeID   string
	InstanceID string
	DeviceName string
	Status     string
}

type Driver interface {
	// GetVolumeMapping lists the block devices that are attached to the instance.
	GetVolumeMapping() ([]*BlockDevice, error)

	// GetInstance retrieves the local instance.
	GetInstance() (*Instance, error)

	// GetVolume returns all volumes for the instance based on either volumeID or volumeName
	// that are available to the instance.
	GetVolume(volumeID, volumeName string) ([]*Volume, error)

	// GetVolumeAttach returns the attachment details based on volumeID or volumeName
	// where the volume is currently attached.
	GetVolumeAttach(volumeID, instanceID string) ([]*VolumeAttachment, error)

	// CreateSnapshot is a synch/async operation that returns snapshots that have been
	// performed based on supplying a snapshotName, source volumeID, and optional description.
	CreateSnapshot(runAsync bool, snapshotName, volumeID, description string) ([]*Snapshot, error)

	// GetSnapshot returns a list of snapshots for a volume based on volumeID, snapshotID, or snapshotName.
	GetSnapshot(volumeID, snapshotID, snapshotName string) ([]*Snapshot, error)

	// RemoveSnapshot will remove a snapshot based on the snapshotID.
	RemoveSnapshot(snapshotID string) error

	// CreateVolume is sync/async and will create an return a new/existing Volume based on volumeID/snapshotID with
	// a name of volumeName and a size in GB.  Optionally based on the storage driver, a volumeType, IOPS, and availabilityZone
	// could be defined.
	CreateVolume(runAsync bool, volumeName string, volumeID string, snapshotID string, volumeType string, IOPS int64, size int64, availabilityZone string) (*Volume, error)

	// RemoveVolume will remove a volume based on volumeID.
	RemoveVolume(volumeID string) error

	// GetDeviceNextAvailable return a device path that will retrieve the next available disk device that can be used.
	GetDeviceNextAvailable() (string, error)

	// AttachVolume returns a list of VolumeAttachments is sync/async that will attach a volume to an instance based on volumeID and instanceID.
	AttachVolume(runAsync bool, volumeID, instanceID string) ([]*VolumeAttachment, error)

	// DetachVolume is sync/async that will detach the volumeID from the local instance or the instanceID.
	DetachVolume(runAsync bool, volumeID string, instanceID string) error

	// CopySnapshot is a sync/async and returns a snapshot that will copy a snapshot based on volumeID/snapshotID/snapshotName and
	// create a new snapshot of desinationSnapshotName in the destinationRegion location.
	CopySnapshot(runAsync bool, volumeID, snapshotID, snapshotName, destinationSnapshotName, destinationRegion string) (*Snapshot, error)
}

type InitFunc func() (Driver, error)

func Register(name string, initFunc InitFunc) error {
	driverInitFuncs[name] = initFunc
	return nil
}

func init() {
	driverInitFuncs = make(map[string]InitFunc)
	drivers = make(map[string]Driver)
	debug = strings.ToUpper(os.Getenv("REXRAY_DEBUG"))
}

func GetDriverNames() []string {
	names := make([]string, 0, len(drivers))
	for n := range drivers {
		names = append(names, n)
	}
	return names
}

func GetDrivers(storageDrivers string) (map[string]Driver, error) {
	var err error
	var storageDriversArr []string
	if storageDrivers != "" {
		storageDriversArr = strings.Split(storageDrivers, ",")
	}

	if debug == "TRUE" {
		fmt.Println(driverInitFuncs)
	}

	for name, initFunc := range driverInitFuncs {
		if len(storageDriversArr) > 0 && !util.StringInSlice(name, storageDriversArr) {
			continue
		}
		drivers[name], err = initFunc()
		if err != nil {
			if debug == "TRUE" {
				fmt.Println(fmt.Sprintf("Info (%s): %s", name, err))
			}
			delete(drivers, name)
		}
	}

	return drivers, nil
}
