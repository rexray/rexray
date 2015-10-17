package core

import (
	"bytes"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/mount"

	"github.com/emccode/rexray/core/errors"
)

// MountInfoArray is an alias of []*mount.Info
type MountInfoArray []*mount.Info

// OSDriver is the interface implemented by types that provide OS introspection
// and management.
type OSDriver interface {
	Driver

	// Shows the existing mount points
	GetMounts(string, string) (MountInfoArray, error)

	// Check whether path is mounted or not
	Mounted(string) (bool, error)

	// Unmount based on a path
	Unmount(string) error

	// Mount based on a device, target, options, label
	Mount(string, string, string, string) error

	// Format a device with a FS type
	Format(string, string, bool) error
}

// OSDriverManager acts as both a OSDriverManager and as an aggregate of OS
// drivers, providing batch methods.
type OSDriverManager interface {
	OSDriver

	// Drivers gets a channel which receives a list of all of the configured
	// OS drivers.
	Drivers() <-chan OSDriver
}

type odm struct {
	rexray  *RexRay
	drivers map[string]OSDriver
}

func (r *odm) Init(rexray *RexRay) error {
	return nil
}

func (r *odm) Name() string {
	var b bytes.Buffer
	for d := range r.Drivers() {
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		b.WriteString(d.Name())
	}
	return b.String()
}

func (r *odm) Drivers() <-chan OSDriver {
	c := make(chan OSDriver)
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

func (r *odm) GetMounts(
	deviceName, mountPoint string) (MountInfoArray, error) {
	for _, d := range r.drivers {
		log.WithFields(log.Fields{
			"deviceName": deviceName,
			"mountPoint": mountPoint,
			"driverName": d.Name()}).Info("getting mounts")
		mounts, err := d.GetMounts(deviceName, mountPoint)
		if err != nil {
			return nil, err
		}
		return mounts, nil
	}
	return nil, errors.ErrNoOSDetected
}

func (r *odm) Mounted(mountPoint string) (bool, error) {
	for _, d := range r.drivers {
		log.WithFields(log.Fields{
			"mountPoint": mountPoint,
			"driverName": d.Name()}).Info("checking filesystem mount")
		return d.Mounted(mountPoint)
	}
	return false, errors.ErrNoOSDetected
}

func (r *odm) Unmount(mountPoint string) error {
	for _, d := range r.drivers {
		log.WithFields(log.Fields{
			"mountPoint": mountPoint,
			"driverName": d.Name()}).Info("unmounting filesystem")
		return d.Unmount(mountPoint)
	}
	return errors.ErrNoOSDetected
}

func (r *odm) Mount(
	device, target, mountOptions, mountLabel string) error {
	for _, d := range r.drivers {
		log.WithFields(log.Fields{
			"device":       device,
			"target":       target,
			"mountOptions": mountOptions,
			"mountLabel":   mountLabel,
			"driverName":   d.Name()}).Info("mounting filesystem")
		return d.Mount(device, target, mountOptions, mountLabel)
	}
	return errors.ErrNoOSDetected
}

func (r *odm) Format(
	deviceName, fsType string, overwriteFs bool) error {
	for _, d := range r.drivers {
		log.WithFields(log.Fields{
			"deviceName":  deviceName,
			"fsType":      fsType,
			"overwriteFs": overwriteFs,
			"driverName":  d.Name()}).Info(
			"formatting if blank or overwriteFs specified")
		return d.Format(deviceName, fsType, overwriteFs)
	}
	return errors.ErrNoOSDetected
}
