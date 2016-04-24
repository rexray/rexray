package drivers

import (
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
)

// NewStorageExecutor is a function that constructs a new StorageExecutors.
type NewStorageExecutor func() StorageExecutor

// NewLocalStorageDriver is a function that constructs a new
// LocalStorageDriver.
type NewLocalStorageDriver func() LocalStorageDriver

// NewRemoteStorageDriver is a function that constructs a new
// RemoteStorageDriver.
type NewRemoteStorageDriver func() RemoteStorageDriver

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

// StorageExecutor is the part of a storage driver that is downloaded at
// runtime by the libStorage client.
type StorageExecutor interface {
	Driver

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
}

// LocalStorageDriver is the client-side storage driver.
type LocalStorageDriver interface {
	Driver

	/***************************************************************************
	**                               Instance                                 **
	***************************************************************************/
	// InstanceInspectBefore may return an error, preventing the operation.
	InstanceInspectBefore(
		ctx context.Context,
		opts types.Store) error

	// InstanceInspectAfter provides an opportunity to inspect/mutate the
	// result.
	InstanceInspectAfter(
		ctx context.Context,
		result *types.Instance)

	/***************************************************************************
	**                               Volumes                                  **
	***************************************************************************/
	// VolumesBefore may return an error, preventing the operation.
	VolumesBefore(
		ctx context.Context,
		opts *VolumesOpts) error

	// VolumesAfter provides an opportunity to inspect/mutate the result.
	VolumesAfter(
		ctx context.Context,
		result []*types.Volume)

	// VolumeInspectBefore may return an error, preventing the operation.
	VolumeInspectBefore(
		ctx context.Context,
		volumeID string,
		opts *VolumeInspectOpts) (*types.Volume, error)

	// VolumeInspectAfter provides an opportunity to inspect/mutate the result.
	VolumeInspectAfter(
		ctx context.Context,
		result *types.Volume)

	// VolumeCreateBefore may return an error, preventing the operation.
	VolumeCreateBefore(
		ctx context.Context,
		name string,
		opts *VolumeCreateOpts) (*types.Volume, error)

	// VolumeCreateAfter provides an opportunity to inspect/mutate the result.
	VolumeCreateAfter(
		ctx context.Context,
		result *types.Volume)

	// VolumeCreateFromSnapshotBefore may return an error, preventing the
	// operation.
	VolumeCreateFromSnapshotBefore(
		ctx context.Context,
		snapshotID, volumeName string,
		opts types.Store) (*types.Volume, error)

	// VolumeCreateFromSnapshotAfter provides an opportunity to inspect/mutate
	// the result.
	VolumeCreateFromSnapshotAfter(
		ctx context.Context,
		result *types.Volume)

	// VolumeCopyBefore may return an error, preventing the operation.
	VolumeCopyBefore(
		ctx context.Context,
		volumeID, volumeName string,
		opts types.Store) (*types.Volume, error)

	// VolumeCopyAfter provides an opportunity to inspect/mutate the result.
	VolumeCopyAfter(
		ctx context.Context,
		result *types.Volume)

	// VolumeSnapshotBefore may return an error, preventing the operation.
	VolumeSnapshotBefore(
		ctx context.Context,
		volumeID, snapshotName string,
		opts types.Store) (*types.Snapshot, error)

	// VolumeSnapshotAfter provides an opportunity to inspect/mutate the result.
	VolumeSnapshotAfter(
		ctx context.Context,
		result *types.Volume)

	// VolumeRemoveBefore may return an error, preventing the operation.
	VolumeRemoveBefore(
		ctx context.Context,
		volumeID string,
		opts types.Store) error

	// VolumeRemoveAfter provides an opportunity to inspect/mutate the result.
	VolumeRemoveAfter(
		ctx context.Context,
		result *types.Volume)

	// VolumeAttachBefore may return an error, preventing the operation.
	VolumeAttachBefore(
		ctx context.Context,
		volumeID string,
		opts *VolumeAttachByIDOpts) (*types.Volume, error)

	// VolumeAttachAfter provides an opportunity to inspect/mutate the result.
	VolumeAttachAfter(
		ctx context.Context,
		result *types.Volume)

	// VolumeDetachBefore may return an error, preventing the operation.
	VolumeDetachBefore(
		ctx context.Context,
		volumeID string,
		opts types.Store) error

	// VolumeDetachAfter provides an opportunity to inspect/mutate the result.
	VolumeDetachAfter(
		ctx context.Context,
		result *types.Volume)

	/***************************************************************************
	**                             Snapshots                                  **
	***************************************************************************/

	// SnapshotsBefore may return an error, preventing the operation.
	SnapshotsBefore(
		ctx context.Context,
		opts types.Store) ([]*types.Snapshot, error)

	// SnapshotsAfter provides an opportunity to inspect/mutate the result.
	SnapshotsAfter(
		ctx context.Context,
		result *types.Volume)

	// SnapshotInspectBefore may return an error, preventing the operation.
	SnapshotInspectBefore(
		ctx context.Context,
		snapshotID string,
		opts types.Store) (*types.Snapshot, error)

	// SnapshotInspectAfter provides an opportunity to inspect/mutate the
	// result.
	SnapshotInspectAfter(
		ctx context.Context,
		result *types.Volume)

	// SnapshotCopyBefore may return an error, preventing the operation.
	SnapshotCopyBefore(
		ctx context.Context,
		snapshotID, snapshotName, destinationID string,
		opts types.Store) (*types.Snapshot, error)

	// SnapshotCopyAfter provides an opportunity to inspect/mutate the result.
	SnapshotCopyAfter(
		ctx context.Context,
		result *types.Volume)

	// SnapshotRemoveBefore may return an error, preventing the operation.
	SnapshotRemoveBefore(
		ctx context.Context,
		snapshotID string,
		opts types.Store) error

	// SnapshotRemoveAfter provides an opportunity to inspect/mutate the result.
	SnapshotRemoveAfter(
		ctx context.Context,
		result *types.Volume)
}

/*
RemoteStorageDriver is a libStorage driver used by the routes to implement the
backend functionality.

Functions that inspect a resource or send an operation to a resource should
always return ErrResourceNotFound if the acted upon resource cannot be found.
*/
type RemoteStorageDriver interface {
	Driver

	// NextDeviceInfo returns the information about the driver's next available
	// device workflow.
	NextDeviceInfo() *types.NextDeviceInfo

	// Type returns the type of storage the driver provides.
	Type() types.StorageType

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
		opts *VolumeCreateOpts) (*types.Volume, error)

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
