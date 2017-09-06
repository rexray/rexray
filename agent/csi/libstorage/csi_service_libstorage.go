package libstorage

import (
	"context"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
	xctx "golang.org/x/net/context"

	gofig "github.com/akutz/gofig/types"
	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
	"github.com/codedellemc/goioc"

	apictx "github.com/codedellemc/rexray/libstorage/api/context"
	apitypes "github.com/codedellemc/rexray/libstorage/api/types"
	apiutils "github.com/codedellemc/rexray/libstorage/api/utils"
)

func init() {
	goioc.Register("libstorage", func() interface{} { return &driver{} })
}

// ctxConfigKey is an interface-wrapped key used to access a possible
// config object in the context given to the provider's Serve function
var ctxConfigKey = interface{}("csi.config")

type driver struct {
	ctx          apitypes.Context
	client       apitypes.Client
	config       gofig.Config
	server       *grpc.Server
	svcName      string
	storType     apitypes.StorageType
	iid          *apitypes.InstanceID
	nodeID       *csi.NodeID
	attTokens    map[string]string
	attTokensRWL sync.RWMutex
}

func (d *driver) Serve(ctx context.Context, lis net.Listener) error {

	d.attTokens = map[string]string{}

	d.ctx = apictx.New(ctx)
	d.client = apictx.MustClient(d.ctx)

	// Check for a gofig.Config in the context.
	if config, ok := d.ctx.Value(ctxConfigKey).(gofig.Config); ok {
		log.Info("init csi libstorage bridge w ctx.config")
		d.config = config
	}

	// Cache the name of the libStorage service for which this bridge
	// is configured.
	svcName, ok := apictx.ServiceName(d.ctx)
	if !ok {
		return errMissingServiceName
	}
	d.svcName = svcName

	// Cache the instance ID for the service.
	iid, err := d.client.Executor().InstanceID(d.ctx, apiutils.NewStore())
	if err != nil {
		return err
	}
	d.iid = iid

	// Cache the storage type of the service.
	storType, err := d.client.Storage().Type(d.ctx)
	if err != nil {
		return err
	}
	d.storType = storType

	// Cache the node ID.
	d.nodeID = toNodeID(d.iid)

	// Create a gRPC server with an idempotent interceptor.
	d.server = grpc.NewServer(
		grpc.UnaryInterceptor(gocsi.NewIdempotentInterceptor(d)))

	csi.RegisterControllerServer(d.server, d)
	csi.RegisterIdentityServer(d.server, d)
	csi.RegisterNodeServer(d.server, d)
	return d.server.Serve(lis)
}

func (d *driver) Stop(ctx context.Context) {
	d.server.Stop()
}

func (d *driver) GracefulStop(ctx context.Context) {
	d.server.GracefulStop()
}

////////////////////////////////////////////////////////////////////////////////
//                            Controller Service                              //
////////////////////////////////////////////////////////////////////////////////

func (d *driver) CreateVolume(
	ctx xctx.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	// Make sure the requested volume capability is supported for
	// the cached storage type.
	if !isVolCapSupported(d.storType, req.VolumeCapabilities...) {
		return gocsi.ErrCreateVolume(
			csi.Error_CreateVolumeError_INVALID_PARAMETER_VALUE,
			"invalid volume capability request"), nil
	}

	opts := &apitypes.VolumeCreateOpts{Opts: apiutils.NewStore()}

	// Determine the requested volume size by selecting the
	// greater of the two values: Required and Limit bytes.
	if req.CapacityRange != nil {
		size := req.CapacityRange.LimitBytes
		if req.CapacityRange.RequiredBytes > size {
			size = req.CapacityRange.RequiredBytes
		}
		if size > 0 {
			if size > math.MaxInt64 {
				size = math.MaxInt64
			}
			opts.Size = addrOfInt64(b2gib(size))
		}
	}

	// Parse the request parameters and if any of the
	// keys match libStorage volume creation options then
	// attempt to use them.
	for k, v := range req.Parameters {
		if strings.EqualFold(k, "availabilityzone") {
			opts.AvailabilityZone = &v
		} else if strings.EqualFold(k, "encrypted") {
			b, _ := strconv.ParseBool(v)
			opts.Encrypted = &b
		} else if strings.EqualFold(k, "encryptionkey") {
			opts.EncryptionKey = &v
		} else if strings.EqualFold(k, "iops") {
			i, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				opts.IOPS = addrOfInt64(int64(i))
			}
		} else if strings.EqualFold(k, "type") {
			opts.Type = &v
		}
	}

	// Use the integration driver to create the volume.
	vol, err := d.client.Integration().Create(d.ctx, req.Name, opts)
	if err != nil {
		return nil, err
	}

	// Transform the created volume into a csi.VolumeInfo
	volInfo := toVolumeInfo(vol)

	return &csi.CreateVolumeResponse{
		Reply: &csi.CreateVolumeResponse_Result_{
			Result: &csi.CreateVolumeResponse_Result{
				VolumeInfo: volInfo,
			},
		},
	}, nil
}

