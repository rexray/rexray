package service

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	gcontext "github.com/gorilla/context"

	"github.com/emccode/libstorage/api"
	"github.com/emccode/libstorage/context"
	"github.com/emccode/libstorage/driver"
	"github.com/emccode/libstorage/service/server"
	"github.com/emccode/libstorage/util"
)

var (
	errNoStorageDetected          = goof.New("no storage detected")
	errDriverBlockDeviceDiscovery = goof.New("no block devices discovered")

	driverCtors     = map[string]driver.NewDriver{}
	driverCtorsLock = &sync.RWMutex{}
)

type service struct {
	config              gofig.Config
	constructedDrivers  map[string]driver.Driver
	initializedDrivers  map[string]driver.Driver
	driversInitialized  bool
	initializingDrivers *sync.Mutex
}

// RegisterDriver is used by drivers to indicate their availability to
// libstorage.
func RegisterDriver(driverName string, ctor driver.NewDriver) {
	driverCtorsLock.Lock()
	defer driverCtorsLock.Unlock()
	driverCtors[driverName] = ctor
}

// Serve starts the reference implementation of a server hosting an
// HTTP/JSON service that implements the libStorage API endpoint.
//
// If the config parameter is nil a default instance is created. The
// libStorage service is served at the address specified by the configuration
// property libstorage.host.
//
// If libstorage.host specifies a tcp network and the provided port is 0,
// aka "select any avaialble port", then after this function exists
// successfully the libstorage.host property is updated to reflect the port
// being used. For example, if the original value is tcp://127.0.0.1:0 and
// the operating system (OS) selects port 3756, then the libstorage.host
// property is updated to tcp://127.0.0.1:3756.
func Serve(config gofig.Config) error {

	si := map[string]*server.ServiceInfo{}

	servers := config.Get("libstorage.service.servers")
	serversMap := servers.(map[string]interface{})
	log.WithField("count", len(serversMap)).Debug("got servers map")

	for name := range serversMap {
		log.WithField("name", name).Debug("processing server config")

		scope := fmt.Sprintf("libstorage.service.servers.%s", name)
		log.WithField("scope", scope).Debug("getting scoped config for server")
		sc := config.Scope(scope)

		service, err := newService(sc)
		if err != nil {
			return err
		}

		si[name] = &server.ServiceInfo{
			Name:    name,
			Service: service,
			Config:  sc,
		}
	}

	if err := server.Serve(si, config); err != nil {
		return err
	}

	return nil
}

// GetRegisteredDriverNames returns the names of the registered drivers.
func (s *service) GetRegisteredDriverNames(
	req *http.Request,
	args *api.GetDriverNamesArgs,
	reply *api.GetDriverNamesReply) error {
	reply.DriverNames = getDriverNames(s.constructedDrivers)
	log.WithField(
		"driverNames", reply.DriverNames).Debug(
		"libStorage.Server.GetRegisteredDriverNames")
	return nil
}

// GetRegisteredDriverNames returns the names of the initialized drivers.
func (s *service) GetInitializedDriverNames(
	req *http.Request,
	args *api.GetDriverNamesArgs,
	reply *api.GetDriverNamesReply) error {
	reply.DriverNames = getDriverNames(s.initializedDrivers)
	log.WithField(
		"driverNames", reply.DriverNames).Debug(
		"libStorage.Server.GetInitializedDriverNames")
	return nil
}

// GetVolumeMapping lists the block devices that are attached to the
func (s *service) GetVolumeMapping(
	req *http.Request,
	args *api.GetVolumeMappingArgs,
	reply *api.GetVolumeMappingReply) error {

	for _, d := range s.initializedDrivers {
		bds, err := d.GetVolumeMapping(s.ctx(req), args)
		if err != nil {
			return errDriverBlockDeviceDiscovery
		}

		for _, bd := range bds {
			reply.BlockDevices = append(reply.BlockDevices, bd)
		}
	}

	if len(reply.BlockDevices) == 0 {
		return errNoStorageDetected
	}

	log.WithField(
		"len(blockDevices)", len(reply.BlockDevices)).Debug(
		"libStorage.Server.GetVolumeMapping")

	return nil
}

// GetInstance retrieves the local instance.
func (s *service) GetInstance(
	req *http.Request,
	args *api.GetInstanceArgs,
	reply *api.GetInstanceReply) error {
	var err error
	for _, d := range s.initializedDrivers {
		if reply.Instance, err = d.GetInstance(s.ctx(req), args); err != nil {
			return err
		}
		return nil
	}
	return errNoStorageDetected
}

