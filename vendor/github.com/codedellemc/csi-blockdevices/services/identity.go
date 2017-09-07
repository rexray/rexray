package services

import (
	"golang.org/x/net/context"

	"github.com/codedellemc/gocsi/csi"
)

func (s *StoragePlugin) GetSupportedVersions(
	ctx context.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	return &csi.GetSupportedVersionsResponse{
		Reply: &csi.GetSupportedVersionsResponse_Result_{
			Result: &csi.GetSupportedVersionsResponse_Result{
				SupportedVersions: CSIVersions,
			},
		},
	}, nil
}

func (s *StoragePlugin) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	return &csi.GetPluginInfoResponse{
		Reply: &csi.GetPluginInfoResponse_Result_{
			Result: &csi.GetPluginInfoResponse_Result{
				Name:          SpName,
				VendorVersion: spVersion,
				Manifest:      nil,
			},
		},
	}, nil
}