func (d *driver) DeleteVolume(
	ctx xctx.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {

	volumeID, ok := req.VolumeId.Values["id"]
	if !ok {
		return gocsi.ErrDeleteVolume(
			csi.Error_DeleteVolumeError_INVALID_VOLUME_ID,
			`missing "id" field`), nil
	}

	opts := &apitypes.VolumeRemoveOpts{Opts: apiutils.NewStore()}

	err := d.client.Storage().VolumeRemove(d.ctx, volumeID, opts)
	if err != nil {
		return nil, err
	}

	return &csi.DeleteVolumeResponse{
		Reply: &csi.DeleteVolumeResponse_Result_{
			Result: &csi.DeleteVolumeResponse_Result{},
		},
	}, nil
}

func (d *driver) ControllerPublishVolume(
	ctx xctx.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	// Make sure the requested volume capability is supported for
	// the cached storage type.
	if !isVolCapSupported(d.storType, req.VolumeCapability) {
		return gocsi.ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_UNSUPPORTED_VOLUME_TYPE,
			"invalid volume capability request"), nil
	}

	iid, err := toInstanceID(req.NodeId)
	if err != nil {
		return gocsi.ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_INVALID_NODE_ID,
			err.Error()), nil
	}
	_ = iid

	//d.client.Storage().VolumeAttach(d.ctx, volumeID, opts)

	return nil, nil
}

func (d *driver) ControllerUnpublishVolume(
	ctx xctx.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	return nil, nil
}

func (d *driver) ValidateVolumeCapabilities(
	ctx xctx.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	return &csi.ValidateVolumeCapabilitiesResponse{
		Reply: &csi.ValidateVolumeCapabilitiesResponse_Result_{
			Result: &csi.ValidateVolumeCapabilitiesResponse_Result{
				Supported: isVolCapSupported(
					d.storType, req.VolumeCapabilities...),
			},
		},
	}, nil
}

func (d *driver) ListVolumes(
	ctx xctx.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	opts := &apitypes.VolumesOpts{Opts: apiutils.NewStore()}

	// Use the storage driver to list the volumes.
	vols, err := d.client.Storage().Volumes(d.ctx, opts)
	if err != nil {
		return nil, err
	}

	// Convert the libStorage volumes to CSI volume info objects.
	entries := make([]*csi.ListVolumesResponse_Result_Entry, len(vols))
	for i, v := range vols {
		entries[i] = &csi.ListVolumesResponse_Result_Entry{
			VolumeInfo: toVolumeInfo(v),
		}
	}

	return &csi.ListVolumesResponse{
		Reply: &csi.ListVolumesResponse_Result_{
			Result: &csi.ListVolumesResponse_Result{
				Entries: entries,
			},
		},
	}, nil
}

func (d *driver) GetCapacity(
	ctx xctx.Context,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	return gocsi.ErrGetCapacity(
		csi.Error_GeneralError_UNDEFINED,
		"call not implemented"), nil
}

