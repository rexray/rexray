package core

import (
	"bytes"
	"sync"

	"github.com/emccode/rexray/core/errors"
)

// BlockDevice provides information about a block-storage device.
type BlockDevice struct {

	// The name of the provider that owns the block device.
	ProviderName string

	// The ID of the instance to which the device is connected.
	InstanceID string

	// The ID of the volume for which the device is mounted.
	VolumeID string

	// The name of the device.
	DeviceName string

	// The region from which the device originates.
	Region string

	// The device status.
	Status string

	// The name of the network on which the device resides.
	NetworkName string
}

// Instance provides information about a storage object.
type Instance struct {

	// The name of the provider that owns the object.
	ProviderName string

	// The ID of the instance to which the object is connected.
	InstanceID string

	// The region from which the object originates.
	Region string

	// The name of the instance.
	Name string
}

// Snapshot provides information about a storage-layer snapshot.
type Snapshot struct {

	// The name of the snapshot.
	Name string

	// The ID of the volume to which the snapshot belongs.
	VolumeID string

	// The snapshot's ID.
	SnapshotID string

	// The size of the volume to which the snapshot belongs/
	VolumeSize string

	// The time at which the request to create the snapshot was submitted.
	StartTime string

	// A description of the snapshot.
	Description string

	// The status of the snapshot.
	Status string
}

// Volume provides information about a storage volume.
type Volume struct {

	// The name of the volume.
	Name string

	// The volume ID.
	VolumeID string

	// The availability zone for which the volume is available.
	AvailabilityZone string

	// The volume status.
	Status string

	// The volume type.
	VolumeType string

	// The volume IOPs.
	IOPS int64

	// The size of the volume.
	Size string

	// The name of the network on which the volume resides.
	NetworkName string

	// The volume's attachments.
	Attachments []*VolumeAttachment
}

// VolumeAttachment provides information about an object attached to a
// storage volume.
type VolumeAttachment struct {

	// The ID of the volume to which the attachment belongs.
	VolumeID string

	// The ID of the instance on which the volume to which the attachment
	// belongs is mounted.
	InstanceID string

	// The name of the device on which the volume to which the object is
	// attached is mounted.
	DeviceName string

	// The status of the attachment.
	Status string
}

// StorageDriver is the interface implemented by types that provide storage
// introspection and management.
type StorageDriver interface {
	Driver

	// GetVolumeMapping lists the block devices that are attached to the
	GetVolumeMapping() ([]*BlockDevice, error)

	// GetInstance retrieves the local instance.
	GetInstance() (*Instance, error)

	// GetVolume returns all volumes for the instance based on either volumeID
	// or volumeName that are available to the instance.
	GetVolume(volumeID, volumeName string) ([]*Volume, error)

	// GetVolumeAttach returns the attachment details based on volumeID or
	// volumeName where the volume is currently attached.
	GetVolumeAttach(volumeID, instanceID string) ([]*VolumeAttachment, error)

	// CreateSnapshot is a synch/async operation that returns snapshots that
	// have been performed based on supplying a snapshotName, source volumeID,
	// and optional description.
	CreateSnapshot(
		runAsync bool,
		snapshotName, volumeID, description string) ([]*Snapshot, error)

	// GetSnapshot returns a list of snapshots for a volume based on volumeID,
	// snapshotID, or snapshotName.
	GetSnapshot(volumeID, snapshotID, snapshotName string) ([]*Snapshot, error)

	// RemoveSnapshot will remove a snapshot based on the snapshotID.
	RemoveSnapshot(snapshotID string) error

	// CreateVolume is sync/async and will create an return a new/existing
	// Volume based on volumeID/snapshotID with a name of volumeName and a size
	// in GB.  Optionally based on the storage driver, a volumeType, IOPS, and
	// availabilityZone could be defined.
	CreateVolume(
		runAsync bool,
		volumeName, volumeID, snapshotID, volumeType string,
		IOPS, size int64,
		availabilityZone string) (*Volume, error)

	// RemoveVolume will remove a volume based on volumeID.
	RemoveVolume(volumeID string) error

	// GetDeviceNextAvailable return a device path that will retrieve the next
	// available disk device that can be used.
	GetDeviceNextAvailable() (string, error)

	// AttachVolume returns a list of VolumeAttachments is sync/async that will
	// attach a volume to an instance based on volumeID and instanceID.
	AttachVolume(
		runAsync bool, volumeID, instanceID string) ([]*VolumeAttachment, error)

	// DetachVolume is sync/async that will detach the volumeID from the local
	// instance or the instanceID.
	DetachVolume(runAsync bool, volumeID string, instanceID string) error

	// CopySnapshot is a sync/async and returns a snapshot that will copy a
	// snapshot based on volumeID/snapshotID/snapshotName and create a new
	// snapshot of desinationSnapshotName in the destinationRegion location.
	CopySnapshot(
		runAsync bool, volumeID, snapshotID, snapshotName,
		destinationSnapshotName, destinationRegion string) (*Snapshot, error)
}

// StorageDriverManager acts as both a StorageDriverManager and as an aggregate
// of storage drivers, providing batch methods.
type StorageDriverManager interface {
	StorageDriver

	// Drivers gets a channel which receives a list of all of the configured
	// storage drivers.
	Drivers() <-chan StorageDriver

	// GetInstances gets the instance for each of the configured drivers.
	GetInstances() ([]*Instance, error)
}

