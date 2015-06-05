package storagedriver

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/emccode/rexray/util"
)

var (
	driverInitFuncs map[string]InitFunc
	drivers         map[string]Driver
	debug           string
)

var (
	ErrDriverInstanceDiscovery = errors.New("Driver Instance discovery failed")
)

var Adapters map[string]Driver

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
	Name        string
	VolumeID    string
	SnapshotID  string
	VolumeSize  string
	StartTime   string
	Description string
	Status      string
}

type Volume struct {
	Name             string
	VolumeID         string
	AvailabilityZone string
	Status           string
	VolumeType       string
	IOPS             int64
	Size             string
	Attachments      []*VolumeAttachment
}

type VolumeAttachment struct {
	VolumeID   string
	InstanceID string
	DeviceName string
	Status     string
}

type Driver interface {
	// Lists the block devices that are attached to the instance
	GetVolumeMapping() (interface{}, error)

	// Get the local instance
	GetInstance() (interface{}, error)

	// Get all Volumes available from infrastructure and storage platform
	GetVolume(string, string) (interface{}, error)

	// Get the currently attached Volumes
	GetVolumeAttach(string, string) (interface{}, error)

	// Create a snpashot of a Volume
	CreateSnapshot(bool, string, string, string) (interface{}, error)

	// Get all Snapshots or specific Snapshots
	GetSnapshot(string, string, string) (interface{}, error)

	// Remove Snapshot
	RemoveSnapshot(string) error

	// Create a Volume from scratch, from a Snaphot, or from another Volume
	CreateVolume(bool, string, string, string, string, int64, int64, string) (interface{}, error)

	// Remove Volume
	RemoveVolume(string) error

	// Get the next available Linux device for attaching external storage
	GetDeviceNextAvailable() (string, error)

	// Attach a Volume to an Instance
	AttachVolume(bool, string, string) (interface{}, error)

	// Detach a Volume from an Instance
	DetachVolume(bool, string, string) error

	// Copy a Snapshot to another region
	CopySnapshot(bool, string, string, string, string, string) (interface{}, error)
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
		if len(storageDriversArr) > 0 && !util.StringInSlice(name, storageDriversArr) {
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
