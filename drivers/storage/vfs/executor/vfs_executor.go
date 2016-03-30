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
	// Name is the name of the storage executor and driver.
	Name = "vfs"
)

// Executor is the storage executor for the VFS storage driver.
type Executor struct {

	// Config is the executor's configuration instance.
	Config gofig.Config
}

func init() {
	gofig.Register(configRegistration())

	registry.RegisterStorageExecutor(Name, newExecutor)
}

func newExecutor() drivers.StorageExecutor {
	return &Executor{}
}

func (d *Executor) Init(config gofig.Config) error {
	d.Config = config
	return nil
}

func (d *Executor) Name() string {
	return Name
}

// InstanceID returns the local system's InstanceID.
func (d *Executor) InstanceID(
	ctx context.Context,
	opts types.Store) (*types.InstanceID, error) {

	hostName, err := getHostName()
	if err != nil {
		return nil, err
	}
	return &types.InstanceID{ID: hostName}, nil
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

func (d *Executor) rootDir() string {
	return d.Config.GetString("vfs.rootDir")
}

func (d *Executor) volumesDir() string {
	return fmt.Sprintf("%s/volumes", d.rootDir())
}

func (d *Executor) snapshotsDir() string {
	return fmt.Sprintf("%s/snapshots", d.rootDir())
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("VFS")
	r.Key(gofig.String, "", "", "", "vfs.rootDir")
	return r
}
