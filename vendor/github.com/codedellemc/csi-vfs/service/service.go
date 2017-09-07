package service

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
	"github.com/codedellemc/gocsi/mount"
)

const (
	// Name is the name of the CSI plug-in.
	Name = "csi-vfs"

	// BindFSEnvVar is the name of the environment variable
	// used to obtain the path to the `bindfs` binary -- a
	// program used to provide bind mounting via FUSE on
	// operating systems that do not natively support bind
	// mounts.
	//
	// If no value is specified and `bindfs` is required, ex.
	// darwin, then `bindfs` is looked up via the path.
	BindFSEnvVar = "X_CSI_VFS_BINDFS"

	// DataDirEnvVar is the name of the environment variable
	// used to obtain the path to the VFS plug-in's data directory.
	//
	// If not specified, the directory defaults to `$HOME/.csi-vfs`.
	DataDirEnvVar = "X_CSI_VFS_DATA"

	// DevDirEnvVar is the name of the environment variable
	// used to obtain the path to the VFS plug-in's `dev` directory.
	//
	// If not specified, the directory defaults to `$X_CSI_VFS_DATA/dev`.
	DevDirEnvVar = "X_CSI_VFS_DEV"

	// MntDirEnvVar is the name of the environment variable
	// used to obtain the path to the VFS plug-in's `mnt` directory.
	//
	// If not specified, the directory defaults to `$X_CSI_VFS_DATA/mnt`.
	MntDirEnvVar = "X_CSI_VFS_MNT"

	// VolDirEnvVar is the name of the environment variable
	// used to obtain the path to the VFS plug-in's `vol` directory.
	//
	// If not specified, the directory defaults to `$X_CSI_VFS_DATA/vol`.
	VolDirEnvVar = "X_CSI_VFS_VOL"

	// VolGlobEnvVar is the name of the environment variable
	// used to obtain the glob pattern used to list the files inside
	// the $X_CSI_VFS_VOL directory. Matching files are considered
	// volumes.
	//
	// If not specified, the glob pattern defaults to `*`.
	//
	// Valid patterns are documented at
	// https://golang.org/pkg/path/filepath/#Match.
	VolGlobEnvVar = "X_CSI_VFS_VOL_GLOB"
)

var (
	// SupportedVersions is a list of the versions this CSI plug-in supports.
	SupportedVersions = []*csi.Version{
		&csi.Version{
			Major: 0,
			Minor: 0,
			Patch: 0,
		},
		&csi.Version{
			Major: 0,
			Minor: 1,
			Patch: 0,
		},
	}
)

// Service is the CSI Virtual File System (VFS) service provider.
type Service interface {
	csi.ControllerServer
	csi.IdentityServer
	csi.NodeServer
}

type service struct {
	bindfs  string
	data    string
	dev     string
	mnt     string
	vol     string
	volGlob string
}

// New returns a new Service using the specified path for the
// plug-in's root data directory.
func New(data, dev, mnt, vol, volGlob, bindfs string) Service {

	InitConfig(&data, &dev, &mnt, &vol, &volGlob, &bindfs)

	log.WithFields(map[string]interface{}{
		"data":    data,
		"dev":     dev,
		"mnt":     mnt,
		"vol":     vol,
		"volGlob": volGlob,
		"bindfs":  bindfs,
	}).Info("created new " + Name + " service")

	return &service{
		data:    data,
		dev:     dev,
		mnt:     mnt,
		vol:     vol,
		volGlob: volGlob,
		bindfs:  bindfs,
	}
}

////////////////////////////////////////////////////////////////////////////////
//                            Controller Service                              //
////////////////////////////////////////////////////////////////////////////////

func (s *service) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	volPath := path.Join(s.vol, req.Name)
	if err := os.MkdirAll(volPath, 0755); err != nil {
		return nil, err
	}

	log.WithField("volPath", volPath).Info("created new volume")
	return &csi.CreateVolumeResponse{
		Reply: &csi.CreateVolumeResponse_Result_{
			Result: &csi.CreateVolumeResponse_Result{
				VolumeInfo: &csi.VolumeInfo{
					Id: &csi.VolumeID{
						Values: map[string]string{
							"path": volPath,
						},
					},
				},
			},
		},
	}, nil
}

