package core

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"github.com/emccode/rexray/core/errors"
	"sync"
)

// VolumeOpts is a map of options used when creating a new volume
type VolumeOpts map[string]string

// Volume is a map of a volume
type VolumeMap map[string]string

// VolumeDriver is the interface implemented by types that provide volume
// introspection and management.
type VolumeDriver interface {
	Driver

	// Mount will return a mount point path when specifying either a volumeName
	// or volumeID.  If a overwriteFs boolean is specified it will overwrite
	// the FS based on newFsType if it is detected that there is no FS present.
	Mount(
		volumeName, volumeID string,
		overwriteFs bool, newFsType string, preempt bool) (string, error)

	// Unmount will unmount the specified volume by volumeName or volumeID.
	Unmount(volumeName, volumeID string) error

	// Path will return the mounted path of the volumeName or volumeID.
	Path(volumeName, volumeID string) (string, error)

	// Create will create a new volume with the volumeName and opts.
	Create(volumeName string, opts VolumeOpts) error

	// Remove will remove a volume of volumeName.
	Remove(volumeName string) error

	// Get will return a specific volume
	Get(volumeName string) (VolumeMap, error)

	// List will return all volumes
	List() ([]VolumeMap, error)

	// Attach will attach a volume based on volumeName to the instance of
	// instanceID.
	Attach(volumeName, instanceID string, force bool) (string, error)

	// Detach will detach a volume based on volumeName to the instance of
	// instanceID.
	Detach(volumeName, instanceID string, force bool) error

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

	// UnmountAll unmounts all volumes.
	UnmountAll() error

	// RemoveAll removes all volumes.
	RemoveAll() error

	// DetachAll detaches all volumes attached to the instance of instanceID.
	DetachAll(instanceID string) error
}

type vdm struct {
	rexray       *RexRay
	drivers      map[string]VolumeDriver
	m            sync.Mutex
	mapUsedCount map[string]*int
}

