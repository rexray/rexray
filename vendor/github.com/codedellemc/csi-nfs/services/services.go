package services

import (
	"os"

	"github.com/codedellemc/gocsi/csi"
)

const (
	// SpName holds the name of the Storage Plugin / driver
	SpName    = "csi-nfs"
	spVersion = "0.1.0"

	mountDirEnvVar = "X_CSI_NFS_MOUNTDIR"
	defaultDir     = "/dev/csi-nfs-mounts"
)

var (
	// CSIVersions holds a slice of compatible CSI spec versions
	CSIVersions = []*csi.Version{
		&csi.Version{
			Major: 0,
			Minor: 0,
			Patch: 0,
		},
	}
)

// StoragePlugin contains parameters for the plugin
type StoragePlugin struct {
	privDir string
}

// Init initializes the plugin based on environment variables
func (sp *StoragePlugin) Init() {
	sp.privDir = defaultDir
	if md := os.Getenv(mountDirEnvVar); md != "" {
		sp.privDir = md
	}
}
