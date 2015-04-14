package storagedriver

import "errors"

import "strings"

var (
	driverInitFuncs map[string]InitFunc
	drivers         map[string]Driver
)

var (
	ErrDriverInstanceIDDiscovery = errors.New("Driver InstanceID discovery failed")
)

type BlockDevice struct {
	ProviderName string
	InstanceID   string
	VolumeID     string
	DeviceName   string
	Region       string
	Status       string
}

type Driver interface {
	GetBlockDeviceMapping() (interface{}, error)
	// GetInstance() (interface{}, error)
}

type InitFunc func() (Driver, error)

func Register(name string, initFunc InitFunc) error {
	driverInitFuncs[name] = initFunc
	return nil
}

func init() {
	driverInitFuncs = make(map[string]InitFunc)
	drivers = make(map[string]Driver)
}

func GetDrivers(storageDrivers string) (map[string]Driver, error) {
	var err error
	var storageDriversArr []string
	if storageDrivers != "" {
		storageDriversArr = strings.Split(storageDrivers, ",")
	}

	for name, initFunc := range driverInitFuncs {
		if len(storageDriversArr) > 0 && !stringInSlice(name, storageDriversArr) {
			continue
		}
		drivers[name], err = initFunc()
		if err != nil {
			delete(drivers, name)
		}
	}

	return drivers, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
