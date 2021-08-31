package libstorage

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"

	xctx "golang.org/x/net/context"

	gofig "github.com/akutz/gofig/types"
	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
	"github.com/thecodeteam/gocsi/mount"
	"github.com/thecodeteam/goioc"

	apictx "github.com/AVENTER-UG/rexray/libstorage/api/context"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	apiutils "github.com/AVENTER-UG/rexray/libstorage/api/utils"
	rrutils "github.com/AVENTER-UG/rexray/util"
)

const (
	// WFDTimeout is the number of seconds to set the timeout to when
	// calling WaitForDevice
	WFDTimeout = 10

	devtmpfs = "devtmpfs"
)

var (
	// ctxConfigKey is an interface-wrapped key used to access a possible
	// config object in the context given to the provider's Serve function
	ctxConfigKey     = interface{}("csi.config")
	ctxExactMountKey = interface{}("exactmount")
	errDirNeeded     = errors.New("target path needs to be a directory")
	errFileNeeded    = errors.New("target path needs to be a file")
)

func init() {
	goioc.Register("libstorage", func() interface{} { return &driver{} })

	mount.BypassSourceFilesystemTypes = []string{
		`(?i)^devtmpfs$`,
		`(?i)^fuse\.`,
		`(?i)^nfs(\d)?$`,
	}
}

type driver struct {
	ctx      apitypes.Context
	client   apitypes.Client
	config   gofig.Config
	server   *grpc.Server
	svcName  string
	storType apitypes.StorageType
	iid      *apitypes.InstanceID
	nodeID   *csi.NodeID
	mntPath  string
}

func (d *driver) Serve(ctx context.Context, lis net.Listener) error {

	d.ctx = apictx.New(ctx)
	d.client = apictx.MustClient(d.ctx)

	// Check for a gofig.Config in the context.
	if config, ok := d.ctx.Value(ctxConfigKey).(gofig.Config); ok {
		d.ctx.Info("init csi libstorage bridge w ctx.config")
		d.config = config
		d.mntPath = d.config.GetString(apitypes.ConfigIgVolOpsMountPath)
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

	szTimeout := d.config.GetString("csi.libstorage.timeout")
	timeout, _ := time.ParseDuration(szTimeout)

	lout := rrutils.NewWriterFor(d.ctx.Infof)
	lerr := rrutils.NewWriterFor(d.ctx.Errorf)

	interceptors := grpc.UnaryInterceptor(gocsi.ChainUnaryServer(
		gocsi.ServerRequestIDInjector,
		gocsi.NewServerRequestLogger(lout, lerr),
		gocsi.NewServerResponseLogger(lout, lerr),
		gocsi.NewIdempotentInterceptor(d, timeout),
	))

	// Create a gRPC server with an idempotent interceptor.
	d.server = grpc.NewServer(interceptors)

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
	if len(req.Parameters) > 0 {
		opts.Opts = apiutils.NewStoreWithVars(req.Parameters)
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

	volumeID, ok := req.VolumeId.Values["id"]
	if !ok {
		return gocsi.ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_INVALID_VOLUME_ID,
			`missing "id" field`), nil
	}

	// Make sure the requested volume capability is supported for
	// the cached storage type.
	if !isVolCapSupported(d.storType, req.VolumeCapability) {
		return gocsi.ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_UNSUPPORTED_VOLUME_TYPE,
			"invalid volume capability request"), nil
	}

	_, err := toInstanceID(req.NodeId)
	if err != nil {
		return gocsi.ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_INVALID_NODE_ID,
			err.Error()), nil
	}

	opts := &apitypes.VolumeAttachOpts{
		Force: d.config.GetBool(apitypes.ConfigIgVolOpsMountPreempt),
		Opts:  apiutils.NewStore(),
	}
	vol, token, err := d.client.Storage().VolumeAttach(d.ctx, volumeID, opts)
	if err != nil {
		return nil, err
	}

	return &csi.ControllerPublishVolumeResponse{
		Reply: &csi.ControllerPublishVolumeResponse_Result_{
			Result: &csi.ControllerPublishVolumeResponse_Result{
				PublishVolumeInfo: &csi.PublishVolumeInfo{
					Values: map[string]string{
						"token":     token,
						"encrypted": fmt.Sprintf("%v", vol.Encrypted),
					},
				},
			},
		},
	}, nil
}

