package libstorage

import (
	"crypto/tls"
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	apiclient "github.com/emccode/libstorage/api/client"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	lstypes "github.com/emccode/libstorage/drivers/storage/libstorage/types"
)

var (
	// EnableInstanceIDHeaders is a flag indicating whether or not the
	// client will automatically send the instance ID header(s) along with
	// storage-related API requests. The default is enabled.
	EnableInstanceIDHeaders = true

	// EnableLocalDevicesHeaders is a flag indicating whether or not the
	// client will automatically send the local devices header(s) along with
	// storage-related API requests. The default is enabled.
	EnableLocalDevicesHeaders = true
)

type driver struct {
	client
}

func newDriver() types.StorageDriver {
	var d lstypes.Driver = &driver{}
	return d
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	logFields := log.Fields{}

	addr := config.GetString(types.ConfigHost)
	ctx = ctx.WithContextSID(types.ContextHost, addr)
	ctx.Debug("got configured host address")

	proto, lAddr, err := gotil.ParseAddress(addr)
	if err != nil {
		return err
	}

	tlsConfig, err := utils.ParseTLSConfig(config, logFields)
	if err != nil {
		return err
	}

	logHTTPReq := config.GetBool(types.ConfigLogHTTPRequests)
	logHTTPRes := config.GetBool(types.ConfigLogHTTPResponses)
	disableKeepAlive := config.GetBool(types.ConfigHTTPDisableKeepAlive)
	lsxPath := config.GetString(types.ConfigExecutorPath)

	d.client = client{
		Client: apiclient.Client{
			Host:         getHost(proto, lAddr, tlsConfig),
			Headers:      http.Header{},
			LogRequests:  logHTTPReq,
			LogResponses: logHTTPRes,
			Client: &http.Client{
				Transport: &http.Transport{
					Dial: func(string, string) (net.Conn, error) {
						if tlsConfig == nil {
							return net.Dial(proto, lAddr)
						}
						return tls.Dial(proto, lAddr, tlsConfig)
					},
					DisableKeepAlives: disableKeepAlive,
				},
			},
		},
		ctx:                ctx,
		config:             config,
		lsxBinPath:         lsxPath,
		enableIIDHeader:    EnableInstanceIDHeaders,
		enableLclDevHeader: EnableLocalDevicesHeaders,
	}

	if err := d.dial(ctx); err != nil {
		return err
	}

	d.ctx.WithFields(logFields).Info(
		"successfully dialed libStorage")

	return nil
}
