package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/gorilla/rpc/json"

	"github.com/emccode/libstorage/api"
	"github.com/emccode/libstorage/model"
	"github.com/emccode/libstorage/util"
)

var (
	netProtoRx = regexp.MustCompile("(?i)tcp")
)

// Client is the reference implementation of the libStorage client.
type Client interface {

	// InitDrivers initializes the drivers on the server.
	//
	// This function can be safely invoked multiple times by multiple clients.
	InitDrivers() (
		registredDriverNames,
		initializedDriverNames []string,
		err error)

	// GetInstanceID gets the instance ID of this host.
	GetInstanceID() (*model.InstanceID, error)

	// GetVolumeMapping lists the block devices that are attached to the instance.
	GetVolumeMapping() ([]*model.BlockDevice, error)

	// GetInstance retrieves the local instance.
	GetInstance() (*model.Instance, error)

	// GetVolume returns all volumes for the instance based on either volumeID
	// or volumeName that are available to the instance.
	GetVolume(volumeID, volumeName string) ([]*model.Volume, error)

	// GetVolumeAttach returns the attachment details based on volumeID or
	// volumeName where the volume is currently attached.
	GetVolumeAttach(volumeID string) ([]*model.VolumeAttachment, error)

	// CreateSnapshot is a synch/async operation that returns snapshots that
	// have been performed based on supplying a snapshotName, source volumeID,
	// and optional description.
	CreateSnapshot(
		snapshotName,
		volumeID,
		description string) ([]*model.Snapshot, error)

	// GetSnapshot returns a list of snapshots for a volume based on volumeID,
	// snapshotID, or snapshotName.
	GetSnapshot(
		volumeID,
		snapshotID,
		snapshotName string) ([]*model.Snapshot, error)

	// RemoveSnapshot will remove a snapshot based on the snapshotID.
	RemoveSnapshot(snapshotID string) error

	// CreateVolume is sync/async and will create an return a new/existing
	// Volume based on volumeID/snapshotID with a name of volumeName and a size
	// in GB.  Optionally based on the storage driver, a volumeType, IOPS, and
	// availabilityZone could be defined.
	CreateVolume(
		volumeName,
		volumeID,
		snapshotID,
		volumeType string,
		IOPS,
		size int64,
		availabilityZone string) (*model.Volume, error)

	// RemoveVolume will remove a volume based on volumeID.
	RemoveVolume(volumeID string) error

	// GetDeviceNextAvailable gets the next device available on this host.
	GetDeviceNextAvailable() (string, error)

	// AttachVolume returns a list of VolumeAttachments is sync/async that will
	// attach a volume to an instance based on volumeID and instanceID.
	AttachVolume(
		nextDeviceName,
		volumeID string) ([]*model.VolumeAttachment, error)

	// DetachVolume is sync/async that will detach the volumeID from the local
	// instance or the instanceID.
	DetachVolume(volumeID string) error

	// CopySnapshot is a sync/async and returns a snapshot that will copy a
	// snapshot based on volumeID/snapshotID/snapshotName and create a new
	// snapshot of desinationSnapshotName in the destinationRegion location.
	CopySnapshot(
		volumeID,
		snapshotID,
		snapshotName,
		destinationSnapshotName,
		destinationRegion string) (*model.Snapshot, error)

	// GetRegisteredDriverNames gets the names of the registered drivers.
	GetRegisteredDriverNames() []string

	// GetInitializedDriverNames gets the names of the initialized drivers.
	GetInitializedDriverNames() []string
}

// Client is the reference client implementation for libStorage.
type client struct {
	config     *gofig.Config
	url        string
	instanceID *model.InstanceID
}

// Dial opens a connection to a remote libStorage serice and returns the client
// that can be used to communicate with said endpoint.
//
// If the config parameter is nil a default instance is created. The
// function dials the libStorage service specified by the configuration
// property libstorage.host.
func Dial(config *gofig.Config) (Client, error) {
	c := &client{
		config: config,
	}
	host := config.GetString("libstorage.host")
	if host == "" {
		return nil, goof.New("libstorage.host is required")
	}
	log.WithField("host", host).Debug("got libStorage host")

	netProto, laddr, err := util.ParseAddress(host)
	if err != nil {
		return nil, err
	}

	if !netProtoRx.MatchString(netProto) {
		return nil, goof.WithField("netProto", netProto, "tcp protocol only")
	}

	c.url = fmt.Sprintf("http://%s/libStorage", laddr)
	log.WithField("url", c.url).Debug("got libStorage service URL")

	log.Debug("begin get libStorage instanceID")
	iid, err := c.GetInstanceID()
	if err != nil {
		return nil, err
	}
	c.instanceID = iid
	log.WithField("instanceID", iid).Debug("end get libStorage instanceID")

	log.WithField("url", c.url).Debug("successfuly dialed libStorage service")
	return c, nil
}