type sdm struct {
	rexray  *RexRay
	drivers map[string]StorageDriver
}

func (r *sdm) Init(rexray *RexRay) error {
	return nil
}

func (r *sdm) Name() string {
	var b bytes.Buffer
	for d := range r.Drivers() {
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		b.WriteString(d.Name())
	}
	return b.String()
}

func (r *sdm) Drivers() <-chan StorageDriver {
	c := make(chan StorageDriver)
	go func() {
		if len(r.drivers) == 0 {
			close(c)
			return
		}
		for _, v := range r.drivers {
			c <- v
		}
		close(c)
	}()
	return c
}

// GetVolumeMapping performs storage introspection and
// returns a listing of block devices from the guest
func (r *sdm) GetVolumeMapping() ([]*BlockDevice, error) {
	var allBlockDevices []*BlockDevice
	for _, driver := range r.drivers {
		blockDevices, err := driver.GetVolumeMapping()
		if err != nil {
			return []*BlockDevice{}, errors.ErrDriverBlockDeviceDiscovery
		}

		if len(blockDevices) > 0 {
			for _, blockDevice := range blockDevices {
				allBlockDevices = append(allBlockDevices, blockDevice)
			}
		}
	}

	return allBlockDevices, nil

}

func (r *sdm) GetInstances() ([]*Instance, error) {
	cI := make(chan *Instance)
	cE := make(chan error)
	defer close(cI)
	defer close(cE)

	done := make(chan int)
	var wg sync.WaitGroup

	wg.Add(len(r.drivers))
	go func() {
		for _, d := range r.drivers {
			go func(d StorageDriver) {
				defer wg.Done()
				var e error
				var i *Instance
				i, e = d.GetInstance()
				if e != nil {
					cE <- e
				} else {
					cI <- i
				}
			}(d)
		}
	}()

	go func() {
		wg.Wait()
		done <- 1
	}()

	var allInstances []*Instance

	for {
		select {
		case i := <-cI:
			allInstances = append(allInstances, i)
		case e := <-cE:
			return nil, e
		case <-done:
			return allInstances, nil
		}
	}
}

func (r *sdm) GetInstance() (*Instance, error) {
	for _, d := range r.drivers {
		return d.GetInstance()
	}
	return nil, errors.ErrNoStorageDrivers
}

func (r *sdm) GetVolume(volumeID, volumeName string) ([]*Volume, error) {
	for _, d := range r.drivers {
		return d.GetVolume(volumeID, volumeName)
	}
	return nil, errors.ErrNoStorageDrivers
}

func (r *sdm) GetSnapshot(volumeID, snapshotID, snapshotName string) ([]*Snapshot, error) {
	for _, d := range r.drivers {
		return d.GetSnapshot(volumeID, snapshotID, snapshotName)
	}
	return nil, errors.ErrNoStorageDrivers
}

func (r *sdm) CreateSnapshot(runAsync bool,
	snapshotName, volumeID, description string) ([]*Snapshot, error) {
	for _, d := range r.drivers {
		return d.CreateSnapshot(runAsync, snapshotName, volumeID, description)
	}
	return nil, errors.ErrNoStorageDrivers
}

func (r *sdm) RemoveSnapshot(snapshotID string) error {
	for _, d := range r.drivers {
		return d.RemoveSnapshot(snapshotID)
	}
	return errors.ErrNoStorageDrivers
}

func (r *sdm) CreateVolume(runAsync bool,
	volumeName, volumeID, snapshotID, volumeType string,
	IOPS, size int64, availabilityZone string) (*Volume, error) {
	for _, d := range r.drivers {
		return d.CreateVolume(
			runAsync, volumeName, volumeID, snapshotID, volumeType,
			IOPS, size, availabilityZone)
	}
	return nil, errors.ErrNoStorageDrivers
}

func (r *sdm) RemoveVolume(volumeID string) error {
	for _, d := range r.drivers {
		return d.RemoveVolume(volumeID)
	}
	return errors.ErrNoStorageDrivers
}

func (r *sdm) AttachVolume(
	runAsync bool,
	volumeID, instanceID string) ([]*VolumeAttachment, error) {
	for _, d := range r.drivers {
		return d.AttachVolume(runAsync, volumeID, instanceID)
	}
	return nil, errors.ErrNoStorageDrivers
}

func (r *sdm) DetachVolume(
	runAsync bool,
	volumeID, instanceID string) error {
	for _, d := range r.drivers {
		return d.DetachVolume(runAsync, volumeID, instanceID)
	}
	return errors.ErrNoStorageDrivers
}

func (r *sdm) GetVolumeAttach(
	volumeID, instanceID string) ([]*VolumeAttachment, error) {
	for _, d := range r.drivers {
		return d.GetVolumeAttach(volumeID, instanceID)
	}
	return nil, errors.ErrNoStorageDrivers
}

func (r *sdm) CopySnapshot(
	runAsync bool,
	volumeID, snapshotID, snapshotName,
	targetSnapshotName, targetRegion string) (*Snapshot, error) {
	for _, d := range r.drivers {
		return d.CopySnapshot(runAsync, volumeID, snapshotID, snapshotName,
			targetSnapshotName, targetRegion)
	}
	return nil, errors.ErrNoStorageDrivers
}

func (r *sdm) GetDeviceNextAvailable() (string, error) {
	for _, d := range r.drivers {
		return d.GetDeviceNextAvailable()
	}
	return "", errors.ErrNoStorageDrivers
}
