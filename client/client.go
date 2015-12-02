package client

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	gjson "github.com/gorilla/rpc/json"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"

	"github.com/emccode/libstorage/api"
)

var (
	netProtoRx = regexp.MustCompile("(?i)tcp")

	letters = []string{
		"a", "b", "c", "d", "e", "f", "g", "h",
		"i", "j", "k", "l", "m", "n", "o", "p"}
)

// Client is the interface for Golang libStorage clients.
type Client interface {

	// GetInstanceID gets the instance ID.
	GetInstanceID(ctx context.Context) (*api.InstanceID, error)

	// GetNextAvailableDeviceName gets the name of the next available device.
	GetNextAvailableDeviceName(
		ctx context.Context,
		args *api.GetNextAvailableDeviceNameArgs) (string, error)

	// GetServiceInfo returns information about the service.
	GetServiceInfo(
		ctx context.Context,
		args *api.GetServiceInfoArgs) (*api.GetServiceInfoReply, error)

	// GetVolumeMapping lists the block devices that are attached to the
	GetVolumeMapping(
		ctx context.Context,
		args *api.GetVolumeMappingArgs) ([]*api.BlockDevice, error)

	// GetInstance retrieves the local instance.
	GetInstance(
		ctx context.Context,
		args *api.GetInstanceArgs) (*api.Instance, error)

	// GetVolume returns all volumes for the instance based on either volumeID
	// or volumeName that are available to the instance.
	GetVolume(
		ctx context.Context,
		args *api.GetVolumeArgs) ([]*api.Volume, error)

	// GetVolumeAttach returns the attachment details based on volumeID or
	// volumeName where the volume is currently attached.
	GetVolumeAttach(
		ctx context.Context,
		args *api.GetVolumeAttachArgs) ([]*api.VolumeAttachment, error)

	// CreateSnapshot is a synch/async operation that returns snapshots that
	// have been performed based on supplying a snapshotName, source volumeID,
	// and optional description.
	CreateSnapshot(
		ctx context.Context,
		args *api.CreateSnapshotArgs) ([]*api.Snapshot, error)

	// GetSnapshot returns a list of snapshots for a volume based on volumeID,
	// snapshotID, or snapshotName.
	GetSnapshot(
		ctx context.Context,
		args *api.GetSnapshotArgs) ([]*api.Snapshot, error)

	// RemoveSnapshot will remove a snapshot based on the snapshotID.
	RemoveSnapshot(
		ctx context.Context,
		args *api.RemoveSnapshotArgs) error

	// CreateVolume is sync/async and will create an return a new/existing
	// Volume based on volumeID/snapshotID with a name of volumeName and a size
	// in GB.  Optionally based on the storage driver, a volumeType, IOPS, and
	// availabilityZone could be defined.
	CreateVolume(
		ctx context.Context,
		args *api.CreateVolumeArgs) (*api.Volume, error)

	// RemoveVolume will remove a volume based on volumeID.
	RemoveVolume(
		ctx context.Context,
		args *api.RemoveVolumeArgs) error

	// AttachVolume returns a list of VolumeAttachments is sync/async that will
	// attach a volume to an instance based on volumeID and ctx.
	AttachVolume(
		ctx context.Context,
		args *api.AttachVolumeArgs) ([]*api.VolumeAttachment, error)

	// DetachVolume is sync/async that will detach the volumeID from the local
	// instance or the ctx.
	DetachVolume(
		ctx context.Context,
		args *api.DetachVolumeArgs) error

	// CopySnapshot is a sync/async and returns a snapshot that will copy a
	// snapshot based on volumeID/snapshotID/snapshotName and create a new
	// snapshot of desinationSnapshotName in the destinationRegion location.
	CopySnapshot(
		ctx context.Context,
		args *api.CopySnapshotArgs) (*api.Snapshot, error)

	// GetClientTool gets the client tool provided by the driver. This tool is
	// executed on the client-side of the connection in order to discover
	// information only available to the client, such as the client's instance
	// ID or a local device map.
	//
	// The client tool is returned as a byte array that's either a binary file
	// or a unicode-encoded, plain-text script file. Use the file extension
	// of the client tool's file name to determine the file type.
	GetClientTool(
		ctx context.Context,
		args *api.GetClientToolArgs) (*api.ClientTool, error)
}

type client struct {
	config           gofig.Config
	url              string
	instanceID       *api.InstanceID
	instanceIDJSON   string
	instanceIDBase64 string
	clientToolPath   string
	logRequests      bool
	logResponses     bool
	nextDevice       *api.NextAvailableDeviceName
}

