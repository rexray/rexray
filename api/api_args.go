package api

/*******************************************************************************
**                              GetDriverNames
*******************************************************************************/

// GetDriverNamesArgs are the arguments expected by the GetDriverNames function.
type GetDriverNamesArgs struct {
	Extensions map[string]interface{}     `json:"extensions"`
	Optional   GetDriverNamesArgsOptional `json:"optional"`
	Required   GetDriverNamesArgsRequired `json:"required"`
}

// GetDriverNamesArgsRequired are the required arguments expected by the
// GetDriverNames function.
type GetDriverNamesArgsRequired struct {
}

// GetDriverNamesArgsOptional are the optional arguments expected by the
// GetDriverNames function.
type GetDriverNamesArgsOptional struct {
}

/*******************************************************************************
**                              GetVolumeMapping
*******************************************************************************/

// GetVolumeMappingArgs are the arguments expected by the GetVolumeMapping
// function.
type GetVolumeMappingArgs struct {
	Extensions map[string]interface{}       `json:"extensions"`
	Optional   GetVolumeMappingArgsOptional `json:"optional"`
	Required   GetVolumeMappingArgsRequired `json:"required"`
}

// GetVolumeMappingArgsRequired are the required arguments expected by the
// GetVolumeMapping function.
type GetVolumeMappingArgsRequired struct {
}

// GetVolumeMappingArgsOptional are the optional arguments expected by the
// GetVolumeMapping function.
type GetVolumeMappingArgsOptional struct {
}

/*******************************************************************************
**                                GetInstance
*******************************************************************************/

// GetInstanceArgs are the arguments expected by the GetInstance function.
type GetInstanceArgs struct {
	Extensions map[string]interface{}  `json:"extensions"`
	Optional   GetInstanceArgsOptional `json:"optional"`
	Required   GetInstanceArgsRequired `json:"required"`
}

// GetInstanceArgsRequired are the required arguments expected by the
// GetInstance function.
type GetInstanceArgsRequired struct {
}

// GetInstanceArgsOptional are the optional arguments expected by the
// GetInstance function.
type GetInstanceArgsOptional struct {
}

/*******************************************************************************
**                                 GetVolume
*******************************************************************************/

// GetVolumeArgs are the arguments expected by the GetVolume function.
type GetVolumeArgs struct {
	Extensions map[string]interface{} `json:"extensions"`
	Optional   GetVolumeArgsOptional  `json:"optional"`
	Required   GetVolumeArgsRequired  `json:"required"`
}

// GetVolumeArgsRequired are the required arguments expected by the
// GetVolume function.
type GetVolumeArgsRequired struct {
}

// GetVolumeArgsOptional are the optional arguments expected by the
// GetVolume function.
type GetVolumeArgsOptional struct {
	VolumeID   string `json:"volumeID"`
	VolumeName string `json:"volumeName"`
}

/*******************************************************************************
**                              GetVolumeAttach
*******************************************************************************/

// GetVolumeAttachArgs are the arguments expected by the GetVolumeAttach
// function.
type GetVolumeAttachArgs struct {
	Extensions map[string]interface{}      `json:"extensions"`
	Optional   GetVolumeAttachArgsOptional `json:"optional"`
	Required   GetVolumeAttachArgsRequired `json:"required"`
}

// GetVolumeAttachArgsRequired are the required arguments expected by the
// GetVolumeAttach function.
type GetVolumeAttachArgsRequired struct {
}

// GetVolumeAttachArgsOptional are the optional arguments expected by the
// GetVolumeAttach function.
type GetVolumeAttachArgsOptional struct {
	VolumeID string `json:"volumeID"`
}

/*******************************************************************************
**                               CreateSnapshot
*******************************************************************************/

// CreateSnapshotArgs are the arguments expected by the CreateSnapshot function.
type CreateSnapshotArgs struct {
	Extensions map[string]interface{}     `json:"extensions"`
	Optional   CreateSnapshotArgsOptional `json:"optional"`
	Required   CreateSnapshotArgsRequired `json:"required"`
}

// CreateSnapshotArgsRequired are the required arguments expected by the
// CreateSnapshot function.
type CreateSnapshotArgsRequired struct {
}

// CreateSnapshotArgsOptional are the optional arguments expected by the
// CreateSnapshot function.
type CreateSnapshotArgsOptional struct {
	Description  string `json:"description"`
	SnapshotName string `json:"snapshotName"`
	VolumeID     string `json:"volumeID"`
}

/*******************************************************************************
**                                GetSnapshot
*******************************************************************************/

// GetSnapshotArgs are the arguments expected by the GetSnapshot function.
type GetSnapshotArgs struct {
	Extensions map[string]interface{}  `json:"extensions"`
	Optional   GetSnapshotArgsOptional `json:"optional"`
	Required   GetSnapshotArgsRequired `json:"required"`
}

// GetSnapshotArgsRequired are the required arguments expected by the
// GetSnapshot function.
type GetSnapshotArgsRequired struct {
}

// GetSnapshotArgsOptional are the optional arguments expected by the
// GetSnapshot function.
type GetSnapshotArgsOptional struct {
	SnapshotID   string `json:"snapshotID"`
	SnapshotName string `json:"snapshotName"`
	VolumeID     string `json:"volumeID"`
}

/*******************************************************************************
**                               RemoveSnapshot
*******************************************************************************/

// RemoveSnapshotArgs are the arguments expected by the RemoveSnapshot function.
type RemoveSnapshotArgs struct {
	Extensions map[string]interface{}     `json:"extensions"`
	Optional   RemoveSnapshotArgsOptional `json:"optional"`
	Required   RemoveSnapshotArgsRequired `json:"required"`
}