func (r *vdm) Init(rexray *RexRay) error {
	if len(r.drivers) == 0 {
		return errors.ErrNoVolumeDrivers
	}
	r.mapUsedCount = make(map[string]*int)
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

// UnmountAll unmounts all volumes.
func (r *vdm) UnmountAll() error {
	for range r.drivers {
		return nil
	}
	return errors.ErrNoVolumesDetected
}

// RemoveAll removes all volumes.
func (r *vdm) RemoveAll() error {
	for range r.drivers {
		return nil
	}
	return errors.ErrNoVolumesDetected
}

// DetachAll detaches all volumes attached to the instance of instanceID.
func (r *vdm) DetachAll(instanceID string) error {
	for range r.drivers {
		return nil
	}
	return errors.ErrNoVolumesDetected
}

func (r *vdm) countUse(volumeName string) {
	r.m.Lock()
	if c, ok := r.mapUsedCount[volumeName]; ok {
		*c++
		log.WithFields(log.Fields{
			"volumeName": volumeName,
			"count":      *c,
		}).Info("set count to")
		r.m.Unlock()
	} else {
		r.m.Unlock()
		r.countInit(volumeName)
		r.countUse(volumeName)
	}
}

func (r *vdm) countInit(volumeName string) {
	r.m.Lock()
	defer r.m.Unlock()

	var c int
	c = 0
	r.mapUsedCount[volumeName] = &c
	log.WithFields(log.Fields{
		"volumeName": volumeName,
		"count":      c,
	}).Info("initialized count")
}

func (r *vdm) countRelease(volumeName string) {
	r.m.Lock()
	defer r.m.Unlock()
	if c, ok := r.mapUsedCount[volumeName]; ok {
		*c--
		log.WithFields(log.Fields{
			"volumeName": volumeName,
			"count":      *c,
		}).Info("released count")
	}
}

func (r *vdm) countExists(volumeName string) bool {
	_, exists := r.mapUsedCount[volumeName]
	log.WithFields(log.Fields{
		"volumeName": volumeName,
		"exists":     exists,
	}).Info("status of count")

	return exists
}

func (r *vdm) countReset(volumeName string) bool {
	r.m.Lock()
	defer r.m.Unlock()

	c, _ := r.mapUsedCount[volumeName]
	if c != nil && *c < 2 {
		log.WithFields(log.Fields{
			"volumeName": volumeName,
			"count":      *c,
		}).Info("count reset")
		*c = 0
		return true
	}
	return false
}

// Mount will return a mount point path when specifying either a volumeName
// or volumeID.  If a overwriteFs boolean is specified it will overwrite
// the FS based on newFsType if it is detected that there is no FS present.
func (r *vdm) Mount(
	volumeName, volumeID string,
	overwriteFs bool, newFsType string, preempt bool) (string, error) {
	for _, d := range r.drivers {
		if !preempt {
			preempt = r.preempt()
		}

		mp, err := d.Mount(volumeName, volumeID, overwriteFs, newFsType, preempt)
		if err != nil {
			return "", err
		}

		r.countUse(volumeName)

		return mp, nil
	}
	return "", errors.ErrNoVolumesDetected
}

// Unmount will unmount the specified volume by volumeName or volumeID.
func (r *vdm) Unmount(volumeName, volumeID string) error {
	for _, d := range r.drivers {
		if r.ignoreUsedCount() || r.countReset(volumeName) || !r.countExists(volumeName) {
			r.countInit(volumeName)
			return d.Unmount(volumeName, volumeID)
		} else {
			r.countRelease(volumeName)
			return nil
		}
	}
	return errors.ErrNoVolumesDetected
}

// Path will return the mounted path of the volumeName or volumeID.
func (r *vdm) Path(volumeName, volumeID string) (string, error) {
	for _, d := range r.drivers {
		return d.Path(volumeName, volumeID)
	}
	return "", errors.ErrNoVolumesDetected
}

// Get will return a specific volume
func (r *vdm) Get(volumeName string) (VolumeMap, error) {
	for _, d := range r.drivers {
		return d.Get(volumeName)
	}
	return nil, errors.ErrNoVolumesDetected
}

// Get will return all volumes
func (r *vdm) List() ([]VolumeMap, error) {
	for _, d := range r.drivers {
		return d.List()
	}
	return nil, errors.ErrNoVolumesDetected
}

// Create will create a new volume with the volumeName and opts.
func (r *vdm) Create(volumeName string, opts VolumeOpts) error {
	for _, d := range r.drivers {
		r.countInit(volumeName)
		return d.Create(volumeName, opts)
	}
	return errors.ErrNoVolumesDetected
}

// Remove will remove a volume of volumeName.
func (r *vdm) Remove(volumeName string) error {
	for _, d := range r.drivers {
		return d.Remove(volumeName)
	}
	return errors.ErrNoVolumesDetected
}

// Attach will attach a volume based on volumeName to the instance of
// instanceID.
func (r *vdm) Attach(volumeName, instanceID string, force bool) (string, error) {
	for _, d := range r.drivers {
		return d.Attach(volumeName, instanceID, force)
	}
	return "", errors.ErrNoVolumesDetected
}

// Detach will detach a volume based on volumeName to the instance of
// instanceID.
func (r *vdm) Detach(volumeName, instanceID string, force bool) error {
	for _, d := range r.drivers {
		return d.Detach(volumeName, instanceID, force)
	}
	return errors.ErrNoVolumesDetected
}

// NetworkName will return an identifier of a volume that is relevant when
// corelating a local device to a device that is the volumeName to the
// local instanceID.
func (r *vdm) NetworkName(volumeName, instanceID string) (string, error) {
	for _, d := range r.drivers {
		return d.NetworkName(volumeName, instanceID)
	}
	return "", errors.ErrNoVolumesDetected
}

func (r *vdm) preempt() bool {
	return r.rexray.Config.GetBool("rexray.volume.mount.preempt")
}

func (r *vdm) ignoreUsedCount() bool {
	return r.rexray.Config.GetBool("rexray.volume.unmount.ignoreusedcount")
}
