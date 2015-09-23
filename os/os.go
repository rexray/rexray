package os

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/mount"
	"github.com/emccode/rexray/config"
	"github.com/emccode/rexray/errors"
	"github.com/emccode/rexray/util"
)

var driverInitFuncs map[string]InitFunc

type InitFunc func(conf *config.Config) (Driver, error)

func init() {
	driverInitFuncs = make(map[string]InitFunc)
}

func Register(name string, initFunc InitFunc) {
	driverInitFuncs[name] = initFunc
}

type OSDriverManager struct {
	Drivers map[string]Driver
	Config  *config.Config
}

func NewOSDriverManager(conf *config.Config) (*OSDriverManager, error) {

	drivers, err := getDrivers(conf)
	if err != nil {
		return nil, err
	}

	if len(drivers) == 0 {
		return nil, errors.New("no os drivers initialized")
	}

	return &OSDriverManager{drivers, conf}, nil
}

func (osdm *OSDriverManager) IsDrivers() bool {
	return len(osdm.Drivers) > 0
}

func getDrivers(conf *config.Config) (map[string]Driver, error) {

	driverNames := conf.OsDrivers

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

func (osdm *OSDriverManager) GetMounts(deviceName, mountPoint string) ([]*mount.Info, error) {

	for _, driver := range osdm.Drivers {
		log.WithFields(log.Fields{
			"deviceName": deviceName,
			"mountPoint": mountPoint,
			"driverName": driver.Name()}).Info("getting mounts")
		mounts, err := driver.GetMounts(deviceName, mountPoint)
		if err != nil {
			return nil, err
		}
		return mounts, nil
	}

	return nil, errors.New("No OS detected")
}

func (osdm *OSDriverManager) Mounted(mountPoint string) (bool, error) {
	for _, driver := range osdm.Drivers {
		log.WithFields(log.Fields{
			"mountPoint": mountPoint,
			"driverName": driver.Name()}).Info("checking filesystem mount")
		return driver.Mounted(mountPoint)
	}
	return false, errors.New("No OS detected")
}

func (osdm *OSDriverManager) Unmount(mountPoint string) error {
	for _, driver := range osdm.Drivers {
		log.WithFields(log.Fields{
			"mountPoint": mountPoint,
			"driverName": driver.Name()}).Info("unmounting filesystem")
		return driver.Unmount(mountPoint)
	}
	return errors.New("No OS detected")
}

func (osdm *OSDriverManager) Mount(device, target, mountOptions, mountLabel string) error {
	for _, driver := range osdm.Drivers {
		log.WithFields(log.Fields{
			"device":       device,
			"target":       target,
			"mountOptions": mountOptions,
			"mountLabel":   mountLabel,
			"driverName":   driver.Name()}).Info("mounting filesystem")
		return driver.Mount(device, target, mountOptions, mountLabel)
	}
	return errors.New("No OS detected")
}

func (osdm *OSDriverManager) Format(deviceName, fsType string, overwriteFs bool) error {
	for _, driver := range osdm.Drivers {
		log.WithFields(log.Fields{
			"deviceName":  deviceName,
			"fsType":      fsType,
			"overwriteFs": overwriteFs,
			"driverName":  driver.Name()}).Info("formatting if blank or overwriteFs specified")
		return driver.Format(deviceName, fsType, overwriteFs)
	}
	return errors.New("No OS detected")
}