func (s *service) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {

	// Get the path to the volume from the volume ID.
	volPath, ok := req.VolumeId.Values["path"]
	if !ok {
		return gocsi.ErrDeleteVolume(
			csi.Error_DeleteVolumeError_INVALID_VOLUME_ID,
			""), nil
	}

	// If the volume does not exist then return an error.
	if !FileExists(volPath) {
		return gocsi.ErrDeleteVolume(
			csi.Error_DeleteVolumeError_VOLUME_DOES_NOT_EXIST,
			volPath), nil
	}

	// Attempt to delete the "device".
	if err := os.RemoveAll(volPath); err != nil {
		log.WithField("path", volPath).WithError(err).Error(
			"delete directory failed")
		return nil, err
	}

	// Indicate the operation was a success.
	log.WithField("volPath", volPath).Info("deleted volume")
	return &csi.DeleteVolumeResponse{
		Reply: &csi.DeleteVolumeResponse_Result_{
			Result: &csi.DeleteVolumeResponse_Result{},
		},
	}, nil
}

func (s *service) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	volPath, ok := req.VolumeId.Values["path"]
	if !ok {
		return gocsi.ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_INVALID_VOLUME_ID,
			""), nil
	}

	if !FileExists(volPath) {
		return gocsi.ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_VOLUME_DOES_NOT_EXIST,
			volPath), nil
	}

	volName := path.Base(volPath)
	devPath := path.Join(s.dev, volName)

	// If the private mount directory for the device does not exist then
	// create it.
	if !FileExists(devPath) {
		if err := os.MkdirAll(devPath, 0755); err != nil {
			log.WithField("path", devPath).WithError(err).Error(
				"create device dir failed")
			return nil, err
		}
		log.WithField("path", devPath).Info("created device dir")
	}

	// Get the mount info to determine if the volume dir is already
	// bind mounted to the device dir.
	minfo, err := mount.GetMounts()
	if err != nil {
		log.WithError(err).Error("failed to get mount info")
		return nil, err
	}
	mounted := false
	for _, i := range minfo {
		// If bindfs is not used then the device path will not match
		// the volume path, otherwise test both the source and target.
		if i.Source == volPath && i.Path == devPath {
			mounted = true
			break
		}
	}

	if mounted {
		log.WithField("path", devPath).Info("already bind mounted")
	} else {
		if err := mount.BindMount(volPath, devPath); err != nil {
			log.WithField("name", volName).WithError(err).Error(
				"bind mount failed")
			return nil, err
		}
		log.WithField("path", devPath).Info("bind mounted volume to device")
	}

	return &csi.ControllerPublishVolumeResponse{
		Reply: &csi.ControllerPublishVolumeResponse_Result_{
			Result: &csi.ControllerPublishVolumeResponse_Result{
				PublishVolumeInfo: &csi.PublishVolumeInfo{
					Values: map[string]string{
						"path": devPath,
					},
				},
			},
		},
	}, nil
}

func (s *service) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	volPath, ok := req.VolumeId.Values["path"]
	if !ok {
		return gocsi.ErrControllerUnpublishVolume(
			csi.Error_ControllerUnpublishVolumeError_INVALID_VOLUME_ID,
			""), nil
	}

	if !FileExists(volPath) {
		return gocsi.ErrControllerUnpublishVolume(
			csi.Error_ControllerUnpublishVolumeError_VOLUME_DOES_NOT_EXIST,
			volPath), nil
	}

	volName := path.Base(volPath)
	devPath := path.Join(s.dev, volName)

	// Get the node's mount information.
	minfo, err := mount.GetMounts()
	if err != nil {
		log.WithError(err).Error("failed to get mount info")
		return nil, err
	}

	// The loop below unmounts the device path if it is mounted.
	for _, i := range minfo {
		// If there is a device that matches the volPath value and
		// a path that matches the devPath value then unmount it as
		// it is the subject of this request.
		if i.Source == volPath && i.Path == devPath {
			if err := mount.Unmount(devPath); err != nil {
				log.WithField("path", devPath).WithError(err).Error(
					"failed to unmount device dir")
				return nil, err
			}
		}
	}

	// If the device path exists then remove it.
	if FileExists(devPath) {
		if err := os.RemoveAll(devPath); err != nil {
			log.WithField("path", devPath).WithError(err).Error(
				"failed to remove device dir")
			return nil, err
		}
	}

	log.WithField("path", devPath).Info("unmount success")

	return &csi.ControllerUnpublishVolumeResponse{
		Reply: &csi.ControllerUnpublishVolumeResponse_Result_{
			Result: &csi.ControllerUnpublishVolumeResponse_Result{},
		},
	}, nil
}

