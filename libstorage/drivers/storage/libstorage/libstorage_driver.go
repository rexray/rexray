package libstorage

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"

	apiclient "github.com/rexray/rexray/libstorage/api/client"
	"github.com/rexray/rexray/libstorage/api/context"
	"github.com/rexray/rexray/libstorage/api/types"
	"github.com/rexray/rexray/libstorage/api/utils"
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
		d.ctx, config, proto, logFields, types.ConfigClient)
	if err != nil {
		return err
	}

	host := getHost(d.ctx, proto, lAddr, tlsConfig)
	lsxPath := config.GetString(types.ConfigExecutorPath)
	cliType := types.ParseClientType(config.GetString(types.ConfigClientType))
	disableKeepAlive := config.GetBool(types.ConfigHTTPDisableKeepAlive)

	logFields["lAddr"] = host
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

			const errMatch = "error matching peer fingerprint"

			// get the fqdn/IP of the endpoint to which the connection
			// is being made in case an ErrKnownHost error occurs
			hostSansPort := lAddr
			if hostParts := strings.Split(lAddr, ":"); len(hostParts) > 1 {
				hostSansPort = hostParts[0]
			}

			peerCerts := conn.ConnectionState().PeerCertificates

			if ok, err := verifyKnownHost(
				d.ctx,
				hostSansPort,
				peerCerts,
				tlsConfig.KnownHost); ok {

				return conn, nil

			} else if err != nil {

				d.ctx.WithError(err).Error(errMatch)
				return nil, err
			}

			if ok, err := verifyKnownHostFiles(
				d.ctx,
				hostSansPort,
				peerCerts,
				tlsConfig.UsrKnownHosts,
				tlsConfig.SysKnownHosts); ok {

				return conn, nil

			} else if err != nil {

				d.ctx.WithError(err).Error(errMatch)
				return nil, err
			}

			return nil, newErrKnownHost(hostSansPort, peerCerts)
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

	d.client = client{
		APIClient:    apiClient,
		ctx:          d.ctx,
		config:       config,
		tlsConfig:    tlsConfig,
		pathConfig:   pathConfig,
		clientType:   cliType,
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

	d.ctx.Info("successfully dialed libStorage server")
	return nil
}
