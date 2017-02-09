package types

import "strconv"

// LibStorageDriverName is the name of the libStorage storage driver.
const LibStorageDriverName = "libstorage"

// NewStorageDriver is a function that constructs a new StorageDriver.
type NewStorageDriver func() StorageDriver

// VolumeAttachmentsTypes is the type of the volume attachments bitmask.
type VolumeAttachmentsTypes int

const (

	// VolumeAttachmentsRequested indicates attachment information is requested.
	VolumeAttachmentsRequested VolumeAttachmentsTypes = 1 << iota // 1

	// VolumeAttachmentsMine indicates attachment information should
	// be returned for volumes attached to the instance specified in the
	// instance ID request header. If this bit is set then the instance ID
	// header is required.
	VolumeAttachmentsMine // 2

	// VolumeAttachmentsDevices indicates an attempt should made to map devices
	// provided via the local devices request header to the appropriate
	// attachment information. If this bit is set then the instance ID and
	// local device headers are required.
	VolumeAttachmentsDevices // 4

	// VolumeAttachmentsAttached indicates only volumes that are attached
	// should be returned.
	VolumeAttachmentsAttached // 8

	// VolumeAttachmentsUnattached indicates only volumes that are unattached
	// should be returned.
	VolumeAttachmentsUnattached // 16
)

const (
	// VolAttNone is the default value. This indicates no attachment
	// information is requested.
	VolAttNone VolumeAttachmentsTypes = 0

	// VolAttFalse is an alias for VolAttNone.
	VolAttFalse = VolAttNone

	// VolAttReq requests attachment information for all retrieved volumes.
	//
	// Mask: 1
	VolAttReq = VolumeAttachmentsRequested

	// VolAttReqForInstance requests attachment information for volumes attached
	// to the instance provided in the instance ID
	//
	// Mask: 1 | 2
	VolAttReqForInstance = VolAttReq | VolumeAttachmentsMine

	// VolAttReqWithDevMapForInstance requests attachment information for
	// volumes attached to the instance provided in the instance ID and perform
	// device mappings where possible.
	//
	// Mask: 1 | 2 | 4
	VolAttReqWithDevMapForInstance = VolAttReqForInstance |
		VolumeAttachmentsDevices

	// VolAttReqOnlyAttachedVols requests attachment information for all
	// retrieved volumes and return only volumes that are attached to some
	// instance.
	//
	// Mask: 1 | 8
	VolAttReqOnlyAttachedVols = VolAttReq | VolumeAttachmentsAttached

	// VolAttReqOnlyUnattachedVols requests attachment information for
	// all retrieved volumes and return only volumes that are not attached to
	// any instance.
	//
	// Mask: 1 | 16
	VolAttReqOnlyUnattachedVols = VolAttReq | VolumeAttachmentsUnattached

	// VolAttReqOnlyVolsAttachedToInstance requests attachment
	// information for all retrieved volumes and return only volumes that
	// attached to the instance provided in the instance ID.
	//
	// Mask: 1 | 2 | 8
	VolAttReqOnlyVolsAttachedToInstance = VolAttReqForInstance |
		VolumeAttachmentsAttached

	// VolAttReqWithDevMapOnlyVolsAttachedToInstance requests attachment
	// information for all retrieved volumes and return only volumes that
	// attached to the instance provided in the instance ID and perform device
	// mappings where possible.
	//
	// Mask: 1 | 2 | 4 | 8
	VolAttReqWithDevMapOnlyVolsAttachedToInstance = VolumeAttachmentsDevices |
		VolAttReqOnlyVolsAttachedToInstance

	// VolAttReqTrue is an alias for
	// VolAttReqWithDevMapOnlyVolsAttachedToInstance.
	VolAttReqTrue = VolAttReqWithDevMapOnlyVolsAttachedToInstance

	// VolumeAttachmentsTrue is an alias for VolAttReqTrue.
	VolumeAttachmentsTrue = VolAttReqTrue

	// VolAttReqOnlyVolsAttachedToInstanceOrUnattachedVols requests attachment
	// information for all retrieved volumes and return only volumes that
	// attached to the instance provided in the instance ID or are not attached
	// to any instance at all. tl;dr - Attached To Me or Available
	//
	// Mask: 1 | 2 | 8 | 16
	VolAttReqOnlyVolsAttachedToInstanceOrUnattachedVols = 0 |
		VolAttReqOnlyVolsAttachedToInstance |
		VolumeAttachmentsUnattached

	// VolAttReqWithDevMapOnlyVolsAttachedToInstanceOrUnattachedVols requests
	// attachment information for all retrieved volumes and return only volumes
	// that attached to the instance provided in the instance ID or are not
	// attached to any instance at all and perform device mappings where
	// possible. tl;dr - Attached To Me With Device Mappings or Available
	//
	// Mask: 1 | 2 | 4 | 8 | 16
	VolAttReqWithDevMapOnlyVolsAttachedToInstanceOrUnattachedVols = 0 |
		VolumeAttachmentsDevices |
		VolAttReqOnlyVolsAttachedToInstanceOrUnattachedVols
)

