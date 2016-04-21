package drivers

import (
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
)

// NewOSDriver is a function that constructs a new OSDriver.
type NewOSDriver func() OSDriver

// DeviceMountOpts are options when mounting a device.
type DeviceMountOpts struct {
	MountOptions string
	MountLabel   string
	Opts         types.Store
}

// DeviceFormatOpts are options when formatting a device.
type DeviceFormatOpts struct {
	NewFSType   string
	OverwriteFS bool
	Opts        types.Store
}

// OSDriver is the interface implemented by types that provide OS introspection
// and management.
type OSDriver interface {
	Driver

	// Mounts get a list of mount points for a local device.
	Mounts(
		ctx context.Context,
		deviceName, mountPoint string,
		opts types.Store) ([]*types.MountInfo, error)

	// Mount mounts a device to a specified path.
	Mount(
		ctx context.Context,
		deviceName, mountPoint string,
		opts *DeviceMountOpts) error

	// Unmount unmounts the underlying device from the specified path.
	Unmount(
		ctx context.Context,
		mountPoint string,
		opts types.Store) error

	// IsMounted checks whether a path is mounted or not
	IsMounted(
		ctx context.Context,
		mountPoint string,
		opts types.Store) (bool, error)

	// Format formats a device.
	Format(
		ctx context.Context,
		deviceName string,
		opts *DeviceFormatOpts) error
}
