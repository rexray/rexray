package types

// NewIntegrationDriver is a function that constructs a new IntegrationDriver.
type NewIntegrationDriver func() IntegrationDriver

// VolumeMountOpts are options for mounting a volume.
type VolumeMountOpts struct {
	OverwriteFS bool
	NewFSType   string
	Preempt     bool
	Opts        Store
}

// IntegrationDriver is the interface implemented to integrate external
// storage consumers, such as Docker, with libStorage.
type IntegrationDriver interface {
	Driver

	// Volumes returns all available volumes.
	Volumes(
		ctx Context,
		opts Store) ([]*Volume, error)

	// Inspect returns a specific volume as identified by the provided
	// volume name.
	Inspect(
		ctx Context,
		volumeName string,
		opts Store) (*Volume, error)

	// Mount will return a mount point path when specifying either a volumeName
	// or volumeID.  If a overwriteFs boolean is specified it will overwrite
	// the FS based on newFsType if it is detected that there is no FS present.
	Mount(
		ctx Context,
		volumeID, volumeName string,
		opts *VolumeMountOpts) (string, *Volume, error)

	// Unmount will unmount the specified volume by volumeName or volumeID.
	Unmount(
		ctx Context,
		volumeID, volumeName string,
		opts Store) error

	// Path will return the mounted path of the volumeName or volumeID.
	Path(
		ctx Context,
		volumeID, volumeName string,
		opts Store) (string, error)

	// Create will create a new volume with the volumeName and opts.
	Create(
		ctx Context,
		volumeName string,
		opts *VolumeCreateOpts) (*Volume, error)

	// Remove will remove a volume of volumeName.
	Remove(
		ctx Context,
		volumeName string,
		opts Store) error

	// Attach will attach a volume based on volumeName to the instance of
	// instanceID.
	Attach(
		ctx Context,
		volumeName string,
		opts *VolumeAttachOpts) (string, error)

	// Detach will detach a volume based on volumeName to the instance of
	// instanceID.
	Detach(
		ctx Context,
		volumeName string,
		opts *VolumeDetachOpts) error

	// NetworkName will return an identifier of a volume that is relevant when
	// corelating a local device to a device that is the volumeName to the
	// local instanceID.
	NetworkName(
		ctx Context,
		volumeName string,
		opts Store) (string, error)
}
