package v2

import (
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/codedellemc/goisilon/api"
)

const (
	namespacePath       = "namespace"
	exportsPath         = "platform/2/protocols/nfs/exports"
	quotaPath           = "platform/2/quota/quotas"
	snapshotsPath       = "platform/2/snapshot/snapshots"
	volumeSnapshotsPath = "/ifs/.snapshot"
)

var (
	debug, _   = strconv.ParseBool(os.Getenv("GOISILON_DEBUG"))
	colonBytes = []byte{byte(':')}
)

func realNamespacePath(c api.Client) string {
	return path.Join(namespacePath, c.VolumesPath())
}

func realExportsPath(c api.Client) string {
	return path.Join(exportsPath, c.VolumesPath())
}

func realVolumeSnapshotPath(c api.Client, name string) string {
	parts := strings.SplitN(realNamespacePath(c), "/ifs/", 2)
	return path.Join(parts[0], volumeSnapshotsPath, name, parts[1])
}
