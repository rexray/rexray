package libstorage

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	apiclient "github.com/emccode/libstorage/api/client"
	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
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
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	logFields := log.Fields{}

	addr := config.GetString(types.ConfigHost)
	d.ctx = ctx.WithValue(context.HostKey, addr)
	ctx.Debug("got configured host address")

	proto, lAddr, err := gotil.ParseAddress(addr)
	if err != nil {
		return err
	}

	tlsConfig, err := utils.ParseTLSConfig(config, logFields)
	if err != nil {
		return err
	}

	host := getHost(proto, lAddr, tlsConfig)
	disableKeepAlive := config.GetBool(types.ConfigHTTPDisableKeepAlive)
	lsxPath := config.GetString(types.ConfigExecutorPath)

	logFields["host"] = host
	logFields["disableKeepAlive"] = disableKeepAlive
	logFields["lsxPath"] = lsxPath

	httpTransport := &http.Transport{
		Dial: func(string, string) (net.Conn, error) {
			if tlsConfig == nil {
				return net.Dial(proto, lAddr)
			}
			return tls.Dial(proto, lAddr, tlsConfig)
		},
		DisableKeepAlives: disableKeepAlive,
	}

	apiClient := apiclient.New(host, httpTransport)
	logReq := config.GetBool(types.ConfigLogHTTPRequests)
	logRes := config.GetBool(types.ConfigLogHTTPResponses)
	apiClient.LogRequests(logReq)
	apiClient.LogResponses(logRes)

	logFields["enableInstanceIDHeaders"] = EnableInstanceIDHeaders
	logFields["enableLocalDevicesHeaders"] = EnableLocalDevicesHeaders
	logFields["logRequests"] = logReq
	logFields["logResponses"] = logRes

	newIIDCache := utils.NewStore
	dur, err := time.ParseDuration(
		config.GetString(types.ConfigClientCacheInstanceID))
	if err != nil {
		logFields["iidCacheDuration"] = dur.String()
		newIIDCache = func() types.Store {
			return utils.NewTTLStore(dur, true)
		}
	}

	d.client = client{
		APIClient:       apiClient,
		ctx:             ctx,
		config:          config,
		serviceCache:    &lss{Store: utils.NewStore()},
		lsxCache:        &lss{Store: utils.NewStore()},
		instanceIDCache: &lss{Store: newIIDCache()},
	}

	if err := d.dial(ctx); err != nil {
		return err
	}

	logFields["server"] = d.ServerName()

	d.ctx.WithFields(logFields).Info(
		"successfully dialed libStorage")

	return nil
}