// ParseVolumeAttachmentTypes parses a value into a VolumeAttachmentsTypes
// value.
func ParseVolumeAttachmentTypes(v interface{}) VolumeAttachmentsTypes {
	switch tv := v.(type) {
	case VolumeAttachmentsTypes:
		return tv
	case int:
		return VolumeAttachmentsTypes(tv)
	case uint:
		return VolumeAttachmentsTypes(tv)
	case int8:
		return VolumeAttachmentsTypes(tv)
	case uint8:
		return VolumeAttachmentsTypes(tv)
	case int16:
		return VolumeAttachmentsTypes(tv)
	case uint16:
		return VolumeAttachmentsTypes(tv)
	case int32:
		return VolumeAttachmentsTypes(tv)
	case uint32:
		return VolumeAttachmentsTypes(tv)
	case int64:
		return VolumeAttachmentsTypes(tv)
	case uint64:
		return VolumeAttachmentsTypes(tv)
	case string:
		if i, err := strconv.ParseInt(tv, 10, 64); err == nil {
			return ParseVolumeAttachmentTypes(i)
		}
		if b, err := strconv.ParseBool(tv); err == nil {
			return ParseVolumeAttachmentTypes(b)
		}
	case bool:
		if tv {
			return VolumeAttachmentsTrue
		}
		return VolumeAttachmentsRequested
	}
	return VolAttNone
}

// RequiresInstanceID returns a flag that indicates whether the attachment
// bit requires an instance ID to perform successfully.
func (v VolumeAttachmentsTypes) RequiresInstanceID() bool {
	return v.Mine() || v.Devices()
}

// Requested returns a flag that indicates attachment information is requested.
func (v VolumeAttachmentsTypes) Requested() bool {
	return v.bitSet(VolumeAttachmentsRequested)
}

// Mine returns a flag that indicates attachment information should
// be returned for volumes attached to the instance specified in the
// instance ID request header. If this bit is set then the instance ID
// header is required.
func (v VolumeAttachmentsTypes) Mine() bool {
	return v.bitSet(VolumeAttachmentsMine)
}

// Devices returns a flag that indicates an attempt should made to map devices
// provided via the local devices request header to the appropriate
// attachment information. If this bit is set then the instance ID and
// local device headers are required.
func (v VolumeAttachmentsTypes) Devices() bool {
	return v.bitSet(VolumeAttachmentsDevices)
}

// Attached returns a flag that indicates only volumes that are attached should
// be returned.
func (v VolumeAttachmentsTypes) Attached() bool {
	return v.bitSet(VolumeAttachmentsAttached)
}

// Unattached returns a flag that indicates only volumes that are unattached
// should be returned.
func (v VolumeAttachmentsTypes) Unattached() bool {
	return v.bitSet(VolumeAttachmentsUnattached)
}

func (v VolumeAttachmentsTypes) bitSet(b VolumeAttachmentsTypes) bool {
	return v&b == b
}

// VolumesOpts are options when inspecting a volume.
type VolumesOpts struct {
	Attachments VolumeAttachmentsTypes
	Opts        Store
}

// VolumeInspectOpts are options when inspecting a volume.
type VolumeInspectOpts struct {
	Attachments VolumeAttachmentsTypes
	Opts        Store
}

// VolumeCreateOpts are options when creating a new volume.
type VolumeCreateOpts struct {
	AvailabilityZone *string
	IOPS             *int64
	Size             *int64
	Type             *string
	Encrypted        *bool
	EncryptionKey    *string
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

// VolumeRemoveOpts are options for removing a volume.
type VolumeRemoveOpts struct {
	Force bool
	Opts  Store
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
		ctx Context) (*NextDeviceInfo, error)

	// Type returns the type of storage the driver provides.
	Type(
		ctx Context) (StorageType, error)

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
		opts *VolumeRemoveOpts) error

	// VolumeAttach attaches a volume and provides a token clients can use
	// to validate that device has appeared locally.
	VolumeAttach(
		ctx Context,
		volumeID string,
		opts *VolumeAttachOpts) (*Volume, string, error)

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

// StorageDriverWithLogin is a StorageDriver with a Login function.
type StorageDriverWithLogin interface {
	StorageDriver

	// Login creates a new connection to the storage platform for the provided
	// context.
	Login(
		ctx Context) (interface{}, error)
}