// GetVolume returns all volumes for the instance based on either volumeID
// or volumeName that are available to the instance.
func (s *service) GetVolume(
	req *http.Request,
	args *api.GetVolumeArgs,
	reply *api.GetVolumeReply) error {
	var err error

	ctx := s.ctx(req)
	unprofile := profile(ctx, &args.Optional.VolumeName)
	defer unprofile()

	for _, d := range s.initializedDrivers {
		if reply.Volumes, err = d.GetVolume(ctx, args); err != nil {
			return err
		}
		if reply.Volumes != nil &&
			!s.config.GetBool("libstorage.profiles.client") {
			for _, v := range reply.Volumes {
				np := strings.Split(v.Name, "-")
				if len(np) == 1 {
					v.Name = np[0]
				} else {
					v.Name = np[1]
				}
			}
		}
		log.WithField(
			"len(volumes)", len(reply.Volumes)).Debug(
			"libStorage.Server.GetVolume")
		return nil
	}
	return errNoStorageDetected
}

// GetVolumeAttach returns the attachment details based on volumeID or
// volumeName where the volume is currently attached.
func (s *service) GetVolumeAttach(
	req *http.Request,
	args *api.GetVolumeAttachArgs,
	reply *api.GetVolumeAttachReply) error {
	var err error
	for _, d := range s.initializedDrivers {
		if reply.Attachments, err = d.GetVolumeAttach(
			s.ctx(req), args); err != nil {
			return err
		}
		return nil
	}
	return errNoStorageDetected
}

// CreateSnapshot is a synch/async operation that returns snapshots that
// have been performed based on supplying a snapshotName, source volumeID,
// and optional description.
func (s *service) CreateSnapshot(
	req *http.Request,
	args *api.CreateSnapshotArgs,
	reply *api.CreateSnapshotReply) error {

	var err error
	for _, d := range s.initializedDrivers {
		if reply.Snapshots, err = d.CreateSnapshot(
			s.ctx(req), args); err != nil {
			return err
		}
		return nil
	}
	return errNoStorageDetected
}

// GetSnapshot returns a list of snapshots for a volume based on volumeID,
// snapshotID, or snapshotName.
func (s *service) GetSnapshot(
	req *http.Request,
	args *api.GetSnapshotArgs,
	reply *api.GetSnapshotReply) error {
	var err error
	for _, d := range s.initializedDrivers {
		if reply.Snapshots, err = d.GetSnapshot(s.ctx(req), args); err != nil {
			return err
		}
		log.WithField(
			"len(snapshots)", len(reply.Snapshots)).Debug(
			"libStorage.Server.GetSnapshot")
		return nil
	}
	return errNoStorageDetected
}

// RemoveSnapshot will remove a snapshot based on the snapshotID.
func (s *service) RemoveSnapshot(
	req *http.Request,
	args *api.RemoveSnapshotArgs,
	reply *api.RemoveSnapshotReply) error {

	var err error
	for _, d := range s.initializedDrivers {
		if err = d.RemoveSnapshot(s.ctx(req), args); err != nil {
			return err
		}
		return nil
	}
	return errNoStorageDetected
}

// CreateVolume is sync/async and will create an return a new/existing
// Volume based on volumeID/snapshotID with a name of volumeName and a size
// in GB.  Optionally based on the storage driver, a volumeType, IOPS, and
// availabilityZone could be defined.
func (s *service) CreateVolume(
	req *http.Request,
	args *api.CreateVolumeArgs,
	reply *api.CreateVolumeReply) error {

	ctx := s.ctx(req)
	unprofile := profile(ctx, &args.Optional.VolumeName)
	defer unprofile()

	var err error
	for _, d := range s.initializedDrivers {
		if reply.Volume, err = d.CreateVolume(ctx, args); err != nil {
			return err
		}
		if !s.config.GetBool("libstorage.profiles.client") {
			np := strings.Split(reply.Volume.Name, "-")
			if len(np) == 1 {
				reply.Volume.Name = np[0]
			} else {
				reply.Volume.Name = np[1]
			}
		}
		return nil
	}
	return errNoStorageDetected
}

// RemoveVolume will remove a volume based on volumeID.
func (s *service) RemoveVolume(
	req *http.Request,
	args *api.RemoveVolumeArgs,
	reply *api.RemoveVolumeReply) error {

	var err error
	for _, d := range s.initializedDrivers {
		if err = d.RemoveVolume(s.ctx(req), args); err != nil {
			return err
		}
		return nil
	}
	return errNoStorageDetected
}

// AttachVolume returns a list of VolumeAttachments is sync/async that will
// attach a volume to an instance based on volumeID and instanceID.
func (s *service) AttachVolume(
	req *http.Request,
	args *api.AttachVolumeArgs,
	reply *api.AttachVolumeReply) error {

	var err error
	for _, d := range s.initializedDrivers {
		if reply.Attachments, err = d.AttachVolume(
			s.ctx(req), args); err != nil {
			return err
		}
		return nil
	}
	return errNoStorageDetected
}

// DetachVolume is sync/async that will detach the volumeID from the local
// instance or the instanceID.
func (s *service) DetachVolume(
	req *http.Request,
	args *api.DetachVolumeArgs,
	reply *api.DetachVolumeReply) error {

	var err error
	for _, d := range s.initializedDrivers {
		if err = d.DetachVolume(s.ctx(req), args); err != nil {
			return err
		}
		return nil
	}
	return errNoStorageDetected
}