// Dial opens a connection to a remote libStorage serice and returns the client
// that can be used to communicate with said endpoint.
//
// If the config parameter is nil a default instance is created. The
// function dials the libStorage service specified by the configuration
// property libstorage.host.
func Dial(
	ctx context.Context,
	config gofig.Config) (Client, error) {

	c := &client{config: config}
	c.logRequests = c.config.GetBool(
		"libstorage.client.http.logging.logrequest")
	c.logResponses = c.config.GetBool(
		"libstorage.client.http.logging.logresponse")

	host := config.GetString("libstorage.host")
	if host == "" {
		return nil, goof.New("libstorage.host is required")
	}
	log.WithField("host", host).Debug("got libStorage host")

	serverName := config.GetString("libstorage.server")
	if serverName == "" {
		return nil, goof.New("libstorage.server is required")
	}
	log.WithField("server", serverName).Debug("got libStorage server name")

	if ctx == nil {
		log.Debug("created empty context for dialer")
		ctx = context.Background()
	}

	netProto, laddr, err := gotil.ParseAddress(host)
	if err != nil {
		return nil, err
	}

	if !netProtoRx.MatchString(netProto) {
		return nil, goof.WithField("netProto", netProto, "tcp protocol only")
	}

	c.url = fmt.Sprintf("http://%s/libStorage/%s", laddr, serverName)
	log.WithField("url", c.url).Debug("got libStorage service URL")

	if err := c.initClientTool(ctx); err != nil {
		return nil, err
	}

	if err := c.initInstanceID(ctx); err != nil {
		return nil, err
	}

	if err := c.initNextAvailableDeviceName(ctx); err != nil {
		return nil, err
	}

	log.WithField("url", c.url).Debug("successfuly dialed libStorage service")
	return c, nil
}

func (c *client) initNextAvailableDeviceName(ctx context.Context) error {
	args := &api.GetNextAvailableDeviceNameArgs{}
	reply := &api.GetNextAvailableDeviceNameReply{}
	if err := c.post(
		ctx, "GetNextAvailableDeviceName", args, reply); err != nil {
		return err
	}
	c.nextDevice = reply.Next
	return nil
}

func (c *client) GetNextAvailableDeviceName(
	ctx context.Context,
	args *api.GetNextAvailableDeviceNameArgs) (string, error) {

	if c.nextDevice.Ignore {
		return "", nil
	}

	blockDevices, err := c.GetVolumeMapping(ctx, &api.GetVolumeMappingArgs{})
	if err != nil {
		return "", err
	}

	lettersInUse := map[string]bool{}

	var rx *regexp.Regexp
	if c.nextDevice.Pattern == "" {
		rx = regexp.MustCompile(
			fmt.Sprintf(`^/dev/%s([a-z])$`, c.nextDevice.Prefix))
	} else {
		rx = regexp.MustCompile(
			fmt.Sprintf(
				`^/dev/%s(%s)$`, c.nextDevice.Prefix, c.nextDevice.Pattern))
	}

	for _, d := range blockDevices {
		m := rx.FindStringSubmatch(d.DeviceName)
		if len(m) > 0 {
			lettersInUse[m[1]] = true
		}
	}

	localDevices, err := c.getLocalDevices(c.nextDevice.Prefix)
	if err != nil {
		return "", err
	}

	for _, d := range localDevices {
		m := rx.FindStringSubmatch(d)
		if len(m) > 0 {
			lettersInUse[m[1]] = true
		}
	}

	for _, l := range letters {
		if !lettersInUse[l] {
			n := fmt.Sprintf("/dev/%s%s", c.nextDevice.Prefix, l)
			log.WithField("name", n).Debug("got next available device name")
			return n, nil
		}
	}

	return "", nil
}

func (c *client) GetInstanceID(ctx context.Context) (*api.InstanceID, error) {
	return c.instanceID, nil
}

func (c *client) GetVolumeMapping(
	ctx context.Context,
	args *api.GetVolumeMappingArgs) ([]*api.BlockDevice, error) {
	reply := &api.GetVolumeMappingReply{}
	if err := c.post(ctx, "GetVolumeMapping", args, reply); err != nil {
		return nil, err
	}
	return reply.BlockDevices, nil
}

func (c *client) GetInstance(
	ctx context.Context,
	args *api.GetInstanceArgs) (*api.Instance, error) {
	reply := &api.GetInstanceReply{}
	if err := c.post(ctx, "GetInstance", args, reply); err != nil {
		return nil, err
	}
	return reply.Instance, nil
}

