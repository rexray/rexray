package gocsi

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sync/atomic"

	"github.com/codedellemc/gocsi/csi"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// ChainUnaryServer chains one or more unary, server interceptors
// together into a left-to-right series that can be provided to a
// new gRPC server.
func ChainUnaryServer(
	i ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {

	switch len(i) {
	case 0:
		return func(
			ctx context.Context,
			req interface{},
			_ *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	case 1:
		return i[0]
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		bc := func(
			cur grpc.UnaryServerInterceptor,
			nxt grpc.UnaryHandler) grpc.UnaryHandler {
			return func(
				curCtx context.Context,
				curReq interface{}) (interface{}, error) {
				return cur(curCtx, curReq, info, nxt)
			}
		}
		c := handler
		for j := len(i) - 1; j >= 0; j-- {
			c = bc(i[j], c)
		}
		return c(ctx, req)
	}
}

type hasGetVersion interface {
	GetVersion() *csi.Version
}

// NewServerRequestVersionValidator initializes a new unary server
// interceptor that validates request versions against the list of
// supported versions.
func NewServerRequestVersionValidator(
	supported []*csi.Version) grpc.UnaryServerInterceptor {

	return (&serverReqVersionValidator{
		supported: supported,
	}).handle
}

type serverReqVersionValidator struct {
	supported []*csi.Version
}

func (v *serverReqVersionValidator) handle(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	// Skip version validation if no supported versions are provided.
	if len(v.supported) == 0 {
		return handler(ctx, req)
	}

	treq, ok := req.(hasGetVersion)
	if !ok {
		return handler(ctx, req)
	}

	rv := treq.GetVersion()

	for _, sv := range v.supported {
		if CompareVersions(rv, sv) == 0 {
			return handler(ctx, req)
		}
	}

	msg := fmt.Sprintf(
		"unsupported request version: %s", SprintfVersion(rv))

	switch req.(type) {
	case *csi.ControllerGetCapabilitiesRequest:
		return ErrControllerGetCapabilities(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.ControllerPublishVolumeRequest:
		return ErrControllerPublishVolumeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.ControllerUnpublishVolumeRequest:
		return ErrControllerUnpublishVolumeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.CreateVolumeRequest:
		return ErrCreateVolumeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.DeleteVolumeRequest:
		return ErrDeleteVolumeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.GetCapacityRequest:
		return ErrGetCapacity(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.GetNodeIDRequest:
		return ErrGetNodeIDGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.GetPluginInfoRequest:
		return ErrGetPluginInfo(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.ListVolumesRequest:
		return ErrListVolumes(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.GetSupportedVersionsRequest:
		panic("Version Check Unsupported for GetSupportedVersions")
	case *csi.NodeGetCapabilitiesRequest:
		return ErrNodeGetCapabilities(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.NodePublishVolumeRequest:
		return ErrNodePublishVolumeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.NodeUnpublishVolumeRequest:
		return ErrNodeUnpublishVolumeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.ProbeNodeRequest:
		return ErrProbeNodeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.ValidateVolumeCapabilitiesRequest:
		return ErrValidateVolumeCapabilitiesGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	}

	panic("Version Check Unsupported")
}

var requestIDVal uint64

// ServerRequestIDInjector is a unary server interceptor that injects
// request contexts with a unique request ID.
func ServerRequestIDInjector(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	return handler(
		context.WithValue(
			ctx,
			requestIDKey,
			atomic.AddUint64(&requestIDVal, 1)),
		req)
}

// NewServerRequestLogger initializes a new unary, server interceptor
// that logs request details.
func NewServerRequestLogger(
	stdout, stderr io.Writer) grpc.UnaryServerInterceptor {

	return (&serverReqLogger{stdout: stdout, stderr: stderr}).handle
}

type serverReqLogger struct {
	stdout io.Writer
	stderr io.Writer
}

var emptyValRX = regexp.MustCompile(
	`^((?:)|(?:\[\])|(?:<nil>)|(?:map\[\]))$`)

func (v *serverReqLogger) handle(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if req == nil {
		return handler(ctx, req)
	}

	w := v.stdout
	fmt.Fprintf(w, "%s: ", info.FullMethod)
	if rid, ok := GetRequestID(ctx); ok {
		fmt.Fprintf(w, "REQ %04d", rid)
	}
	rprintReqOrRep(w, req)
	fmt.Fprintln(w)

	return handler(ctx, req)
}

func rprintReqOrRep(w io.Writer, obj interface{}) {
	rv := reflect.ValueOf(obj).Elem()
	tv := rv.Type()
	nf := tv.NumField()
	printedColon := false
	printComma := false
	for i := 0; i < nf; i++ {
		sv := fmt.Sprintf("%v", rv.Field(i).Interface())
		if emptyValRX.MatchString(sv) {
			continue
		}
		if printComma {
			fmt.Fprintf(w, ", ")
		}
		if !printedColon {
			fmt.Fprintf(w, ": ")
			printedColon = true
		}
		printComma = true
		fmt.Fprintf(w, "%s=%s", tv.Field(i).Name, sv)
	}
}

// NewServerResponseLogger initializes a new unary, server interceptor
// that logs reply details.
func NewServerResponseLogger(
	stdout, stderr io.Writer) grpc.UnaryServerInterceptor {

	return (&serverRepLogger{stdout: stdout, stderr: stderr}).handle
}

type serverRepLogger struct {
	stdout io.Writer
	stderr io.Writer
}

type hasGetError interface {
	GetError() *csi.Error
}

func (v *serverRepLogger) handle(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if req == nil {
		return handler(ctx, req)
	}

	w := v.stdout
	b := &bytes.Buffer{}

	fmt.Fprintf(b, "%s: ", info.FullMethod)
	if rid, ok := GetRequestID(ctx); ok {
		fmt.Fprintf(b, "REP %04d", rid)
	}

	rep, err := handler(ctx, req)
	if err != nil {
		fmt.Fprintf(b, ": %v", &Error{
			FullMethod: info.FullMethod,
			Code:       ErrorNoCode,
			InnerError: err,
		})
		fmt.Fprintln(w, b.String())
		return rep, err
	}

	if rep == nil {
		return nil, nil
	}

	var gocsiErr error

	switch trep := rep.(type) {

	// Controller
	case *csi.CreateVolumeResponse:
		gocsiErr = CheckResponseErrCreateVolume(
			ctx, info.FullMethod, trep)
	case *csi.DeleteVolumeResponse:
		gocsiErr = CheckResponseErrDeleteVolume(
			ctx, info.FullMethod, trep)
	case *csi.ControllerPublishVolumeResponse:
		gocsiErr = CheckResponseErrControllerPublishVolume(
			ctx, info.FullMethod, trep)
	case *csi.ControllerUnpublishVolumeResponse:
		gocsiErr = CheckResponseErrControllerUnpublishVolume(
			ctx, info.FullMethod, trep)
	case *csi.ValidateVolumeCapabilitiesResponse:
		gocsiErr = CheckResponseErrValidateVolumeCapabilities(
			ctx, info.FullMethod, trep)
	case *csi.ListVolumesResponse:
		gocsiErr = CheckResponseErrListVolumes(
			ctx, info.FullMethod, trep)
	case *csi.GetCapacityResponse:
		gocsiErr = CheckResponseErrGetCapacity(
			ctx, info.FullMethod, trep)
	case *csi.ControllerGetCapabilitiesResponse:
		gocsiErr = CheckResponseErrControllerGetCapabilities(
			ctx, info.FullMethod, trep)

	// Identity
	case *csi.GetSupportedVersionsResponse:
		gocsiErr = CheckResponseErrGetSupportedVersions(
			ctx, info.FullMethod, trep)
	case *csi.GetPluginInfoResponse:
		gocsiErr = CheckResponseErrGetPluginInfo(
			ctx, info.FullMethod, trep)

	// Node
	case *csi.NodePublishVolumeResponse:
		gocsiErr = CheckResponseErrNodePublishVolume(
			ctx, info.FullMethod, trep)
	case *csi.NodeUnpublishVolumeResponse:
		gocsiErr = CheckResponseErrNodeUnpublishVolume(
			ctx, info.FullMethod, trep)
	case *csi.GetNodeIDResponse:
		gocsiErr = CheckResponseErrGetNodeID(
			ctx, info.FullMethod, trep)
	case *csi.ProbeNodeResponse:
		gocsiErr = CheckResponseErrProbeNode(
			ctx, info.FullMethod, trep)
	case *csi.NodeGetCapabilitiesResponse:
		gocsiErr = CheckResponseErrNodeGetCapabilities(
			ctx, info.FullMethod, trep)
	}

	// Check to see if the reply has an error or is an error itself.
	if gocsiErr != nil {
		fmt.Fprintf(b, ": %v", gocsiErr)
		fmt.Fprintln(w, b.String())
		return rep, err
	}

	// At this point the reply must be valid. Format and print it.
	rprintReqOrRep(b, rep)
	fmt.Fprintln(w, b.String())
	return rep, err
}

// ServerRequestValidator is a UnaryServerInterceptor that validates
// server request data.
func ServerRequestValidator(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	switch treq := req.(type) {

	// Controller
	case *csi.CreateVolumeRequest:
		rep, err := sreqvCreateVolume(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.DeleteVolumeRequest:
		rep, err := sreqvDeleteVolume(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.ControllerPublishVolumeRequest:
		rep, err := sreqvControllerPublishVolume(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.ControllerUnpublishVolumeRequest:
		rep, err := sreqvControllerUnpublishVolume(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.ValidateVolumeCapabilitiesRequest:
		rep, err := sreqvValidateVolumeCapabilities(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.ListVolumesRequest:
		rep, err := sreqvListVolumes(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.GetCapacityRequest:
		rep, err := sreqvGetCapacity(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.ControllerGetCapabilitiesRequest:
		rep, err := sreqvControllerGetCapabilities(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}

	// Identity
	case *csi.GetSupportedVersionsRequest:
		rep, err := sreqvGetSupportedVersions(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.GetPluginInfoRequest:
		rep, err := sreqvGetPluginInfo(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}

	// Node
	case *csi.NodePublishVolumeRequest:
		rep, err := sreqvNodePublishVolume(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.NodeUnpublishVolumeRequest:
		rep, err := sreqvNodeUnpublishVolume(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.GetNodeIDRequest:
		rep, err := sreqvGetNodeID(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.ProbeNodeRequest:
		rep, err := sreqvProbeNode(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.NodeGetCapabilitiesRequest:
		rep, err := sreqvNodeGetCapabilities(
			ctx, info.FullMethod, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	}

	return handler(ctx, req)
}

////////////////////////////////////////////////////////////////////////////////
//                      SERVER REQUEST - CONTROLLER                           //
////////////////////////////////////////////////////////////////////////////////

func sreqvCreateVolume(
	ctx context.Context,
	method string,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	if req.Name == "" {
		return ErrCreateVolume(
			csi.Error_CreateVolumeError_INVALID_VOLUME_NAME,
			"missing name"), nil
	}

	if len(req.VolumeCapabilities) == 0 {
		return ErrCreateVolume(
			csi.Error_CreateVolumeError_UNKNOWN,
			"missing volume capabilities"), nil
	}

	for i, cap := range req.VolumeCapabilities {
		if cap.AccessMode == nil {
			return ErrCreateVolume(
				csi.Error_CreateVolumeError_UNKNOWN,
				fmt.Sprintf("missing access mode: index %d", i)), nil
		}
		atype := cap.GetAccessType()
		if atype == nil {
			return ErrCreateVolume(
				csi.Error_CreateVolumeError_UNKNOWN,
				fmt.Sprintf("missing access type: index %d", i)), nil
		}
		switch tatype := atype.(type) {
		case *csi.VolumeCapability_Block:
			if tatype.Block == nil {
				return ErrCreateVolume(
					csi.Error_CreateVolumeError_UNKNOWN,
					fmt.Sprintf("missing block type: index %d", i)), nil
			}
		case *csi.VolumeCapability_Mount:
			if tatype.Mount == nil {
				return ErrCreateVolume(
					csi.Error_CreateVolumeError_UNKNOWN,
					fmt.Sprintf("missing mount type: index %d", i)), nil
			}
		default:
			return ErrCreateVolume(
				csi.Error_CreateVolumeError_UNKNOWN,
				fmt.Sprintf(
					"invalid access type: index %d, type=%T",
					i, atype)), nil
		}
	}

	if req.UserCredentials != nil && len(req.UserCredentials.Data) == 0 {
		return ErrCreateVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"empty credentials package specified"), nil
	}

	return nil, nil
}

func sreqvDeleteVolume(
	ctx context.Context,
	method string,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {

	if req.VolumeId == nil {
		return ErrDeleteVolume(
			csi.Error_DeleteVolumeError_INVALID_VOLUME_ID,
			"missing id obj"), nil
	}

	if len(req.VolumeId.Values) == 0 {
		return ErrDeleteVolume(
			csi.Error_DeleteVolumeError_INVALID_VOLUME_ID,
			"missing id map"), nil
	}

	if req.UserCredentials != nil && len(req.UserCredentials.Data) == 0 {
		return ErrDeleteVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"empty credentials package specified"), nil
	}

	return nil, nil
}

func sreqvControllerPublishVolume(
	ctx context.Context,
	method string,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	if req.VolumeId == nil {
		return ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_INVALID_VOLUME_ID,
			"missing id obj"), nil
	}

	if len(req.VolumeId.Values) == 0 {
		return ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_INVALID_VOLUME_ID,
			"missing id map"), nil
	}

	if req.VolumeCapability == nil {
		return ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_UNKNOWN,
			"missing volume capability"), nil
	}

	if req.VolumeCapability.AccessMode == nil {
		return ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_UNKNOWN,
			"missing access mode"), nil
	}
	atype := req.VolumeCapability.GetAccessType()
	if atype == nil {
		return ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_UNKNOWN,
			"missing access type"), nil
	}
	switch tatype := atype.(type) {
	case *csi.VolumeCapability_Block:
		if tatype.Block == nil {
			return ErrControllerPublishVolume(
				csi.Error_ControllerPublishVolumeError_UNKNOWN,
				"missing block type"), nil
		}
	case *csi.VolumeCapability_Mount:
		if tatype.Mount == nil {
			return ErrControllerPublishVolume(
				csi.Error_ControllerPublishVolumeError_UNKNOWN,
				"missing mount type"), nil
		}
	default:
		return ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_UNKNOWN,
			fmt.Sprintf("invalid access type: %T", atype)), nil
	}

	if req.UserCredentials != nil && len(req.UserCredentials.Data) == 0 {
		return ErrControllerPublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"empty credentials package specified"), nil
	}

	return nil, nil
}

func sreqvControllerUnpublishVolume(
	ctx context.Context,
	method string,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	if req.VolumeId == nil {
		return ErrControllerUnpublishVolume(
			csi.Error_ControllerUnpublishVolumeError_INVALID_VOLUME_ID,
			"missing id obj"), nil
	}

	if len(req.VolumeId.Values) == 0 {
		return ErrControllerUnpublishVolume(
			csi.Error_ControllerUnpublishVolumeError_INVALID_VOLUME_ID,
			"missing id map"), nil
	}

	if req.UserCredentials != nil && len(req.UserCredentials.Data) == 0 {
		return ErrControllerUnpublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"empty credentials package specified"), nil
	}

	return nil, nil
}

func sreqvValidateVolumeCapabilities(
	ctx context.Context,
	method string,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	if req.VolumeInfo == nil {
		return ErrValidateVolumeCapabilities(
			csi.Error_ValidateVolumeCapabilitiesError_INVALID_VOLUME_INFO,
			"missing volume info"), nil
	}

	if req.VolumeInfo.Id == nil {
		return ErrValidateVolumeCapabilities(
			csi.Error_ValidateVolumeCapabilitiesError_UNKNOWN,
			"missing id obj"), nil
	}

	if len(req.VolumeInfo.Id.Values) == 0 {
		return ErrValidateVolumeCapabilities(
			csi.Error_ValidateVolumeCapabilitiesError_UNKNOWN,
			"missing id map"), nil
	}

	if len(req.VolumeCapabilities) == 0 {
		return ErrValidateVolumeCapabilities(
			csi.Error_ValidateVolumeCapabilitiesError_UNKNOWN,
			"missing volume capabilities"), nil
	}

	for i, cap := range req.VolumeCapabilities {
		if cap.AccessMode == nil {
			return ErrValidateVolumeCapabilities(
				csi.Error_ValidateVolumeCapabilitiesError_UNKNOWN,
				fmt.Sprintf("missing access mode: index %d", i)), nil
		}
		atype := cap.GetAccessType()
		if atype == nil {
			return ErrValidateVolumeCapabilities(
				csi.Error_ValidateVolumeCapabilitiesError_UNKNOWN,
				fmt.Sprintf("missing access type: index %d", i)), nil
		}
		switch tatype := atype.(type) {
		case *csi.VolumeCapability_Block:
			if tatype.Block == nil {
				return ErrValidateVolumeCapabilities(
					csi.Error_ValidateVolumeCapabilitiesError_UNKNOWN,
					fmt.Sprintf("missing block type: index %d", i)), nil
			}
		case *csi.VolumeCapability_Mount:
			if tatype.Mount == nil {
				return ErrValidateVolumeCapabilities(
					csi.Error_ValidateVolumeCapabilitiesError_UNKNOWN,
					fmt.Sprintf("missing mount type: index %d", i)), nil
			}
		default:
			return ErrValidateVolumeCapabilities(
				csi.Error_ValidateVolumeCapabilitiesError_UNKNOWN,
				fmt.Sprintf(
					"invalid access type: index %d, type=%T",
					i, atype)), nil
		}
	}

	return nil, nil
}

func sreqvListVolumes(
	ctx context.Context,
	method string,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	return nil, nil
}

func sreqvGetCapacity(
	ctx context.Context,
	method string,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	if len(req.VolumeCapabilities) == 0 {
		return nil, nil
	}

	for i, cap := range req.VolumeCapabilities {
		if cap.AccessMode == nil {
			return ErrGetCapacity(
				csi.Error_GeneralError_UNDEFINED,
				fmt.Sprintf("missing access mode: index %d", i)), nil
		}
		atype := cap.GetAccessType()
		if atype == nil {
			return ErrGetCapacity(
				csi.Error_GeneralError_UNDEFINED,
				fmt.Sprintf("missing access type: index %d", i)), nil
		}
		switch tatype := atype.(type) {
		case *csi.VolumeCapability_Block:
			if tatype.Block == nil {
				return ErrGetCapacity(
					csi.Error_GeneralError_UNDEFINED,
					fmt.Sprintf("missing block type: index %d", i)), nil
			}
		case *csi.VolumeCapability_Mount:
			if tatype.Mount == nil {
				return ErrGetCapacity(
					csi.Error_GeneralError_UNDEFINED,
					fmt.Sprintf("missing mount type: index %d", i)), nil
			}
		default:
			return ErrGetCapacity(
				csi.Error_GeneralError_UNDEFINED,
				fmt.Sprintf(
					"invalid access type: index %d, type=%T",
					i, atype)), nil
		}
	}

	return nil, nil
}

func sreqvControllerGetCapabilities(
	ctx context.Context,
	method string,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////
//                        SERVER REQUEST - IDENTITY                           //
////////////////////////////////////////////////////////////////////////////////

func sreqvGetSupportedVersions(
	ctx context.Context,
	method string,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	return nil, nil
}

func sreqvGetPluginInfo(
	ctx context.Context,
	method string,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////
//                         SERVER REQUEST - NODE                              //
////////////////////////////////////////////////////////////////////////////////

func sreqvNodePublishVolume(
	ctx context.Context,
	method string,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	if req.VolumeId == nil {
		return ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_INVALID_VOLUME_ID,
			"missing id obj"), nil
	}

	if len(req.VolumeId.Values) == 0 {
		return ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_INVALID_VOLUME_ID,
			"missing id map"), nil
	}

	if req.VolumeCapability == nil {
		return ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_UNKNOWN,
			"missing volume capability"), nil
	}

	if req.VolumeCapability.AccessMode == nil {
		return ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_UNKNOWN,
			"missing access mode"), nil
	}
	atype := req.VolumeCapability.GetAccessType()
	if atype == nil {
		return ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_UNKNOWN,
			"missing access type"), nil
	}
	switch tatype := atype.(type) {
	case *csi.VolumeCapability_Block:
		if tatype.Block == nil {
			return ErrNodePublishVolume(
				csi.Error_NodePublishVolumeError_UNKNOWN,
				"missing block type"), nil
		}
	case *csi.VolumeCapability_Mount:
		if tatype.Mount == nil {
			return ErrNodePublishVolume(
				csi.Error_NodePublishVolumeError_UNKNOWN,
				"missing mount type"), nil
		}
	default:
		return ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_UNKNOWN,
			fmt.Sprintf("invalid access type: %T", atype)), nil
	}

	if req.TargetPath == "" {
		return ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_UNKNOWN,
			"missing target path"), nil
	}

	if req.UserCredentials != nil && len(req.UserCredentials.Data) == 0 {
		return ErrNodePublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"empty credentials package specified"), nil
	}

	return nil, nil
}

func sreqvNodeUnpublishVolume(
	ctx context.Context,
	method string,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	if req.VolumeId == nil {
		return ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_INVALID_VOLUME_ID,
			"missing id obj"), nil
	}

	if len(req.VolumeId.Values) == 0 {
		return ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_INVALID_VOLUME_ID,
			"missing id map"), nil
	}

	if req.TargetPath == "" {
		return ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_UNKNOWN,
			"missing target path"), nil
	}

	if req.UserCredentials != nil && len(req.UserCredentials.Data) == 0 {
		return ErrNodeUnpublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"empty credentials package specified"), nil
	}

	return nil, nil
}

func sreqvGetNodeID(
	ctx context.Context,
	method string,
	req *csi.GetNodeIDRequest) (
	*csi.GetNodeIDResponse, error) {

	return nil, nil
}

func sreqvProbeNode(
	ctx context.Context,
	method string,
	req *csi.ProbeNodeRequest) (
	*csi.ProbeNodeResponse, error) {

	return nil, nil
}

func sreqvNodeGetCapabilities(
	ctx context.Context,
	method string,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	return nil, nil
}