func (s *service) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	volPath, ok := req.VolumeInfo.Id.Values["path"]
	if !ok {
		return gocsi.ErrValidateVolumeCapabilities(
			csi.Error_ValidateVolumeCapabilitiesError_INVALID_VOLUME_INFO,
			"invalid volume id"), nil
	}

	if !FileExists(volPath) {
		return gocsi.ErrValidateVolumeCapabilities(
			csi.Error_ValidateVolumeCapabilitiesError_VOLUME_DOES_NOT_EXIST,
			volPath), nil
	}

	// If any of the requested capabilities are related to block
	// devices then indicate lack of support.
	supported := true
	message := ""
	for _, vc := range req.VolumeCapabilities {
		if vc.GetBlock() != nil {
			supported = false
			message = "raw device access is not supported"
			break
		}
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Reply: &csi.ValidateVolumeCapabilitiesResponse_Result_{
			Result: &csi.ValidateVolumeCapabilitiesResponse_Result{
				Supported: supported,
				Message:   message,
			},
		},
	}, nil
}

func (s *service) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	fileNames, err := filepath.Glob(s.volGlob)
	if err != nil {
		return nil, err
	}
	entries := []*csi.ListVolumesResponse_Result_Entry{}
	for _, fname := range fileNames {
		entries = append(entries,
			&csi.ListVolumesResponse_Result_Entry{
				VolumeInfo: &csi.VolumeInfo{
					Id: &csi.VolumeID{
						Values: map[string]string{
							"path": path.Join(s.vol, fname),
						},
					},
				},
			})
	}

	return &csi.ListVolumesResponse{
		Reply: &csi.ListVolumesResponse_Result_{
			Result: &csi.ListVolumesResponse_Result{
				Entries: entries,
			},
		},
	}, nil
}

func (s *service) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	return &csi.GetCapacityResponse{
		Reply: &csi.GetCapacityResponse_Result_{
			Result: &csi.GetCapacityResponse_Result{
				AvailableCapacity: getAvailableBytes(s.dev),
			},
		},
	}, nil
}

func (s *service) ControllerGetCapabilities(
	ctx context.Context,
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

func (s *service) GetSupportedVersions(
	ctx context.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	return &csi.GetSupportedVersionsResponse{
		Reply: &csi.GetSupportedVersionsResponse_Result_{
			Result: &csi.GetSupportedVersionsResponse_Result{
				SupportedVersions: SupportedVersions,
			},
		},
	}, nil
}

func (s *service) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	return &csi.GetPluginInfoResponse{
		Reply: &csi.GetPluginInfoResponse_Result_{
			Result: &csi.GetPluginInfoResponse_Result{
				Name:          Name,
				VendorVersion: "0.1.4",
			},
		},
	}, nil
}

////////////////////////////////////////////////////////////////////////////////
//                                Node Service                                //
////////////////////////////////////////////////////////////////////////////////

