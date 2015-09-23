package os

import (
	"errors"

	"github.com/docker/docker/pkg/mount"
)

var (
	ErrDriverInstanceDiscovery = errors.New("Driver Instance discovery failed")
)

type Driver interface {
	// Name returns the name of the driver
	Name() string

	// Shows the existing mount points
	GetMounts(string, string) ([]*mount.Info, error)

	// Check whether path is mounted or not
	Mounted(string) (bool, error)

	// Unmount based on a path
	Unmount(string) error

	// Mount based on a device, target, options, label
	Mount(string, string, string, string) error

	// Format a device with a FS type
	Format(string, string, bool) error
}
