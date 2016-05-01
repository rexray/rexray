package executor

import (
	"encoding/json"

	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
)

const (
	// Name is the name of the driver.
	Name = "mock"
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
	registry.RegisterStorageExecutor(Name, newExecutor)
}

func newExecutor() types.StorageExecutor {
	return NewExecutor()
}

// NewExecutor returns a new executor.
func NewExecutor() *Executor {
	return &Executor{
		name:       Name,
		instanceID: GetInstanceID(),
	}
}

func (d *Executor) Init(ctx types.Context, config gofig.Config) error {
	d.Config = config
	return nil
}

func (d *Executor) Name() string {
	return d.name
}

// InstanceID returns the local system's InstanceID.
func (d *Executor) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {
	return d.instanceID, nil
}

// NextDevice returns the next available device.
func (d *Executor) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "/dev/xvde", nil
}

// LocalDevices returns a map of the system's local devices.
func (d *Executor) LocalDevices(
	ctx types.Context,
	opts types.Store) (map[string]string, error) {
	return map[string]string{
		"/dev/xvda": "/var/log",
		"/dev/xvdb": "/home",
		"/dev/xvdc": "/net/share",
		"/dev/xvdd": "/var/lib/backup",
	}, nil
}

// GetInstanceID gets the mock instance ID.
func GetInstanceID() *types.InstanceID {
	return &types.InstanceID{
		ID:       "12345",
		Metadata: instanceIDMetadata(),
	}
}

func instanceIDMetadata() json.RawMessage {
	metadata, _ := json.Marshal(map[string]interface{}{
		"min":     0,
		"max":     10,
		"rad":     "cool",
		"totally": "tubular",
	})
	return metadata
}
