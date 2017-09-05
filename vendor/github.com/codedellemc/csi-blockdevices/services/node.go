package services

import (
	"os"
	"path/filepath"

	"golang.org/x/net/context"

	log "github.com/sirupsen/logrus"

	"github.com/codedellemc/csi-blockdevices/block"
	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
)

func (s *StoragePlugin) NodePublishVolume(
	ctx context.Context,
	in *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {

	idm := in.GetVolumeId().GetValues()
	target := in.GetTargetPath()
	ro := in.GetReadonly()
	vc := in.GetVolumeCapability()
	am := vc.GetAccessMode()

	name, ok := idm["name"]
	if !ok {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_INVALID_VOLUME_ID,
			"name key missing from volumeID"), nil
	}

	dev, err := block.GetDeviceInDir(s.DevDir, name)
	if err != nil {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_VOLUME_DOES_NOT_EXIST,
			err.Error()), nil
	}

	if mv := vc.GetMount(); mv != nil {
		fs := mv.GetFsType()
		mf := mv.GetMountFlags()

		return s.handleMountVolume(dev, target, fs, mf, ro, am)
	}
	if bv := vc.GetBlock(); bv != nil {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_UNSUPPORTED_VOLUME_TYPE,
			"Block type not yet implemented"), nil

	}

	return gocsi.ErrNodePublishVolume(
		csi.Error_NodePublishVolumeError_UNSUPPORTED_VOLUME_TYPE,
		"No supported volume type received"), nil
}

func (s *StoragePlugin) NodeUnpublishVolume(
	ctx context.Context,
	in *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {

	idm := in.GetVolumeId().GetValues()
	target := in.GetTargetPath()

	name, ok := idm["name"]
	if !ok {
		return gocsi.ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_INVALID_VOLUME_ID,
			"name key missing from volumeID"), nil
	}

	dev, err := block.GetDeviceInDir(s.DevDir, name)
	if err != nil {
		return gocsi.ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_VOLUME_DOES_NOT_EXIST,
			err.Error()), nil
	}

	// check to see if volume is really mounted at target
	mnts, err := block.GetDevMounts(dev.RealDev)
	if err != nil {
		return gocsi.ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_UNMOUNT_ERROR,
			err.Error()), nil
	}

	if len(mnts) > 0 {
		// device is mounted somewhere. could be target, other targets,
		// or private mount
		var (
			idx       int
			m         block.MountPoint
			unmounted = false
		)
		for idx, m = range mnts {
			if m.Path == target {
				if err := block.Unmount(target); err != nil {
					return gocsi.ErrNodeUnpublishVolume(
						csi.Error_NodeUnpublishVolumeError_UNMOUNT_ERROR,
						err.Error()), nil
				}
				unmounted = true
				break
			}
		}
		if unmounted {
			mnts = append(mnts[:idx], mnts[idx+1:]...)
		}
	}

	// remove private mount if we can
	privTgt := s.getPrivateMountPoint(dev)
	if len(mnts) == 1 && mnts[0].Path == privTgt {
		if err := block.Unmount(privTgt); err != nil {
			return gocsi.ErrNodeUnpublishVolume(
				csi.Error_NodeUnpublishVolumeError_UNMOUNT_ERROR,
				err.Error()), nil
		}
		os.Remove(privTgt)
	}

	// TODO handle block volume type (links)

	return &csi.NodeUnpublishVolumeResponse{
		Reply: &csi.NodeUnpublishVolumeResponse_Result_{
			Result: &csi.NodeUnpublishVolumeResponse_Result{},
		},
	}, nil
}

func (s *StoragePlugin) GetNodeID(
	ctx context.Context,
	in *csi.GetNodeIDRequest) (*csi.GetNodeIDResponse, error) {

	return &csi.GetNodeIDResponse{
		Reply: &csi.GetNodeIDResponse_Result_{
			// Return nil ID because it's not used by the
			// controller
			Result: &csi.GetNodeIDResponse_Result{},
		},
	}, nil
}

func (s *StoragePlugin) ProbeNode(
	ctx context.Context,
	in *csi.ProbeNodeRequest) (*csi.ProbeNodeResponse, error) {

	if err := block.Supported(); err != nil {
		return gocsi.ErrProbeNode(
			csi.Error_ProbeNodeError_MISSING_REQUIRED_HOST_DEPENDENCY,
			err.Error()), nil
	}

	return &csi.ProbeNodeResponse{
		Reply: &csi.ProbeNodeResponse_Result_{
			Result: &csi.ProbeNodeResponse_Result{},
		},
	}, nil
}