func (s *service) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	volPath, ok := req.VolumeId.Values["path"]
	if !ok {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_INVALID_VOLUME_ID,
			""), nil
	}

	if !FileExists(volPath) {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_VOLUME_DOES_NOT_EXIST,
			volPath), nil
	}

	volName := path.Base(volPath)
	devPath := path.Join(s.dev, volName)
	mntPath := path.Join(s.mnt, volName)
	tgtPath := req.TargetPath
	resolveSymlink(&tgtPath)

	if !FileExists(devPath) {
		log.WithField("path", devPath).Error(
			"must call ControllerPublishVolume first")
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_UNKNOWN,
			"must call ControllerPublishVolume first"), nil
	}

	// If the private mount directory for the device does not exist then
	// create it.
	if !FileExists(mntPath) {
		if err := os.MkdirAll(mntPath, 0755); err != nil {
			log.WithField("path", mntPath).WithError(err).Error(
				"create private mount dir failed")
			return nil, err
		}
		log.WithField("path", mntPath).Info("created private mount dir")
	}

	// Get the mount info to determine if the device is already mounted
	// into the private mount directory.
	minfo, err := mount.GetMounts()
	if err != nil {
		log.WithError(err).Error("failed to get mount info")
		return nil, err
	}
	isPrivMounted := false
	isTgtMounted := false
	for _, i := range minfo {
		if i.Source == devPath && i.Path == mntPath {
			isPrivMounted = true
		}
		if i.Source == mntPath && i.Path == tgtPath {
			isTgtMounted = true
		}
	}

	success := &csi.NodePublishVolumeResponse{
		Reply: &csi.NodePublishVolumeResponse_Result_{
			Result: &csi.NodePublishVolumeResponse_Result{},
		},
	}

	if isTgtMounted {
		log.WithField("path", tgtPath).Info("already mounted")
		return success, nil
	}

	// If the devie is not already mounted into the private mount
	// area then go ahead and mount it.
	if !isPrivMounted {
		if err := mount.BindMount(devPath, mntPath); err != nil {
			log.WithField("path", mntPath).WithError(err).Error(
				"create private mount failed")
			return nil, err
		}
		log.WithField("path", mntPath).Info("created private mount")
	}

	// Create the bind mount options from the requested access mode.
	var opts []string
	if vc := req.VolumeCapability; vc != nil {
		if am := req.VolumeCapability.AccessMode; am != nil {
			switch am.Mode {
			case csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER:
				opts = []string{"rw"}
			case csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY:
				opts = []string{"ro"}
			default:
				return gocsi.ErrNodePublishVolume(
					csi.Error_NodePublishVolumeError_UNSUPPORTED_MOUNT_FLAGS,
					fmt.Sprintf("unsupported access mode: %v", am.Mode)), nil
			}
		}
	}

	// Ensure the directory for the request's target path exists.
	if err := os.MkdirAll(tgtPath, 0755); err != nil {
		return nil, err
	}

	// Bind mount the private mount to the requested target path with
	// the requested access mode.
	if err := mount.BindMount(mntPath, tgtPath, opts...); err != nil {
		return nil, err
	}

	return success, nil
}

func (s *service) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	volPath, ok := req.VolumeId.Values["path"]
	if !ok {
		return gocsi.ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_INVALID_VOLUME_ID,
			""), nil
	}

	if !FileExists(volPath) {
		return gocsi.ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_VOLUME_DOES_NOT_EXIST,
			volPath), nil
	}

	volName := path.Base(volPath)
	mntPath := path.Join(s.mnt, volName)
	tgtPath := req.TargetPath
	resolveSymlink(&tgtPath)

	// Get the node's mount information.
	minfo, err := mount.GetMounts()
	if err != nil {
		log.WithError(err).Error("failed to get mount info")
		return nil, err
	}

	// The loop below does two things:
	//
	//   1. It unmounts the target path if it is mounted.
	//   2. It counts how many times the volume is mounted.
	mountCount := 0
	for _, i := range minfo {

		// If there is a device that matches the mntPath value then
		// increment the number of times this volume is mounted on
		// this node.
		if i.Source == mntPath {
			mountCount++
		}

		// If there is a device that matches the mntPath value and
		// a path that matches the tgtPath value then unmount it as
		// it is the subject of this request.
		if i.Source == mntPath && i.Path == tgtPath {
			if err := mount.Unmount(tgtPath); err != nil {
				log.WithField("path", tgtPath).WithError(err).Error(
					"failed to unmount target path")
				return nil, err
			}
			log.WithField("path", tgtPath).Info("unmounted target path")
			mountCount--
		}
	}

	log.WithFields(map[string]interface{}{
		"name":  volName,
		"count": mountCount,
	}).Info("volume mount info")

	// If the target path exists then remove it.
	if FileExists(tgtPath) {
		if err := os.RemoveAll(tgtPath); err != nil {
			log.WithField("path", tgtPath).WithError(err).Error(
				"failed to remove target path")
			return nil, err
		}
		log.WithField("path", tgtPath).Info("removed target path")
	}

	// If the volume is no longer mounted anywhere else on this node then
	// unmount the volume's private mount as well.
	if mountCount == 0 {
		if err := mount.Unmount(mntPath); err != nil {
			log.WithField("path", mntPath).WithError(err).Error(
				"failed to unmount private mount")
			return nil, err
		}
		log.WithField("path", mntPath).Info("unmounted private mount")
		if err := os.RemoveAll(mntPath); err != nil {
			log.WithField("path", mntPath).WithError(err).Error(
				"failed to remove private mount")
			return nil, err
		}
		log.WithField("path", mntPath).Info("removed private mount")
	}

	return &csi.NodeUnpublishVolumeResponse{
		Reply: &csi.NodeUnpublishVolumeResponse_Result_{
			Result: &csi.NodeUnpublishVolumeResponse_Result{},
		},
	}, nil
}

