package executor

import (
	"fmt"
	"os"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

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

	devFilePath string
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

	d.devFilePath = fmt.Sprintf("%s/dev", d.RootDir())
	if !gotil.FileExists(d.devFilePath) {

	}

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

// RootDir returns the path to the VFS root directory.
func (d *Executor) RootDir() string {
	return d.Config.GetString("vfs.root")
}

func configRegistration() *gofig.Registration {

	defaultRootDir := fmt.Sprintf("%s/libstorage-vfs", os.TempDir())
	r := gofig.NewRegistration("VFS")
	r.Key(gofig.String, "", "", defaultRootDir, "vfs.root")
	return r
}