func (c *client) GetVolume(
	ctx context.Context,
	args *api.GetVolumeArgs) ([]*api.Volume, error) {
	reply := &api.GetVolumeReply{}
	if err := c.post(ctx, "GetVolume", args, reply); err != nil {
		return nil, err
	}
	return reply.Volumes, nil
}

func (c *client) GetVolumeAttach(
	ctx context.Context,
	args *api.GetVolumeAttachArgs) ([]*api.VolumeAttachment, error) {
	reply := &api.GetVolumeAttachReply{}
	if err := c.post(ctx, "GetVolumeAttach", args, reply); err != nil {
		return nil, err
	}
	return reply.Attachments, nil
}

func (c *client) CreateSnapshot(
	ctx context.Context,
	args *api.CreateSnapshotArgs) ([]*api.Snapshot, error) {

	reply := &api.CreateSnapshotReply{}
	if err := c.post(ctx, "CreateSnapshot", args, reply); err != nil {
		return nil, err
	}
	return reply.Snapshots, nil
}

func (c *client) GetSnapshot(
	ctx context.Context,
	args *api.GetSnapshotArgs) ([]*api.Snapshot, error) {
	reply := &api.GetSnapshotReply{}
	if err := c.post(ctx, "GetSnapshot", args, reply); err != nil {
		return nil, err
	}
	return reply.Snapshots, nil
}

func (c *client) RemoveSnapshot(
	ctx context.Context,
	args *api.RemoveSnapshotArgs) error {
	reply := &api.RemoveSnapshotReply{}
	if err := c.post(ctx, "RemoveSnapshot", args, reply); err != nil {
		return err
	}
	return nil
}

func (c *client) CreateVolume(
	ctx context.Context,
	args *api.CreateVolumeArgs) (*api.Volume, error) {
	reply := &api.CreateVolumeReply{}
	if err := c.post(ctx, "CreateVolume", args, reply); err != nil {
		return nil, err
	}
	return reply.Volume, nil
}

func (c *client) RemoveVolume(
	ctx context.Context,
	args *api.RemoveVolumeArgs) error {
	reply := &api.RemoveVolumeReply{}
	if err := c.post(ctx, "RemoveVolume", args, reply); err != nil {
		return err
	}
	return nil
}

func (c *client) AttachVolume(
	ctx context.Context,
	args *api.AttachVolumeArgs) ([]*api.VolumeAttachment, error) {
	if args.Required.NextDeviceName == "" {

	}
	reply := &api.AttachVolumeReply{}
	if err := c.post(ctx, "AttachVolume", args, reply); err != nil {
		return nil, err
	}
	return reply.Attachments, nil
}

func (c *client) DetachVolume(
	ctx context.Context,
	args *api.DetachVolumeArgs) error {
	reply := &api.DetachVolumeReply{}
	if err := c.post(ctx, "DetachVolume", args, reply); err != nil {
		return err
	}
	return nil
}

func (c *client) CopySnapshot(
	ctx context.Context,
	args *api.CopySnapshotArgs) (*api.Snapshot, error) {
	reply := &api.CopySnapshotReply{}
	if err := c.post(ctx, "CopySnapshot", args, reply); err != nil {
		return nil, err
	}
	return reply.Snapshot, nil
}

func (c *client) GetServiceInfo(
	ctx context.Context,
	args *api.GetServiceInfoArgs) (*api.GetServiceInfoReply, error) {
	reply := &api.GetServiceInfoReply{}
	if err := c.post(ctx, "GetServiceInfo", args, reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (c *client) GetClientTool(
	ctx context.Context,
	args *api.GetClientToolArgs) (*api.ClientTool, error) {
	reply := &api.GetClientToolReply{}
	if err := c.post(ctx, "GetClientTool", args, reply); err != nil {
		return nil, err
	}
	return reply.ClientTool, nil
}

func (c *client) post(
	ctx context.Context,
	method string,
	args interface{},
	reply interface{}) error {

	method = fmt.Sprintf("libStorage.%s", method)

	log.WithFields(log.Fields{
		"url":    c.url,
		"method": method,
	}).Debug("begin libStorage method")

	m, err := encReq(method, args)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewReader(m))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if c.instanceIDBase64 != "" {
		req.Header.Set("libStorage-InstanceID", c.instanceIDBase64)
	}

	c.logRequest(log.StandardLogger().Writer(), req)

	res, err := ctxhttp.Do(ctx, nil, req)
	if err != nil {
		return err
	}

	c.logResponse(log.StandardLogger().Writer(), res)

	if err := gjson.DecodeClientResponse(res.Body, reply); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"url":    c.url,
		"method": method,
	}).Debug("end libStorage method")

	return nil
}

