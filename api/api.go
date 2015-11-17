package api

import (
	"net/http"

	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/model"
)

// ServiceEndpoint is the libStorage API service-side endpoing.
type ServiceEndpoint interface {
	// InitDrivers initializes the drivers for the LibStorage instance.
	InitDrivers(
		req *http.Request,
		args *InitDriversArgs,
		reply *InitDriversReply) error

	// GetRegisteredDriverNames gets the names of the registered drivers.
	GetRegisteredDriverNames(req *http.Request,
		args *GetDriverNamesArgs,
		reply *GetDriverNamesReply) error

	// GetInitializedDriverNames gets the names of the initialized drivers.
	GetInitializedDriverNames(req *http.Request,
		args *GetDriverNamesArgs,
		reply *GetDriverNamesReply) error

	// GetVolumeMapping lists the block devices that are attached to the
	// instance.
	GetVolumeMapping(
		req *http.Request,
		args *GetVolumeMappingArgs,
		reply *GetVolumeMappingReply) error

	// GetInstance retrieves the local instance.
	GetInstance(
		req *http.Request,
		args *GetInstanceArgs,
		reply *GetInstanceReply) error

	// GetVolume returns all volumes for the instance based on either volumeID
	// or volumeName that are available to the instance.
	GetVolume(
		req *http.Request,
		args *GetVolumeArgs,
		reply *GetVolumeReply) error

	// GetVolumeAttach returns the attachment details based on volumeID or
	// volumeName where the volume is currently attached.
	GetVolumeAttach(
		req *http.Request,
		args *GetVolumeAttachArgs,
		reply *GetVolumeAttachReply) error

	// CreateSnapshot is a synch/async operation that returns snapshots that
	// have been performed based on supplying a snapshotName, source volumeID,
	// and optional description.
	CreateSnapshot(
		req *http.Request,
		args *CreateSnapshotArgs,
		reply *CreateSnapshotReply) error

	// GetSnapshot returns a list of snapshots for a volume based on volumeID,
	// snapshotID, or snapshotName.
	GetSnapshot(
		req *http.Request,
		args *GetSnapshotArgs,
		reply *GetSnapshotReply) error

	// RemoveSnapshot will remove a snapshot based on the snapshotID.
	RemoveSnapshot(
		req *http.Request,
		args *RemoveSnapshotArgs,
		reply *RemoveSnapshotReply) error

	// CreateVolume is sync/async and will create an return a new/existing
	// Volume based on volumeID/snapshotID with a name of volumeName and a size
	// in GB.  Optionally based on the storage driver, a volumeType, IOPS, and
	// availabilityZone could be defined.
	CreateVolume(
		req *http.Request,
		args *CreateVolumeArgs,
		reply *CreateVolumeReply) error

	// RemoveVolume will remove a volume based on volumeID.
	RemoveVolume(
		req *http.Request,
		args *RemoveVolumeArgs,
		reply *RemoveVolumeReply) error

	// AttachVolume returns a list of VolumeAttachments is sync/async that will
	// attach a volume to an instance based on volumeID and instanceID.
	AttachVolume(
		req *http.Request,
		args *AttachVolumeArgs,
		reply *AttachVolumeReply) error

	// DetachVolume is sync/async that will detach the volumeID from the local
	// instance or the instanceID.
	DetachVolume(
		req *http.Request,
		args *DetachVolumeArgs,
		reply *DetachVolumeReply) error

	// CopySnapshot is a sync/async and returns a snapshot that will copy a
	// snapshot based on volumeID/snapshotID/snapshotName and create a new
	// snapshot of desinationSnapshotName in the destinationRegion location.
	CopySnapshot(
		req *http.Request,
		args *CopySnapshotArgs,
		reply *CopySnapshotReply) error

	// GetClientToolName gets the file name of the tool this driver provides
	// to be executed on the client-side in order to discover a client's
	// instance ID and next, available device name.
	//
	// Use the function GetClientTool to get the actual tool.
	GetClientToolName(
		req *http.Request,
		args *GetClientToolNameArgs,
		reply *GetClientToolNameReply) error

	// GetClientTool gets the file  for the tool this driver provides
	// to be executed on the client-side in order to discover a client's
	// instance ID and next, available device name.
	//
	// This function returns a byte array that will be either a binary file
	// or a unicode-encoded, plain-text script file. Use the file extension
	// of the client tool's file name to determine the file type.
	//
	// The function GetClientToolName can be used to get the file name.
	GetClientTool(
		req *http.Request,
		args *GetClientToolArgs,
		reply *GetClientToolReply) error
}

// InitDriversArgs are the arguments expected by the InitDrivers function.
type InitDriversArgs struct {
	Config *gofig.Config `json:"config"`
}

// InitDriversReply is the reply from the InitDrivers function.
type InitDriversReply struct {
	RegisteredDriverNames  []string `json:"registeredDriverNames"`
	InitializedDriverNames []string `json:"initializedDriverNames"`
}

// GetDriverNamesArgs are the arguments expected by the GetDriverNames function.
type GetDriverNamesArgs struct {
}

// GetDriverNamesReply is the reply from the GetDriverNames function.
type GetDriverNamesReply struct {
	DriverNames []string `json:"driverNames"`
}

// GetVolumeMappingArgs are the arguments expected by the GetVolumeMapping
// function.
type GetVolumeMappingArgs struct {
	InstanceID *model.InstanceID `json:"instanceID"`
}

// GetVolumeMappingReply is the reply from the GetVolumeMapping function.
type GetVolumeMappingReply struct {
	BlockDevices []*model.BlockDevice `json:"blockDevices"`
}

