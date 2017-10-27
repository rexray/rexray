package v1

import (
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/thecodeteam/goisilon/api"
)

const (
	namespacePath       = "namespace"
	exportsPath         = "platform/1/protocols/nfs/exports"
	quotaPath           = "platform/1/quota/quotas"
	snapshotsPath       = "platform/1/snapshot/snapshots"
	volumesnapshotsPath = "/ifs/.snapshot"
)

var (
	debug, _ = strconv.ParseBool(os.Getenv("GOISILON_DEBUG"))
)

func realNamespacePath(client api.Client) string {
	return path.Join(namespacePath, client.VolumesPath())
}

func realexportsPath(client api.Client) string {
	return path.Join(exportsPath, client.VolumesPath())
}

func realVolumeSnapshotPath(client api.Client, name string) string {
	parts := strings.SplitN(realNamespacePath(client), "/ifs/", 2)
	return path.Join(parts[0], volumesnapshotsPath, name, parts[1])
}
