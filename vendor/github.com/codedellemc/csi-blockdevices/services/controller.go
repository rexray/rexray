package services

import (
	"golang.org/x/net/context"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
	log "github.com/sirupsen/logrus"

	"github.com/codedellemc/csi-blockdevices/block"
)

func (s *StoragePlugin) ControllerGetCapabilities(
	ctx context.Context,
	in *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {

	return &csi.ControllerGetCapabilitiesResponse{
		Reply: &csi.ControllerGetCapabilitiesResponse_Result_{
			Result: &csi.ControllerGetCapabilitiesResponse_Result{
				Capabilities: []*csi.ControllerServiceCapability{
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

func (s *StoragePlugin) CreateVolume(
	ctx context.Context,
	in *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	return gocsi.ErrCreateVolume(
		csi.Error_CreateVolumeError_CALL_NOT_IMPLEMENTED,
		"CreateVolume not valid for Block Devices"), nil
}

func (s *StoragePlugin) DeleteVolume(
	ctx context.Context,
	in *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	return gocsi.ErrDeleteVolume(
		csi.Error_DeleteVolumeError_CALL_NOT_IMPLEMENTED,
		"DeleteVolume not valid for Block Devices"), nil
}

func (s *StoragePlugin) ControllerPublishVolume(
	ctx context.Context,
	in *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {

	return gocsi.ErrControllerPublishVolume(
		csi.Error_ControllerPublishVolumeError_CALL_NOT_IMPLEMENTED,
		"ControllerPublishVolume not valid for Block Devices"), nil
}

func (s *StoragePlugin) ControllerUnpublishVolume(
	ctx context.Context,
	in *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {

	return gocsi.ErrControllerUnpublishVolume(
		csi.Error_ControllerUnpublishVolumeError_CALL_NOT_IMPLEMENTED,
		"ControllerUnpublishVolume not valid for Block Devices"), nil
}

func (s *StoragePlugin) ValidateVolumeCapabilities(
	ctx context.Context,
	in *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	r := &csi.ValidateVolumeCapabilitiesResponse{
		Reply: &csi.ValidateVolumeCapabilitiesResponse_Result_{
			Result: &csi.ValidateVolumeCapabilitiesResponse_Result{
				Supported: true,
			},
		},
	}

	volID := in.GetVolumeInfo().GetId().GetValues()
	volName, ok := volID["name"]
	if !ok {
		return gocsi.ErrValidateVolumeCapabilities(
			csi.Error_ValidateVolumeCapabilitiesError_INVALID_VOLUME_INFO,
			"Invalid volume ID"), nil
	}

	dev, err := block.GetDeviceInDir(s.DevDir, volName)
	if err != nil {
		log.WithError(err).Error("device does not appear to exist")
		return gocsi.ErrValidateVolumeCapabilities(
			csi.Error_ValidateVolumeCapabilitiesError_VOLUME_DOES_NOT_EXIST,
			""), nil
	}

	cpcty := in.GetVolumeInfo().GetCapacityBytes()
	if cpcty != 0 {
		if cpcty > dev.Capacity {
			return &csi.ValidateVolumeCapabilitiesResponse{
				Reply: &csi.ValidateVolumeCapabilitiesResponse_Error{
					Error: &csi.Error{
						Value: &csi.Error_GeneralError_{
							GeneralError: &csi.Error_GeneralError{
								ErrorCode:          csi.Error_GeneralError_UNKNOWN,
								ErrorDescription:   "volume too small",
								CallerMustNotRetry: true,
							},
						},
					},
				},
			}, nil
		}
	}

	for _, c := range in.VolumeCapabilities {
		if t := c.GetMount(); t != nil {
			// If a filesystem is given, it must be xfs or ext4
			fs := t.GetFsType()
			if fs != "" {
				hostFSs, err := block.GetHostFileSystems("")
				if err != nil {
					return gocsi.ErrValidateVolumeCapabilities(
						csi.Error_ValidateVolumeCapabilitiesError_UNSUPPORTED_FS_TYPE,
						err.Error()), nil
				}
				if !contains(hostFSs, fs) {
					return gocsi.ErrValidateVolumeCapabilities(
						csi.Error_ValidateVolumeCapabilitiesError_UNSUPPORTED_FS_TYPE,
						"no host support for fstype"), nil
				}
			}
			// TODO: Check mount flags
			//for _, f := range t.GetMountFlags() {}
		}
		if t := c.GetAccessMode(); t != nil {
			if t.GetMode() != csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER ||
				t.GetMode() != csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY {
				return &csi.ValidateVolumeCapabilitiesResponse{
					Reply: &csi.ValidateVolumeCapabilitiesResponse_Error{
						Error: &csi.Error{
							Value: &csi.Error_GeneralError_{
								GeneralError: &csi.Error_GeneralError{
									ErrorCode:          csi.Error_GeneralError_UNKNOWN,
									ErrorDescription:   "invalid access mode",
									CallerMustNotRetry: true,
								},
							},
						},
					},
				}, nil
			}
		}
	}

	return r, nil
}

func (s *StoragePlugin) ListVolumes(
	ctx context.Context,
	in *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {

	vols, err := block.ListDevices(s.DevDir)
	if err != nil {
		return gocsi.ErrListVolumes(
			csi.Error_GeneralError_UNDEFINED,
			err.Error()), nil
	}

	entries := []*csi.ListVolumesResponse_Result_Entry{}
	for _, v := range vols {
		entries = append(entries, &csi.ListVolumesResponse_Result_Entry{
			VolumeInfo: &csi.VolumeInfo{
				Id: &csi.VolumeID{
					Values: map[string]string{
						"name": v.Name,
					},
				},
				CapacityBytes: v.Capacity,
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

func (s *StoragePlugin) GetCapacity(
	ctx context.Context,
	in *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {

	return gocsi.ErrGetCapacity(
		csi.Error_GeneralError_UNDEFINED,
		"GetCapacity not implemented for Block Devices"), nil
}

func contains(list []string, item string) bool {
	for _, x := range list {
		if x == item {
			return true
		}
	}
	return false
}
