// +build mock

package mock

import (
	"fmt"

	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/drivers"
)

type driver struct {
	config         gofig.Config
	name           string
	instanceID     *types.InstanceID
	nextDeviceInfo *types.NextDeviceInfo
	volumes        []*types.Volume
	snapshots      []*types.Snapshot
	storageType    types.StorageType
}

const (
	// Driver1Name is the name of the first mock driver.
	Driver1Name = "mockDriver1"

	// Driver2Name is the name of the second mock driver.
	Driver2Name = "mockDriver2"

	// Driver3Name is the name of the third mock driver.
	Driver3Name = "mockDriver3"
)

func init() {
	registry.RegisterStorageDriver(Driver1Name, newMockDriver1)
	registry.RegisterStorageDriver(Driver2Name, newMockDriver2)
	registry.RegisterStorageDriver(Driver3Name, newMockDriver3)
}

func newMockDriver1() drivers.StorageDriver {
	return newMockDriver(Driver1Name)
}

func newMockDriver2() drivers.StorageDriver {
	return newMockDriver(Driver2Name)
}

func newMockDriver3() drivers.StorageDriver {
	return newMockDriver(Driver3Name)
}

func newMockDriver(name string) drivers.StorageDriver {
	d := &driver{
		name:        name,
		storageType: types.Block,
		instanceID:  getInstanceID(name),
	}

	d.nextDeviceInfo = &types.NextDeviceInfo{
		Prefix:  "xvd",
		Pattern: `\w`,
		Ignore:  getDeviceIgnore(name),
	}

	d.volumes = []*types.Volume{
		&types.Volume{
			Name:             d.pwn("Volume 0"),
			ID:               d.pwn("vol-000"),
			AvailabilityZone: d.pwn("zone-000"),
			Type:             "gold",
			Size:             10240,
		},
		&types.Volume{
			Name:             d.pwn("Volume 1"),
			ID:               d.pwn("vol-001"),
			AvailabilityZone: d.pwn("zone-001"),
			Type:             "gold",
			Size:             40960,
		},
		&types.Volume{
			Name:             d.pwn("Volume 2"),
			ID:               d.pwn("vol-002"),
			AvailabilityZone: d.pwn("zone-002"),
			Type:             "gold",
			Size:             163840,
		},
	}

	d.snapshots = []*types.Snapshot{
		&types.Snapshot{
			Name:     d.pwn("Snapshot 0"),
			ID:       d.pwn("snap-000"),
			VolumeID: d.pwn("vol-000"),
		},
		&types.Snapshot{
			Name:     d.pwn("Snapshot 1"),
			ID:       d.pwn("snap-001"),
			VolumeID: d.pwn("vol-001"),
		},
		&types.Snapshot{
			Name:     d.pwn("Snapshot 2"),
			ID:       d.pwn("snap-002"),
			VolumeID: d.pwn("vol-002"),
		},
	}

	return d
}

func (d *driver) Init(config gofig.Config) error {
	d.config = config
	return nil
}

func (d *driver) Name() string {
	return d.name
}

func (d *driver) Type() types.StorageType {
	return d.storageType
}

func pwn(name, v string) string {
	return fmt.Sprintf("%s-%s", name, v)
}

func (d *driver) pwn(v string) string {
	return fmt.Sprintf("%s-%s", d.name, v)
}

func getInstanceID(name string) *types.InstanceID {
	return &types.InstanceID{
		ID:       pwn(name, "InstanceID"),
		Metadata: instanceIDMetadata(),
	}
}

func instanceIDMetadata() map[string]interface{} {
	return map[string]interface{}{
		"min":     0,
		"max":     10,
		"rad":     "cool",
		"totally": "tubular",
	}
}

func getDeviceIgnore(driver string) bool {
	if driver == Driver2Name {
		return true
	}
	return false
}

func getDeviceName(driver string) string {
	var deviceName string
	switch driver {
	case Driver1Name:
		deviceName = "/dev/xvdb"
	case Driver2Name:
		deviceName = "/dev/xvda"
	case Driver3Name:
		deviceName = "/dev/xvdc"
	}
	return deviceName
}

func getNextDeviceName(driver string) string {
	var deviceName string
	switch driver {
	case Driver1Name:
		deviceName = "/dev/xvdc"
	case Driver3Name:
		deviceName = "/dev/xvdb"
	}
	return deviceName
}
