package client

import (
	log "github.com/Sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	gocontext "golang.org/x/net/context"

	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/api/utils"
	apicnfg "github.com/codedellemc/libstorage/api/utils/config"

	// load the local imports
	_ "github.com/codedellemc/libstorage/imports/local"
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
func New(goCtx gocontext.Context, config gofig.Config) (types.Client, error) {

	if config == nil {
		var err error
		if config, err = apicnfg.NewConfig(); err != nil {
			return nil, err
		}
	}

	config = config.Scope(types.ConfigClient)
	types.BackCompat(config)

	var (
		c   *client
		err error
	)

	c = &client{ctx: context.New(goCtx), config: config}
	c.ctx = c.ctx.WithValue(context.ClientKey, c)

	logFields := log.Fields{}
	logConfig, err := utils.ParseLoggingConfig(
		config, logFields, types.ConfigClient)
	if err != nil {
		return nil, err
	}

	// always update the server context's log level
	context.SetLogLevel(c.ctx, logConfig.Level)
	c.ctx.WithFields(logFields).Info("configured logging")

	if config.IsSet(types.ConfigService) {
		c.ctx = c.ctx.WithValue(
			context.ServiceKey, config.GetString(types.ConfigService))
	}

	storDriverName := config.GetString(types.ConfigStorageDriver)
	if storDriverName == "" {
		c.ctx.Warn("no storage driver found")
	} else {
		if c.sd, err = registry.NewStorageDriver(storDriverName); err != nil {
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
		c.ctx.Info("storage driver initialized")
	}

	// if the API or XLI are nil, then the storage driver is not the libStorage
	// storage driver, and we should jump avoid any more initialization
	if c.api == nil || c.xli == nil {
		c.ctx.Info("created libStorage client")
		return c, nil
	}

	osDriverName := config.GetString(types.ConfigOSDriver)
	if osDriverName == "" {
		c.ctx.Warn("no os driver found")
	} else {
		if c.od, err = registry.NewOSDriver(osDriverName); err != nil {
			return nil, err
		}
		if err = c.od.Init(c.ctx, config); err != nil {
			return nil, err
		}
		c.ctx.Info("os driver initialized")
	}

	intDriverName := config.GetString(types.ConfigIntegrationDriver)
	if intDriverName == "" {
		c.ctx.Warn("no integration driver found")
	} else {
		if c.id, err = registry.NewIntegrationDriver(
			intDriverName); err != nil {
			return nil, err
		}
		if err := c.id.Init(c.ctx, config); err != nil {
			return nil, err
		}
		c.ctx.Info("integration driver initialized")
	}

	c.ctx.Info("created libStorage client")
	return c, nil
}
