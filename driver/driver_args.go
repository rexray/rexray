// +build OMIT

package driver

// GetVolumeMappingArgs are the arguments expected by the GetVolumeMapping
// function.
type GetVolumeMappingArgs struct {
	Extensions map[string]interface{} `json:"extensions"`
}

// GetInstanceArgs are the arguments expected by the GetInstance function.
type GetInstanceArgs struct {
	Extensions map[string]interface{} `json:"extensions"`
}

// GetVolumeArgs are the arguments expected by the GetVolume function.
type GetVolumeArgs struct {
	VolumeID   string                 `json:"volumeID"`
	VolumeName string                 `json:"volumeName"`
	Extensions map[string]interface{} `json:"extensions"`
}

// GetVolumeAttachArgs are the arguments expected by the GetVolumeAttach
// function.
type GetVolumeAttachArgs struct {
	VolumeID   string                 `json:"volumeID"`
	Extensions map[string]interface{} `json:"extensions"`
}

// CreateSnapshotArgs are the arguments expected by the CreateSnapshot function.
type CreateSnapshotArgs struct {
	Description  string                 `json:"description"`
	SnapshotName string                 `json:"snapshotName"`
	VolumeID     string                 `json:"volumeID"`
	Extensions   map[string]interface{} `json:"extensions"`
}

// GetSnapshotArgs are the arguments expected by the GetSnapshot function.
type GetSnapshotArgs struct {
	SnapshotID   string                 `json:"snapshotID"`
	SnapshotName string                 `json:"snapshotName"`
	VolumeID     string                 `json:"volumeID"`
	Extensions   map[string]interface{} `json:"extensions"`
}

// RemoveSnapshotArgs are the arguments expected by the RemoveSnapshot function.
type RemoveSnapshotArgs struct {
	SnapshotID string                 `json:"snapshotID"`
	Extensions map[string]interface{} `json:"extensions"`
}

// CreateVolumeArgs are the arguments expected by the CreateVolume function.
type CreateVolumeArgs struct {
	VolumeID         string                 `json:"volumeID"`
	AvailabilityZone string                 `json:"availabilityZone"`
	VolumeName       string                 `json:"volumeName"`
	Size             int64                  `json:"size"`
	SnapshotID       string                 `json:"snapshotID"`
	IOPS             int64                  `json:"iops"`
	VolumeType       string                 `json:"volumeType"`
	Extensions       map[string]interface{} `json:"extensions"`
}

// RemoveVolumeArgs are the arguments expected by the RemoveVolume function.
type RemoveVolumeArgs struct {
	VolumeID   string                 `json:"volumeID"`
	Extensions map[string]interface{} `json:"extensions"`
}

// AttachVolumeArgs are the arguments expected by the AttachVolume function.
type AttachVolumeArgs struct {
	NextDeviceName string                 `json:"nextDeviceName"`
	VolumeID       string                 `json:"volumeID"`
	Extensions     map[string]interface{} `json:"extensions"`
}

// DetachVolumeArgs are the arguments expected by the DetachVolume function.
type DetachVolumeArgs struct {
	VolumeID   string                 `json:"volumeID"`
	Extensions map[string]interface{} `json:"extensions"`
}

// CopySnapshotArgs are the arguments expected by the CopySnapshot function.
type CopySnapshotArgs struct {
	DestinationRegion       string                 `json:"destinationRegion"`
	DestinationSnapshotName string                 `json:"destinationSnapshotName"`
	SnapshotID              string                 `json:"snapshotID"`
	SnapshotName            string                 `json:"snapshotName"`
	VolumeID                string                 `json:"volumeID"`
	Extensions              map[string]interface{} `json:"extensions"`
}

// GetClientToolNameArgs are the arguments expected by the GetClientToolName
// function.
type GetClientToolNameArgs struct {
	Extensions map[string]interface{} `json:"extensions"`
}

// GetClientToolArgs are the arguments expected by the GetClientTool function.
type GetClientToolArgs struct {
	Extensions map[string]interface{} `json:"extensions"`
}
