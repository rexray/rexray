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

	log.WithField("pathCache", r.pathCache()).
		Debug("checking volume path cache setting")

	if r.pathCache() {
		_, _ = r.List()
	}

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
			"moduleName": r.rexray.Context,
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
		"moduleName": r.rexray.Context,
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
			"moduleName": r.rexray.Context,
			"volumeName": volumeName,
			"count":      *c,
		}).Info("released count")
	}
}

func (r *vdm) countExists(volumeName string) bool {
	fields := log.Fields{
		"moduleName": r.rexray.Context,
		"volumeName": volumeName,
	}

	var c *int
	var exists bool
	if c, exists = r.mapUsedCount[volumeName]; exists {
		fields["count"] = *c
	}
	fields["exists"] = exists

	log.WithFields(fields).Info("status of count")

	return exists
}

func (r *vdm) countReset(volumeName string) bool {
	r.m.Lock()
	defer r.m.Unlock()

	c, _ := r.mapUsedCount[volumeName]
	if c != nil && *c < 2 {
		log.WithFields(log.Fields{
			"moduleName": r.rexray.Context,
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
		log.WithFields(log.Fields{
			"moduleName":  r.rexray.Context,
			"driverName":  d.Name(),
			"volumeName":  volumeName,
			"volumeID":    volumeID,
			"overwriteFs": overwriteFs,
			"newFsType":   newFsType,
			"preempt":     preempt}).Info("vdm.Mount")

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
		log.WithFields(log.Fields{
			"moduleName": r.rexray.Context,
			"driverName": d.Name(),
			"volumeName": volumeName,
			"volumeID":   volumeID}).Info("vdm.Unmount")

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
		fields := log.Fields{
			"moduleName": r.rexray.Context,
			"driverName": d.Name(),
			"volumeName": volumeName,
			"volumeID":   volumeID}

		log.WithFields(fields).Info("vdm.Path")

		if !r.pathCache() {
			return d.Path(volumeName, volumeID)
		}

		if _, ok := r.mapUsedCount[volumeName]; !ok {
			log.WithFields(fields).Debug("skipping path lookup")
			return "", nil
		}

		return d.Path(volumeName, volumeID)
	}
	return "", errors.ErrNoVolumesDetected
}

// Get will return a specific volume
func (r *vdm) Get(volumeName string) (VolumeMap, error) {
	for _, d := range r.drivers {
		log.WithFields(log.Fields{
			"moduleName": r.rexray.Context,
			"driverName": d.Name(),
			"volumeName": volumeName}).Info("vdm.Get")

		return d.Get(volumeName)
	}
	return nil, errors.ErrNoVolumesDetected
}

func (volMap *VolumeMap) get(key string) (string, *VolumeMap) {
	if val, ok := (*volMap)[key]; ok && val != "" {
		return val, volMap
	}
	return "", nil
}

// Get will return all volumes
func (r *vdm) List() ([]VolumeMap, error) {
	for _, d := range r.drivers {
		log.WithFields(log.Fields{
			"moduleName": r.rexray.Context,
			"driverName": d.Name()}).Info("vdm.List")

		list, err := d.List()
		if err != nil {
			return nil, err
		}

		if !r.pathCache() {
			return list, nil
		}

		var listNew []VolumeMap

		for _, vol := range list {
			var name string
			var gvol *VolumeMap
			if name, gvol = vol.get("Name"); gvol == nil {
				continue
			}

			if _, ok := r.mapUsedCount[name]; !ok {
				if _, gvol := vol.get("Mountpoint"); gvol != nil {
					r.countInit(name)
				}
			}

			listNew = append(listNew, *gvol)
		}

		return listNew, nil

	}
	return nil, errors.ErrNoVolumesDetected
}

// Create will create a new volume with the volumeName and opts.
func (r *vdm) Create(volumeName string, opts VolumeOpts) error {
	for _, d := range r.drivers {
		log.WithFields(log.Fields{
			"moduleName": r.rexray.Context,
			"driverName": d.Name(),
			"volumeName": volumeName,
			"opts":       opts}).Info("vdm.Create")

		r.countInit(volumeName)
		return d.Create(volumeName, opts)
	}
	return errors.ErrNoVolumesDetected
}

// Remove will remove a volume of volumeName.
func (r *vdm) Remove(volumeName string) error {
	for _, d := range r.drivers {
		log.WithFields(log.Fields{
			"moduleName": r.rexray.Context,
			"driverName": d.Name(),
			"volumeName": volumeName}).Info("vdm.Remove")

		return d.Remove(volumeName)
	}
	return errors.ErrNoVolumesDetected
}

// Attach will attach a volume based on volumeName to the instance of
// instanceID.
func (r *vdm) Attach(volumeName, instanceID string, force bool) (string, error) {
	for _, d := range r.drivers {
		log.WithFields(log.Fields{
			"moduleName": r.rexray.Context,
			"driverName": d.Name(),
			"volumeName": volumeName,
			"instanceID": instanceID,
			"force":      force}).Info("vdm.Attach")

		return d.Attach(volumeName, instanceID, force)
	}
	return "", errors.ErrNoVolumesDetected
}

// Detach will detach a volume based on volumeName to the instance of
// instanceID.
func (r *vdm) Detach(volumeName, instanceID string, force bool) error {
	for _, d := range r.drivers {
		log.WithFields(log.Fields{
			"moduleName": r.rexray.Context,
			"driverName": d.Name(),
			"volumeName": volumeName,
			"instanceID": instanceID,
			"force":      force}).Info("vdm.Detach")
		return d.Detach(volumeName, instanceID, force)
	}
	return errors.ErrNoVolumesDetected
}

// NetworkName will return an identifier of a volume that is relevant when
// corelating a local device to a device that is the volumeName to the
// local instanceID.
func (r *vdm) NetworkName(volumeName, instanceID string) (string, error) {
	for _, d := range r.drivers {
		log.WithFields(log.Fields{
			"moduleName": r.rexray.Context,
			"driverName": d.Name(),
			"volumeName": volumeName,
			"instanceID": instanceID}).Info("vdm.NetworkName")
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

func (r *vdm) pathCache() bool {
	return r.rexray.Config.GetBool("rexray.volume.path.cache")
}