// InitDrivers initializes the drivers on the server.
//
// This function can be safely invoked multiple times by multiple clients.
func (c *client) InitDrivers() (
	registredDriverNames,
	initializedDriverNames []string,
	err error) {

	args := api.InitDriversArgs{
		Config: c.config,
	}
	reply := api.InitDriversReply{}
	if err := c.post("libStorage.InitDrivers", &args, &reply); err != nil {
		return nil, nil, err
	}
	log.WithFields(log.Fields{
		"registeredDriverNames":  reply.RegisteredDriverNames,
		"initializedDriverNames": reply.InitializedDriverNames,
	}).Debug("libStorage.Client.InitDrivers")
	return reply.RegisteredDriverNames, reply.InitializedDriverNames, nil
}

// GetInstanceID gets the instance ID of this host.
func (c *client) GetInstanceID() (*model.InstanceID, error) {
	return &model.InstanceID{
		ID: "TODO",
	}, nil
}

// GetVolumeMapping lists the block devices that are attached to the instance.
func (c *client) GetVolumeMapping() ([]*model.BlockDevice, error) {

	args := api.GetVolumeMappingArgs{
		InstanceID: c.instanceID,
	}
	reply := api.GetVolumeMappingReply{}
	if err := c.post("libStorage.GetVolumeMapping", &args, &reply); err != nil {
		return nil, err
	}
	log.WithField(
		"len(blockDevices)", len(reply.BlockDevices)).Debug(
		"libStorage.Client.GetVolumeMapping")
	return reply.BlockDevices, nil
}

// GetInstance retrieves the local instance.
func (c *client) GetInstance() (*model.Instance, error) {

	args := api.GetInstanceArgs{
		InstanceID: c.instanceID,
	}
	reply := api.GetInstanceReply{}
	if err := c.post("libStorage.GetInstance", &args, &reply); err != nil {
		return nil, err
	}
	return reply.Instance, nil
}

// GetVolume returns all volumes for the instance based on either volumeID
// or volumeName that are available to the instance.
func (c *client) GetVolume(
	volumeID,
	volumeName string) ([]*model.Volume, error) {

	args := api.GetVolumeArgs{
		InstanceID: c.instanceID,
		VolumeID:   volumeID,
		VolumeName: volumeName,
	}
	reply := api.GetVolumeReply{}
	if err := c.post("libStorage.GetVolume", &args, &reply); err != nil {
		return nil, err
	}
	log.WithField(
		"len(volumes)", len(reply.Volumes)).Debug(
		"libStorage.Client.GetVolume")
	return reply.Volumes, nil
}

func (c *client) GetVolumeAttach(
	volumeID string) ([]*model.VolumeAttachment, error) {

	args := api.GetVolumeAttachArgs{
		InstanceID: c.instanceID,
		VolumeID:   volumeID,
	}
	reply := api.GetVolumeAttachReply{}
	if err := c.post("libStorage.GetVolumeAttach", &args, &reply); err != nil {
		return nil, err
	}
	return reply.Attachments, nil
}

func (c *client) CreateSnapshot(
	snapshotName,
	volumeID,
	description string) ([]*model.Snapshot, error) {

	args := api.CreateSnapshotArgs{
		InstanceID:   c.instanceID,
		SnapshotName: snapshotName,
		VolumeID:     volumeID,
		Description:  description,
	}
	reply := api.CreateSnapshotReply{}
	if err := c.post("libStorage.CreateSnapshot", &args, &reply); err != nil {
		return nil, err
	}
	return reply.Snapshots, nil
}

func (c *client) GetSnapshot(
	volumeID,
	snapshotID,
	snapshotName string) ([]*model.Snapshot, error) {

	args := api.GetSnapshotArgs{
		InstanceID:   c.instanceID,
		VolumeID:     volumeID,
		SnapshotID:   snapshotID,
		SnapshotName: snapshotName,
	}
	reply := api.GetSnapshotReply{}
	if err := c.post("libStorage.GetSnapshot", &args, &reply); err != nil {
		return nil, err
	}
	log.WithField(
		"len(snapshots)", len(reply.Snapshots)).Debug(
		"libStorage.Client.GetSnapshot")
	return reply.Snapshots, nil
}

