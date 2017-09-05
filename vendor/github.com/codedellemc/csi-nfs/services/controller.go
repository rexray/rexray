package services

import (
	"golang.org/x/net/context"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
)

func (s *StoragePlugin) ControllerGetCapabilities(
	ctx context.Context,
	in *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {

	return &csi.ControllerGetCapabilitiesResponse{
		Reply: &csi.ControllerGetCapabilitiesResponse_Result_{
			Result: &csi.ControllerGetCapabilitiesResponse_Result{
				Capabilities: []*csi.ControllerServiceCapability{},
			},
		},
	}, nil
}

func (s *StoragePlugin) CreateVolume(
	ctx context.Context,
	in *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	return gocsi.ErrCreateVolume(
		csi.Error_CreateVolumeError_CALL_NOT_IMPLEMENTED,
		"CreateVolume not valid for NFS"), nil
}

func (s *StoragePlugin) DeleteVolume(
	ctx context.Context,
	in *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	return gocsi.ErrDeleteVolume(
		csi.Error_DeleteVolumeError_CALL_NOT_IMPLEMENTED,
		"DeleteVolume not valid for NFS"), nil
}

func (s *StoragePlugin) ControllerPublishVolume(
	ctx context.Context,
	in *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {

	return gocsi.ErrControllerPublishVolume(
		csi.Error_ControllerPublishVolumeError_CALL_NOT_IMPLEMENTED,
		"ControllerPublishVolume not valid for NFS"), nil
}

func (s *StoragePlugin) ControllerUnpublishVolume(
	ctx context.Context,
	in *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {

	return gocsi.ErrControllerUnpublishVolume(
		csi.Error_ControllerUnpublishVolumeError_CALL_NOT_IMPLEMENTED,
		"ControllerUnpublishVolume not valid for NFS"), nil
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

	for _, c := range in.VolumeCapabilities {
		if t := c.GetBlock(); t != nil {
			r.GetResult().Supported = false
			break
		}
		if t := c.GetMount(); t != nil {
			// If a filesystem is given, it must be NFS
			fs := t.GetFsType()
			if fs != "" && fs != "nfs" {
				r.GetResult().Supported = false
				break
			}
			// TODO: Check mount flags
			//for _, f := range t.GetMountFlags() {}

		}
	}

	return r, nil
}

func (s *StoragePlugin) ListVolumes(
	ctx context.Context,
	in *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {

	return gocsi.ErrListVolumes(
		csi.Error_GeneralError_UNDEFINED,
		"ListVolumes not implemented for NFS"), nil
}

func (s *StoragePlugin) GetCapacity(
	ctx context.Context,
	in *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {

	return gocsi.ErrGetCapacity(
		csi.Error_GeneralError_UNDEFINED,
		"GetCapacity not implemented for NFS"), nil
}
