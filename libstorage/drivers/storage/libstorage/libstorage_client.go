package libstorage

import (
	"errors"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
)

type client struct {
	types.APIClient
	ctx             types.Context
	config          gofig.Config
	tlsConfig       *types.TLSConfig
	pathConfig      *types.PathConfig
	clientType      types.ClientType
	lsxCache        *lss
	serviceCache    *lss
	supportedCache  *lss
	instanceIDCache types.Store
}

var errExecutorNotSupported = errors.New("executor not supported")

func (c *client) isController() bool {
	return c.clientType == types.ControllerClient
}

func (c *client) dial(ctx types.Context) error {

	svcInfos, err := c.Services(ctx)
	if err != nil {
		return err
	}

	// controller clients do not have any additional dialer logic
	if c.isController() {
		return nil
	}

	store := utils.NewStore()
	c.ctx = c.ctx.WithValue(context.ServerKey, c.ServerName())

	for service := range svcInfos {
		ctx := c.ctx.WithValue(context.ServiceKey, service)
		ctx.Info("initializing supported cache")
		lsxSO, err := c.Supported(ctx, store)
		if err != nil {
			return goof.WithError("error initializing supported cache", err)
		}

		if lsxSO == types.LSXSOpNone {
			ctx.Warn("executor not supported")
			continue
		}

		ctx.Info("initializing instance ID cache")
		if _, err := c.InstanceID(ctx, store); err != nil {
			if err == types.ErrNotImplemented {
				ctx.WithError(err).Warn("cannot get instance ID")
				continue
			}
			return goof.WithError("error initializing instance ID cache", err)
		}
	}

	return nil
}

func getHost(
	ctx types.Context,
	proto, lAddr string, tlsConfig *types.TLSConfig) string {

	if tlsConfig != nil && tlsConfig.ServerName != "" {
		ctx.WithField("getHost", tlsConfig.ServerName).Debug(
			`getHost tlsConfig != nil && tlsConfig.ServerName != ""`)
		return tlsConfig.ServerName
	} else if proto == "unix" {
		ctx.WithField("getHost", "libstorage-server").Debug(
			`getHost proto == "unix"`)
		return "libstorage-server"
	} else {
		ctx.WithField("getHost", lAddr).Debug(
			`getHost lAddr`)
		return lAddr
	}
}

func (c *client) getServiceInfo(service string) (*types.ServiceInfo, error) {

	if si := c.serviceCache.GetServiceInfo(service); si != nil {
		return si, nil
	}
	return nil, goof.WithField("name", service, "unknown service")
}