func (c *client) RemoveSnapshot(snapshotID string) error {

	args := api.RemoveSnapshotArgs{
		InstanceID: c.instanceID,
		SnapshotID: snapshotID,
	}
	reply := api.RemoveSnapshotReply{}
	if err := c.post("libStorage.RemoveSnapshot", &args, &reply); err != nil {
		return err
	}
	return nil
}

func (c *client) CreateVolume(
	volumeName,
	volumeID,
	snapshotID,
	volumeType string,
	IOPS,
	size int64,
	availabilityZone string) (*model.Volume, error) {

	args := api.CreateVolumeArgs{
		InstanceID:       c.instanceID,
		VolumeName:       volumeName,
		VolumeID:         volumeID,
		SnapshotID:       snapshotID,
		VolumeType:       volumeType,
		IOPS:             IOPS,
		Size:             size,
		AvailabilityZone: availabilityZone,
	}
	reply := api.CreateVolumeReply{}
	if err := c.post("libStorage.CreateVolume", &args, &reply); err != nil {
		return nil, err
	}
	return reply.Volume, nil
}

func (c *client) RemoveVolume(volumeID string) error {

	args := api.RemoveVolumeArgs{
		InstanceID: c.instanceID,
		VolumeID:   volumeID,
	}
	reply := api.RemoveVolumeReply{}
	if err := c.post("libStorage.RemoveVolume", &args, &reply); err != nil {
		return err
	}
	return nil
}

func (c *client) GetDeviceNextAvailable() (string, error) {
	return "", nil
}

func (c *client) AttachVolume(
	nextDeviceName,
	volumeID string) ([]*model.VolumeAttachment, error) {

	args := api.AttachVolumeArgs{
		InstanceID:     c.instanceID,
		NextDeviceName: nextDeviceName,
		VolumeID:       volumeID,
	}
	reply := api.AttachVolumeReply{}
	if err := c.post("libStorage.AttachVolume", &args, &reply); err != nil {
		return nil, err
	}
	return reply.Attachments, nil
}

func (c *client) DetachVolume(volumeID string) error {

	args := api.DetachVolumeArgs{
		InstanceID: c.instanceID,
		VolumeID:   volumeID,
	}
	reply := api.DetachVolumeReply{}
	if err := c.post("libStorage.DetachVolume", &args, &reply); err != nil {
		return err
	}
	return nil
}

func (c *client) CopySnapshot(
	volumeID,
	snapshotID,
	snapshotName,
	destinationSnapshotName,
	destinationRegion string) (*model.Snapshot, error) {
	args := api.CopySnapshotArgs{
		InstanceID:              c.instanceID,
		VolumeID:                volumeID,
		SnapshotID:              snapshotID,
		SnapshotName:            snapshotName,
		DestinationSnapshotName: destinationSnapshotName,
		DestinationRegion:       destinationRegion,
	}
	reply := api.CopySnapshotReply{}
	if err := c.post("libStorage.CopySnapshot", &args, &reply); err != nil {
		return nil, err
	}
	return reply.Snapshot, nil
}

// GetRegisteredDriverNames gets the names of the registered drivers.
func (c *client) GetRegisteredDriverNames() []string {
	args := api.GetDriverNamesArgs{}
	reply := api.GetDriverNamesReply{}
	if err := c.post(
		"libStorage.GetRegisteredDriverNames", &args, &reply); err != nil {
		log.Error(err)
		return nil
	}
	return reply.DriverNames
}

// GetInitializedDriverNames gets the names of the initialized drivers.
func (c *client) GetInitializedDriverNames() []string {
	args := api.GetDriverNamesArgs{}
	reply := api.GetDriverNamesReply{}
	if err := c.post(
		"libStorage.GetInitializedDriverNames", &args, &reply); err != nil {
		log.Error(err)
		return nil
	}
	return reply.DriverNames
}

func (c *client) post(
	method string,
	args interface{},
	reply interface{}) error {

	log.WithFields(log.Fields{
		"url":    c.url,
		"method": method,
	}).Debug("posting to libstorage")

	m, err := encReq(method, args)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(m))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	hc := &http.Client{}
	res, err := hc.Do(req)
	if err != nil {
		return err
	}
	if err := json.DecodeClientResponse(res.Body, reply); err != nil {
		return err
	}
	return nil
}

func encReq(method string, args interface{}) ([]byte, error) {
	enc, err := json.EncodeClientRequest(method, args)
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
	if err := json.DecodeClientResponse(b, reply); err != nil {
		return err
	}
	return nil
}
