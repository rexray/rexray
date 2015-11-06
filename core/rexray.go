package core

import (
	log "github.com/Sirupsen/logrus"

	"github.com/emccode/rexray/core/config"
	"github.com/emccode/rexray/util"
)

// RexRay is the library's entrance type and storage management platform.
type RexRay struct {
	Config  *config.Config
	OS      OSDriverManager
	Volume  VolumeDriverManager
	Storage StorageDriverManager
	drivers map[string]Driver
}

// New creates a new REX-Ray instance and configures it with the
// provided configuration instance.
func New(conf *config.Config) *RexRay {

	if conf == nil {
		conf = config.New()
	}

	r := &RexRay{
		Config:  conf,
		drivers: map[string]Driver{},
	}

	for name, ctor := range driverCtors {
		r.drivers[name] = ctor()
		log.WithField("driverName", name).Debug("constructed driver")
	}

	return r
}

// InitDrivers initializes the drivers for the REX-Ray platform.
func (r *RexRay) InitDrivers() error {

	od := map[string]OSDriver{}
	vd := map[string]VolumeDriver{}
	sd := map[string]StorageDriver{}

	log.Info(r.Config.Get("osDrivers"))
	log.Info(r.Config.Get("volumeDrivers"))
	log.Info(r.Config.Get("storageDrivers"))

	osDrivers := r.Config.GetStringSlice("osDrivers")
	volDrivers := r.Config.GetStringSlice("volumeDrivers")
	storDrivers := r.Config.GetStringSlice("storageDrivers")

	log.WithFields(log.Fields{
		"osDrivers":      osDrivers,
		"volumeDrivers":  volDrivers,
		"storageDrivers": storDrivers,
	}).Debug("core get drivers")

	for n, d := range r.drivers {
		switch td := d.(type) {
		case OSDriver:
			if util.StringInSlice(n, osDrivers) {
				if err := d.Init(r); err != nil {
					log.WithFields(log.Fields{
						"driverName": n,
						"error":      err}).Debug("error initializing driver")
					continue
				}
				od[n] = td
			}
		case VolumeDriver:
			if util.StringInSlice(n, volDrivers) {
				if err := d.Init(r); err != nil {
					log.WithFields(log.Fields{
						"driverName": n,
						"error":      err}).Debug("error initializing driver")
					continue
				}
				vd[n] = td
			}
		case StorageDriver:
			if util.StringInSlice(n, storDrivers) {
				if err := d.Init(r); err != nil {
					log.WithFields(log.Fields{
						"driverName": n,
						"error":      err}).Debug("error initializing driver")
					continue
				}
				sd[n] = td
			}
		}
	}

	r.OS = &odm{
		rexray:  r,
		drivers: od,
	}

	r.Volume = &vdm{
		rexray:  r,
		drivers: vd,
	}

	r.Storage = &sdm{
		rexray:  r,
		drivers: sd,
	}

	if err := r.OS.Init(r); err != nil {
		return err
	}

	if err := r.Volume.Init(r); err != nil {
		return err
	}

	if err := r.Storage.Init(r); err != nil {
		return err
	}

	return nil
}

// DriverNames returns a list of the registered driver names.
func (r *RexRay) DriverNames() <-chan string {
	c := make(chan string)
	go func() {
		for n := range r.drivers {
			c <- n
		}
		close(c)
	}()
	return c
}
