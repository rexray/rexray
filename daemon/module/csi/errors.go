// +build !rexray_build_type_client
// +build !rexray_build_type_controller
// +build csi

package csi

import "github.com/codedellemc/rexray/daemon/module/csi/csi"

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
				Value: &csi.Error_ControllerPublishVolumeVolumeError{
					ControllerPublishVolumeVolumeError: &csi.Error_ControllerPublishVolumeError{
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
				Value: &csi.Error_ControllerUnpublishVolumeVolumeError{
					ControllerUnpublishVolumeVolumeError: &csi.Error_ControllerUnpublishVolumeError{
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
