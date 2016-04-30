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

	// OS returns the client's OS driver instance.
	OS() types.OSDriver

	// Storage returns the client's storage driver instance.
	Storage() types.StorageDriver

	// IntegrationDriver returns the client's integration driver instance.
	Integration() types.IntegrationDriver
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
	osDriverKey          = "libstorage.client.driver.os"
	storageDriverKey     = "libstorage.client.driver.storage"
	integrationDriverKey = "libstorage.client.driver.integration"
)

func registerConfig() {
	r := gofig.NewRegistration("libStorage Client")
	r.Key(gofig.String, "", runtime.GOOS, "", osDriverKey)
	r.Key(gofig.String, "", libstor.Name, "", storageDriverKey)
	r.Key(gofig.String, "", "docker", "", integrationDriverKey)
	gofig.Register(r)
}
