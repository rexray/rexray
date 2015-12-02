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
	"github.com/akutz/gotil"
	gcontext "github.com/gorilla/context"

	"github.com/emccode/libstorage/api"
	"github.com/emccode/libstorage/context"
	"github.com/emccode/libstorage/driver"
	"github.com/emccode/libstorage/service/server"
)

var (
	errNoStorageDetected          = goof.New("no storage detected")
	errDriverBlockDeviceDiscovery = goof.New("no block devices discovered")

	driverCtors     = map[string]driver.NewDriver{}
	driverCtorsLock = &sync.RWMutex{}
)

type service struct {
	name               string
	config             gofig.Config
	constructedDrivers map[string]driver.Driver
	driver             driver.Driver
	driversInitialized bool
	initializingDriver *sync.Mutex
}

// RegisterDriver is used by drivers to indicate their availability to
// libstorage.
func RegisterDriver(driverName string, ctor driver.NewDriver) {
	driverCtorsLock.Lock()
	defer driverCtorsLock.Unlock()
	driverCtors[strings.ToLower(driverName)] = ctor
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

		service, err := newService(name, sc)
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

func (s *service) GetServiceInfo(
	req *http.Request,
	args *api.GetServiceInfoArgs,
	reply *api.GetServiceInfoReply) error {
	reply.Name = s.name
	reply.Driver = s.driver.Name()
	reply.RegisteredDrivers = s.getRegisteredDriverNames()
	return nil
}

func (s *service) GetNextAvailableDeviceName(
	req *http.Request,
	args *api.GetNextAvailableDeviceNameArgs,
	reply *api.GetNextAvailableDeviceNameReply) (err error) {
	reply.Next, err =
		s.driver.GetNextAvailableDeviceName(s.ctx(req), args)
	return nil
}

func (s *service) GetVolumeMapping(
	req *http.Request,
	args *api.GetVolumeMappingArgs,
	reply *api.GetVolumeMappingReply) error {

	bds, err := s.driver.GetVolumeMapping(s.ctx(req), args)
	if err != nil {
		return errDriverBlockDeviceDiscovery
	}

	for _, bd := range bds {
		reply.BlockDevices = append(reply.BlockDevices, bd)
	}

	if len(reply.BlockDevices) == 0 {
		return errNoStorageDetected
	}

	log.WithField(
		"len(blockDevices)", len(reply.BlockDevices)).Debug(
		"libStorage.Server.GetVolumeMapping")

	return nil
}

func (s *service) GetInstance(
	req *http.Request,
	args *api.GetInstanceArgs,
	reply *api.GetInstanceReply) (err error) {
	reply.Instance, err = s.driver.GetInstance(s.ctx(req), args)
	return
}

func (s *service) GetVolume(
	req *http.Request,
	args *api.GetVolumeArgs,
	reply *api.GetVolumeReply) (err error) {

	ctx := s.ctx(req)
	unprofile := profile(ctx, &args.Optional.VolumeName)
	defer unprofile()

	if reply.Volumes, err = s.driver.GetVolume(ctx, args); err != nil {
		return
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
	return
}

func (s *service) GetVolumeAttach(
	req *http.Request,
	args *api.GetVolumeAttachArgs,
	reply *api.GetVolumeAttachReply) (err error) {
	reply.Attachments, err = s.driver.GetVolumeAttach(s.ctx(req), args)
	return
}

func (s *service) CreateSnapshot(
	req *http.Request,
	args *api.CreateSnapshotArgs,
	reply *api.CreateSnapshotReply) (err error) {
	reply.Snapshots, err = s.driver.CreateSnapshot(s.ctx(req), args)
	return
}

func (s *service) GetSnapshot(
	req *http.Request,
	args *api.GetSnapshotArgs,
	reply *api.GetSnapshotReply) (err error) {
	reply.Snapshots, err = s.driver.GetSnapshot(s.ctx(req), args)
	return
}

func (s *service) RemoveSnapshot(
	req *http.Request,
	args *api.RemoveSnapshotArgs,
	reply *api.RemoveSnapshotReply) (err error) {
	err = s.driver.RemoveSnapshot(s.ctx(req), args)
	return
}

func (s *service) CreateVolume(
	req *http.Request,
	args *api.CreateVolumeArgs,
	reply *api.CreateVolumeReply) (err error) {

	ctx := s.ctx(req)
	unprofile := profile(ctx, &args.Optional.VolumeName)
	defer unprofile()
	if reply.Volume, err = s.driver.CreateVolume(ctx, args); err != nil {
		return
	}
	if s.config.GetBool("libstorage.profiles.client") {
		return
	}

	np := strings.Split(reply.Volume.Name, "-")
	if len(np) == 1 {
		reply.Volume.Name = np[0]
	} else {
		reply.Volume.Name = np[1]
	}
	return
}

func (s *service) RemoveVolume(
	req *http.Request,
	args *api.RemoveVolumeArgs,
	reply *api.RemoveVolumeReply) (err error) {
	err = s.driver.RemoveVolume(s.ctx(req), args)
	return
}

func (s *service) AttachVolume(
	req *http.Request,
	args *api.AttachVolumeArgs,
	reply *api.AttachVolumeReply) (err error) {
	reply.Attachments, err = s.driver.AttachVolume(s.ctx(req), args)
	return
}

func (s *service) DetachVolume(
	req *http.Request,
	args *api.DetachVolumeArgs,
	reply *api.DetachVolumeReply) (err error) {
	err = s.driver.DetachVolume(s.ctx(req), args)
	return
}

func (s *service) CopySnapshot(
	req *http.Request,
	args *api.CopySnapshotArgs,
	reply *api.CopySnapshotReply) (err error) {
	reply.Snapshot, err = s.driver.CopySnapshot(s.ctx(req), args)
	return
}

func (s *service) GetClientTool(
	req *http.Request,
	args *api.GetClientToolArgs,
	reply *api.GetClientToolReply) (err error) {
	reply.ClientTool, err = s.driver.GetClientTool(s.ctx(req), args)

	// calculate the checksum of the tool if the tool is set and no checksum
	// exists for it
	if len(reply.ClientTool.Data) > 0 && reply.ClientTool.MD5Checksum == "" {
		reply.ClientTool.MD5Checksum = fmt.Sprintf("%x",
			md5.Sum(reply.ClientTool.Data))
	}
	return
}

func newService(name string, config gofig.Config) (*service, error) {
	s := &service{
		name:               name,
		config:             config,
		constructedDrivers: map[string]driver.Driver{},
		initializingDriver: &sync.Mutex{},
	}

	driverCtorsLock.RLock()
	defer driverCtorsLock.RUnlock()
	for name, ctor := range driverCtors {
		d := ctor(config)
		s.constructedDrivers[name] = d
		log.WithField("driverName", name).Debug("constructed driver")
	}

	if err := s.initDriver(); err != nil {
		return nil, err
	}

	return s, nil
}

func getGroupForIP(ip, groupMap string) string {
	gmp := strings.Split(groupMap, "=")
	group := gmp[0]
	ipMap := strings.Split(gmp[1], ",")
	if gotil.StringInSlice(ip, ipMap) {
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

// initDriver initializes the drivers for the libStorage instance.
func (s *service) initDriver() error {

	driverName := s.config.GetString("libstorage.driver")
	log.WithField("driverName", driverName).Debug("libstorage driver")

	var ok bool
	s.driver, ok = s.constructedDrivers[strings.ToLower(driverName)]
	if !ok {
		return errNoStorageDetected
	}

	if err := s.driver.Init(); err != nil {
		return goof.WithFields(goof.Fields{
			"name":  s.driver.Name(),
			"error": err,
		}, "error initializing libstorage driver")
	}

	log.WithFields(log.Fields{
		"registeredDriverNames": s.getRegisteredDriverNames(),
		"driverName":            s.driver.Name(),
	}).Debug("libStorage.Server.InitDrivers")

	return nil
}

func (s *service) getRegisteredDriverNames() []string {
	driverNames := []string{}
	for dn := range s.constructedDrivers {
		driverNames = append(driverNames, dn)
	}
	return driverNames
}
