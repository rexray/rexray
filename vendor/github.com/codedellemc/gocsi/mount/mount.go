package mount

import "errors"

// Most of this file is based on k8s.io/pkg/util/mount

// ErrNotImplemented is returned when a platform does not implement
// the contextual function.
var ErrNotImplemented = errors.New("not implemented")

// Info is information about a single mount point.
type Info struct {
	Device string
	Path   string
	Type   string
	Opts   []string
	Freq   int
	Pass   int
}

// GetDiskFormat uses 'lsblk' to see if the given disk is unformatted.
func GetDiskFormat(disk string) (string, error) {
	return getDiskFormat(disk)
}

// FormatAndMount uses unix utils to format and mount the given disk.
func FormatAndMount(source, target, fsType string, options ...string) error {
	return formatAndMount(source, target, fsType, options)
}

// Mount mounts source to target as fstype with given options.
//
// The parameters 'source' and 'fstype' must be empty strings in case they
// are not required, e.g. for remount, or for an auto filesystem type where
// the kernel handles fstype automatically.
//
// The 'options' parameter is a list of options. Please see mount(8) for
// more information. If no options are required then please invoke Mount
// with an empty or nil argument.
func Mount(source, target, fsType string, options ...string) error {
	return mount(source, target, fsType, options)
}

// BindMount behaves like Mount was called with a "bind" flag set
// in the options list.
func BindMount(source, target string, options ...string) error {
	if options == nil {
		options = []string{"bind"}
	} else {
		options = append(options, "bind")
	}
	return mount(source, target, "", options)
}

// Unmount unmounts the target.
func Unmount(target string) error {
	return unmount(target)
}

// GetMounts returns a slice of all the mounted filesystems
func GetMounts() ([]*Info, error) {
	return getMounts()
}

// GetDevMounts returns a slice of all mounts for dev
func GetDevMounts(dev string) ([]*Info, error) {
	return getDevMounts(dev)
}

func contains(list []string, item string) bool {
	for _, x := range list {
		if x == item {
			return true
		}
	}
	return false
}