func encReq(method string, args interface{}) ([]byte, error) {
	enc, err := gjson.EncodeClientRequest(method, args)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decRes(body io.Reader, reply interface{}) error {
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	log.WithField("json", string(buf)).Debug("response body")
	b := bufio.NewReader(bytes.NewBuffer(buf))
	if err := gjson.DecodeClientResponse(b, reply); err != nil {
		return err
	}
	return nil
}

func (c *client) logRequest(
	w io.Writer,
	req *http.Request) {

	if !c.logRequests {
		return
	}

	fmt.Fprintln(w, "")
	fmt.Fprint(w, "    -------------------------- ")
	fmt.Fprint(w, "HTTP REQUEST (CLIENT)")
	fmt.Fprintln(w, " -------------------------")

	buf, err := httputil.DumpRequest(req, true)
	if err != nil {
		return
	}

	gotil.WriteIndented(w, buf)
}

func (c *client) logResponse(
	w io.Writer,
	res *http.Response) {

	if !c.logResponses {
		return
	}

	fmt.Fprint(w, "    -------------------------- ")
	fmt.Fprint(w, "HTTP RESPONSE (CLIENT)")
	fmt.Fprintln(w, " -------------------------")

	buf, err := httputil.DumpResponse(res, true)
	if err != nil {
		return
	}

	gotil.WriteIndented(w, buf)
}

func (c *client) execClientToolInstanceID(
	ctx context.Context) ([]byte, error) {
	log.WithField("path", c.clientToolPath).Debug(
		"executing client tool GetInstanceID")
	return exec.Command(c.clientToolPath, "GetInstanceID").Output()
}

func (c *client) execClientToolNextDevID(ctx context.Context) ([]byte, error) {

	log.WithField("path", c.clientToolPath).Debug(
		"executing client tool GetNextAvailableDeviceName")

	bds, err := c.GetVolumeMapping(ctx, &api.GetVolumeMappingArgs{})
	if err != nil {
		return nil, err
	}

	bdsJSON, err := json.MarshalIndent(bds, "", "  ")
	if err != nil {
		return nil, err
	}

	env := os.Environ()
	env = append(env, fmt.Sprintf("BLOCK_DEVICES_JSON=%s", string(bdsJSON)))

	cmd := exec.Command(c.clientToolPath, "GetNextAvailableDeviceName")
	cmd.Env = env
	return cmd.Output()
}

func (c *client) initClientTool(ctx context.Context) error {

	args := &api.GetClientToolArgs{
		Optional: api.GetClientToolArgsOptional{
			OmitBinary: true,
		},
	}
	clientTool, err := c.GetClientTool(ctx, args)
	if err != nil {
		return err
	}

	if gotil.FileExistsInPath(clientTool.Name) {
		c.clientToolPath = clientTool.Name
		log.WithField("path", c.clientToolPath).Debug(
			"client tool exists in path")
		return nil
	}

	args.Optional.OmitBinary = false
	clientTool, err = c.GetClientTool(ctx, args)
	if err != nil {
		return err
	}

	toolDir := c.config.GetString("libstorage.client.tooldir")
	toolPath := fmt.Sprintf("%s/%s", toolDir, clientTool.Name)
	log.WithField("path", toolPath).Debug("writing client tool")

	if err := ioutil.WriteFile(
		toolPath, clientTool.Data, 0755); err != nil {
		return err
	}

	c.clientToolPath = toolPath
	log.WithField("path", c.clientToolPath).Debug(
		"client tool path initialized")
	return nil
}

func (c *client) initInstanceID(ctx context.Context) error {

	log.Debug("begin get libStorage instanceID")

	iidJSON, err := c.execClientToolInstanceID(ctx)
	if err != nil {
		return err
	}

	err = json.Unmarshal(iidJSON, &c.instanceID)
	if err != nil {
		return err
	}
	c.instanceIDJSON = string(iidJSON)
	c.instanceIDBase64 = base64.URLEncoding.EncodeToString(iidJSON)

	log.WithFields(log.Fields{
		"instanceID":       c.instanceID,
		"instanceIDJSON":   c.instanceIDJSON,
		"instanceIDBase64": c.instanceIDBase64,
	}).Debug("end get libStorage instanceID")
	return nil
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
