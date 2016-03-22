package drivers

import (
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
)

// NewStorageDriver is a function that constructs a new StorageDriver.
type NewStorageDriver func() StorageDriver

// VolumesOpts are options when inspecting a volume.
type VolumesOpts struct {
	Attachments bool
	Opts        types.Store
}

// VolumeInspectOpts are options when inspecting a volume.
type VolumeInspectOpts struct {
	Attachments bool
	Opts        types.Store
}

// VolumeCreateOpts are options when creating a new volume.
type VolumeCreateOpts struct {
	AvailabilityZone *string
	IOPS             *int64
	Size             *int64
	Type             *string
	Opts             types.Store
}

// VolumeAttachByIDOpts are options for attaching a volume by its ID.
type VolumeAttachByIDOpts struct {
	NextDevice *string
	Opts       types.Store
}

/*
StorageDriver is a libStorage driver used by the routes to implement the backend
functionality.

Functions that inspect a resource or send an operation to a resource should
always return ErrResourceNotFound if the acted upon resource cannot be found.
*/
type StorageDriver interface {
	Driver

	// Type returns the type of storage the driver provides.
	Type() types.StorageType

	// NextDeviceInfo returns the information about the driver's next available
	// device workflow.
	NextDeviceInfo() *types.NextDeviceInfo

	/***************************************************************************
	**                              Executor                                  **
	***************************************************************************/
	// InstanceID returns the local system's InstanceID.
	InstanceID(
		ctx context.Context,
		opts types.Store) (*types.InstanceID, error)

	// NextDevice returns the next available device.
	NextDevice(
		ctx context.Context,
		opts types.Store) (string, error)

	// LocalDevices returns a map of the system's local devices.
	LocalDevices(
		ctx context.Context,
		opts types.Store) (map[string]string, error)

	/***************************************************************************
	**                               Instance                                 **
	***************************************************************************/
	// InstanceInspect returns an instance.
	InstanceInspect(
		ctx context.Context,
		opts types.Store) (*types.Instance, error)

	/***************************************************************************
	**                               Volumes                                  **
	***************************************************************************/
	// Volumes returns all volumes or a filtered list of volumes.
	Volumes(
		ctx context.Context,
		opts *VolumesOpts) ([]*types.Volume, error)

	// VolumeInspect inspects a single volume.
	VolumeInspect(
		ctx context.Context,
		volumeID string,
		opts *VolumeInspectOpts) (*types.Volume, error)

	// VolumeCreate creates a new volume.
	VolumeCreate(
		ctx context.Context,
		name string,
		opts *VolumeCreateOpts) (*types.Volume, error)

	// VolumeCreateFromSnapshot creates a new volume from an existing snapshot.
	VolumeCreateFromSnapshot(
		ctx context.Context,
		snapshotID, volumeName string,
		opts types.Store) (*types.Volume, error)

	// VolumeCopy copies an existing volume.
	VolumeCopy(
		ctx context.Context,
		volumeID, volumeName string,
		opts types.Store) (*types.Volume, error)

	// VolumeSnapshot snapshots a volume.
	VolumeSnapshot(
		ctx context.Context,
		volumeID, snapshotName string,
		opts types.Store) (*types.Snapshot, error)

	// VolumeRemove removes a volume.
	VolumeRemove(
		ctx context.Context,
		volumeID string,
		opts types.Store) error

	// VolumeAttach attaches a volume.
	VolumeAttach(
		ctx context.Context,
		volumeID string,
		opts *VolumeAttachByIDOpts) (*types.Volume, error)

	// VolumeDetach detaches a volume.
	VolumeDetach(
		ctx context.Context,
		volumeID string,
		opts types.Store) error

	/***************************************************************************
	**                             Snapshots                                  **
	***************************************************************************/
	// Snapshots returns all volumes or a filtered list of snapshots.
	Snapshots(ctx context.Context, opts types.Store) ([]*types.Snapshot, error)

	// SnapshotInspect inspects a single snapshot.
	SnapshotInspect(
		ctx context.Context,
		snapshotID string,
		opts types.Store) (*types.Snapshot, error)

	// SnapshotCopy copies an existing snapshot.
	SnapshotCopy(
		ctx context.Context,
		snapshotID, snapshotName, destinationID string,
		opts types.Store) (*types.Snapshot, error)

	// SnapshotRemove removes a snapshot.
	SnapshotRemove(
		ctx context.Context,
		snapshotID string,
		opts types.Store) error
}
