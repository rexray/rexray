package executor

import (
	"fmt"

	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
)

const (
	// Name1 is the name of the first mock driver.
	Name1 = "mock1"

	// Name2 is the name of the second mock driver.
	Name2 = "mock2"

	// Name3 is the name of the third mock driver.
	Name3 = "mock3"
)

var (
	NextDeviceVals = []string{"/dev/mock0", "/dev/mock1", "/dev/mock2"}
)

// Executor is the storage executor for the VFS storage driver.
type Executor struct {
	Config     gofig.Config
	name       string
	instanceID *types.InstanceID
}

func init() {
	registry.RegisterStorageExecutor(Name1, newExecutor1)
	registry.RegisterStorageExecutor(Name2, newExecutor2)
	registry.RegisterStorageExecutor(Name3, newExecutor3)
}

func newExecutor1() drivers.StorageExecutor {
	return NewExecutor(Name1)
}

func newExecutor2() drivers.StorageExecutor {
	return NewExecutor(Name2)
}

func newExecutor3() drivers.StorageExecutor {
	return NewExecutor(Name3)
}

// NewExecutor returns a new executor.
func NewExecutor(name string) *Executor {
	return &Executor{
		name:       name,
		instanceID: getInstanceID(name),
	}
}

func (d *Executor) Init(config gofig.Config) error {
	d.Config = config
	return nil
}

func (d *Executor) Name() string {
	return d.name
}

// InstanceID returns the local system's InstanceID.
func (d *Executor) InstanceID(
	ctx context.Context,
	opts types.Store) (*types.InstanceID, error) {
	return d.instanceID, nil
}

// NextDevice returns the next available device.
func (d *Executor) NextDevice(
	ctx context.Context,
	opts types.Store) (string, error) {
	return "", nil
}

// LocalDevices returns a map of the system's local devices.
func (d *Executor) LocalDevices(
	ctx context.Context,
	opts types.Store) (map[string]string, error) {
	return nil, nil
}

func pwn(name, v string) string {
	return fmt.Sprintf("%s-%s", name, v)
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
