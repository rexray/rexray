package client

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	"golang.org/x/net/context/ctxhttp"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/api/utils"
)

// Client is the base interface for the libStorage API client..
type Client interface {

	// Services returns a map of the configured Services.
	Services() (apihttp.ServicesMap, error)

	// ServiceInspect returns information about a service.
	ServiceInspect(service string) (*types.ServiceInfo, error)

	// Volumes returns a list of all Volumes for all Services.
	Volumes() (apihttp.ServiceVolumeMap, error)

	// VolumeInspect gets information about a single volume.
	VolumeInspect(
		service, volumeID string, attachments bool) (*types.Volume, error)

	// VolumeCreate creates a single volume.
	VolumeCreate(
		service string,
		request *apihttp.VolumeCreateRequest) (*types.Volume, error)

	// VolumeRemove removes a single volume.
	VolumeRemove(
		service, volumeID string) error

	// VolumeSnapshot creates a single snapshot.
	VolumeSnapshot(
		service string,
		volumeID string,
		request *apihttp.VolumeSnapshotRequest) (*types.Snapshot, error)

	// Volumes returns a list of all Snapshots for all services.
	Snapshots() (apihttp.ServiceSnapshotMap, error)

	// ServiceSnapshots returns a single snapshot from a service.
	SnapshotsByService(
		service string) (apihttp.SnapshotMap, error)

	// VolumeInspect gets information about a single snapshot.
	SnapshotInspect(
		service, snapshotID string) (*types.Snapshot, error)

	// SnapshotCreate creates a single volume from a snapshot.
	SnapshotCreate(
		service, snapshotID string,
		request *apihttp.VolumeCreateRequest) (*types.Volume, error)

	// SnapshotRemove removes a single snapshot.
	SnapshotRemove(
		service, snapshotID string) error

	// SnapshotCopy copies a snapshot to a new snapshot.
	SnapshotCopy(
		service, snapshotID string,
		request *apihttp.SnapshotCopyRequest) (*types.Snapshot, error)
}

// APIClient is the extended interface for the libStorage API client.
type APIClient interface {
	Client

	// Root returns a list of root resources.
	Root() (apihttp.RootResources, error)

	// Executors returns information about the executors.
	Executors() (apihttp.ExecutorsMap, error)

	// ExecutorGet downloads an executor.
	ExecutorGet(name string) (io.ReadCloser, error)

	// ExecutorHead returns information about an executor.
	ExecutorHead(name string) (*types.ExecutorInfo, error)
}

type client struct {
	config       gofig.Config
	httpClient   *http.Client
	proto        string
	laddr        string
	tlsConfig    *tls.Config
	logRequests  bool
	logResponses bool
	ctx          context.Context
}

// Dial opens a connection to a remote libStorage serice and returns the client
// that can be used to communicate with said endpoint.
//
// If the config parameter is nil a default instance is created. The
// function dials the libStorage service specified by the configuration
// property libstorage.host.
func Dial(
	ctx context.Context,
	config gofig.Config) (APIClient, error) {

	c := &client{config: config}
	c.logRequests = c.config.GetBool(
		"libstorage.client.http.logging.logrequest")
	c.logResponses = c.config.GetBool(
		"libstorage.client.http.logging.logresponse")

	logFields := log.Fields{}

	host := config.GetString("libstorage.host")
	if host == "" {
		return nil, goof.New("libstorage.host is required")
	}

	tlsConfig, tlsFields, err :=
		utils.ParseTLSConfig(config.Scope("libstorage.client"))
	if err != nil {
		return nil, err
	}
	c.tlsConfig = tlsConfig
	for k, v := range tlsFields {
		logFields[k] = v
	}

	cProto, cLaddr, err := gotil.ParseAddress(host)
	if err != nil {
		return nil, err
	}
	c.proto = cProto
	c.laddr = cLaddr

	if ctx == nil {
		log.Debug("created empty context for client")
		ctx = context.Background()
	}
	ctx = ctx.WithContextID("host", host)

	c.httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: func(proto, addr string) (conn net.Conn, err error) {
				if tlsConfig == nil {
					return net.Dial(cProto, cLaddr)
				}
				return tls.Dial(cProto, cLaddr, tlsConfig)
			},
		},
	}

	ctx.Log().WithFields(logFields).Debug("configured client")

	c.ctx = ctx
	return c, nil
}

