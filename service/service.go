package service

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	golog "log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"golang.org/x/net/context"

	"github.com/emccode/libstorage/api"
	"github.com/emccode/libstorage/driver"
	"github.com/emccode/libstorage/model"
	"github.com/emccode/libstorage/util"
)

const (
	lowerhex = "0123456789abcdef"
)

var (
	errNoStorageDetected          = goof.New("no storage detected")
	errDriverBlockDeviceDiscovery = goof.New("no block devices discovered")

	driverCtors     = map[string]driver.NewDriver{}
	driverCtorsLock = &sync.RWMutex{}
)

type service struct {
	config              *gofig.Config
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
func Serve(config *gofig.Config) (err error) {
	ls := newService(config)

	host := config.GetString("libstorage.host")
	if host == "" {
		host = "tcp://127.0.0.1:0"
	}

	s := rpc.NewServer()
	codec := json.NewCodec()
	s.RegisterCodec(codec, "application/json")
	s.RegisterCodec(codec, "application/json;charset=UTF-8")
	if err = s.RegisterService(ls, "libStorage"); err != nil {
		return
	}

	var netProto, laddr string
	if netProto, laddr, err = util.ParseAddress(host); err != nil {
		return
	}

	var l net.Listener
	log.WithField("host", host).Debug("ready to listen")
	if l, err = net.Listen(netProto, laddr); err != nil {
		return
	}

	var logReq, logRes bool
	var stdout, stderr io.WriteCloser

	doLogs := config.GetBool("libstorage.service.http.logging.enabled")
	if doLogs {
		stdout = getLogIO("libstorage.service.http.logging.out", config)
		logReq = config.GetBool("libstorage.service.http.logging.logrequest")
		logRes = config.GetBool("libstorage.service.http.logging.logresponse")
	}

	handleHTTP := http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			if !doLogs {
				s.ServeHTTP(w, req)
				return
			}

			var reqDump []byte
			if logReq {
				if reqDump, err = httputil.DumpRequest(req, true); err != nil {
					log.Error(err)
				}
			}

			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, req)

			logRequest(logReq, stdout, rec, req, reqDump)
			if logRes {
				fmt.Fprintln(stdout, "")
				logResponse(stdout, rec, req)
				fmt.Fprintln(stdout, "")
			}

			w.WriteHeader(rec.Code)
			for k, v := range rec.HeaderMap {
				w.Header()[k] = v
			}
			w.Write(rec.Body.Bytes())
		})

	mux := http.NewServeMux()
	mux.Handle("/libStorage", handleHTTP)

	hs := &http.Server{
		Addr:           laddr,
		Handler:        mux,
		ReadTimeout:    getReadTimeout(config) * time.Second,
		WriteTimeout:   getWriteTimeout(config) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if doLogs {
		hs.ErrorLog = golog.New(stderr, "", 0)
	}

	go func() {
		defer func() {
			close(stdout)
			close(stderr)

			r := recover()
			switch tr := r.(type) {
			case error:
				log.Panic(
					"unhandled exception when serving libStorage service", tr)
			}
		}()

		if err = hs.Serve(l); err != nil {
			log.Panic("error serving libStorage service", err)
		}
	}()

	updatedHost := fmt.Sprintf("%s://%s", l.Addr().Network(), l.Addr().String())
	if updatedHost != host {
		host = updatedHost
		config.Set("libstorage.host", host)
	}
	log.WithField("host", host).Debug("listening")

	return
}

func close(toClose io.Closer) {
	if toClose == nil {
		return
	}
	if err := toClose.Close(); err != nil {
		log.Error(err)
	}
}

func logRequest(
	l bool,
	w io.Writer,
	rec *httptest.ResponseRecorder,
	req *http.Request,
	reqDump []byte) {

	cll := buildCommonLogLine(
		req, *req.URL, time.Now(), rec.Code, rec.Body.Len())
	fmt.Fprintln(w, string(cll))

	if !l || len(reqDump) == 0 {
		return
	}

	fmt.Fprintln(w, "")
	fmt.Fprint(w, "    ------------------------------ ")
	fmt.Fprint(w, "HTTP REQUEST")
	fmt.Fprintln(w, " ------------------------------")
	writeIndented(w, reqDump)
}

func logResponse(
	w io.Writer,
	rec *httptest.ResponseRecorder,
	req *http.Request) {

	fmt.Fprint(w, "    ------------------------------ ")
	fmt.Fprint(w, "HTTP RESPONSE")
	fmt.Fprintln(w, " -----------------------------")

	for k, v := range rec.HeaderMap {
		fmt.Fprintf(w, "    %s=%s\n", k, strings.Join(v, ","))
	}
	fmt.Fprintln(w, "")
	writeIndented(w, rec.Body.Bytes())
}

