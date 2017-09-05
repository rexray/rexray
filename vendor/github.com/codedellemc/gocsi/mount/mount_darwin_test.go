// +build darwin

package mount_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/codedellemc/gocsi/mount"
)

func TestGetMounts(t *testing.T) {
	mounts, err := mount.GetMounts()
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	for _, m := range mounts {
		t.Logf("%+v", m)
	}
}

func TestBindMount(t *testing.T) {
	src, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	tgt, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		mount.Unmount(tgt)
		os.RemoveAll(tgt)
		os.RemoveAll(src)
	}()
	if err := mount.BindMount(src, tgt); err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	t.Logf("bind mount success: source=%s, target=%s", src, tgt)
	TestGetMounts(t)
}
