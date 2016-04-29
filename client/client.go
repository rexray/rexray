package client

import (
	"runtime"

	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	libstor "github.com/emccode/libstorage/drivers/storage/libstorage"

	// load the local imports
	_ "github.com/emccode/libstorage/imports/local"
)

func init() {
	registerConfig()
}

// Client is the libStorage client.
type Client interface {

	// API returns the underlying libStorage API client.
	API() libstor.Client

	// Driver returns the client's underlying storage driver.
	Driver() types.StorageDriver

	// List returns all volumes attached to this instance.
	List(service string) ([]*types.Volume, error)

	// Inspect returns a specific volume as identified by the provided volume
	// name.
	Inspect(service, volumeName string) (*types.Volume, error)

	// Mount will return a mount point path when specifying either a volumeName
	// or volumeID.  If overwriteFS is true the file system will be overwritten
	// based on the newFSType if there is not an existing filesystem.
	Mount(
		service, volumeID, volumeName string,
		opts *types.VolumeMountOpts) (string, *types.Volume, error)

	// Unmount unmounts the volume.
	Unmount(service, volumeID, volumeName string) error

	// Path returns the mounted path of the volume.
	Path(service, volumeID, volumeName string) (string, error)

	// Create creates a new volume with the specified name and options.
	Create(
		service, volumeName string,
		opts *types.VolumeCreateOpts) (*types.Volume, error)

	// Remove removes the volume.
	Remove(service, volumeName string) error

	// Attach attaches the volume.
	Attach(
		service, volumeName string,
		opts *types.VolumeAttachOpts) (string, error)

	// Detach detaches the volume.
	Detach(service, volumeName string, opts *types.VolumeDetachOpts) error

	// NetworkName returns an identifier of a volume that is relevant when
	// correlating a local device to a device that is the volumeName to the
	// local instanceID.
	NetworkName(service, volumeName string) (string, error)
}

type client struct {
	config gofig.Config
	sd     types.StorageDriver
	od     types.OSDriver
	id     types.IntegrationDriver
	ctx    types.Context
	lsc    libstor.Client
}

// New returns a new libStorage client.
func New(config gofig.Config) (Client, error) {

	ctx := context.Background()

	osDriverName := config.GetString(osDriverKey)
	od, err := registry.NewOSDriver(osDriverName)
	if err != nil {
		return nil, err
	}
	if err := od.Init(config); err != nil {
		return nil, err
	}
	ctx = ctx.WithContextID("osDriver", osDriverName)
	ctx.Info("os driver initialized")

	integrationDriverName := config.GetString(integrationDriverKey)
	id, err := registry.NewIntegrationDriver(integrationDriverName)
	if err != nil {
		return nil, err
	}
	if err := id.Init(config); err != nil {
		return nil, err
	}
	ctx = ctx.WithContextID("integrationDriver", integrationDriverName)
	ctx.Info("integration driver initialized")

	storageDriverName := config.GetString(storageDriverKey)
	sd, err := registry.NewStorageDriver(storageDriverName)
	if err != nil {
		return nil, err
	}
	if err := sd.Init(config); err != nil {
		return nil, err
	}
	ctx = ctx.WithContextID("storageDriver", storageDriverName)
	ctx.Info("storage driver initialized")

	c := &client{
		od:     od,
		sd:     sd,
		id:     id,
		ctx:    ctx,
		config: config,
	}

	if lsd, ok := c.sd.(libstor.Driver); ok {
		c.lsc = lsd.API()
	}

	ctx.Info("created libStorage client")
	return c, nil
}

const (
	osDriverKey          = "libstorage.client.types.os"
	storageDriverKey     = "libstorage.client.types.storage"
	integrationDriverKey = "libstorage.client.types.integration"
)

func registerConfig() {
	r := gofig.NewRegistration("libStorage Client")
	r.Key(gofig.String, "", runtime.GOOS, "", osDriverKey)
	r.Key(gofig.String, "", libstor.Name, "", storageDriverKey)
	r.Key(gofig.String, "", "docker", "", integrationDriverKey)
	gofig.Register(r)
}
