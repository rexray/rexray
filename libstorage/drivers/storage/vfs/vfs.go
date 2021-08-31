package vfs

import (
	"path"

	gofig "github.com/akutz/gofig/types"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

const (
	// Name is the name of the driver.
	Name = "vfs"
)

func init() {
	registry.RegisterConfigReg(
		"VFS",
		func(ctx types.Context, r gofig.ConfigRegistration) {
			vfsRoot := path.Join(context.MustPathConfig(ctx).Lib, "vfs")
			r.Key(
				gofig.String,
				"",
				vfsRoot,
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
