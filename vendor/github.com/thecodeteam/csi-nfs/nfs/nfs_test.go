package nfs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/thecodeteam/gocsi/mount"
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

	err = mount.Mount(nfsHost+":"+nfsPath, mntPath, "nfs", "")
	if err != nil {
		t.Fatal(err)
	}

	err = mount.Unmount(mntPath)
	if err != nil {
		t.Fatal(err)
	}
}
