package libstorage

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"

	apiclient "github.com/codedellemc/libstorage/api/client"
	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/api/utils"
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
	d.ctx.Debug("got configured host address")

	if tok := config.GetString(types.ConfigClientAuthToken); len(tok) > 0 {
		if gotil.FileExists(tok) {
			d.ctx.WithField("tokenFilePath", tok).Debug(
				"reading client token file")
			buf, err := ioutil.ReadFile(tok)
			if err != nil {
				d.ctx.WithField("tokenFilePath", tok).WithError(err).Error(
					"error reading client token file")
				return err
			}
			tok = string(buf)
		}
		d.ctx = d.ctx.WithValue(context.EncodedAuthTokenKey, tok)
		d.ctx.WithField("encodedToken", tok).Debug("got configured auth token")
		logFields["encodedToken"] = tok
	}

	proto, lAddr, err := gotil.ParseAddress(addr)
	if err != nil {
		return err
	}

	tlsConfig, err := utils.ParseTLSConfig(
		d.ctx, config, logFields, types.ConfigClient)
	if err != nil {
		return err
	}

	host := getHost(proto, lAddr, tlsConfig)
	lsxPath := config.GetString(types.ConfigExecutorPath)
	cliType := types.ParseClientType(config.GetString(types.ConfigClientType))
	disableKeepAlive := config.GetBool(types.ConfigHTTPDisableKeepAlive)

	logFields["host"] = host
	logFields["lsxPath"] = lsxPath
	logFields["clientType"] = cliType
	logFields["disableKeepAlive"] = disableKeepAlive

	httpTransport := &http.Transport{
		Dial: func(string, string) (net.Conn, error) {
			if tlsConfig == nil {
				conn, err := net.Dial(proto, lAddr)
				if err != nil {
					return nil, err
				}
				d.ctx.Debug("successful connection")
				return conn, nil
			}

			conn, err := tls.Dial(proto, lAddr, &tlsConfig.Config)
			if err != nil {
				return nil, err
			}

			if !tlsConfig.VerifyPeers {
				d.ctx.Debug("successful tls connection; not verifying peers")
				return conn, nil
			}

			if err := verifyKnownHost(
				d.ctx,
				conn.ConnectionState().PeerCertificates,
				tlsConfig.KnownHost); err != nil {

				d.ctx.WithError(err).Error("error matching peer fingerprint")
				return nil, err
			}

			if err := verifyKnownHostFiles(
				d.ctx,
				conn.ConnectionState().PeerCertificates,
				tlsConfig.UsrKnownHosts,
				tlsConfig.SysKnownHosts); err != nil {

				d.ctx.WithError(err).Error("error matching known host")
				return nil, err
			}

			return conn, nil
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

	pathConfig := context.MustPathConfig(d.ctx)

	lsxMutexPath := path.Join(pathConfig.Run, "lsx.lock")
	logFields["lsxMutexPath"] = lsxMutexPath

	d.client = client{
		APIClient:    apiClient,
		ctx:          d.ctx,
		config:       config,
		tlsConfig:    tlsConfig,
		pathConfig:   pathConfig,
		clientType:   cliType,
		lsxMutexPath: lsxMutexPath,
		serviceCache: &lss{Store: utils.NewStore()},
	}

	if d.clientType == types.IntegrationClient {

		newIIDCache := utils.NewStore
		dur, err := time.ParseDuration(
			config.GetString(types.ConfigClientCacheInstanceID))
		if err != nil {
			logFields["iidCacheDuration"] = dur.String()
			newIIDCache = func() types.Store {
				return utils.NewTTLStore(dur, true)
			}
		}

		d.lsxCache = &lss{Store: utils.NewStore()}
		d.supportedCache = &lss{Store: utils.NewStore()}
		d.instanceIDCache = newIIDCache()
	}

	d.ctx.WithFields(logFields).Info("created libStorage client")

	if err := d.dial(d.ctx); err != nil {
		return err
	}

	d.ctx.Info("successefully dialed libStorage server")
	return nil
}
