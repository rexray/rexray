package libstorage

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	apiclient "github.com/emccode/libstorage/api/client"
	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	"github.com/emccode/libstorage/api/utils/paths"
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

const (
	clientScope          = "libstorage.client"
	hostKey              = "libstorage.host"
	logEnabledKey        = clientScope + ".http.logging.enabled"
	logOutKey            = clientScope + ".http.logging.out"
	logErrKey            = clientScope + ".http.logging.err"
	logRequestsKey       = clientScope + ".http.logging.logrequest"
	logResponsesKey      = clientScope + ".http.logging.logresponse"
	disableKeepAlivesKey = clientScope + ".http.disableKeepAlives"
	lsxOffline           = clientScope + ".executor.offline"

	// LSXPathKey is the configuration key for the libStorage executor
	// binary path.
	LSXPathKey = clientScope + ".executor.path"
)

func registerConfig() {
	r := gofig.NewRegistration("libStorage Storage Driver")
	lsxBinPath := fmt.Sprintf("%s/%s", paths.UsrDirPath(), types.LSX)
	r.Key(gofig.String, "", lsxBinPath, "", LSXPathKey)
	r.Key(gofig.Bool, "", false, "", lsxOffline)
	r.Key(gofig.Bool, "", false, "", logEnabledKey)
	r.Key(gofig.String, "", "", "", logOutKey)
	r.Key(gofig.String, "", "", "", logErrKey)
	r.Key(gofig.Bool, "", false, "", logRequestsKey)
	r.Key(gofig.Bool, "", false, "", logResponsesKey)
	r.Key(gofig.Bool, "", false, "", disableKeepAlivesKey)
	gofig.Register(r)
}

type driver struct {
	client
}

func newDriver() types.StorageDriver {
	var d Driver = &driver{}
	return d
}

func (d *driver) Init(config gofig.Config) error {
	logFields := log.Fields{}

	addr := config.GetString(hostKey)
	ctx := context.Background().WithContextID("host", addr)

	proto, lAddr, err := gotil.ParseAddress(addr)
	if err != nil {
		return err
	}

	tlsConfig, err := utils.ParseTLSConfig(
		config.Scope(clientScope), logFields)
	if err != nil {
		return err
	}

	d.client = client{
		Client: apiclient.Client{
			Host:         getHost(proto, lAddr, tlsConfig),
			Headers:      http.Header{},
			LogRequests:  config.GetBool(logRequestsKey),
			LogResponses: config.GetBool(logResponsesKey),
			Client: &http.Client{
				Transport: &http.Transport{
					Dial: func(string, string) (net.Conn, error) {
						if tlsConfig == nil {
							return net.Dial(proto, lAddr)
						}
						return tls.Dial(proto, lAddr, tlsConfig)
					},
					DisableKeepAlives: config.GetBool(disableKeepAlivesKey),
				},
			},
		},
		ctx:                ctx,
		config:             config,
		lsxBinPath:         config.GetString(LSXPathKey),
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
