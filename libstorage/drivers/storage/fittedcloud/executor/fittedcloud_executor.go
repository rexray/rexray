package executor

import (
	gofig "github.com/akutz/gofig/types"

	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"

	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/fittedcloud"
	fcUtils "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/fittedcloud/utils"
)

// driver is the storage executor for the ec2 storage driver.
type driver struct {
	name   string
	config gofig.Config
}

func init() {
	registry.RegisterStorageExecutor(fittedcloud.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{name: fittedcloud.Name}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	// ensure backwards compatibility with ebs and ec2 in config
	d.config = config
	return nil
}

func (d *driver) Name() string {
	return d.name
}

// Supported returns a flag indicating whether or not the platform
// implementing the executor is valid for the host on which the executor
// resides.
func (d *driver) Supported(
	ctx types.Context,
	opts types.Store) (bool, error) {
	return fcUtils.IsEC2Instance(ctx)
}

// InstanceID returns the instance ID from the current instance from metadata
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {
	return fcUtils.InstanceID(ctx)
}

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return fcUtils.NextDevice(ctx, opts)
}

// LocalDevices retrieves device paths currently attached and/or mounted
func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {
	return fcUtils.LocalDevices(ctx, opts)
}
