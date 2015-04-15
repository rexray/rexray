package storagedriver

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	driverInitFuncs map[string]InitFunc
	drivers         map[string]Driver
	debug           string
)

var (
	ErrDriverInstanceDiscovery = errors.New("Driver Instance discovery failed")
)

type BlockDevice struct {
	ProviderName string
	InstanceID   string
	VolumeID     string
	DeviceName   string
	Region       string
	Status       string
}

type Instance struct {
	ProviderName string
	InstanceID   string
	Region       string
	Name         string
}

type Snapshot struct {
	VolumeID    string
	SnapshotID  string
	VolumeSize  string
	StartTime   string
	Description string
	Status      string
}

type Volume struct {
	VolumeID         string
	AvailabilityZone string
	Status           string
	VolumeType       string
	IOPS             int64
	Size             string
	Attachments      []VolumeAttachment
}

type VolumeAttachment struct {
	VolumeID   string
	InstanceID string
	DeviceName string
	Status     string
}

type Driver interface {
	GetBlockDeviceMapping() (interface{}, error)
	GetInstance() (interface{}, error)
	GetVolume(string) (interface{}, error)
	GetVolumeAttach(string, string) (interface{}, error)
	GetSnapshot(string, string) (interface{}, error)
	// CreateSnapshot(bool, string, string) (interface{}, error)
	// CreateSnapshotVolume(bool, string) (string, error)
	// RemoveSnapshot(string) error
	// CreateVolume(bool, string, string, int64, int64) (interface{}, error)
	// GetDeviceNextAvailable() (string, error)
	// AttachVolume(bool, string, string) (interface{}, error)
	// DetachVolume(bool, string) error
	// RemoveVolume(string) error
}

type InitFunc func() (Driver, error)

func Register(name string, initFunc InitFunc) error {
	driverInitFuncs[name] = initFunc
	return nil
}

func init() {
	driverInitFuncs = make(map[string]InitFunc)
	drivers = make(map[string]Driver)
	debug = strings.ToUpper(os.Getenv("REXRAY_DEBUG"))
}

func GetDrivers(storageDrivers string) (map[string]Driver, error) {
	var err error
	var storageDriversArr []string
	if storageDrivers != "" {
		storageDriversArr = strings.Split(storageDrivers, ",")
	}

	if debug == "TRUE" {
		fmt.Println(driverInitFuncs)
	}

	for name, initFunc := range driverInitFuncs {
		if len(storageDriversArr) > 0 && !stringInSlice(name, storageDriversArr) {
			continue
		}
		drivers[name], err = initFunc()
		if err != nil {
			if debug == "TRUE" {
				fmt.Println(fmt.Sprintf("Info (%s): %s", name, err))
			}
			delete(drivers, name)
		}
	}

	return drivers, nil
}
