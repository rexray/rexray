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
	gjson "github.com/gorilla/rpc/json"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"

	"github.com/emccode/libstorage/api"
	"github.com/emccode/libstorage/util"
)

var (
	netProtoRx = regexp.MustCompile("(?i)tcp")
)

// Client is the reference implementation of the libStorage client.
type Client interface {

	// GetInstanceID gets the instance ID.
	GetInstanceID(ctx context.Context) (*api.InstanceID, error)

	// GetNextAvailableDeviceName gets the name of the next available device.
	GetNextAvailableDeviceName(ctx context.Context) (string, error)

	// GetRegisteredDriverNames gets the names of the registered drivers.
	GetRegisteredDriverNames(
		ctx context.Context,
		args *api.GetDriverNamesArgs) ([]string, error)

	// GetInitializedDriverNames gets the names of the initialized drivers.
	GetInitializedDriverNames(
		ctx context.Context,
		args *api.GetDriverNamesArgs) ([]string, error)

	// GetVolumeMapping lists the block devices that are attached to the
	// instance.
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
	// attach a volume to an instance based on volumeID and instanceID.
	AttachVolume(
		ctx context.Context,
		args *api.AttachVolumeArgs) ([]*api.VolumeAttachment, error)

	// DetachVolume is sync/async that will detach the volumeID from the local
	// instance or the instanceID.
	DetachVolume(
		ctx context.Context,
		args *api.DetachVolumeArgs) error

	// CopySnapshot is a sync/async and returns a snapshot that will copy a
	// snapshot based on volumeID/snapshotID/snapshotName and create a new
	// snapshot of desinationSnapshotName in the destinationRegion location.
	CopySnapshot(
		ctx context.Context,
		args *api.CopySnapshotArgs) (*api.Snapshot, error)

	// GetClientToolName gets the file name of the tool this driver provides
	// to be executed on the client-side in order to discover a client's
	// instance ID and next, available device name.
	//
	// Use the function GetClientTool to get the actual tool.
	GetClientToolName(
		ctx context.Context,
		args *api.GetClientToolNameArgs) (string, error)

	// GetClientTool gets the file  for the tool this driver provides
	// to be executed on the client-side in order to discover a client's
	// instance ID and next, available device name.
	//
	// This function returns a byte array that will be either a binary file
	// or a unicode-encoded, plain-text script file. Use the file extension
	// of the client tool's file name to determine the file type.
	//
	// The function GetClientToolName can be used to get the file name.
	GetClientTool(
		ctx context.Context,
		args *api.GetClientToolArgs) ([]byte, error)
}

// Client is the reference client implementation for libStorage.
type client struct {
	config           gofig.Config
	url              string
	instanceID       *api.InstanceID
	instanceIDJSON   string
	instanceIDBase64 string
	clientToolPath   string
	logRequests      bool
	logResponses     bool
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

	netProto, laddr, err := util.ParseAddress(host)
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

	log.WithField("url", c.url).Debug("successfuly dialed libStorage service")
	return c, nil
}

func (c *client) initClientTool(ctx context.Context) error {

	toolName, err := c.GetClientToolName(ctx, &api.GetClientToolNameArgs{})
	if err != nil {
		return err
	}

	if util.FileExistsInPath(toolName) {
		c.clientToolPath = toolName
		log.WithField("path", c.clientToolPath).Debug(
			"client tool exists in path")
		return nil
	}

	toolBuf, err := c.GetClientTool(ctx, &api.GetClientToolArgs{})
	if err != nil {
		return err
	}

	toolDir := c.config.GetString("libstorage.client.tooldir")
	toolPath := fmt.Sprintf("%s/%s", toolDir, toolName)
	log.WithField("path", toolPath).Debug("writing client tool")

	if err := ioutil.WriteFile(toolPath, toolBuf, 0755); err != nil {
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

func (c *client) GetNextAvailableDeviceName(
	ctx context.Context) (string, error) {
	out, err := c.execClientToolNextDevID(ctx)
	if err != nil {
		return "", err
	}
	return string(out), nil
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

// GetInstanceID gets the instance ID.
func (c *client) GetInstanceID(ctx context.Context) (*api.InstanceID, error) {
	return c.instanceID, nil
}

// GetVolumeMapping lists the block devices that are attached to the instance.
func (c *client) GetVolumeMapping(
	ctx context.Context,
	args *api.GetVolumeMappingArgs) ([]*api.BlockDevice, error) {
	reply := &api.GetVolumeMappingReply{}
	if err := c.post(ctx, "GetVolumeMapping", args, reply); err != nil {
		return nil, err
	}
	return reply.BlockDevices, nil
}

// GetInstance retrieves the local instance.
func (c *client) GetInstance(
	ctx context.Context,
	args *api.GetInstanceArgs) (*api.Instance, error) {
	reply := &api.GetInstanceReply{}
	if err := c.post(ctx, "GetInstance", args, reply); err != nil {
		return nil, err
	}
	return reply.Instance, nil
}

// GetVolume returns all volumes for the instance based on either volumeID
// or volumeName that are available to the instance.
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

// GetRegisteredDriverNames gets the names of the registered drivers.
func (c *client) GetRegisteredDriverNames(
	ctx context.Context,
	args *api.GetDriverNamesArgs) ([]string, error) {
	reply := &api.GetDriverNamesReply{}
	if err := c.post(ctx, "GetRegisteredDriverNames", args, reply); err != nil {
		return nil, err
	}
	return reply.DriverNames, nil
}

// GetInitializedDriverNames gets the names of the initialized drivers.
func (c *client) GetInitializedDriverNames(
	ctx context.Context,
	args *api.GetDriverNamesArgs) ([]string, error) {
	reply := &api.GetDriverNamesReply{}
	if err := c.post(ctx, "GetInitializedDriverNames", args, reply); err != nil {
		return nil, err
	}
	return reply.DriverNames, nil
}

// GetClientToolName gets the file name of the tool this driver provides
// to be executed on the client-side in order to discover a client's
// instance ID and next, available device name.
//
// Use the function GetClientTool to get the actual tool.
func (c *client) GetClientToolName(
	ctx context.Context,
	args *api.GetClientToolNameArgs) (string, error) {
	reply := &api.GetClientToolNameReply{}
	if err := c.post(ctx, "GetClientToolName", args, reply); err != nil {
		return "", err
	}
	return reply.ClientToolName, nil
}

// GetClientTool gets the file  for the tool this driver provides
// to be executed on the client-side in order to discover a client's
// instance ID and next, available device name.
//
// This function returns a byte array that will be either a binary file
// or a unicode-encoded, plain-text script file. Use the file extension
// of the client tool's file name to determine the file type.
//
// The function GetClientToolName can be used to get the file name.
func (c *client) GetClientTool(
	ctx context.Context,
	args *api.GetClientToolArgs) ([]byte, error) {
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

	util.WriteIndented(w, buf)
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

	util.WriteIndented(w, buf)
}
