package vfs

import (
	"path"

	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"

	"github.com/codedellemc/libstorage/api/types"
)

const (
	// Name is the name of the driver.
	Name = "vfs"
)

func init() {
	defaultRootDir := types.Lib.Join("vfs")
	r := gofigCore.NewRegistration("VFS")
	r.Key(gofig.String, "", defaultRootDir, "", "vfs.root")
	gofigCore.Register(r)
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