func (c *client) Root() (apihttp.RootResources, error) {
	reply := apihttp.RootResources{}
	if _, err := c.httpGet("/", &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (c *client) Services() (apihttp.ServicesMap, error) {
	reply := apihttp.ServicesMap{}
	if _, err := c.httpGet("/services", &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (c *client) ServiceInspect(name string) (*types.ServiceInfo, error) {
	reply := &types.ServiceInfo{}
	if _, err := c.httpGet(
		fmt.Sprintf("/services/%s", name), &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (c *client) SnapshotsByService(
	service string) (apihttp.SnapshotMap, error) {
	reply := apihttp.SnapshotMap{}
	if _, err := c.httpGet(
		fmt.Sprintf("/snapshots/%s", service), &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (c *client) Volumes() (apihttp.ServiceVolumeMap, error) {
	reply := apihttp.ServiceVolumeMap{}
	if _, err := c.httpGet("/volumes", &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (c *client) VolumeInspect(
	service, volumeID string, attachments bool) (*types.Volume, error) {

	reply := types.Volume{}
	if _, err := c.httpGet(
		fmt.Sprintf("/volumes/%s/%s", service, volumeID), &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *client) VolumeCreate(
	service string,
	request *apihttp.VolumeCreateRequest) (*types.Volume, error) {

	reply := types.Volume{}
	if _, err := c.httpPost(
		fmt.Sprintf("/volumes/%s", service), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *client) VolumeRemove(
	service, volumeID string) error {

	if _, err := c.httpDelete(
		fmt.Sprintf("/volumes/%s/%s", service, volumeID), nil); err != nil {
		return err
	}
	return nil
}

func (c *client) VolumeSnapshot(
	service string,
	volumeID string,
	request *apihttp.VolumeSnapshotRequest) (*types.Snapshot, error) {

	reply := types.Snapshot{}
	if _, err := c.httpPost(
		fmt.Sprintf("/volumes/%s/%s?snapshot", service, volumeID), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *client) Snapshots() (apihttp.ServiceSnapshotMap, error) {
	reply := apihttp.ServiceSnapshotMap{}
	if _, err := c.httpGet("/snapshots", &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (c *client) SnapshotInspect(
	service, snapshotID string) (*types.Snapshot, error) {
	reply := types.Snapshot{}
	if _, err := c.httpGet(
		fmt.Sprintf("/snapshots/%s/%s", service, snapshotID), &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *client) SnapshotCreate(
	service, snapshotID string,
	request *apihttp.VolumeCreateRequest) (*types.Volume, error) {
	reply := types.Volume{}
	if _, err := c.httpPost(
		fmt.Sprintf("/snapshots/%s/%s?create", service, snapshotID), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *client) SnapshotRemove(
	service, snapshotID string) error {
	if _, err := c.httpDelete(
		fmt.Sprintf("/snapshots/%s/%s", service, snapshotID), nil); err != nil {
		return err
	}
	return nil
}

func (c *client) SnapshotCopy(
	service, snapshotID string,
	request *apihttp.SnapshotCopyRequest) (*types.Snapshot, error) {
	reply := types.Snapshot{}
	if _, err := c.httpPost(
		fmt.Sprintf("/snapshots/%s/%s?copy", service, snapshotID), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (c *client) Executors() (apihttp.ExecutorsMap, error) {
	reply := apihttp.ExecutorsMap{}
	if _, err := c.httpGet("/executors", &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (c *client) ExecutorHead(name string) (*types.ExecutorInfo, error) {
	res, err := c.httpHead(fmt.Sprintf("/executors/%s", name))
	if err != nil {
		return nil, err
	}

	size, err := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return nil, err
	}

	buf, err := base64.StdEncoding.DecodeString(res.Header.Get("Content-MD5"))
	if err != nil {
		return nil, err
	}

	return &types.ExecutorInfo{
		Name:        name,
		Size:        size,
		MD5Checksum: fmt.Sprintf("%x", buf),
	}, nil
}

func (c *client) ExecutorGet(name string) (io.ReadCloser, error) {
	res, err := c.httpGet(fmt.Sprintf("/executors/%s", name), nil)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (c *client) httpDo(
	method, path string, payload, reply interface{}) (*http.Response, error) {

	reqBody, err := encPayload(payload)
	if err != nil {
		return nil, err
	}

	host := c.laddr
	if c.proto == "unix" {
		host = "libstorage-server"
	}
	if c.tlsConfig != nil && c.tlsConfig.ServerName != "" {
		host = c.tlsConfig.ServerName
	}

	url := fmt.Sprintf("http://%s%s", host, path)
	c.ctx.Log().WithField("url", url).Debug("built request url")
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	c.logRequest(req)

	res, err := ctxhttp.Do(c.ctx, c.httpClient, req)
	if err != nil {
		return nil, err
	}
	c.logResponse(res)

	if res.StatusCode > 299 {
		je := &types.JSONError{}
		if err := json.NewDecoder(res.Body).Decode(je); err != nil {
			return res, goof.WithField("status", res.StatusCode, "http error")
		}
		return res, je
	}

	if req.Method != http.MethodHead && reply != nil {
		if err := decRes(res.Body, reply); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (c *client) httpGet(
	path string, reply interface{}) (*http.Response, error) {
	return c.httpDo("GET", path, nil, reply)
}

func (c *client) httpHead(
	path string) (*http.Response, error) {
	return c.httpDo("HEAD", path, nil, nil)
}

func (c *client) httpPost(
	path string,
	payload interface{}, reply interface{}) (*http.Response, error) {
	return c.httpDo("POST", path, payload, reply)
}

func (c *client) httpDelete(
	path string, reply interface{}) (*http.Response, error) {
	return c.httpDo("DELETE", path, nil, reply)
}

func encPayload(payload interface{}) (io.Reader, error) {
	if payload == nil {
		return nil, nil
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf), nil
}

func decRes(body io.Reader, reply interface{}) error {
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(buf, reply); err != nil {
		return err
	}
	return nil
}

func (c *client) logRequest(req *http.Request) {

	if !c.logRequests {
		return
	}

	w := log.StandardLogger().Writer()

	fmt.Fprintln(w, "")
	fmt.Fprint(w, "    -------------------------- ")
	fmt.Fprint(w, "HTTP REQUEST (CLIENT)")
	fmt.Fprintln(w, " -------------------------")

	buf, err := httputil.DumpRequest(req, true)
	if err != nil {
		return
	}

	gotil.WriteIndented(w, buf)
	fmt.Fprintln(w)
}

func (c *client) logResponse(res *http.Response) {

	if !c.logResponses {
		return
	}

	w := log.StandardLogger().Writer()

	fmt.Fprintln(w)
	fmt.Fprint(w, "    -------------------------- ")
	fmt.Fprint(w, "HTTP RESPONSE (CLIENT)")
	fmt.Fprintln(w, " -------------------------")

	buf, err := httputil.DumpResponse(
		res,
		res.Header.Get("Content-Type") != "application/octet-stream")
	if err != nil {
		return
	}

	bw := &bytes.Buffer{}
	gotil.WriteIndented(bw, buf)

	scanner := bufio.NewScanner(bw)
	for {
		if !scanner.Scan() {
			break
		}
		fmt.Fprintln(w, scanner.Text())
	}
}

func (c *client) getLocalDevices(prefix string) ([]string, error) {

	path := c.config.GetString("libstorage.client.localdevicesfile")
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	deviceNames := []string{}
	scan := bufio.NewScanner(bytes.NewReader(buf))

	rx := regexp.MustCompile(fmt.Sprintf(`^.+?\s(%s\w+)$`, prefix))
	for scan.Scan() {
		l := scan.Text()
		m := rx.FindStringSubmatch(l)
		if len(m) > 0 {
			deviceNames = append(deviceNames, fmt.Sprintf("/dev/%s", m[1]))
		}
	}

	return deviceNames, nil
}
