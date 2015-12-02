package api

import (
	"net/http"
)

// ServiceEndpoint is the interface for the libStorage service/API.
type ServiceEndpoint interface {

	// GetServiceInfo returns information about the service.
	GetServiceInfo(
		req *http.Request,
		args *GetServiceInfoArgs,
		reply *GetServiceInfoReply) error

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

	// GetNextAvailableDeviceName gets the driver's NextAvailableDeviceName
	// information.
	GetNextAvailableDeviceName(
		req *http.Request,
		args *GetNextAvailableDeviceNameArgs,
		reply *GetNextAvailableDeviceNameReply) error

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

	// GetClientTool gets the client tool provided by the driver. This tool is
	// executed on the client-side of the connection in order to discover
	// information only available to the client, such as the client's instance
	// ID or a local device map.
	//
	// The client tool is returned as a byte array that's either a binary file
	// or a unicode-encoded, plain-text script file. Use the file extension
	// of the client tool's file name to determine the file type.
	GetClientTool(
		req *http.Request,
		args *GetClientToolArgs,
		reply *GetClientToolReply) error
}
