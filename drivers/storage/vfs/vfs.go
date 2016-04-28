package vfs

import (
	"fmt"

	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/utils/paths"
)

const (
	// Name is the name of the driver.
	Name = "vfs"
)

func init() {
	configRegistration()
}

func configRegistration() {
	defaultRootDir := fmt.Sprintf("%s/vfs", paths.UsrDirPath())
	r := gofig.NewRegistration("VFS")
	r.Key(gofig.String, "", defaultRootDir, "", "vfs.root")
	gofig.Register(r)
}

// RootDir returns the path to the VFS root directory.
func RootDir(config gofig.Config) string {
	return config.GetString("vfs.root")
}

// DeviceFilePath returns the path to the VFS devices file.
func DeviceFilePath(config gofig.Config) string {
	return fmt.Sprintf("%s/dev", RootDir(config))
}

// VolumesDirPath returns the path to the VFS volumes directory.
func VolumesDirPath(config gofig.Config) string {
	return fmt.Sprintf("%s/vol", RootDir(config))
}

// SnapshotsDirPath returns the path to the VFS volumes directory.
func SnapshotsDirPath(config gofig.Config) string {
	return fmt.Sprintf("%s/snap", RootDir(config))
}
