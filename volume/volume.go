package volume

import (
	log "github.com/Sirupsen/logrus"

	"github.com/emccode/rexray/config"
	"github.com/emccode/rexray/errors"
	osm "github.com/emccode/rexray/os"
	"github.com/emccode/rexray/storage"
	"github.com/emccode/rexray/util"
)

var driverInitFuncs map[string]InitFunc

type InitFunc func(
	osDriverManager *osm.OSDriverManager,
	storageDriverManager *storage.StorageDriverManager) (Driver, error)

func init() {
	driverInitFuncs = make(map[string]InitFunc)
}

func Register(name string, initFunc InitFunc) {
	driverInitFuncs[name] = initFunc
}

type VolumeDriverManager struct {
	Drivers map[string]Driver
}

func NewVolumeDriverManager(
	conf *config.Config,
	osDriverManager *osm.OSDriverManager,
	storageDriverManager *storage.StorageDriverManager) (*VolumeDriverManager, error) {

	vd, vdErr := getDrivers(conf, osDriverManager, storageDriverManager)
	if vdErr != nil {
		return nil, vdErr
	}

	if len(vd) == 0 {
		return nil, errors.New("no volume drivers initialized")
	}

	return &VolumeDriverManager{
		Drivers: vd,
	}, nil
}

func (vdm *VolumeDriverManager) IsDrivers() bool {
	return len(vdm.Drivers) > 0
}

func getDrivers(
	conf *config.Config,
	osdm *osm.OSDriverManager,
	sdm *storage.StorageDriverManager) (map[string]Driver, error) {

	driverNames := conf.VolumeDrivers

	log.WithFields(log.Fields{
		"driverInitFuncs": driverInitFuncs,
		"driverNames":     driverNames}).Debug("getting driver instances")

	drivers := map[string]Driver{}

	for name, initFunc := range driverInitFuncs {
		if len(driverNames) > 0 && !util.StringInSlice(name, driverNames) {
			continue
		}

		var initErr error
		drivers[name], initErr = initFunc(osdm, sdm)
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

func (vdm *VolumeDriverManager) Mount(volumeName, volumeID string, overwriteFs bool, newFsType string) (string, error) {
	for _, driver := range vdm.Drivers {
		return driver.Mount(volumeName, volumeID, overwriteFs, newFsType)
	}
	return "", errors.New("no volume manager specified")
}

func (vdm *VolumeDriverManager) Unmount(volumeName, volumeID string) error {
	for _, driver := range vdm.Drivers {
		return driver.Unmount(volumeName, volumeID)
	}
	return errors.New("no volume manager specified")
}

func (vdm *VolumeDriverManager) Path(volumeName, volumeID string) (string, error) {
	for _, driver := range vdm.Drivers {
		return driver.Path(volumeName, volumeID)
	}
	return "", errors.New("no volume manager specified")
}

func (vdm *VolumeDriverManager) Create(volumeName string) error {
	for _, driver := range vdm.Drivers {
		return driver.Create(volumeName)
	}
	return errors.New("no volume manager specified")
}

func (vdm *VolumeDriverManager) Remove(volumeName string) error {
	for _, driver := range vdm.Drivers {
		return driver.Remove(volumeName)
	}
	return errors.New("no volume manager specified")
}

func (vdm *VolumeDriverManager) Attach(volumeName, instanceID string) (string, error) {
	for _, driver := range vdm.Drivers {
		return driver.Attach(volumeName, instanceID)
	}
	return "", errors.New("no volume manager specified")
}

func (vdm *VolumeDriverManager) Detach(volumeName, instanceID string) error {
	for _, driver := range vdm.Drivers {
		return driver.Detach(volumeName, instanceID)
	}
	return errors.New("no volume manager specified")
}

func (vdm *VolumeDriverManager) NetworkName(volumeName, instanceID string) (string, error) {
	for _, driver := range vdm.Drivers {
		return driver.NetworkName(volumeName, instanceID)
	}
	return "", errors.New("no volume manager specified")
}
