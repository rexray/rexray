// +build !rexray_build_type_client
// +build !rexray_build_type_controller
// +build csi

package csi

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"
	apitypes "github.com/codedellemc/libstorage/api/types"
	"google.golang.org/grpc"
	xctx "golang.org/x/net/context"

	"github.com/codedellemc/rexray/daemon/module"
	"github.com/codedellemc/rexray/daemon/module/csi/csi"
)

const (
	modName = "csi"
)

type mod struct {
	lsc    apitypes.Client
	ctx    apitypes.Context
	config gofig.Config
	name   string
	addr   string
	desc   string
	cs     *service
	gs     *grpc.Server
}

var (
	separators  = regexp.MustCompile(`[ &_=+:]`)
	dashes      = regexp.MustCompile(`[\-]+`)
	illegalPath = regexp.MustCompile(`[^[:alnum:]\~\-\./]`)
)

func init() {
	module.RegisterModule(modName, newModule)
}

func newModule(
	ctx apitypes.Context,
	c *module.Config) (module.Module, error) {

	host := strings.Trim(c.Address, " ")

	if host == "" {
		return nil, errors.New("error: host is required")
	}

	c.Address = host
	config := c.Config

	return &mod{
		ctx:    ctx,
		config: config,
		lsc:    c.Client,
		name:   c.Name,
		desc:   c.Description,
		addr:   host,
		cs:     &service{},
	}, nil
}

func (m *mod) Start() error {

	proto, addr, err := gotil.ParseAddress(m.Address())
	if err != nil {
		return err
	}

	// ensure the sock file directory is created & remove
	// any stale sock files with the same path
	if proto == "unix" {
		os.MkdirAll(filepath.Dir(addr), 0755)
		os.RemoveAll(addr)
	}

	// create a listener
	l, err := net.Listen(proto, addr)
	if err != nil {
		return err
	}

	// create a grpc server
	m.gs = grpc.NewServer()
	csi.RegisterControllerServer(m.gs, m.cs)
	csi.RegisterIdentityServer(m.gs, m.cs)
	csi.RegisterNodeServer(m.gs, m.cs)

	go func() {
		err := m.gs.Serve(l)
		if proto == "unix" {
			os.RemoveAll(addr)
		}
		if err != grpc.ErrServerStopped {
			panic(err)
		}
	}()

	return nil
}

func (m *mod) Stop() error {
	m.gs.GracefulStop()
	return nil
}

func (m *mod) Name() string {
	return m.name
}

func (m *mod) Description() string {
	return m.desc
}

func (m *mod) Address() string {
	return m.addr
}

////////////////////////////////////////////////////////////////////////////////
//                            CSI Service Struct                              //
////////////////////////////////////////////////////////////////////////////////

type service struct {
}

////////////////////////////////////////////////////////////////////////////////
//                            Controller Service                              //
////////////////////////////////////////////////////////////////////////////////

func (s *service) CreateVolume(
	ctx xctx.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	if len(req.GetName()) == 0 {
		// INVALID_VOLUME_NAME
		return ErrCreateVolume(3, "missing name"), nil
	}

	return nil, nil
}

func (s *service) DeleteVolume(
	ctx xctx.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {

	idObj := req.GetVolumeId()
	if idObj == nil {
		// INVALID_VOLUME_ID
		return ErrDeleteVolume(3, "missing id obj"), nil
	}

	idVals := idObj.GetValues()
	if len(idVals) == 0 {
		// INVALID_VOLUME_ID
		return ErrDeleteVolume(3, "missing id map"), nil
	}

	return nil, nil
}

func (s *service) ControllerPublishVolume(
	ctx xctx.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	idObj := req.GetVolumeId()
	if idObj == nil {
		// INVALID_VOLUME_ID
		return ErrControllerPublishVolume(3, "missing id obj"), nil
	}

	idVals := idObj.GetValues()
	if len(idVals) == 0 {
		// INVALID_VOLUME_ID
		return ErrControllerPublishVolume(3, "missing id map"), nil
	}

	return nil, nil
}

func (s *service) ControllerUnpublishVolume(
	ctx xctx.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	idObj := req.GetVolumeId()
	if idObj == nil {
		// INVALID_VOLUME_ID
		return ErrControllerUnpublishVolume(3, "missing id obj"), nil
	}

	idVals := idObj.GetValues()
	if len(idVals) == 0 {
		// INVALID_VOLUME_ID
		return ErrControllerUnpublishVolume(3, "missing id map"), nil
	}

	return nil, nil
}

func (s *service) ValidateVolumeCapabilities(
	ctx xctx.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	return nil, nil
}

func (s *service) ListVolumes(
	ctx xctx.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	return nil, nil
}

func (s *service) GetCapacity(
	ctx xctx.Context,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	return nil, nil
}

func (s *service) ControllerGetCapabilities(
	ctx xctx.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////
//                             Identity Service                               //
////////////////////////////////////////////////////////////////////////////////

func (s *service) GetSupportedVersions(
	ctx xctx.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	return nil, nil
}

func (s *service) GetPluginInfo(
	ctx xctx.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////
//                               Node Service                                 //
////////////////////////////////////////////////////////////////////////////////

func (s *service) NodePublishVolume(
	ctx xctx.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	idObj := req.GetVolumeId()
	if idObj == nil {
		// MISSING_REQUIRED_FIELD
		return ErrNodePublishVolumeGeneral(3, "missing id obj"), nil
	}

	idVals := idObj.GetValues()
	if len(idVals) == 0 {
		// MISSING_REQUIRED_FIELD
		return ErrNodePublishVolumeGeneral(3, "missing id map"), nil
	}

	return nil, nil
}

func (s *service) NodeUnpublishVolume(
	ctx xctx.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	idObj := req.GetVolumeId()
	if idObj == nil {
		// MISSING_REQUIRED_FIELD
		return ErrNodeUnpublishVolumeGeneral(3, "missing id obj"), nil
	}

	idVals := idObj.GetValues()
	if len(idVals) == 0 {
		// MISSING_REQUIRED_FIELD
		return ErrNodeUnpublishVolumeGeneral(3, "missing id map"), nil
	}

	return nil, nil
}

func (s *service) GetNodeID(
	ctx xctx.Context,
	req *csi.GetNodeIDRequest) (
	*csi.GetNodeIDResponse, error) {

	return nil, nil
}

func (s *service) ProbeNode(
	ctx xctx.Context,
	req *csi.ProbeNodeRequest) (
	*csi.ProbeNodeResponse, error) {

	return nil, nil
}

func (s *service) NodeGetCapabilities(
	ctx xctx.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	return nil, nil
}