func (s *service) GetNodeID(
	ctx context.Context,
	req *csi.GetNodeIDRequest) (
	*csi.GetNodeIDResponse, error) {

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &csi.GetNodeIDResponse{
		Reply: &csi.GetNodeIDResponse_Result_{
			Result: &csi.GetNodeIDResponse_Result{
				NodeId: &csi.NodeID{
					Values: map[string]string{"hostname": hostname},
				},
			},
		},
	}, nil
}

func (s *service) ProbeNode(
	ctx context.Context,
	req *csi.ProbeNodeRequest) (
	*csi.ProbeNodeResponse, error) {

	switch runtime.GOOS {
	case "linux":
		break
	case "darwin":
		if _, err := exec.LookPath(s.bindfs); err != nil {
			return gocsi.ErrProbeNode(
				csi.Error_ProbeNodeError_MISSING_REQUIRED_HOST_DEPENDENCY,
				s.bindfs), nil
		}
	default:
		return gocsi.ErrProbeNode(
			csi.Error_ProbeNodeError_MISSING_REQUIRED_HOST_DEPENDENCY,
			fmt.Sprintf("unsupported operating system: %s", runtime.GOOS)), nil
	}

	return &csi.ProbeNodeResponse{
		Reply: &csi.ProbeNodeResponse_Result_{
			Result: &csi.ProbeNodeResponse_Result{},
		},
	}, nil
}

func (s *service) NodeGetCapabilities(
	ctx context.Context,
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

// FileExists returns a flag indicating whether or not a file
// path exists.
func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return true
	}
	return false
}

func resolveSymlink(symPath *string) error {
	realPath, err := filepath.EvalSymlinks(*symPath)
	if err != nil {
		return err
	}
	*symPath = realPath
	return nil
}

// InitConfig initializes several CSI-VFS configuration properties.
func InitConfig(
	data, dev, mnt, vol, volGlob, bindfs *string) {

	if *data == "" {
		*data = os.Getenv(DataDirEnvVar)
	}
	if *data == "" {
		if v := os.Getenv("HOME"); v != "" {
			*data = path.Join(v, ".csi-vfs")
		} else if v := os.Getenv("USER_PROFILE"); v != "" {
			*data = path.Join(v, ".csi-vfs")
		}
	}
	os.MkdirAll(*data, 0755)
	resolveSymlink(data)

	if *dev == "" {
		*dev = os.Getenv(DevDirEnvVar)
	}
	if *dev == "" {
		*dev = path.Join(*data, "dev")
	}
	os.MkdirAll(*dev, 0755)
	resolveSymlink(dev)

	if *mnt == "" {
		*mnt = os.Getenv(MntDirEnvVar)
	}
	if *mnt == "" {
		*mnt = path.Join(*data, "mnt")
	}
	os.MkdirAll(*mnt, 0755)
	resolveSymlink(mnt)

	if *vol == "" {
		*vol = os.Getenv(VolDirEnvVar)
	}
	if *vol == "" {
		*vol = path.Join(*data, "vol")
	}
	os.MkdirAll(*vol, 0755)
	resolveSymlink(vol)

	if volGlob != nil {
		if *volGlob == "" {
			*volGlob = os.Getenv(VolGlobEnvVar)
		}
		if *volGlob == "" {
			*volGlob = "*"
		}
		*volGlob = path.Join(*dev, *volGlob)
	}

	if bindfs != nil {
		if *bindfs == "" {
			*bindfs = os.Getenv(BindFSEnvVar)
		}
		if *bindfs == "" {
			*bindfs = "bindfs"
		}
	}
}
