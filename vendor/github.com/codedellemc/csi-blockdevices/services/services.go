package services

import (
	"os"
	"path/filepath"

	"github.com/codedellemc/gocsi/csi"
)

const (
	SpName    = "csi-blockdevices"
	spVersion = "0.1.0"

	blockDirEnvVar = "BDPLUGIN_DEVDIR"

	defaultDevDir = "/dev/disk/csi-blockdevices"
)

var (
	CSIVersions = []*csi.Version{
		&csi.Version{
			Major: 0,
			Minor: 1,
			Patch: 0,
		},
	}
)

type StoragePlugin struct {
	DevDir  string
	privDir string
}

func (sp *StoragePlugin) Init() {
	sp.DevDir = defaultDevDir
	if dd := os.Getenv(blockDirEnvVar); dd != "" {
		sp.DevDir = dd
	}

	sp.privDir = filepath.Join(sp.DevDir, ".mounts")
}
