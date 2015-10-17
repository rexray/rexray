package core

import (
	"bytes"

	"github.com/emccode/rexray/core/errors"
)

// VolumeOpts is a map of options used when creating a new volume
type VolumeOpts map[string]string

// VolumeDriver is the interface implemented by types that provide volume
// introspection and management.
type VolumeDriver interface {
	Driver

	// Mount will return a mount point path when specifying either a volumeName
	// or volumeID.  If a overwriteFs boolean is specified it will overwrite
	// the FS based on newFsType if it is detected that there is no FS present.
	Mount(
		volumeName, volumeID string,
		overwriteFs bool, newFsType string) (string, error)

	// Unmount will unmount the specified volume by volumeName or volumeID.
	Unmount(volumeName, volumeID string) error

	// Path will return the mounted path of the volumeName or volumeID.
	Path(volumeName, volumeID string) (string, error)

	// Create will create a new volume with the volumeName and opts.
	Create(volumeName string, opts VolumeOpts) error

	// Remove will remove a volume of volumeName.
	Remove(volumeName string) error

	// Attach will attach a volume based on volumeName to the instance of
	// instanceID.
	Attach(volumeName, instanceID string) (string, error)

	// Detach will detach a volume based on volumeName to the instance of
	// instanceID.
	Detach(volumeName, instanceID string) error

	// NetworkName will return an identifier of a volume that is relevant when
	// corelating a local device to a device that is the volumeName to the
	// local instanceID.
	NetworkName(volumeName, instanceID string) (string, error)
}

// VolumeDriverManager acts as both a VolumeDriver and as an aggregate of
// volume drivers, providing batch methods.
type VolumeDriverManager interface {
	VolumeDriver

	// Drivers gets a channel which receives a list of all of the configured
	// volume drivers.
	Drivers() <-chan VolumeDriver

	// Unmounts unmounts all volumes.
	UnmountAll() error

	// RemoveAll removes all volumes.
	RemoveAll() error

	// DetachAll detaches all volumes attached to the instance of instanceID.
	DetachAll(instanceID string) error
}

type vdm struct {
	rexray  *RexRay
	drivers map[string]VolumeDriver
}

func (r *vdm) Init(rexray *RexRay) error {
	return nil
}

func (r *vdm) Name() string {
	var b bytes.Buffer
	for d := range r.Drivers() {
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		b.WriteString(d.Name())
	}
	return b.String()
}

func (r *vdm) Drivers() <-chan VolumeDriver {
	c := make(chan VolumeDriver)
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

// Unmounts unmounts all volumes.
func (r *vdm) UnmountAll() error {
	return nil
}

// RemoveAll removes all volumes.
func (r *vdm) RemoveAll() error {
	return nil
}

// DetachAll detaches all volumes attached to the instance of instanceID.
func (r *vdm) DetachAll(instanceID string) error {
	return nil
}

// Mount will return a mount point path when specifying either a volumeName
// or volumeID.  If a overwriteFs boolean is specified it will overwrite
// the FS based on newFsType if it is detected that there is no FS present.
func (r *vdm) Mount(
	volumeName, volumeID string,
	overwriteFs bool, newFsType string) (string, error) {
	for _, d := range r.drivers {
		return d.Mount(volumeName, volumeID, overwriteFs, newFsType)
	}
	return "", errors.ErrNoStorageDrivers
}

// Unmount will unmount the specified volume by volumeName or volumeID.
func (r *vdm) Unmount(volumeName, volumeID string) error {
	for _, d := range r.drivers {
		return d.Unmount(volumeName, volumeID)
	}
	return errors.ErrNoStorageDrivers
}

// Path will return the mounted path of the volumeName or volumeID.
func (r *vdm) Path(volumeName, volumeID string) (string, error) {
	for _, d := range r.drivers {
		return d.Path(volumeName, volumeID)
	}
	return "", errors.ErrNoStorageDrivers
}

// Create will create a new volume with the volumeName and opts.
func (r *vdm) Create(volumeName string, opts VolumeOpts) error {
	for _, d := range r.drivers {
		return d.Create(volumeName, opts)
	}
	return errors.ErrNoStorageDrivers
}

// Remove will remove a volume of volumeName.
func (r *vdm) Remove(volumeName string) error {
	for _, d := range r.drivers {
		return d.Remove(volumeName)
	}
	return errors.ErrNoStorageDrivers
}

// Attach will attach a volume based on volumeName to the instance of
// instanceID.
func (r *vdm) Attach(volumeName, instanceID string) (string, error) {
	for _, d := range r.drivers {
		return d.Attach(volumeName, instanceID)
	}
	return "", errors.ErrNoStorageDrivers
}

// Detach will detach a volume based on volumeName to the instance of
// instanceID.
func (r *vdm) Detach(volumeName, instanceID string) error {
	for _, d := range r.drivers {
		return d.Detach(volumeName, instanceID)
	}
	return errors.ErrNoStorageDrivers
}

// NetworkName will return an identifier of a volume that is relevant when
// corelating a local device to a device that is the volumeName to the
// local instanceID.
func (r *vdm) NetworkName(volumeName, instanceID string) (string, error) {
	for _, d := range r.drivers {
		return d.NetworkName(volumeName, instanceID)
	}
	return "", errors.ErrNoStorageDrivers
}