func (s *StoragePlugin) NodeGetCapabilities(
	ctx context.Context,
	in *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {

	return &csi.NodeGetCapabilitiesResponse{
		Reply: &csi.NodeGetCapabilitiesResponse_Result_{
			Result: &csi.NodeGetCapabilitiesResponse_Result{
				Capabilities: []*csi.NodeServiceCapability{},
			},
		},
	}, nil
}

// mkdir creates the directory specified by path if needed.
// return pair is a bool flag of whether dir was created, and an error
func mkdir(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			log.WithField("dir", path).WithError(
				err).Error("Unable to create dir")
			return false, err
		}
		log.WithField("path", path).Debug("created directory")
		return true, nil
	}
	return false, nil
}

func (s *StoragePlugin) handleMountVolume(
	dev *block.Device,
	target string,
	fs string,
	mf []string,
	ro bool,
	am *csi.VolumeCapability_AccessMode) (*csi.NodePublishVolumeResponse, error) {

	// Make sure PrivDir exists
	if _, err := mkdir(s.privDir); err != nil {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_MOUNT_ERROR,
			"Unable to create private mount dir"), nil
	}

	// Path to mount device to
	privTgt := s.getPrivateMountPoint(dev)

	f := log.Fields{
		"name":         dev.Name,
		"volumePath":   dev.FullPath,
		"device":       dev.RealDev,
		"target":       target,
		"privateMount": privTgt,
	}

	// Check if device is already mounted
	mnts, err := block.GetDevMounts(dev.RealDev)
	if err != nil {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_MOUNT_ERROR,
			"could not reliably determine existing mount status"), nil
	}

	if len(mnts) == 0 {
		// Device isn't mounted anywhere, do the private mount
		log.WithFields(f).Debug("attempting mount to private area")

		// Make sure private mount point exists
		created, err := mkdir(privTgt)
		if err != nil {
			return gocsi.ErrNodePublishVolume(
				csi.Error_NodePublishVolumeError_MOUNT_ERROR,
				"Unable to create private mount point"), nil
		}
		if !created {
			log.WithFields(f).Debug("private mount target already exists")

			// The place where our device is supposed to be mounted
			// already exists, but we also know that our device is not mounted anywhere
			// Either something didn't clean up correctly, or something else is mounted
			// If the directory is not in use, it's okay to re-use it. But make sure
			// it's not in use first

			for _, m := range mnts {
				if m.Path == privTgt {
					log.WithFields(f).WithField("mountedDevice", m.Device).Error(
						"mount point already in use by device")
					return gocsi.ErrNodePublishVolume(
						csi.Error_NodePublishVolumeError_MOUNT_ERROR,
						"Unable to use private mount point"), nil
				}
			}
		}

		// If read-only access mode, we don't allow formatting
		if am.GetMode() == csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY {
			mf = append(mf, "ro")
			if err := block.Mount(dev.FullPath, privTgt, fs, mf); err != nil {
				return gocsi.ErrNodePublishVolume(
					csi.Error_NodePublishVolumeError_MOUNT_ERROR,
					err.Error()), nil
			}
		} else if am.GetMode() == csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER {
			if err := block.FormatAndMount(dev.FullPath, privTgt, fs, mf); err != nil {
				return gocsi.ErrNodePublishVolume(
					csi.Error_NodePublishVolumeError_MOUNT_ERROR,
					err.Error()), nil
			}
		} else {
			return gocsi.ErrNodePublishVolume(
				csi.Error_NodePublishVolumeError_MOUNT_ERROR,
				"Invalid access mode"), nil
		}
	} else {
		// Device is already mounted. Need to ensure that it is already
		// mounted to the expected private mount, with correct rw/ro perms
		mounted := false
		for _, m := range mnts {
			if m.Path == privTgt {
				mounted = true
				rwo := "rw"
				if am.GetMode() == csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY {
					rwo = "ro"
				}
				if contains(m.Opts, rwo) {
					break
				} else {
					return gocsi.ErrNodePublishVolume(
						csi.Error_NodePublishVolumeError_MOUNT_ERROR,
						"access mode conflicts with existing mounts"), nil
				}
			}
		}
		if !mounted {
			return gocsi.ErrNodePublishVolume(
				csi.Error_NodePublishVolumeError_MOUNT_ERROR,
				"device in use by external entity"), nil
		}
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
	if ro {
		mf = append(mf, "ro")
	}
	mf = append(mf, "bind")
	if err := block.Mount(privTgt, target, "", mf); err != nil {
		//if err := SafeUnmnt(privTgt); err != nil {
		//	log.WithFields(f).WithError(err).Error(
		//		"Unable to umount from private dir")
		//}
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_MOUNT_ERROR,
			err.Error()), nil
	}

	// Mount successful
	return &csi.NodePublishVolumeResponse{
		Reply: &csi.NodePublishVolumeResponse_Result_{
			Result: &csi.NodePublishVolumeResponse_Result{},
		},
	}, nil
}

func (s *StoragePlugin) getPrivateMountPoint(dev *block.Device) string {
	return filepath.Join(s.privDir, dev.Name)
}
