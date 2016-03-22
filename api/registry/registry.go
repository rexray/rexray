// Package registry is the central hub for Drivers and other types that
// follow the init-time registration.
package registry

import (
	"strings"
	"sync"

	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types/drivers"
)

var (
	storDriverCtors    = map[string]drivers.NewStorageDriver{}
	storDriverCtorsRWL = &sync.RWMutex{}

	osDriverCtors    = map[string]drivers.NewOSDriver{}
	osDriverCtorsRWL = &sync.RWMutex{}

	intDriverCtors    = map[string]drivers.NewIntegrationDriver{}
	intDriverCtorsRWL = &sync.RWMutex{}

	routers    = []httputils.Router{}
	routersRWL = &sync.RWMutex{}
)

// RegisterRouter registers a Router.
func RegisterRouter(router httputils.Router) {
	routersRWL.Lock()
	defer routersRWL.Unlock()
	routers = append(routers, router)
}

// RegisterStorageDriver registers a StorageDriver.
func RegisterStorageDriver(name string, ctor drivers.NewStorageDriver) {
	storDriverCtorsRWL.Lock()
	defer storDriverCtorsRWL.Unlock()
	storDriverCtors[strings.ToLower(name)] = ctor
}

// RegisterOSDriver registers a OSDriver.
func RegisterOSDriver(name string, ctor drivers.NewOSDriver) {
	osDriverCtorsRWL.Lock()
	defer osDriverCtorsRWL.Unlock()
	osDriverCtors[strings.ToLower(name)] = ctor
}

// RegisterIntegrationDriver registers a IntegrationDriver.
func RegisterIntegrationDriver(name string, ctor drivers.NewIntegrationDriver) {
	intDriverCtorsRWL.Lock()
	defer intDriverCtorsRWL.Unlock()
	intDriverCtors[strings.ToLower(name)] = ctor
}

// NewStorageDriver returns a new instance of the driver specified by the
// driver name.
func NewStorageDriver(name string) (drivers.StorageDriver, error) {

	var ok bool
	var ctor drivers.NewStorageDriver

	func() {
		storDriverCtorsRWL.RLock()
		defer storDriverCtorsRWL.RUnlock()
		ctor, ok = storDriverCtors[strings.ToLower(name)]
	}()

	if !ok {
		return nil, goof.WithField("driver", name, "invalid driver name")
	}

	return ctor(), nil
}

// NewOSDriver returns a new instance of the driver specified by the
// driver name.
func NewOSDriver(name string) (drivers.OSDriver, error) {

	var ok bool
	var ctor drivers.NewOSDriver

	func() {
		osDriverCtorsRWL.RLock()
		defer osDriverCtorsRWL.RUnlock()
		ctor, ok = osDriverCtors[strings.ToLower(name)]
	}()

	if !ok {
		return nil, goof.WithField("driver", name, "invalid driver name")
	}

	return ctor(), nil
}

// NewIntegrationDriver returns a new instance of the driver specified by the
// driver name.
func NewIntegrationDriver(name string) (drivers.IntegrationDriver, error) {

	var ok bool
	var ctor drivers.NewIntegrationDriver

	func() {
		intDriverCtorsRWL.RLock()
		defer intDriverCtorsRWL.RUnlock()
		ctor, ok = intDriverCtors[strings.ToLower(name)]
	}()

	if !ok {
		return nil, goof.WithField("driver", name, "invalid driver name")
	}

	return ctor(), nil
}

// StorageDrivers returns a channel on which new instances of all registered
// storage drivers can be received.
func StorageDrivers() <-chan drivers.StorageDriver {
	c := make(chan drivers.StorageDriver)
	go func() {
		storDriverCtorsRWL.RLock()
		defer storDriverCtorsRWL.RUnlock()
		for _, ctor := range storDriverCtors {
			c <- ctor()
		}
		close(c)
	}()
	return c
}

// OSDrivers returns a channel on which new instances of all registered
// OS drivers can be received.
func OSDrivers() <-chan drivers.OSDriver {
	c := make(chan drivers.OSDriver)
	go func() {
		osDriverCtorsRWL.RLock()
		defer osDriverCtorsRWL.RUnlock()
		for _, ctor := range osDriverCtors {
			c <- ctor()
		}
		close(c)
	}()
	return c
}

// IntegrationDrivers returns a channel on which new instances of all registered
// integration drivers can be received.
func IntegrationDrivers() <-chan drivers.IntegrationDriver {
	c := make(chan drivers.IntegrationDriver)
	go func() {
		intDriverCtorsRWL.RLock()
		defer intDriverCtorsRWL.RUnlock()
		for _, ctor := range intDriverCtors {
			c <- ctor()
		}
		close(c)
	}()
	return c
}

// Routers returns a channel on which new instances of all registered routers
// can be received.
func Routers() <-chan httputils.Router {
	c := make(chan httputils.Router)
	go func() {
		routersRWL.RLock()
		defer routersRWL.RUnlock()
		for _, r := range routers {
			c <- r
		}
		close(c)
	}()
	return c
}