func writeIndented(w io.Writer, b []byte) {
	s := bufio.NewScanner(bytes.NewReader(b))
	for s.Scan() {
		if _, err := fmt.Fprintf(w, "    %s\n", s.Text()); err != nil {
			log.Error(err)
			return
		}
	}
}

func getReadTimeout(config *gofig.Config) time.Duration {
	t := config.GetInt("libstorage.service.readtimeout")
	if t == 0 {
		return 60
	}
	return time.Duration(t)
}

func getWriteTimeout(config *gofig.Config) time.Duration {
	t := config.GetInt("libstorage.service.writetimeout")
	if t == 0 {
		return 60
	}
	return time.Duration(t)
}

func getLogIO(propName string, config *gofig.Config) io.WriteCloser {
	if path := config.GetString(propName); path != "" {
		logio, err := os.OpenFile(
			path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Error(err)
		}
		log.WithFields(log.Fields{
			"logType": propName,
			"logPath": path,
		}).Debug("using log file")
		return logio
	}
	return log.StandardLogger().Writer()
}

func newService(config *gofig.Config) api.ServiceEndpoint {
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

	return s
}

func profile(ctx context.Context, val string) string {
	profileVal := val
	p := ctx.Value("profile")
	switch tp := p.(type) {
	case string:
		profileVal = fmt.Sprintf("%s-%s", tp, val)
	}
	log.WithFields(log.Fields{
		"old": val,
		"new": profileVal,
	}).Debug("prefixed value w/ profile")
	return profileVal
}

func hash(text string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(text)))
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

