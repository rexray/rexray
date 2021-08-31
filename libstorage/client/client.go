package client

import (
	gofig "github.com/akutz/gofig/types"
	log "github.com/sirupsen/logrus"
	gocontext "golang.org/x/net/context"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	apicnfg "github.com/AVENTER-UG/rexray/libstorage/api/utils/config"

	// load the config
	_ "github.com/AVENTER-UG/rexray/libstorage/imports/config"

	// load the libStorage storage executors
	_ "github.com/AVENTER-UG/rexray/libstorage/imports/executors"

	// load the libStorage storage driver
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/libstorage"
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

	if goCtx == nil {
		goCtx = context.Background()
	}

	ctx := context.New(goCtx)

	if _, ok := context.PathConfig(ctx); !ok {
		pathConfig := utils.NewPathConfig()
		ctx = ctx.WithValue(context.PathConfigKey, pathConfig)
		registry.ProcessRegisteredConfigs(ctx)
	}

	if config == nil {
		var err error
		if config, err = apicnfg.NewConfig(ctx); err != nil {
			return nil, err
		}
	}

	config = config.Scope(types.ConfigClient)
	types.BackCompat(config)

	var (
		c   *client
		err error
	)

	c = &client{ctx: ctx, config: config}
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

	if v := config.GetString(types.ConfigService); v != "" {
		c.ctx = c.ctx.WithValue(context.ServiceKey, v)
		c.ctx.WithField("serviceName", v).Info("set client service name")
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
