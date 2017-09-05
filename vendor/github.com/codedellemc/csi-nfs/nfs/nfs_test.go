package nfs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/codedellemc/csi-blockdevices/block"
)

var (
	nfsHost = os.Getenv("NFS_HOST")
	nfsPath = os.Getenv("NFS_PATH")
	mntPath string
)

func TestMountUnmount(t *testing.T) {
	if len(nfsHost) == 0 || len(nfsPath) == 0 {
		t.Skip("nfs server details not set")
	}

	mntPath, err := ioutil.TempDir("", "csinfsplugin")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v\n", err)
	}
	defer os.Remove(mntPath)

	err = block.Mount(nfsHost+":"+nfsPath, mntPath, "nfs", []string{})
	if err != nil {
		t.Fatal(err)
	}

	err = block.Unmount(mntPath)
	if err != nil {
		t.Fatal(err)
	}
}