func ctx(
	config *gofig.Config,
	req *http.Request,
	instanceID *model.InstanceID) context.Context {
	ctx := util.WithInstanceID(instanceID)
	if !config.GetBool("libstorage.profiles.enabled") {
		log.Debug("not using server-side profiles")
		return ctx
	}

	ra := req.RemoteAddr
	rap := strings.Split(ra, ":")
	ip := rap[0]

	g := ""
	grps := config.GetStringSlice("libstorage.profiles.groups")
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

// InitDrivers initializes the drivers for the LibStorage instance.
func (s *service) InitDrivers(
	req *http.Request,
	args *api.InitDriversArgs,
	reply *api.InitDriversReply) error {

	// exit super quick without locking if the drivers are initialized
	if s.driversInitialized {
		return nil
	}

	s.initializingDrivers.Lock()
	defer s.initializingDrivers.Unlock()

	// check again to see if the flag is true since there's a small
	// possibility it could be true between the first check and acquiring a
	// lock
	if s.driversInitialized {
		return nil
	}

	// indicate that the drivers have been initialized before they are since
	// if they fail initialization there is no point in subsequent attempts
	s.driversInitialized = true

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

	reply.RegisteredDriverNames = getDriverNames(s.constructedDrivers)
	reply.InitializedDriverNames = getDriverNames(s.initializedDrivers)

	log.WithFields(log.Fields{
		"registeredDriverNames":  reply.RegisteredDriverNames,
		"initializedDriverNames": reply.InitializedDriverNames,
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
		bds, err := d.GetVolumeMapping(ctx(s.config, req, args.InstanceID))
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
		if reply.Instance, err = d.GetInstance(
			ctx(s.config, req, args.InstanceID)); err != nil {
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

	ctx := ctx(s.config, req, args.InstanceID)
	volName := profile(ctx, args.VolumeName)

	for _, d := range s.initializedDrivers {
		if reply.Volumes, err = d.GetVolume(ctx,
			args.VolumeID, volName); err != nil {
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
			ctx(s.config, req, args.InstanceID), args.VolumeID); err != nil {
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

	ctx := ctx(s.config, req, args.InstanceID)

	var err error
	for _, d := range s.initializedDrivers {
		if reply.Snapshots, err = d.CreateSnapshot(
			ctx,
			args.SnapshotName,
			args.VolumeID,
			args.Description); err != nil {
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
		if reply.Snapshots, err = d.GetSnapshot(
			ctx(s.config, req, args.InstanceID),
			args.VolumeID,
			args.SnapshotID,
			args.SnapshotName); err != nil {
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

	ctx := ctx(s.config, req, args.InstanceID)

	var err error
	for _, d := range s.initializedDrivers {
		if err = d.RemoveSnapshot(
			ctx, args.SnapshotID); err != nil {
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

	ctx := ctx(s.config, req, args.InstanceID)
	volName := profile(ctx, args.VolumeName)

	var err error
	for _, d := range s.initializedDrivers {
		if reply.Volume, err = d.CreateVolume(
			ctx,
			volName,
			args.VolumeID,
			args.SnapshotID,
			args.VolumeType,
			args.IOPS,
			args.Size,
			args.AvailabilityZone); err != nil {
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

	ctx := ctx(s.config, req, args.InstanceID)

	var err error
	for _, d := range s.initializedDrivers {
		if err = d.RemoveVolume(
			ctx, args.VolumeID); err != nil {
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
	ctx := ctx(s.config, req, args.InstanceID)
	var err error
	for _, d := range s.initializedDrivers {
		if reply.Attachments, err = d.AttachVolume(
			ctx,
			args.NextDeviceName,
			args.VolumeID); err != nil {
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
	ctx := ctx(s.config, req, args.InstanceID)
	var err error
	for _, d := range s.initializedDrivers {
		if err = d.DetachVolume(ctx, args.VolumeID); err != nil {
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
	ctx := ctx(s.config, req, args.InstanceID)
	var err error
	for _, d := range s.initializedDrivers {
		if reply.Snapshot, err = d.CopySnapshot(
			ctx,
			args.VolumeID,
			args.SnapshotID,
			args.SnapshotName,
			args.DestinationSnapshotName,
			args.DestinationRegion); err != nil {
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
	ctx := ctx(s.config, req, args.InstanceID)
	var err error
	for _, d := range s.initializedDrivers {
		if reply.ClientToolName, err = d.GetClientToolName(ctx); err != nil {
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
	ctx := ctx(s.config, req, args.InstanceID)
	var err error
	for _, d := range s.initializedDrivers {
		if reply.ClientTool, err = d.GetClientTool(ctx); err != nil {
			return err
		}
		return nil
	}
	return errNoStorageDetected
}

// buildCommonLogLine builds a log entry for req in Apache Common Log Format.
// ts is the timestamp with which the entry should be logged.
// status and size are used to provide the response HTTP status and size.
//
// This function was taken from the Gorilla toolkit's handlers.go file.
func buildCommonLogLine(
	req *http.Request,
	url url.URL,
	ts time.Time,
	status int,
	size int) []byte {

	username := "-"
	if url.User != nil {
		if name := url.User.Username(); name != "" {
			username = name
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)

	if err != nil {
		host = req.RemoteAddr
	}

	uri := req.RequestURI

	// Requests using the CONNECT method over HTTP/2.0 must use
	// the authority field (aka r.Host) to identify the target.
	// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
	if req.ProtoMajor == 2 && req.Method == "CONNECT" {
		uri = req.Host
	}
	if uri == "" {
		uri = url.RequestURI()
	}

	buf := make([]byte, 0, 3*(len(host)+len(username)+
		len(req.Method)+len(uri)+len(req.Proto)+50)/2)
	buf = append(buf, host...)
	buf = append(buf, " - "...)
	buf = append(buf, username...)
	buf = append(buf, " ["...)
	buf = append(buf, ts.Format("02/Jan/2006:15:04:05 -0700")...)
	buf = append(buf, `] "`...)
	buf = append(buf, req.Method...)
	buf = append(buf, " "...)
	buf = appendQuoted(buf, uri)
	buf = append(buf, " "...)
	buf = append(buf, req.Proto...)
	buf = append(buf, `" `...)
	buf = append(buf, strconv.Itoa(status)...)
	buf = append(buf, " "...)
	buf = append(buf, strconv.Itoa(size)...)
	return buf
}

func appendQuoted(buf []byte, s string) []byte {
	var runeTmp [utf8.UTFMax]byte
	for width := 0; len(s) > 0; s = s[width:] {
		r := rune(s[0])
		width = 1
		if r >= utf8.RuneSelf {
			r, width = utf8.DecodeRuneInString(s)
		}
		if width == 1 && r == utf8.RuneError {
			buf = append(buf, `\x`...)
			buf = append(buf, lowerhex[s[0]>>4])
			buf = append(buf, lowerhex[s[0]&0xF])
			continue
		}
		if r == rune('"') || r == '\\' { // always backslashed
			buf = append(buf, '\\')
			buf = append(buf, byte(r))
			continue
		}
		if strconv.IsPrint(r) {
			n := utf8.EncodeRune(runeTmp[:], r)
			buf = append(buf, runeTmp[:n]...)
			continue
		}
		switch r {
		case '\a':
			buf = append(buf, `\a`...)
		case '\b':
			buf = append(buf, `\b`...)
		case '\f':
			buf = append(buf, `\f`...)
		case '\n':
			buf = append(buf, `\n`...)
		case '\r':
			buf = append(buf, `\r`...)
		case '\t':
			buf = append(buf, `\t`...)
		case '\v':
			buf = append(buf, `\v`...)
		default:
			switch {
			case r < ' ':
				buf = append(buf, `\x`...)
				buf = append(buf, lowerhex[s[0]>>4])
				buf = append(buf, lowerhex[s[0]&0xF])
			case r > utf8.MaxRune:
				r = 0xFFFD
				fallthrough
			case r < 0x10000:
				buf = append(buf, `\u`...)
				for s := 12; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			default:
				buf = append(buf, `\U`...)
				for s := 28; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			}
		}
	}
	return buf

}
