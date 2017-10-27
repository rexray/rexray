package gocsi

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi/csi"
)

// ChainUnaryClient chains one or more unary, client interceptors
// together into a left-to-right series that can be provided to a
// new gRPC client.
func ChainUnaryClient(
	i ...grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor {

	switch len(i) {
	case 0:
		return func(
			ctx context.Context,
			method string,
			req, rep interface{},
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, rep, cc, opts...)
		}
	case 1:
		return i[0]
	}

	return func(
		ctx context.Context,
		method string,
		req, rep interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		bc := func(
			cur grpc.UnaryClientInterceptor,
			nxt grpc.UnaryInvoker) grpc.UnaryInvoker {

			return func(
				curCtx context.Context,
				curMethod string,
				curReq, curRep interface{},
				curCC *grpc.ClientConn,
				curOpts ...grpc.CallOption) error {

				return cur(
					curCtx,
					curMethod,
					curReq, curRep,
					curCC, nxt,
					curOpts...)
			}
		}

		c := invoker
		for j := len(i) - 1; j >= 0; j-- {
			c = bc(i[j], c)
		}

		return c(ctx, method, req, rep, cc, opts...)
	}
}

// ClientCheckReponseError is a unary, client validator that checks a
// reply's message to see if it contains an error and transforms it
// into an *Error object, which adheres to Go's Error interface.
func ClientCheckReponseError(
	ctx context.Context,
	method string,
	req, rep interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	// Invoke the call and check the reply for an error.
	if err := invoker(ctx, method, req, rep, cc, opts...); err != nil {
		return &Error{
			FullMethod: method,
			InnerError: err,
		}
	}

	switch trep := rep.(type) {

	// Controller
	case *csi.CreateVolumeResponse:
		if err := CheckResponseErrCreateVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.DeleteVolumeResponse:
		if err := CheckResponseErrDeleteVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ControllerPublishVolumeResponse:
		if err := CheckResponseErrControllerPublishVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ControllerUnpublishVolumeResponse:
		if err := CheckResponseErrControllerUnpublishVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ValidateVolumeCapabilitiesResponse:
		if err := CheckResponseErrValidateVolumeCapabilities(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ListVolumesResponse:
		if err := CheckResponseErrListVolumes(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.GetCapacityResponse:
		if err := CheckResponseErrGetCapacity(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ControllerGetCapabilitiesResponse:
		if err := CheckResponseErrControllerGetCapabilities(
			ctx, method, trep); err != nil {
			return err
		}

	// Identity
	case *csi.GetSupportedVersionsResponse:
		if err := CheckResponseErrGetSupportedVersions(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.GetPluginInfoResponse:
		if err := CheckResponseErrGetPluginInfo(
			ctx, method, trep); err != nil {
			return err
		}

	// Node
	case *csi.NodePublishVolumeResponse:
		if err := CheckResponseErrNodePublishVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.NodeUnpublishVolumeResponse:
		if err := CheckResponseErrNodeUnpublishVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.GetNodeIDResponse:
		if err := CheckResponseErrGetNodeID(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ProbeNodeResponse:
		if err := CheckResponseErrProbeNode(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.NodeGetCapabilitiesResponse:
		if err := CheckResponseErrNodeGetCapabilities(
			ctx, method, trep); err != nil {
			return err
		}
	}

	return nil
}

// ClientResponseValidator is a unary, client validator for validating
// replies from a CSI plug-in.
func ClientResponseValidator(
	ctx context.Context,
	method string,
	req, rep interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	// Invoke the call and validate the reply.
	if err := invoker(ctx, method, req, rep, cc, opts...); err != nil {
		return &Error{
			FullMethod: method,
			InnerError: err,
		}
	}

	// Do not validate the reply if it has an error.
	if trep, ok := rep.(hasGetError); ok {
		if trep.GetError() != nil {
			return nil
		}
	}

	switch trep := rep.(type) {

	// Controller
	case *csi.CreateVolumeResponse:
		if err := crepvCreateVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.DeleteVolumeResponse:
		if err := crepvDeleteVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ControllerPublishVolumeResponse:
		if err := crepvControllerPublishVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ControllerUnpublishVolumeResponse:
		if err := crepvControllerUnpublishVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ValidateVolumeCapabilitiesResponse:
		if err := crepvValidateVolumeCapabilities(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ListVolumesResponse:
		if err := crepvListVolumes(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.GetCapacityResponse:
		if err := crepvGetCapacity(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ControllerGetCapabilitiesResponse:
		if err := crepvControllerGetCapabilities(
			ctx, method, trep); err != nil {
			return err
		}

	// Identity
	case *csi.GetSupportedVersionsResponse:
		if err := crepvGetSupportedVersions(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.GetPluginInfoResponse:
		if err := crepvGetPluginInfo(
			ctx, method, trep); err != nil {
			return err
		}

	// Node
	case *csi.NodePublishVolumeResponse:
		if err := crepvNodePublishVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.NodeUnpublishVolumeResponse:
		if err := crepvNodeUnpublishVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.GetNodeIDResponse:
		if err := crepvGetNodeID(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ProbeNodeResponse:
		if err := crepvProbeNode(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.NodeGetCapabilitiesResponse:
		if err := crepvNodeGetCapabilities(
			ctx, method, trep); err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
//                     CLIENT RESPONSE - CONTROLLER                           //
////////////////////////////////////////////////////////////////////////////////

func crepvCreateVolume(
	ctx context.Context,
	method string,
	rep *csi.CreateVolumeResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	if rep.GetResult().VolumeInfo == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilVolumeInfo,
		}
	}

	if rep.GetResult().VolumeInfo.Id == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilVolumeID,
		}
	}

	return nil
}

func crepvDeleteVolume(
	ctx context.Context,
	method string,
	rep *csi.DeleteVolumeResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	return nil
}

func crepvControllerPublishVolume(
	ctx context.Context,
	method string,
	rep *csi.ControllerPublishVolumeResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	if rep.GetResult().PublishVolumeInfo == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilPublishVolumeInfo,
		}
	}

	if len(rep.GetResult().PublishVolumeInfo.Values) == 0 {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrEmptyPublishVolumeInfo,
		}
	}

	return nil
}

func crepvControllerUnpublishVolume(
	ctx context.Context,
	method string,
	rep *csi.ControllerUnpublishVolumeResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	return nil
}

func crepvValidateVolumeCapabilities(
	ctx context.Context,
	method string,
	rep *csi.ValidateVolumeCapabilitiesResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	return nil
}

func crepvListVolumes(
	ctx context.Context,
	method string,
	rep *csi.ListVolumesResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	return nil
}

func crepvGetCapacity(
	ctx context.Context,
	method string,
	rep *csi.GetCapacityResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	return nil
}

func crepvControllerGetCapabilities(
	ctx context.Context,
	method string,
	rep *csi.ControllerGetCapabilitiesResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
//                       CLIENT RESPONSE - IDENTITY                           //
////////////////////////////////////////////////////////////////////////////////

func crepvGetSupportedVersions(
	ctx context.Context,
	method string,
	rep *csi.GetSupportedVersionsResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	return nil
}

func crepvGetPluginInfo(
	ctx context.Context,
	method string,
	rep *csi.GetPluginInfoResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
//                        CLIENT RESPONSE - NODE                              //
////////////////////////////////////////////////////////////////////////////////

func crepvNodePublishVolume(
	ctx context.Context,
	method string,
	rep *csi.NodePublishVolumeResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	return nil
}

func crepvNodeUnpublishVolume(
	ctx context.Context,
	method string,
	rep *csi.NodeUnpublishVolumeResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	return nil
}

func crepvGetNodeID(
	ctx context.Context,
	method string,
	rep *csi.GetNodeIDResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	if rep.GetResult().NodeId == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilNodeID,
		}
	}

	return nil
}

func crepvProbeNode(
	ctx context.Context,
	method string,
	rep *csi.ProbeNodeResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	return nil
}

func crepvNodeGetCapabilities(
	ctx context.Context,
	method string,
	rep *csi.NodeGetCapabilitiesResponse) error {

	if rep.GetResult() == nil {
		return &Error{
			Code:       ErrorNoCode,
			FullMethod: method,
			InnerError: ErrNilResult,
		}
	}

	return nil
}
