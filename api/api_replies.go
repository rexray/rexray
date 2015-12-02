package api

// GetServiceInfoReply is the reply from the GetServiceInfo function.
type GetServiceInfoReply struct {
	Name              string   `json:"name"`
	Driver            string   `json:"driver"`
	RegisteredDrivers []string `json:"registeredDrivers"`
}

// GetNextAvailableDeviceNameReply is the reply from the
// GetNextAvailableDeviceName function.
type GetNextAvailableDeviceNameReply struct {
	Next *NextAvailableDeviceName `json:"next"`
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

// GetClientToolReply is the reply from the GetClientTool function.
type GetClientToolReply struct {
	ClientTool *ClientTool `json:"clientTool"`
}
