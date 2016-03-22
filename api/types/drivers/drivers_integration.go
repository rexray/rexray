package drivers

import (
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
)

// NewIntegrationDriver is a function that constructs a new IntegrationDriver.
type NewIntegrationDriver func() IntegrationDriver

// VolumeMountOpts are options for mounting a volume.
type VolumeMountOpts struct {
	OverwriteFS bool
	NewFSType   string
	Preempt     bool
	Opts        types.Store
}

// VolumeAttachByNameOpts are options for attaching a volume by its name.
type VolumeAttachByNameOpts struct {
	Force bool
	Opts  types.Store
}

// VolumeDetachByNameOpts are options for detaching a volume by its name.
type VolumeDetachByNameOpts struct {
	Force bool
	Opts  types.Store
}

// IntegrationDriver is the interface implemented to integrate external
// storage consumers, such as Docker, with libStorage.
type IntegrationDriver interface {
	Driver

	// Volumes returns all available volumes.
	Volumes(
		ctx context.Context,
		opts types.Store) ([]map[string]string, error)

	// Inspect returns a specific volume as identified by the provided
	// volume name.
	Inspect(
		ctx context.Context,
		volumeName string,
		opts types.Store) (map[string]string, error)

	// Mount will return a mount point path when specifying either a volumeName
	// or volumeID.  If a overwriteFs boolean is specified it will overwrite
	// the FS based on newFsType if it is detected that there is no FS present.
	Mount(
		ctx context.Context,
		volumeID, volumeName string,
		opts *VolumeMountOpts) (string, error)

	// Unmount will unmount the specified volume by volumeName or volumeID.
	Unmount(
		ctx context.Context,
		volumeID, volumeName string,
		opts types.Store) error

	// Path will return the mounted path of the volumeName or volumeID.
	Path(
		ctx context.Context,
		volumeID, volumeName string,
		opts types.Store) (string, error)

	// Create will create a new volume with the volumeName and opts.
	Create(
		ctx context.Context,
		volumeName string,
		opts types.Store) error

	// Remove will remove a volume of volumeName.
	Remove(
		ctx context.Context,
		volumeName string,
		opts types.Store) error

	// Attach will attach a volume based on volumeName to the instance of
	// instanceID.
	Attach(
		ctx context.Context,
		volumeName string,
		opts *VolumeAttachByNameOpts) (string, error)

	// Detach will detach a volume based on volumeName to the instance of
	// instanceID.
	Detach(
		ctx context.Context,
		volumeName string,
		opts *VolumeDetachByNameOpts) error

	// NetworkName will return an identifier of a volume that is relevant when
	// corelating a local device to a device that is the volumeName to the
	// local instanceID.
	NetworkName(
		ctx context.Context,
		volumeName string,
		opts types.Store) (string, error)
}