func (d *driver) ControllerUnpublishVolume(
	ctx xctx.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	volumeID, ok := req.VolumeId.Values["id"]
	if !ok {
		return gocsi.ErrControllerUnpublishVolume(
			csi.Error_ControllerUnpublishVolumeError_INVALID_VOLUME_ID,
			`missing "id" field`), nil
	}

	_, err := toInstanceID(req.NodeId)
	if err != nil {
		return gocsi.ErrControllerUnpublishVolume(
			csi.Error_ControllerUnpublishVolumeError_INVALID_NODE_ID,
			err.Error()), nil
	}

	opts := &apitypes.VolumeDetachOpts{Opts: apiutils.NewStore()}

	_, err = d.client.Storage().VolumeDetach(d.ctx, volumeID, opts)
	if err != nil {
		return nil, err
	}

	return &csi.ControllerUnpublishVolumeResponse{
		Reply: &csi.ControllerUnpublishVolumeResponse_Result_{
			Result: &csi.ControllerUnpublishVolumeResponse_Result{},
		},
	}, nil
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

	// If isMountInfoRequested is true then set the VolumesOpts to
	// request attachment information for this instance.
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

	volumeID, ok := req.VolumeId.Values["id"]
	if !ok {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_INVALID_VOLUME_ID,
			`missing "id" field`), nil
	}
	target := req.GetTargetPath()
	vi := req.GetPublishVolumeInfo()

	st, err := os.Stat(target)
	if os.IsNotExist(err) {
		return nil, errMissingTargetPath
	}

	// Make sure the requested volume capability is supported for
	// the cached storage type.
	if !isVolCapSupported(d.storType, req.VolumeCapability) {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_UNSUPPORTED_VOLUME_TYPE,
			"invalid volume capability request"), nil
	}

	token, ok := vi.Values["token"]
	if !ok {
		return gocsi.ErrNodePublishVolume(
				csi.Error_NodePublishVolumeError_MOUNT_ERROR,
				"missing device token from publish volume info"),
			nil
	}

	opts := &apitypes.WaitForDeviceOpts{
		Token:   token,
		Timeout: WFDTimeout * time.Second,
		LocalDevicesOpts: apitypes.LocalDevicesOpts{
			Opts:     apiutils.NewStore(),
			ScanType: apitypes.DeviceScanQuick,
		},
	}
	found, devs, err := d.client.Executor().WaitForDevice(d.ctx, opts)
	if err != nil {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_MOUNT_ERROR,
			fmt.Sprintf("device not found: %v", err)), nil
	}
	if !found {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_MOUNT_ERROR,
			fmt.Sprintf("device not found: token=%s", token)), nil
	}
	dev, ok := devs.DeviceMap[token]
	if !ok {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_MOUNT_ERROR,
			fmt.Sprintf("device not found: token=%s: missing", token)), nil
	}

	if mv := req.VolumeCapability.GetMount(); mv != nil {
		if !st.IsDir() {
			return nil, errDirNeeded
		}
		return d.handleMountVolume(
			volumeID, target, mv, mv.MountFlags, dev, req.Readonly)
	}
	if bv := req.VolumeCapability.GetBlock(); bv != nil {
		if st.IsDir() {
			return nil, errFileNeeded
		}
		return d.handleBlockVolume(target, dev)
	}

	return gocsi.ErrNodePublishVolume(
		csi.Error_NodePublishVolumeError_UNSUPPORTED_VOLUME_TYPE,
		"No supported volume type received"), nil
}

func (d *driver) NodeUnpublishVolume(
	ctx xctx.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	volumeID, ok := req.VolumeId.Values["id"]
	if !ok {
		return gocsi.ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_INVALID_VOLUME_ID,
			`missing "id" field`), nil
	}

	target := req.GetTargetPath()

	opts := &apitypes.VolumeInspectOpts{Opts: apiutils.NewStore()}
	opts.Attachments = apitypes.VolAttReqWithDevMapOnlyVolsAttachedToInstance
	vol, err := d.client.Storage().VolumeInspect(d.ctx, volumeID, opts)
	if err != nil {
		if err.Error() == "resource not found" {
			// Volume is not attached
			return &csi.NodeUnpublishVolumeResponse{
				Reply: &csi.NodeUnpublishVolumeResponse_Result_{
					Result: &csi.NodeUnpublishVolumeResponse_Result{},
				},
			}, nil
		}
		return nil, err
	}
	if vol == nil {
		return &csi.NodeUnpublishVolumeResponse{
			Reply: &csi.NodeUnpublishVolumeResponse_Result_{
				Result: &csi.NodeUnpublishVolumeResponse_Result{},
			},
		}, nil
	}
	dev := vol.Attachments[0].DeviceName
	mntPath := d.config.GetString(apitypes.ConfigIgVolOpsMountPath)
	privTgt := path.Join(mntPath, vol.Name)

	mnts, err := mount.GetMounts()
	if err != nil {
		return gocsi.ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_UNMOUNT_ERROR,
			err.Error()), nil
	}

	mt := false
	mp := false
	bt := false
	for _, m := range mnts {
		if m.Device == dev {
			if m.Path == privTgt {
				mp = true
			} else if m.Path == target {
				mt = true
			}
		}
		if m.Device == devtmpfs && m.Path == target {
			bt = true
		}
	}

	if mt || bt {
		if err := mount.Unmount(target); err != nil {
			return gocsi.ErrNodeUnpublishVolume(
				csi.Error_NodeUnpublishVolumeError_UNMOUNT_ERROR,
				err.Error()), nil
		}
	}

	if mp {
		if err := d.unmountPrivMount(dev, privTgt); err != nil {
			return gocsi.ErrNodeUnpublishVolume(
				csi.Error_NodeUnpublishVolumeError_UNMOUNT_ERROR,
				err.Error()), nil
		}
	}

	return &csi.NodeUnpublishVolumeResponse{
		Reply: &csi.NodeUnpublishVolumeResponse_Result_{
			Result: &csi.NodeUnpublishVolumeResponse_Result{},
		},
	}, nil
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

