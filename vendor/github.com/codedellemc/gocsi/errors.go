package gocsi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/codedellemc/gocsi/csi"
)

// ErrorNoCode is the value that indicates the error code is not set.
const ErrorNoCode int32 = -1

// Error is a client-side representation of the error message
// data returned from a CSI endpoint.
type Error struct {
	// Code is the error code.
	Code int32
	// Description is the error description.
	Description string
	// FullMethod is the full name of the CSI method that returned the error.
	FullMethod string
	// InnerError is an optional inner error wrapped by this error.
	InnerError error
}

func (e *Error) Error() string {
	w := &bytes.Buffer{}
	fmt.Fprintf(w, "%s failed", e.ErrorMethod())
	if e.Code > ErrorNoCode {
		fmt.Fprintf(w, ": %d", e.Code)
	}
	if e.Description != "" {
		fmt.Fprintf(w, ": %s", e.Description)
	}
	if e.InnerError != nil {
		fmt.Fprintf(w, ": %v", e.InnerError)
	}
	return w.String()
}

func (e *Error) String() string {
	return e.Error()
}

// MarshalText encodes the receiver into UTF-8-encoded text
// and returns the result.
func (e *Error) MarshalText() (text []byte, err error) {
	return []byte(e.String()), nil
}

// ErrorCode returns the CSI error code.
func (e *Error) ErrorCode() int32 {
	return e.Code
}

// ErrorDescription returns the CSI error description.
func (e *Error) ErrorDescription() string {
	return e.Description
}

// ErrorFullMethod returns  the full name of the CSI method that
// returned the error.
func (e *Error) ErrorFullMethod() string {
	return e.FullMethod
}

// ErrorMethod returns the name-only of the CSI method that
// returned the error.
func (e *Error) ErrorMethod() string {
	parts := strings.Split(e.FullMethod, "/")
	return parts[len(parts)-1]
}

// ErrEmptyServices occurs when a Server's Services list is empty.
var ErrEmptyServices = errors.New("services list is empty")

// ErrMissingCSIEndpoint occurs when the value for the environment
// variable CSI_ENDPOINT is not set.
var ErrMissingCSIEndpoint = errors.New("missing CSI_ENDPOINT")

// ErrInvalidCSIEndpoint occurs when the value for the environment
// variable CSI_ENDPOINT is an invalid network address.
var ErrInvalidCSIEndpoint = errors.New("invalid CSI_ENDPOINT")

// ErrNilVolumeInfo occurs when a gRPC call returns a nil VolumeInfo.
var ErrNilVolumeInfo = errors.New("volumeInfo is nil")

// ErrNilVolumeID occurs when a gRPC call returns a VolumeInfo with
// a nil Id field.
var ErrNilVolumeID = errors.New("volumeInfo.Id is nil")

// ErrNilPublishVolumeInfo occurs when a gRPC call returns
// a nil PublishVolumeInfo.
var ErrNilPublishVolumeInfo = errors.New("publishVolumeInfo is nil")

// ErrEmptyPublishVolumeInfo occurs when a gRPC call returns
// a PublishVolumeInfo with an empty Values field.
var ErrEmptyPublishVolumeInfo = errors.New("publishVolumeInfo is empty")

// ErrNilNodeID occurs when a gRPC call returns
// a nil NodeID.
var ErrNilNodeID = errors.New("nodeID is nil")

// ErrNilSupportedVersions occurs when a gRPC call returns nil SupportedVersions
var ErrNilSupportedVersions = errors.New("supportedVersions is nil")

// ErrVersionRequired occurs when an RPC call is made with a nil
// version argument.
var ErrVersionRequired = errors.New("version is required")

// ErrVolumeIDRequired occurs when an RPC call is made with a nil
// volumeID argument.
var ErrVolumeIDRequired = errors.New("volumeID is required")

// ErrVolumeInfoRequired occurs when an RPC call is made with a nil
// volumeI argument.
var ErrVolumeInfoRequired = errors.New("volumeInfo is required")

// ErrVolumeCapabilityRequired occurs when an RPC call is made with
// a nil volumeCapability argument.
var ErrVolumeCapabilityRequired = errors.New("volumeCapability is required")

// ErrInvalidTargetPath occurs when an RPC call is made with
// an invalid targetPath argument.
var ErrInvalidTargetPath = errors.New("invalid targetPath")

// ErrNilResult occurs when a gRPC call returns a nil Result.
var ErrNilResult = errors.New("result is nil")

// ErrInvalidProvider is returned from NewService if the
// specified provider name is unknown.
var ErrInvalidProvider = errors.New("invalid service provider")