// RemoveSnapshotArgsRequired are the required arguments expected by the
// RemoveSnapshot function.
type RemoveSnapshotArgsRequired struct {
	SnapshotID string `json:"snapshotID"`
}

// RemoveSnapshotArgsOptional are the optional arguments expected by the
// RemoveSnapshot function.
type RemoveSnapshotArgsOptional struct {
}

/*******************************************************************************
**                                CreateVolume
*******************************************************************************/

// CreateVolumeArgs are the arguments expected by the CreateVolume function.
type CreateVolumeArgs struct {
	Extensions map[string]interface{}   `json:"extensions"`
	Optional   CreateVolumeArgsOptional `json:"optional"`
	Required   CreateVolumeArgsRequired `json:"required"`
}

// CreateVolumeArgsRequired are the required arguments expected by the
// CreateVolume function.
type CreateVolumeArgsRequired struct {
}

// CreateVolumeArgsOptional are the optional arguments expected by the
// CreateVolume function.
type CreateVolumeArgsOptional struct {
	AvailabilityZone string `json:"availabilityZone"`
	IOPS             int64  `json:"iops"`
	Size             int64  `json:"size"`
	SnapshotID       string `json:"snapshotID"`
	VolumeID         string `json:"volumeID"`
	VolumeName       string `json:"volumeName"`
	VolumeType       string `json:"volumeType"`
}

/*******************************************************************************
**                              RemoveVolume
*******************************************************************************/

// RemoveVolumeArgs are the arguments expected by the RemoveVolume function.
type RemoveVolumeArgs struct {
	Extensions map[string]interface{}   `json:"extensions"`
	Optional   RemoveVolumeArgsOptional `json:"optional"`
	Required   RemoveVolumeArgsRequired `json:"required"`
}

// RemoveVolumeArgsRequired are the required arguments expected by the
// RemoveVolume function.
type RemoveVolumeArgsRequired struct {
	VolumeID string `json:"volumeID"`
}

// RemoveVolumeArgsOptional are the optional arguments expected by the
// RemoveVolume function.
type RemoveVolumeArgsOptional struct {
}

/*******************************************************************************
**                               AttachVolume
*******************************************************************************/

// AttachVolumeArgs are the arguments expected by the AttachVolume function.
type AttachVolumeArgs struct {
	Extensions map[string]interface{}   `json:"extensions"`
	Optional   AttachVolumeArgsOptional `json:"optional"`
	Required   AttachVolumeArgsRequired `json:"required"`
}

// AttachVolumeArgsRequired are the required arguments expected by the
// AttachVolume function.
type AttachVolumeArgsRequired struct {
	NextDeviceName string `json:"nextDeviceName"`
	VolumeID       string `json:"volumeID"`
}

// AttachVolumeArgsOptional are the optional arguments expected by the
// AttachVolume function.
type AttachVolumeArgsOptional struct {
}

/*******************************************************************************
**                               DetachVolume
*******************************************************************************/

// DetachVolumeArgs are the arguments expected by the DetachVolume function.
type DetachVolumeArgs struct {
	Extensions map[string]interface{}   `json:"extensions"`
	Optional   DetachVolumeArgsOptional `json:"optional"`
	Required   DetachVolumeArgsRequired `json:"required"`
}

// DetachVolumeArgsRequired are the required arguments expected by the
// DetachVolume function.
type DetachVolumeArgsRequired struct {
	VolumeID string `json:"volumeID"`
}

// DetachVolumeArgsOptional are the optional arguments expected by the
// DetachVolume function.
type DetachVolumeArgsOptional struct {
}

/*******************************************************************************
**                                CopySnapshot
*******************************************************************************/

// CopySnapshotArgs are the arguments expected by the CopySnapshot function.
type CopySnapshotArgs struct {
	Extensions map[string]interface{}   `json:"extensions"`
	Optional   CopySnapshotArgsOptional `json:"optional"`
	Required   CopySnapshotArgsRequired `json:"required"`
}

// CopySnapshotArgsRequired are the required arguments expected by the
// CopySnapshot function.
type CopySnapshotArgsRequired struct {
}

// CopySnapshotArgsOptional are the optional arguments expected by the
// CopySnapshot function.
type CopySnapshotArgsOptional struct {
	DestinationRegion       string `json:"destinationRegion"`
	DestinationSnapshotName string `json:"destinationSnapshotName"`
	SnapshotID              string `json:"snapshotID"`
	SnapshotName            string `json:"snapshotName"`
	VolumeID                string `json:"volumeID"`
}

/*******************************************************************************
**                              GetClientToolName
*******************************************************************************/

// GetClientToolNameArgs are the arguments expected by the GetClientToolName
// function.
type GetClientToolNameArgs struct {
	Extensions map[string]interface{}        `json:"extensions"`
	Optional   GetClientToolNameArgsOptional `json:"optional"`
	Required   GetClientToolNameArgsRequired `json:"required"`
}

// GetClientToolNameArgsRequired are the required arguments expected by the
// GetClientToolName function.
type GetClientToolNameArgsRequired struct {
}

// GetClientToolNameArgsOptional are the optional arguments expected by the
// GetClientToolName function.
type GetClientToolNameArgsOptional struct {
}

/*******************************************************************************
**                                GetClientTool
*******************************************************************************/

// GetClientToolArgs are the arguments expected by the GetClientTool function.
type GetClientToolArgs struct {
	Extensions map[string]interface{}    `json:"extensions"`
	Optional   GetClientToolArgsOptional `json:"optional"`
	Required   GetClientToolArgsRequired `json:"required"`
}

// GetClientToolArgsRequired are the required arguments expected by the
// GetClientTool function.
type GetClientToolArgsRequired struct {
}

// GetClientToolArgsOptional are the optional arguments expected by the
// GetClientTool function.
type GetClientToolArgsOptional struct {
}
