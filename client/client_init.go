package client

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	apiclient "github.com/emccode/libstorage/api/client"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/api/utils"
	apiconfig "github.com/emccode/libstorage/api/utils/config"
	"github.com/emccode/libstorage/api/utils/paths"
	"github.com/emccode/libstorage/api/utils/semaphore"
)

var (
	lsxMutex semaphore.Semaphore
)

func init() {
	var err error
	for {
		lsxMutex, err = semaphore.Open(types.LSX, false, 0644, 1)
		if err != nil {
			log.WithError(err).Warn(err)
		} else {
			break
		}
	}

	registerConfig()
}

const (
	clientScope          = "libstorage.client"
	hostKey              = "libstorage.host"
	logEnabledKey        = "libstorage.client.http.logging.enabled"
	logOutKey            = "libstorage.client.http.logging.out"
	logErrKey            = "libstorage.client.http.logging.err"
	logRequestsKey       = "libstorage.client.http.logging.logrequest"
	logResponsesKey      = "libstorage.client.http.logging.logresponse"
	disableKeepAlivesKey = "libstorage.client.http.disableKeepAlives"
	lsxOffline           = "libstorage.client.executor.offline"

	// LSXPathKey is the configuration key for the libStorage executor
	// binary path.
	LSXPathKey = "libstorage.client.executor.path"
)

type lsc struct {
	apiclient.Client
	config             gofig.Config
	svcInfo            apihttp.ServicesMap
	lsxInfo            apihttp.ExecutorsMap
	lsxBinPath         string
	ctx                context.Context
	enableIIDHeader    bool
	enableLclDevHeader bool
}

// New returns a new Client.
func New(config gofig.Config) (Client, error) {

	logFields := log.Fields{}

	if config == nil {
		var err error
		if config, err = apiconfig.NewConfig(); err != nil {
			return nil, err
		}
	}

	addr := config.GetString(hostKey)
	ctx := context.WithContextID(context.Background(), "host", addr)

	proto, lAddr, err := gotil.ParseAddress(addr)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := utils.ParseTLSConfig(
		config.Scope(clientScope), logFields)
	if err != nil {
		return nil, err
	}

	c := &lsc{
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
		config:             config,
		ctx:                ctx,
		enableIIDHeader:    EnableInstanceIDHeaders,
		enableLclDevHeader: EnableLocalDevicesHeaders,
		lsxBinPath:         config.GetString(LSXPathKey),
	}

	if err := c.updateServiceInfo(); err != nil {
		return nil, err
	}

	c.ctx = c.ctx.WithContextID(context.ContextKeyServerName, c.ServerName())

	if !config.GetBool(lsxOffline) {
		if err := c.updateExecutorInfo(); err != nil {
			return nil, err
		}

		if err := c.updateExecutor(); err != nil {
			return nil, err
		}
	}

	if err := c.updateInstanceIDs(); err != nil {
		return nil, err
	}

	if err := c.updateLocalDevices(); err != nil {
		return nil, err
	}

	ctx.Log().WithFields(logFields).Debug("created new libStorage client")

	return c, nil
}

// Close releases system resources.
func Close() error {
	return lsxMutex.Close()
}

func (c *lsc) InstanceIDCallback(callback func(service string) (*types.InstanceID, error)) {

}

func (c *lsc) EnableInstanceIDHeaders(enabled bool) {
	c.enableIIDHeader = enabled
}

func (c *lsc) EnableLocalDevicesHeaders(enabled bool) {
	c.enableLclDevHeader = enabled
}

func (c *lsc) API() *apiclient.Client {
	return &c.Client
}

func (c *lsc) ServerName() string {
	return c.Client.ServerName
}

func getHost(proto, lAddr string, tlsConfig *tls.Config) string {
	if tlsConfig != nil && tlsConfig.ServerName != "" {
		return tlsConfig.ServerName
	} else if proto == "unix" {
		return "libstorage-server"
	} else {
		return lAddr
	}
}

func (c *lsc) updateServiceInfo() error {
	ctx := c.getTXCTX()
	ctx.Log().Debug("getting service information")
	svcInfo, err := c.Client.Services(ctx)
	if err != nil {
		return err
	}
	c.svcInfo = svcInfo
	return nil
}

type iidHeader struct {
	driverName string
	headerName string
	headerValu string
}

func (c *lsc) updateInstanceIDs() error {
	if !c.enableIIDHeader {
		return nil
	}

	ctx := c.getTXCTX()
	ctx.Log().Debug("getting instance IDs")
	cache := map[string]*iidHeader{}

	for service, si := range c.svcInfo {

		if _, ok := cache[si.Driver.Name]; ok {
			continue
		}

		iid, err := c.instanceID(ctx, service)
		if err != nil {
			return err
		}

		var h *iidHeader

		if len(iid.Metadata) == 0 {
			h = &iidHeader{
				headerName: apihttp.InstanceIDHeader,
				headerValu: iid.ID,
			}
		} else {
			jBuf, err := json.Marshal(iid)
			if err != nil {
				return err
			}
			h = &iidHeader{
				headerName: apihttp.InstanceID64Header,
				headerValu: base64.StdEncoding.EncodeToString(jBuf),
			}
		}

		h.driverName = si.Driver.Name
		cache[h.driverName] = h
	}

	for _, h := range cache {
		c.Client.Headers.Add(
			h.headerName,
			fmt.Sprintf("%s=%s", h.driverName, h.headerValu))
	}

	return nil
}

type ldHeader struct {
	driverName string
	headerName string
	headerValu map[string]string
}

func (c *lsc) updateLocalDevices() error {
	if !c.enableLclDevHeader {
		return nil
	}

	ctx := c.getTXCTX()
	ctx.Log().Debug("getting local devices")

	cache := map[string]*ldHeader{}

	for service, si := range c.svcInfo {

		if _, ok := cache[si.Driver.Name]; ok {
			continue
		}

		ldm, err := c.localDevices(ctx, service)
		if err != nil {
			return err
		}

		h := &ldHeader{
			driverName: si.Driver.Name,
			headerName: apihttp.LocalDevicesHeader,
			headerValu: ldm,
		}

		cache[h.driverName] = h
	}

	for _, h := range cache {
		buf := &bytes.Buffer{}

		fmt.Fprintf(buf, "%s=", h.driverName)
		for device, mountPoint := range h.headerValu {
			fmt.Fprintf(buf, "%s=%s, ", device, mountPoint)
		}

		if buf.Len() > (len(h.driverName) + 1) {
			buf.Truncate(buf.Len() - 2)
		}

		c.Client.Headers.Add(h.headerName, buf.String())
	}

	return nil
}

func (c *lsc) getServiceInfo(service string) (*types.ServiceInfo, error) {
	si, ok := c.svcInfo[strings.ToLower(service)]
	if !ok {
		return nil, goof.WithField("name", service, "unknown service")
	}
	return si, nil
}

func registerConfig() {
	r := gofig.NewRegistration("libStorage Client")
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
