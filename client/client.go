package client

import (
	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	lstypes "github.com/emccode/libstorage/drivers/storage/libstorage/types"

	// load the local imports
	_ "github.com/emccode/libstorage/imports/local"
)

// Client is the libStorage client.
type Client interface {

	// API returns the underlying libStorage API client.
	API() lstypes.Client

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
	lsc    lstypes.Client
}

// New returns a new libStorage client.
func New(config gofig.Config) (Client, error) {

	ctx := context.Background().WithConfig(config)

	osDriverName := ctx.GetString(types.ConfigOSDriver)
	od, err := registry.NewOSDriver(osDriverName)
	if err != nil {
		return nil, err
	}
	if err := od.Init(ctx, config); err != nil {
		return nil, err
	}
	ctx = ctx.WithOSDriver(
		od,
	).WithContextSID(
		types.ContextOSDriver, osDriverName,
	)
	ctx.Info("os driver initialized")

	storageDriverName := ctx.GetString(types.ConfigStorageDriver)
	sd, err := registry.NewStorageDriver(storageDriverName)
	if err != nil {
		return nil, err
	}
	if err := sd.Init(ctx, config); err != nil {
		return nil, err
	}
	ctx = ctx.WithStorageDriver(
		sd,
	).WithContextSID(
		types.ContextStorageDriver, storageDriverName,
	)
	ctx.Info("storage driver initialized")

	integrationDriverName := ctx.GetString(types.ConfigIntegrationDriver)
	id, err := registry.NewIntegrationDriver(integrationDriverName)
	if err != nil {
		return nil, err
	}
	if err := id.Init(ctx, config); err != nil {
		return nil, err
	}
	ctx = ctx.WithIntegrationDriver(
		id,
	).WithContextSID(
		types.ContextIntegrationDriver, integrationDriverName,
	)
	ctx.Info("integration driver initialized")

	c := &client{
		od:     od,
		sd:     sd,
		id:     id,
		ctx:    ctx,
		config: config,
	}
	if lsd, ok := c.sd.(lstypes.Driver); ok {
		c.lsc = lsd.API()
	}

	ctx.Info("created libStorage client")
	return c, nil
}
