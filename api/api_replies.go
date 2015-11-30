package api

// GetDriverNamesReply is the reply from the GetDriverNames function.
type GetDriverNamesReply struct {
	DriverNames []string `json:"driverNames"`
}

// GetVolumeMappingReply is the reply from the GetVolumeMapping function.
type GetVolumeMappingReply struct {
	BlockDevices []*BlockDevice `json:"blockDevices"`
}

// GetInstanceReply is the reply from the GetInstance function.
type GetInstanceReply struct {
	Instance *Instance `json:"instance"`
}

// GetVolumeReply is the reply from the GetVolume function.
type GetVolumeReply struct {
	Volumes []*Volume `json:"volumes"`
}

// GetVolumeAttachReply is the reply from the GetVolumeAttach function.
type GetVolumeAttachReply struct {
	Attachments []*VolumeAttachment `json:"attachments"`
}

// CreateSnapshotReply is the reply from the CreateSnapshot function.
type CreateSnapshotReply struct {
	Snapshots []*Snapshot `json:"snapshots"`
}

// GetSnapshotReply is the reply from the GetSnapshot function.
type GetSnapshotReply struct {
	Snapshots []*Snapshot `json:"snapshots"`
}

// RemoveSnapshotReply is the reply from the RemoveSnapshot function.
type RemoveSnapshotReply struct {
}

// CreateVolumeReply is the reply from the CreateVolume function.
type CreateVolumeReply struct {
	Volume *Volume `json:"volume"`
}

// RemoveVolumeReply is the reply from the RemoveVolume function.
type RemoveVolumeReply struct {
}

// AttachVolumeReply is the reply from the AttachVolume function.
type AttachVolumeReply struct {
	Attachments []*VolumeAttachment `json:"attachments"`
}

// DetachVolumeReply is the reply from the DetachVolume function.
type DetachVolumeReply struct {
}

// CopySnapshotReply is the reply from the CopySnapshot function.
type CopySnapshotReply struct {
	Snapshot *Snapshot `json:"snapshot"`
}

// GetClientToolNameReply is the reply from the GetClientToolName function.
type GetClientToolNameReply struct {
	ClientToolName string `json:"clientToolName"`
}

// GetClientToolReply is the reply from the GetClientTool function.
type GetClientToolReply struct {
	ClientTool []byte `json:"clientTool"`
}