////////////////////////////////////////////////////////////////////////////////
//                             Controller Service                             //
////////////////////////////////////////////////////////////////////////////////

// ErrListVolumes returns a ListVolumesResponse with a GeneralError.
func ErrListVolumes(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.ListVolumesResponse {

	return &csi.ListVolumesResponse{
		Reply: &csi.ListVolumesResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrCreateVolume returns a CreateVolumeResponse with a CreateVolumeError.
func ErrCreateVolume(
	code csi.Error_CreateVolumeError_CreateVolumeErrorCode,
	msg string) *csi.CreateVolumeResponse {

	return &csi.CreateVolumeResponse{
		Reply: &csi.CreateVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_CreateVolumeError_{
					CreateVolumeError: &csi.Error_CreateVolumeError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrCreateVolumeGeneral returns a CreateVolumeResponse with a GeneralError.
func ErrCreateVolumeGeneral(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.CreateVolumeResponse {

	return &csi.CreateVolumeResponse{
		Reply: &csi.CreateVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrDeleteVolume returns a DeleteVolumeResponse with a DeleteVolumeError.
func ErrDeleteVolume(
	code csi.Error_DeleteVolumeError_DeleteVolumeErrorCode,
	msg string) *csi.DeleteVolumeResponse {

	return &csi.DeleteVolumeResponse{
		Reply: &csi.DeleteVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_DeleteVolumeError_{
					DeleteVolumeError: &csi.Error_DeleteVolumeError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrDeleteVolumeGeneral returns a DeleteVolumeResponse with a GeneralError.
func ErrDeleteVolumeGeneral(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.DeleteVolumeResponse {

	return &csi.DeleteVolumeResponse{
		Reply: &csi.DeleteVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrControllerPublishVolume returns a
// ControllerPublishVolumeResponse with a
// ControllerPublishVolumeVolumeError.
func ErrControllerPublishVolume(
	code csi.Error_ControllerPublishVolumeError_ControllerPublishVolumeErrorCode,
	msg string) *csi.ControllerPublishVolumeResponse {

	return &csi.ControllerPublishVolumeResponse{
		Reply: &csi.ControllerPublishVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_ControllerPublishVolumeError_{
					ControllerPublishVolumeError: &csi.Error_ControllerPublishVolumeError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrControllerPublishVolumeGeneral returns a
// ControllerPublishVolumeResponse with a
// GeneralError.
func ErrControllerPublishVolumeGeneral(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.ControllerPublishVolumeResponse {

	return &csi.ControllerPublishVolumeResponse{
		Reply: &csi.ControllerPublishVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrControllerUnpublishVolume returns a
// ControllerUnpublishVolumeResponse with a
// ControllerUnpublishVolumeVolumeError.
func ErrControllerUnpublishVolume(
	code csi.Error_ControllerUnpublishVolumeError_ControllerUnpublishVolumeErrorCode,
	msg string) *csi.ControllerUnpublishVolumeResponse {

	return &csi.ControllerUnpublishVolumeResponse{
		Reply: &csi.ControllerUnpublishVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_ControllerUnpublishVolumeError_{
					ControllerUnpublishVolumeError: &csi.Error_ControllerUnpublishVolumeError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrControllerUnpublishVolumeGeneral returns a
// ControllerUnpublishVolumeResponse with a
// GeneralError.
func ErrControllerUnpublishVolumeGeneral(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.ControllerUnpublishVolumeResponse {

	return &csi.ControllerUnpublishVolumeResponse{
		Reply: &csi.ControllerUnpublishVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrValidateVolumeCapabilities returns a
// ValidateVolumeCapabilitiesResponse with a
// ValidateVolumeCapabilitiesError.
func ErrValidateVolumeCapabilities(
	code csi.Error_ValidateVolumeCapabilitiesError_ValidateVolumeCapabilitiesErrorCode,
	msg string) *csi.ValidateVolumeCapabilitiesResponse {

	return &csi.ValidateVolumeCapabilitiesResponse{
		Reply: &csi.ValidateVolumeCapabilitiesResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_ValidateVolumeCapabilitiesError_{
					ValidateVolumeCapabilitiesError: &csi.Error_ValidateVolumeCapabilitiesError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrValidateVolumeCapabilitiesGeneral returns a
// ValidateVolumeCapabilitiesResponse with a
// GeneralError.
func ErrValidateVolumeCapabilitiesGeneral(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.ValidateVolumeCapabilitiesResponse {

	return &csi.ValidateVolumeCapabilitiesResponse{
		Reply: &csi.ValidateVolumeCapabilitiesResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrGetCapacity returns a
// GetCapacityResponse with a
// GeneralError.
func ErrGetCapacity(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.GetCapacityResponse {

	return &csi.GetCapacityResponse{
		Reply: &csi.GetCapacityResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrControllerGetCapabilities returns a
// ControllerGetCapabilitiesResponse with a
// GeneralError.
func ErrControllerGetCapabilities(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.ControllerGetCapabilitiesResponse {

	return &csi.ControllerGetCapabilitiesResponse{
		Reply: &csi.ControllerGetCapabilitiesResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
//                             Identity Service                               //
////////////////////////////////////////////////////////////////////////////////

// ErrGetSupportedVersions returns a
// GetSupportedVersionsResponse with a
// GeneralError.
func ErrGetSupportedVersions(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.GetSupportedVersionsResponse {

	return &csi.GetSupportedVersionsResponse{
		Reply: &csi.GetSupportedVersionsResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrGetPluginInfo returns a
// GetPluginInfoResponse with a
// GeneralError.
func ErrGetPluginInfo(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.GetPluginInfoResponse {

	return &csi.GetPluginInfoResponse{
		Reply: &csi.GetPluginInfoResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
//                               Node Service                                 //
////////////////////////////////////////////////////////////////////////////////

// ErrNodePublishVolume returns a
// NodePublishVolumeResponse with a
// NodePublishVolumeError.
func ErrNodePublishVolume(
	code csi.Error_NodePublishVolumeError_NodePublishVolumeErrorCode,
	msg string) *csi.NodePublishVolumeResponse {

	return &csi.NodePublishVolumeResponse{
		Reply: &csi.NodePublishVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_NodePublishVolumeError_{
					NodePublishVolumeError: &csi.Error_NodePublishVolumeError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrNodePublishVolumeGeneral returns a
// NodePublishVolumeResponse with a
// GeneralError.
func ErrNodePublishVolumeGeneral(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.NodePublishVolumeResponse {

	return &csi.NodePublishVolumeResponse{
		Reply: &csi.NodePublishVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrNodeUnpublishVolume returns a
// NodeUnpublishVolumeResponse with a
// NodeUnpublishVolumeError.
func ErrNodeUnpublishVolume(
	code csi.Error_NodeUnpublishVolumeError_NodeUnpublishVolumeErrorCode,
	msg string) *csi.NodeUnpublishVolumeResponse {

	return &csi.NodeUnpublishVolumeResponse{
		Reply: &csi.NodeUnpublishVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_NodeUnpublishVolumeError_{
					NodeUnpublishVolumeError: &csi.Error_NodeUnpublishVolumeError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrNodeUnpublishVolumeGeneral returns a
// NodeUnpublishVolumeResponse with a
// GeneralError.
func ErrNodeUnpublishVolumeGeneral(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.NodeUnpublishVolumeResponse {

	return &csi.NodeUnpublishVolumeResponse{
		Reply: &csi.NodeUnpublishVolumeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrGetNodeID returns a
// GetNodeIDResponse with a
// GetNodeIDError.
func ErrGetNodeID(
	code csi.Error_GetNodeIDError_GetNodeIDErrorCode,
	msg string) *csi.GetNodeIDResponse {

	return &csi.GetNodeIDResponse{
		Reply: &csi.GetNodeIDResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GetNodeIdError{
					GetNodeIdError: &csi.Error_GetNodeIDError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrGetNodeIDGeneral returns a
// GetNodeIDResponse with a
// GeneralError.
func ErrGetNodeIDGeneral(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.GetNodeIDResponse {

	return &csi.GetNodeIDResponse{
		Reply: &csi.GetNodeIDResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrProbeNode returns a
// ProbeNodeResponse with a
// ProbeNodeError.
func ErrProbeNode(
	code csi.Error_ProbeNodeError_ProbeNodeErrorCode,
	msg string) *csi.ProbeNodeResponse {

	return &csi.ProbeNodeResponse{
		Reply: &csi.ProbeNodeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_ProbeNodeError_{
					ProbeNodeError: &csi.Error_ProbeNodeError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrProbeNodeGeneral returns a
// ProbeNodeResponse with a
// GeneralError.
func ErrProbeNodeGeneral(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.ProbeNodeResponse {

	return &csi.ProbeNodeResponse{
		Reply: &csi.ProbeNodeResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

// ErrNodeGetCapabilities returns a
// NodeGetCapabilitiesResponse with a
// GeneralError.
func ErrNodeGetCapabilities(
	code csi.Error_GeneralError_GeneralErrorCode,
	msg string) *csi.NodeGetCapabilitiesResponse {

	return &csi.NodeGetCapabilitiesResponse{
		Reply: &csi.NodeGetCapabilitiesResponse_Error{
			Error: &csi.Error{
				Value: &csi.Error_GeneralError_{
					GeneralError: &csi.Error_GeneralError{
						ErrorCode:        code,
						ErrorDescription: msg,
					},
				},
			},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
//                       RESPONSE ERROR - CONTROLLER                          //
////////////////////////////////////////////////////////////////////////////////

// CheckResponseErrCreateVolume returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrCreateVolume(
	ctx context.Context,
	method string,
	response *csi.CreateVolumeResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMCreateVolume
	}

	if err := rErr.GetCreateVolumeError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

// CheckResponseErrDeleteVolume returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrDeleteVolume(
	ctx context.Context,
	method string,
	response *csi.DeleteVolumeResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMDeleteVolume
	}

	if err := rErr.GetDeleteVolumeError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

// CheckResponseErrControllerPublishVolume returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrControllerPublishVolume(
	ctx context.Context,
	method string,
	response *csi.ControllerPublishVolumeResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMControllerPublishVolume
	}

	if err := rErr.GetControllerPublishVolumeError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

// CheckResponseErrControllerUnpublishVolume returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrControllerUnpublishVolume(
	ctx context.Context,
	method string,
	response *csi.ControllerUnpublishVolumeResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMControllerUnpublishVolume
	}

	if err := rErr.GetControllerUnpublishVolumeError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

// CheckResponseErrValidateVolumeCapabilities returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrValidateVolumeCapabilities(
	ctx context.Context,
	method string,
	response *csi.ValidateVolumeCapabilitiesResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMValidateVolumeCapabilities
	}

	if err := rErr.GetValidateVolumeCapabilitiesError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

// CheckResponseErrListVolumes returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrListVolumes(
	ctx context.Context,
	method string,
	response *csi.ListVolumesResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMListVolumes
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

// CheckResponseErrGetCapacity returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrGetCapacity(
	ctx context.Context,
	method string,
	response *csi.GetCapacityResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMGetCapacity
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

// CheckResponseErrControllerGetCapabilities returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrControllerGetCapabilities(
	ctx context.Context,
	method string,
	response *csi.ControllerGetCapabilitiesResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMControllerGetCapabilities
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

////////////////////////////////////////////////////////////////////////////////
//                       RESPONSE ERROR - IDENTITY                            //
////////////////////////////////////////////////////////////////////////////////

// CheckResponseErrGetSupportedVersions returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrGetSupportedVersions(
	ctx context.Context,
	method string,
	response *csi.GetSupportedVersionsResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMGetSupportedVersions
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

// CheckResponseErrGetPluginInfo returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrGetPluginInfo(
	ctx context.Context,
	method string,
	response *csi.GetPluginInfoResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMGetPluginInfo
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

////////////////////////////////////////////////////////////////////////////////
//                        RESPONSE ERROR - NODE                               //
////////////////////////////////////////////////////////////////////////////////

// CheckResponseErrNodePublishVolume returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrNodePublishVolume(
	ctx context.Context,
	method string,
	response *csi.NodePublishVolumeResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMNodePublishVolume
	}

	if err := rErr.GetNodePublishVolumeError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

// CheckResponseErrNodeUnpublishVolume returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrNodeUnpublishVolume(
	ctx context.Context,
	method string,
	response *csi.NodeUnpublishVolumeResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMNodeUnpublishVolume
	}

	if err := rErr.GetNodeUnpublishVolumeError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

// CheckResponseErrGetNodeID returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrGetNodeID(
	ctx context.Context,
	method string,
	response *csi.GetNodeIDResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMGetNodeID
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

// CheckResponseErrProbeNode returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrProbeNode(
	ctx context.Context,
	method string,
	response *csi.ProbeNodeResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMProbeNode
	}

	if err := rErr.GetProbeNodeError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}

// CheckResponseErrNodeGetCapabilities returns a Go error for the
// error message inside of a CSI response if present; otherwise nil
// is returned.
func CheckResponseErrNodeGetCapabilities(
	ctx context.Context,
	method string,
	response *csi.NodeGetCapabilitiesResponse) error {

	rErr := response.GetError()
	if rErr == nil {
		return nil
	}

	if method == "" {
		method = FMNodeGetCapabilities
	}

	if err := rErr.GetGeneralError(); err != nil {
		return &Error{
			FullMethod:  method,
			Code:        int32(err.ErrorCode),
			Description: err.ErrorDescription,
		}
	}

	return &Error{
		FullMethod:  method,
		Code:        ErrorNoCode,
		Description: rErr.String(),
	}
}
