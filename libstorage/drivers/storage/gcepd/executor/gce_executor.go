package executor

import (
	"io/ioutil"
	"path"
	"regexp"

	gofig "github.com/akutz/gofig/types"

	"github.com/rexray/rexray/libstorage/api/registry"
	"github.com/rexray/rexray/libstorage/api/types"
	"github.com/rexray/rexray/libstorage/drivers/storage/gcepd"
	gceUtils "github.com/rexray/rexray/libstorage/drivers/storage/gcepd/utils"
)

const (
	diskIDPath = "/dev/disk/by-id"
	diskPrefix = "Google_PersistentDisk_"
)

// driver is the storage executor for the storage driver.
type driver struct {
	config gofig.Config
}

func init() {
	registry.RegisterStorageExecutor(gcepd.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	return nil
}

func (d *driver) Name() string {
	return gcepd.Name
}

// Supported returns a flag indicating whether or not the platform
// implementing the executor is valid for the host on which the executor
// resides.
func (d *driver) Supported(
	ctx types.Context,
	opts types.Store) (bool, error) {

	return gceUtils.IsGCEInstance(ctx)
}

// InstanceID returns the instance ID from the current instance from metadata
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	return gceUtils.InstanceID(ctx)
}

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {

	return "", types.ErrNotImplemented
}

// Retrieve device paths currently attached and/or mounted
func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	files, err := ioutil.ReadDir(diskIDPath)
	if err != nil {
		return nil, err
	}

	persistentDiskRX, err := regexp.Compile(
		diskPrefix + `(` + gceUtils.DiskNameRX + `)`)
	if err != nil {
		return nil, err
	}

	attachedDisks, err := gceUtils.GetDisks(ctx)
	if err != nil {
		return nil, err
	}

	ld := &types.LocalDevices{Driver: d.Name()}
	devMap := map[string]string{}
	for _, f := range files {
		if persistentDiskRX.MatchString(f.Name()) {
			matches := persistentDiskRX.FindStringSubmatch(f.Name())
			volID := matches[1]
			if _, ok := attachedDisks[volID]; ok && volID != "" {
				devMap[volID] = path.Join(diskIDPath, f.Name())
			}
		}
	}

	if len(devMap) > 0 {
		ld.DeviceMap = devMap
	}

	return ld, nil
}