// CopySnapshot is a sync/async and returns a snapshot that will copy a
// snapshot based on volumeID/snapshotID/snapshotName and create a new
// snapshot of desinationSnapshotName in the destinationRegion location.
func (s *service) CopySnapshot(
	req *http.Request,
	args *api.CopySnapshotArgs,
	reply *api.CopySnapshotReply) error {

	var err error
	for _, d := range s.initializedDrivers {
		if reply.Snapshot, err = d.CopySnapshot(s.ctx(req), args); err != nil {
			return err
		}
		return nil
	}
	return errNoStorageDetected
}

// GetClientToolName gets the file name of the tool this driver provides
// to be executed on the client-side in order to discover a client's
// instance ID and next, available device name.
//
// Use the function GetClientTool to get the actual tool.
func (s *service) GetClientToolName(
	req *http.Request,
	args *api.GetClientToolNameArgs,
	reply *api.GetClientToolNameReply) error {

	var err error
	for _, d := range s.initializedDrivers {
		if reply.ClientToolName, err = d.GetClientToolName(
			s.ctx(req), args); err != nil {
			return err
		}
		return nil
	}
	return errNoStorageDetected
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
func (s *service) GetClientTool(
	req *http.Request,
	args *api.GetClientToolArgs,
	reply *api.GetClientToolReply) error {

	var err error
	for _, d := range s.initializedDrivers {
		if reply.ClientTool, err = d.GetClientTool(
			s.ctx(req), args); err != nil {
			return err
		}
		return nil
	}
	return errNoStorageDetected
}

func newService(config gofig.Config) (*service, error) {
	s := &service{
		config:              config,
		constructedDrivers:  map[string]driver.Driver{},
		initializedDrivers:  map[string]driver.Driver{},
		initializingDrivers: &sync.Mutex{},
	}

	driverCtorsLock.RLock()
	defer driverCtorsLock.RUnlock()
	for name, ctor := range driverCtors {
		d := ctor(config)
		s.constructedDrivers[name] = d
		log.WithField("driverName", name).Debug("constructed driver")
	}

	if err := s.initDrivers(); err != nil {
		return nil, err
	}

	return s, nil
}

func getGroupForIP(ip, groupMap string) string {
	gmp := strings.Split(groupMap, "=")
	group := gmp[0]
	ipMap := strings.Split(gmp[1], ",")
	if util.StringInSlice(ip, ipMap) {
		return group
	}
	return ""
}

func profile(ctx context.Context, val *string) func() {
	originalVal := *val
	p := ctx.Value("profile")
	switch tp := p.(type) {
	case string:
		*val = fmt.Sprintf("%s-%s", tp, originalVal)
	}
	log.WithFields(log.Fields{
		"old": originalVal,
		"new": *val,
	}).Debug("prefixed value w/ profile")
	return func() { *val = originalVal }
}

func hash(text string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(text)))
}

func (s *service) ctx(req *http.Request) context.Context {

	ctx := context.Background()
	iid := gcontext.Get(req, "instanceID")
	switch tiid := iid.(type) {
	case (*api.InstanceID):
		ctx = context.WithInstanceID(ctx, tiid)
	}

	if !s.config.GetBool("libstorage.profiles.enabled") {
		log.Debug("not using server-side profiles")
		return ctx
	}

	ra := req.RemoteAddr
	rap := strings.Split(ra, ":")
	ip := rap[0]

	g := ""
	grps := s.config.GetStringSlice("libstorage.profiles.groups")
	if grps != nil {
		for _, gmap := range grps {
			if g = getGroupForIP(ip, gmap); g != "" {
				break
			}
		}
	}

	var profile string
	if profile = g; profile == "" {
		profile = hash(ip)
	}
	log.WithField("profile", profile).Debug("using server-side profile")
	return context.WithValue(ctx, "profile", profile)
}

// initDrivers initializes the drivers for the libStorage instance.
func (s *service) initDrivers() error {

	drivers := s.config.GetStringSlice("libstorage.drivers")
	log.WithField("drivers", drivers).Debug("libstorage get drivers")

	for n, d := range s.constructedDrivers {
		if util.StringInSlice(n, drivers) {
			if err := d.Init(); err != nil {
				log.WithFields(log.Fields{
					"name":  n,
					"error": err}).Debug(
					"error initializing libstorage driver")
				continue
			}
			log.WithField("name", n).Debug("initialized libstorage driver")
			s.initializedDrivers[n] = d
		}
	}

	log.WithFields(log.Fields{
		"registeredDriverNames":  getDriverNames(s.constructedDrivers),
		"initializedDriverNames": getDriverNames(s.initializedDrivers),
	}).Debug("libStorage.Server.InitDrivers")

	return nil
}

func getDriverNames(drivers map[string]driver.Driver) []string {
	driverNames := []string{}
	for dn := range drivers {
		driverNames = append(driverNames, dn)
	}
	return driverNames
}
