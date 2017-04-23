// +build !libstorage_storage_driver libstorage_storage_driver_vfs

package vfs

import (
	"path"

	gofig "github.com/akutz/gofig/types"

	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"
)

const (
	// Name is the name of the driver.
	Name = "vfs"
)

func init() {
	registry.RegisterConfigReg(
		"VFS",
		func(ctx types.Context, r gofig.ConfigRegistration) {
			r.Key(
				gofig.String,
				"",
				path.Join(context.MustPathConfig(ctx).Lib, "vfs"),
				"",
				"vfs.root")
		})
}

// RootDir returns the path to the VFS root directory.
func RootDir(config gofig.Config) string {
	return config.GetString("vfs.root")
}

// DeviceFilePath returns the path to the VFS devices file.
func DeviceFilePath(config gofig.Config) string {
	return path.Join(RootDir(config), "dev")
}

// VolumesDirPath returns the path to the VFS volumes directory.
func VolumesDirPath(config gofig.Config) string {
	return path.Join(RootDir(config), "vol")
}

// SnapshotsDirPath returns the path to the VFS volumes directory.
func SnapshotsDirPath(config gofig.Config) string {
	return path.Join(RootDir(config), "snap")
}