func (d *driver) unmountPrivMount(
	device string,
	target string) error {

	d.ctx.WithField("path", target).Debug(
		"checking if private mount can be unmounted")

	mnts, err := mount.GetDevMounts(device)
	if err != nil {
		return err
	}

	// remove private mount if we can
	if len(mnts) == 1 && mnts[0].Path == target {
		if err := mount.Unmount(target); err != nil {
			return err
		}
		os.Remove(target)
	}
	return nil
}

func (d *driver) handleMountVolume(
	volumeID string,
	target string,
	mv *csi.VolumeCapability_MountVolume,
	mf []string,
	dev string,
	ro bool) (*csi.NodePublishVolumeResponse, error) {

	// Format and mount and device via integration driver
	opts := &apitypes.VolumeMountOpts{Opts: apiutils.NewStore()}
	fs := mv.GetFsType()
	if fs == "" {
		fs = d.config.GetString(apitypes.ConfigIgVolOpsCreateDefaultFsType)
	}
	opts.NewFSType = fs

	mctx := d.ctx.WithValue(ctxExactMountKey, true)

	mnt, _, err := d.client.Integration().Mount(
		mctx, volumeID, "", opts)
	if err != nil {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_MOUNT_ERROR,
			err.Error()), nil
	}

	mnts, err := mount.GetDevMounts(dev)
	if err != nil {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_MOUNT_ERROR,
			"could not reliably determine existing mount status"), nil
	}

	// Private mount in place, now bind mount to target path
	// If mounts already existed for this device, check if mount to
	// target path was already there
	if len(mnts) > 0 {
		for _, m := range mnts {
			if m.Path == target {
				// volume already published to target
				// if mount options look good, do nothing
				rwo := "rw"
				if ro {
					rwo = "ro"
				}
				if !contains(m.Opts, rwo) {
					return gocsi.ErrNodePublishVolume(
						csi.Error_NodePublishVolumeError_MOUNT_ERROR,
						"volume previously published with different options"), nil

				}
				// Existing mount satisfied requested
				return &csi.NodePublishVolumeResponse{
					Reply: &csi.NodePublishVolumeResponse_Result_{
						Result: &csi.NodePublishVolumeResponse_Result{},
					},
				}, nil
			}
		}

	}
	// bind mount to target
	if ro {
		mf = append(mf, "ro")
	}
	if err := mount.BindMount(mnt, target, mf...); err != nil {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_MOUNT_ERROR,
			err.Error()), nil
	}
	return &csi.NodePublishVolumeResponse{
		Reply: &csi.NodePublishVolumeResponse_Result_{
			Result: &csi.NodePublishVolumeResponse_Result{},
		},
	}, nil
}

func (d *driver) handleBlockVolume(
	target string,
	dev string) (*csi.NodePublishVolumeResponse, error) {

	f := map[string]interface{}{
		"target": target,
		"device": dev,
	}

	// Check if device is already mounted
	mnts, err := mount.GetDevMounts(dev)
	if err != nil {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_MOUNT_ERROR,
			"could not reliably determine existing mount status"), nil
	}

	if len(mnts) > 0 {
		// Our devices is mounted *somewhere* If we find it mounted to
		// our target path, return success for the existing mount
		// Other mounts are ignored, and left for the sysadmin and
		// clustered filesystems to handleBlockVolume
		for _, m := range mnts {
			if m.Path == target {
				// Existing mount satisfies request
				d.ctx.WithFields(f).Debug("mount already in place")
				return &csi.NodePublishVolumeResponse{
					Reply: &csi.NodePublishVolumeResponse_Result_{
						Result: &csi.NodePublishVolumeResponse_Result{},
					},
				}, nil
			}
		}
	}

	if err := mount.BindMount(dev, target, ""); err != nil {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_MOUNT_ERROR,
			err.Error()), nil
	}
	return &csi.NodePublishVolumeResponse{
		Reply: &csi.NodePublishVolumeResponse_Result_{
			Result: &csi.NodePublishVolumeResponse_Result{},
		},
	}, nil
}

func contains(list []string, item string) bool {
	for _, x := range list {
		if x == item {
			return true
		}
	}
	return false
}