func (d *driver) ControllerGetCapabilities(
	ctx xctx.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	return &csi.ControllerGetCapabilitiesResponse{
		Reply: &csi.ControllerGetCapabilitiesResponse_Result_{
			Result: &csi.ControllerGetCapabilitiesResponse_Result{
				Capabilities: []*csi.ControllerServiceCapability{
					&csi.ControllerServiceCapability{
						Type: &csi.ControllerServiceCapability_Rpc{
							Rpc: &csi.ControllerServiceCapability_RPC{
								Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
							},
						},
					},
					&csi.ControllerServiceCapability{
						Type: &csi.ControllerServiceCapability_Rpc{
							Rpc: &csi.ControllerServiceCapability_RPC{
								Type: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
							},
						},
					},
					&csi.ControllerServiceCapability{
						Type: &csi.ControllerServiceCapability_Rpc{
							Rpc: &csi.ControllerServiceCapability_RPC{
								Type: csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
							},
						},
					},
				},
			},
		},
	}, nil
}

////////////////////////////////////////////////////////////////////////////////
//                             Identity Service                               //
////////////////////////////////////////////////////////////////////////////////

func (d *driver) GetSupportedVersions(
	ctx xctx.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	return &csi.GetSupportedVersionsResponse{
		Reply: &csi.GetSupportedVersionsResponse_Result_{
			Result: &csi.GetSupportedVersionsResponse_Result{
				SupportedVersions: []*csi.Version{
					&csi.Version{
						Major: 0,
						Minor: 0,
						Patch: 0,
					},
				},
			},
		},
	}, nil
}

func (d *driver) GetPluginInfo(
	ctx xctx.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	return &csi.GetPluginInfoResponse{
		Reply: &csi.GetPluginInfoResponse_Result_{
			Result: &csi.GetPluginInfoResponse_Result{
				Name:          "REX-Ray CSI Bridge",
				VendorVersion: "0.0.0",
				Manifest:      nil,
			},
		},
	}, nil
}

////////////////////////////////////////////////////////////////////////////////
//                               Node Service                                 //
////////////////////////////////////////////////////////////////////////////////

func (d *driver) NodePublishVolume(
	ctx xctx.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	// Make sure the requested volume capability is supported for
	// the cached storage type.
	if !isVolCapSupported(d.storType, req.VolumeCapability) {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_UNSUPPORTED_VOLUME_TYPE,
			"invalid volume capability request"), nil
	}

	return nil, nil
}

func (d *driver) NodeUnpublishVolume(
	ctx xctx.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	return nil, nil
}

func (d *driver) GetNodeID(
	ctx xctx.Context,
	req *csi.GetNodeIDRequest) (
	*csi.GetNodeIDResponse, error) {

	return &csi.GetNodeIDResponse{
		Reply: &csi.GetNodeIDResponse_Result_{
			Result: &csi.GetNodeIDResponse_Result{
				NodeId: d.nodeID,
			},
		},
	}, nil
}

func (d *driver) ProbeNode(
	ctx xctx.Context,
	req *csi.ProbeNodeRequest) (
	*csi.ProbeNodeResponse, error) {

	sup, err := d.client.Executor().Supported(d.ctx, apiutils.NewStore())
	if err != nil {
		return nil, err
	}

	// If the supported mask is anything but zero then return
	// a positive result.
	if sup > apitypes.LSXSOpNone {
		return &csi.ProbeNodeResponse{
			Reply: &csi.ProbeNodeResponse_Result_{
				Result: &csi.ProbeNodeResponse_Result{},
			},
		}, nil
	}

	// Return an error indicating no support.
	return gocsi.ErrProbeNode(
			csi.Error_ProbeNodeError_MISSING_REQUIRED_HOST_DEPENDENCY, ""),
		nil
}

func (d *driver) NodeGetCapabilities(
	ctx xctx.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	return &csi.NodeGetCapabilitiesResponse{
		Reply: &csi.NodeGetCapabilitiesResponse_Result_{
			Result: &csi.NodeGetCapabilitiesResponse_Result{
				Capabilities: []*csi.NodeServiceCapability{
					&csi.NodeServiceCapability{
						Type: &csi.NodeServiceCapability_Rpc{
							Rpc: &csi.NodeServiceCapability_RPC{
								Type: csi.NodeServiceCapability_RPC_UNKNOWN,
							},
						},
					},
				},
			},
		},
	}, nil
}
