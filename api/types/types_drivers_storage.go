package types

// LibStorageDriverName is the name of the libStorage storage driver.
const LibStorageDriverName = "libstorage"

// NewStorageExecutor is a function that constructs a new StorageExecutors.
type NewStorageExecutor func() StorageExecutor

// NewStorageDriver is a function that constructs a new StorageDriver.
type NewStorageDriver func() StorageDriver

// VolumesOpts are options when inspecting a volume.
type VolumesOpts struct {
	Attachments bool
	Opts        Store
}

// VolumeInspectOpts are options when inspecting a volume.
type VolumeInspectOpts struct {
	Attachments bool
	Opts        Store
}

// VolumeCreateOpts are options when creating a new volume.
type VolumeCreateOpts struct {
	AvailabilityZone *string
	IOPS             *int64
	Size             *int64
	Type             *string
	Opts             Store
}

// VolumeAttachOpts are options for attaching a volume.
type VolumeAttachOpts struct {
	NextDevice *string
	Force      bool
	Opts       Store
}

// VolumeDetachOpts are options for detaching a volume.
type VolumeDetachOpts struct {
	Force bool
	Opts  Store
}

// StorageExecutor is the part of a storage driver that is downloaded at
// runtime by the libStorage client.
type StorageExecutor interface {
	Driver

	// InstanceID returns the local system's InstanceID.
	InstanceID(
		ctx Context,
		opts Store) (*InstanceID, error)

	// NextDevice returns the next available device.
	NextDevice(
		ctx Context,
		opts Store) (string, error)

	// LocalDevices returns a map of the system's local devices.
	LocalDevices(
		ctx Context,
		opts Store) (map[string]string, error)
}

// StorageDriverManager is the management wrapper for a StorageDriver.
type StorageDriverManager interface {
	StorageDriver

	// Driver returns the underlying driver.
	Driver() StorageDriver
}

/*
StorageDriver is a libStorage driver used by the routes to implement the
backend functionality.

Functions that inspect a resource or send an operation to a resource should
always return ErrResourceNotFound if the acted upon resource cannot be found.
*/
type StorageDriver interface {
	Driver

	// NextDeviceInfo returns the information about the driver's next available
	// device workflow.
	NextDeviceInfo(
		ctx Context) *NextDeviceInfo

	// Type returns the type of storage the driver provides.
	Type(
		ctx Context) StorageType

	// InstanceInspect returns an instance.
	InstanceInspect(
		ctx Context,
		opts Store) (*Instance, error)

	// Volumes returns all volumes or a filtered list of volumes.
	Volumes(
		ctx Context,
		opts *VolumesOpts) ([]*Volume, error)

	// VolumeInspect inspects a single volume.
	VolumeInspect(
		ctx Context,
		volumeID string,
		opts *VolumeInspectOpts) (*Volume, error)

	// VolumeCreate creates a new volume.
	VolumeCreate(
		ctx Context,
		name string,
		opts *VolumeCreateOpts) (*Volume, error)

	// VolumeCreateFromSnapshot creates a new volume from an existing snapshot.
	VolumeCreateFromSnapshot(
		ctx Context,
		snapshotID,
		volumeName string,
		opts *VolumeCreateOpts) (*Volume, error)

	// VolumeCopy copies an existing volume.
	VolumeCopy(
		ctx Context,
		volumeID,
		volumeName string,
		opts Store) (*Volume, error)

	// VolumeSnapshot snapshots a volume.
	VolumeSnapshot(
		ctx Context,
		volumeID,
		snapshotName string,
		opts Store) (*Snapshot, error)

	// VolumeRemove removes a volume.
	VolumeRemove(
		ctx Context,
		volumeID string,
		opts Store) error

	// VolumeAttach attaches a volume.
	VolumeAttach(
		ctx Context,
		volumeID string,
		opts *VolumeAttachOpts) (*Volume, error)

	// VolumeDetach detaches a volume.
	VolumeDetach(
		ctx Context,
		volumeID string,
		opts *VolumeDetachOpts) (*Volume, error)

	// Snapshots returns all volumes or a filtered list of snapshots.
	Snapshots(
		ctx Context,
		opts Store) ([]*Snapshot, error)

	// SnapshotInspect inspects a single snapshot.
	SnapshotInspect(
		ctx Context,
		snapshotID string,
		opts Store) (*Snapshot, error)

	// SnapshotCopy copies an existing snapshot.
	SnapshotCopy(
		ctx Context,
		snapshotID,
		snapshotName,
		destinationID string,
		opts Store) (*Snapshot, error)

	// SnapshotRemove removes a snapshot.
	SnapshotRemove(
		ctx Context,
		snapshotID string,
		opts Store) error
}
