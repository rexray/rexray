package client

import (
	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	apicnfg "github.com/emccode/libstorage/api/utils/config"

	// load the local imports
	_ "github.com/emccode/libstorage/imports/local"
)

type client struct {
	config gofig.Config
	sd     types.StorageDriver
	od     types.OSDriver
	id     types.IntegrationDriver
	ctx    types.Context
	api    types.APIClient
	xli    types.StorageExecutorCLI
}

// New returns a new libStorage client.
func New(config gofig.Config) (types.Client, error) {

	if config == nil {
		var err error
		if config, err = apicnfg.NewConfig(); err != nil {
			return nil, err
		}
	}

	config = config.Scope(types.ConfigClient)

	var (
		c   *client
		err error
	)

	c = &client{ctx: context.Background(), config: config}
	c.ctx = c.ctx.WithClient(c)

	if config.IsSet(types.ConfigService) {
		c.ctx = c.ctx.WithServiceName(config.GetString(types.ConfigService))
	}

	storageDriverName := config.GetString(types.ConfigStorageDriver)
	if c.sd, err = registry.NewStorageDriver(storageDriverName); err != nil {
		return nil, err
	}
	if err = c.sd.Init(c.ctx, config); err != nil {
		return nil, err
	}
	if papi, ok := c.sd.(types.ProvidesAPIClient); ok {
		c.api = papi.API()
	}
	if pxli, pxliOk := c.sd.(types.ProvidesStorageExecutorCLI); pxliOk {
		c.xli = pxli.XCLI()
	}

	c.ctx = c.ctx.WithContextSID(types.ContextStorageDriver, storageDriverName)
	c.ctx.Info("storage driver initialized")

	// if the API or XLI are nil, then the storage driver is not the libStorage
	// storage driver, and we should jump avoid any more initialization
	if c.api == nil || c.xli == nil {
		c.ctx.Info("created libStorage client")
		return c, nil
	}

	osDriverName := config.GetString(types.ConfigOSDriver)
	if c.od, err = registry.NewOSDriver(osDriverName); err != nil {
		return nil, err
	}
	if err = c.od.Init(c.ctx, config); err != nil {
		return nil, err
	}
	c.ctx = c.ctx.WithContextSID(types.ContextOSDriver, osDriverName)
	c.ctx.Info("os driver initialized")

	integrationDriverName := config.GetString(types.ConfigIntegrationDriver)
	if c.id, err = registry.NewIntegrationDriver(
		integrationDriverName); err != nil {
		return nil, err
	}
	if err := c.id.Init(c.ctx, config); err != nil {
		return nil, err
	}
	c.ctx = c.ctx.WithContextSID(
		types.ContextIntegrationDriver, integrationDriverName)
	c.ctx.Info("integration driver initialized")

	c.ctx.Info("created libStorage client")
	return c, nil
}
