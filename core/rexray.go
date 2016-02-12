package core

import (
	log "github.com/Sirupsen/logrus"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"
)

// RexRay is the library's entrance type and storage management platform.
type RexRay struct {
	Config  gofig.Config
	OS      OSDriverManager
	Volume  VolumeDriverManager
	Storage StorageDriverManager
	Context string
	drivers map[string]Driver
}

// New creates a new REX-Ray instance and configures it with the
// provided configuration instance.
func New(conf gofig.Config) *RexRay {

	if conf == nil {
		conf = gofig.New()
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

	log.Info(r.Config.Get("rexray.osDrivers"))
	log.Info(r.Config.Get("rexray.volumeDrivers"))
	log.Info(r.Config.Get("rexray.storageDrivers"))

	osDrivers := r.Config.GetStringSlice("rexray.osDrivers")
	volDrivers := r.Config.GetStringSlice("rexray.volumeDrivers")
	storDrivers := r.Config.GetStringSlice("rexray.storageDrivers")

	log.WithFields(log.Fields{
		"osDrivers":      osDrivers,
		"volumeDrivers":  volDrivers,
		"storageDrivers": storDrivers,
	}).Debug("core get drivers")

	for n, d := range r.drivers {
		switch td := d.(type) {
		case OSDriver:
			if gotil.StringInSlice(n, osDrivers) {
				if err := d.Init(r); err != nil {
					log.WithFields(log.Fields{
						"driverName": n,
						"error":      err}).Debug("error initializing driver")
					continue
				}
				od[n] = td
			}
		case VolumeDriver:
			if gotil.StringInSlice(n, volDrivers) {
				if err := d.Init(r); err != nil {
					log.WithFields(log.Fields{
						"driverName": n,
						"error":      err}).Debug("error initializing driver")
					continue
				}
				vd[n] = td
			}
		case StorageDriver:
			if gotil.StringInSlice(n, storDrivers) {
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
