package gocsi

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi/csi"
)

const (
	// FMGetSupportedVersions is the full method name for the
	// eponymous RPC message.
	FMGetSupportedVersions = "/" + Namespace +
		".Identity/" +
		"GetSupportedVersions"

	// FMGetPluginInfo is the full method name for the
	// eponymous RPC message.
	FMGetPluginInfo = "/" + Namespace +
		".Identity/" +
		"GetPluginInfo"
)

// GetSupportedVersions issues a
// GetSupportedVersions request
// to a CSI controller.
func GetSupportedVersions(
	ctx context.Context,
	c csi.IdentityClient,
	callOpts ...grpc.CallOption) ([]*csi.Version, error) {

	req := &csi.GetSupportedVersionsRequest{}

	res, err := c.GetSupportedVersions(ctx, req, callOpts...)
	if err != nil {
		return nil, err
	}

	return res.GetResult().SupportedVersions, nil
}

// GetPluginInfo issues a
// GetPluginInfo request
// to a CSI controller.
func GetPluginInfo(
	ctx context.Context,
	c csi.IdentityClient,
	version *csi.Version,
	callOpts ...grpc.CallOption) (*csi.GetPluginInfoResponse_Result, error) {

	req := &csi.GetPluginInfoRequest{
		Version: version,
	}

	res, err := c.GetPluginInfo(ctx, req, callOpts...)
	if err != nil {
		return nil, err
	}

	return res.GetResult(), nil
}
