package client

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/emccode/libstorage/api/types"
)

// Client is the libStorage API client.
type client struct {
	http.Client
	host                      string
	logRequests               bool
	logResponses              bool
	serverName                string
	headers                   http.Header
	headersRWL                *sync.RWMutex
	enableInstanceIDHeaders   bool
	enableLocalDevicesHeaders bool
}

// New returns a new API client.
func New(host string, transport *http.Transport) types.APIClient {
	return &client{
		Client: http.Client{
			Transport: transport,
		},
		host:       host,
		headers:    http.Header{},
		headersRWL: &sync.RWMutex{},
	}
}

func (c *client) ServerName() string {
	return c.serverName
}

func (c *client) LogRequests(enabled bool) {
	c.logRequests = enabled
}

func (c *client) LogResponses(enabled bool) {
	c.logResponses = enabled
}

func (c *client) AddHeader(key, value string) {
	c.AddHeaderForDriver("", key, value)
}

func (c *client) AddHeaderForDriver(driverName, key, value string) {
	c.headersRWL.Lock()
	defer c.headersRWL.Unlock()

	var (
		ckey = http.CanonicalHeaderKey(key)
		vals = c.headers[ckey]
		xist = -1
		vrgx *regexp.Regexp
	)

	if vals == nil {
		vals = []string{}
	}

	if driverName == "" {
		vrgx = regexp.MustCompile(fmt.Sprintf(`(?i)%s`, value))
	} else {
		vrgx = regexp.MustCompile(fmt.Sprintf(`(?i)%s=.*`, driverName))
		value = fmt.Sprintf("%s=%s", driverName, value)
	}

	for x, v := range vals {
		if vrgx.MatchString(v) {
			xist = x
			break
		}
	}

	if xist >= 0 {
		vals = append(vals[:xist], vals[xist+1:]...)
	}
	vals = append(vals, value)

	c.headers[ckey] = vals
}

func (c *client) EnableInstanceIDHeaders(enabled bool) {
	c.enableInstanceIDHeaders = enabled
}

func (c *client) EnableLocalDevicesHeaders(enabled bool) {
	c.enableLocalDevicesHeaders = enabled
}