// GetInstanceArgs are the arguments expected by the GetInstance function.
type GetInstanceArgs struct {
	InstanceID *model.InstanceID `json:"instanceID"`
}

// GetInstanceReply is the reply from the GetInstance function.
type GetInstanceReply struct {
	Instance *model.Instance `json:"instance"`
}

// GetVolumeArgs are the arguments expected by the GetVolume function.
type GetVolumeArgs struct {
	InstanceID *model.InstanceID `json:"instanceID"`
	VolumeID   string            `json:"volumeID"`
	VolumeName string            `json:"volumeName"`
}

// GetVolumeReply is the reply from the GetVolume function.
type GetVolumeReply struct {
	Volumes []*model.Volume `json:"volumes"`
}

// GetVolumeAttachArgs are the arguments expected by the GetVolumeAttach
// function.
type GetVolumeAttachArgs struct {
	InstanceID *model.InstanceID `json:"instanceID"`
	VolumeID   string            `json:"volumeID"`
}

// GetVolumeAttachReply is the reply from the GetVolumeAttach function.
type GetVolumeAttachReply struct {
	Attachments []*model.VolumeAttachment `json:"attachments"`
}

// CreateSnapshotArgs are the arguments expected by the CreateSnapshot function.
type CreateSnapshotArgs struct {
	InstanceID   *model.InstanceID `json:"instanceID"`
	SnapshotName string            `json:"snapshotName"`
	VolumeID     string            `json:"volumeID"`
	Description  string            `json:"description"`
}

// CreateSnapshotReply is the reply from the CreateSnapshot function.
type CreateSnapshotReply struct {
	Snapshots []*model.Snapshot `json:"snapshots"`
}

// GetSnapshotArgs are the arguments expected by the GetSnapshot function.
type GetSnapshotArgs struct {
	InstanceID   *model.InstanceID `json:"instanceID"`
	VolumeID     string            `json:"volumeID"`
	SnapshotID   string            `json:"snapshotID"`
	SnapshotName string            `json:"snapshotName"`
}

// GetSnapshotReply is the reply from the GetSnapshot function.
type GetSnapshotReply struct {
	Snapshots []*model.Snapshot `json:"snapshots"`
}

// RemoveSnapshotArgs are the arguments expected by the RemoveSnapshot function.
type RemoveSnapshotArgs struct {
	InstanceID *model.InstanceID `json:"instanceID"`
	SnapshotID string            `json:"snapshotID"`
}

// RemoveSnapshotReply is the reply from the RemoveSnapshot function.
type RemoveSnapshotReply struct {
}

// CreateVolumeArgs are the arguments expected by the CreateVolume function.
type CreateVolumeArgs struct {
	InstanceID       *model.InstanceID `json:"instanceID"`
	VolumeName       string            `json:"volumeName"`
	VolumeID         string            `json:"volumeID"`
	SnapshotID       string            `json:"snapshotID"`
	VolumeType       string            `json:"volumeType"`
	IOPS             int64             `json:"iops"`
	Size             int64             `json:"size"`
	AvailabilityZone string            `json:"availabilityZone"`
}

// CreateVolumeReply is the reply from the CreateVolume function.
type CreateVolumeReply struct {
	Volume *model.Volume `json:"volume"`
}

// RemoveVolumeArgs are the arguments expected by the RemoveVolume function.
type RemoveVolumeArgs struct {
	InstanceID *model.InstanceID `json:"instanceID"`
	VolumeID   string            `json:"volumeID"`
}

// RemoveVolumeReply is the reply from the RemoveVolume function.
type RemoveVolumeReply struct {
}

// AttachVolumeArgs are the arguments expected by the AttachVolume function.
type AttachVolumeArgs struct {
	InstanceID     *model.InstanceID `json:"instanceID"`
	NextDeviceName string            `json:"nextDeviceName"`
	VolumeID       string            `json:"volumeID"`
}

// AttachVolumeReply is the reply from the AttachVolume function.
type AttachVolumeReply struct {
	Attachments []*model.VolumeAttachment `json:"attachments"`
}

// DetachVolumeArgs are the arguments expected by the DetachVolume function.
type DetachVolumeArgs struct {
	InstanceID *model.InstanceID `json:"instanceID"`
	VolumeID   string            `json:"volumeID"`
}

// DetachVolumeReply is the reply from the DetachVolume function.
type DetachVolumeReply struct {
}

// CopySnapshotArgs are the arguments expected by the CopySnapshot function.
type CopySnapshotArgs struct {
	InstanceID              *model.InstanceID `json:"instanceID"`
	VolumeID                string            `json:"volumeID"`
	SnapshotID              string            `json:"snapshotID"`
	SnapshotName            string            `json:"snapshotName"`
	DestinationSnapshotName string            `json:"destinationSnapshotName"`
	DestinationRegion       string            `json:"destinationRegion"`
}

// CopySnapshotReply is the reply from the CopySnapshot function.
type CopySnapshotReply struct {
	Snapshot *model.Snapshot `json:"snapshot"`
}

// GetClientToolNameArgs are the arguments expected by the GetClientToolName
// function.
type GetClientToolNameArgs struct {
	InstanceID *model.InstanceID `json:"instanceID"`
}

// GetClientToolNameReply is the reply from the GetClientToolName function.
type GetClientToolNameReply struct {
	ClientToolName string `json:"clientToolName"`
}

// GetClientToolArgs are the arguments expected by the GetClientTool function.
type GetClientToolArgs struct {
	InstanceID *model.InstanceID `json:"instanceID"`
}

// GetClientToolReply is the reply from the GetClientTool function.
type GetClientToolReply struct {
	ClientTool []byte `json:"clientTool"`
}
