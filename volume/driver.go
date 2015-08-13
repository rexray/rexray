package volume

type VolumeOpts map[string]string

type Driver interface {
	// Mount will return a mount point path when specifying either a volumeName or volumeID.  If a overwriteFs boolean
	// is specified it will overwrite the FS based on newFsType if it is detected that there is no FS present.
	Mount(volumeName, volumeID string, overwriteFs bool, newFsType string) (string, error)

	// Unmount will unmount the specified volume by volumeName or volumeID.
	Unmount(volumeName, volumeID string) error

	// Path will return the mounted path of the volumeName or volumeID.
	Path(volumeName, volumeID string) (string, error)

	// Create will create a new volume with the volumeName and opts.
	Create(volumeName string, opts VolumeOpts) error

	// Remove will remove a volume of volumeName.
	Remove(volumeName string) error

	// Attach will attach a volume based on volumeName to the instance of instanceID.
	Attach(volumeName, instanceID string) (string, error)

	// Detach will detach a volume based on volumeName to the instance of instanceID.
	Detach(volumeName, instanceID string) error

	// NetworkName will return an identifier of a volume that is relevant when corelating a
	// local device to a device that is the volumeName to the local instanceID.
	NetworkName(volumeName, instanceID string) (string, error)
}
