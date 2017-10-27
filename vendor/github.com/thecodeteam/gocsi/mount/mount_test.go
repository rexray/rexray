package mount

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestBindMount(t *testing.T) {
	src, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	tgt, err := ioutil.TempDir("", "")
	if err != nil {
		os.RemoveAll(src)
		t.Fatal(err)
	}
	if err := EvalSymlinks(&src); err != nil {
		os.RemoveAll(tgt)
		os.RemoveAll(src)
		t.Fatal(err)
	}
	if err := EvalSymlinks(&tgt); err != nil {
		os.RemoveAll(tgt)
		os.RemoveAll(src)
		t.Fatal(err)
	}
	defer func() {
		Unmount(tgt)
		os.RemoveAll(tgt)
		os.RemoveAll(src)
	}()
	if err := BindMount(src, tgt); err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	t.Logf("bind mount success: source=%s, target=%s", src, tgt)
	mounts, err := GetMounts()
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	success := false
	for _, m := range mounts {
		if m.Source == src && m.Path == tgt {
			success = true
		}
		t.Logf("%+v", m)
	}
	if !success {
		t.Errorf("unable to find bind mount: src=%s, tgt=%s", src, tgt)
		t.Fail()
	}
}

func TestGetMounts(t *testing.T) {
	mounts, err := GetMounts()
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	for _, m := range mounts {
		t.Logf("%+v", m)
	}
}

func TestMountArgs(t *testing.T) {
	tests := []struct {
		src    string
		tgt    string
		fst    string
		opts   []string
		result string
	}{
		{
			src:    "localhost:/data",
			tgt:    "/mnt",
			fst:    "nfs",
			result: "-t nfs localhost:/data /mnt",
		},
		{
			src:    "localhost:/data",
			tgt:    "/mnt",
			result: "localhost:/data /mnt",
		},
		{
			src:    "localhost:/data",
			tgt:    "/mnt",
			fst:    "nfs",
			opts:   []string{"tcp", "vers=4"},
			result: "-t nfs -o tcp,vers=4 localhost:/data /mnt",
		},
		{
			src:    "/dev/disk/mydisk",
			tgt:    "/mnt/mydisk",
			fst:    "xfs",
			opts:   []string{"ro", "noatime", "ro"},
			result: "-t xfs -o ro,noatime /dev/disk/mydisk /mnt/mydisk",
		},
		{
			src:    "/dev/sdc",
			tgt:    "/mnt",
			opts:   []string{"rw", "", "noatime"},
			result: "-o rw,noatime /dev/sdc /mnt",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(st *testing.T) {
			st.Parallel()
			opts := makeMountArgs(tt.src, tt.tgt, tt.fst, tt.opts)
			optsStr := strings.Join(opts, " ")
			if optsStr != tt.result {
				t.Errorf("Formatting of mount args incorrect, got: %s want: %s",
					optsStr, tt.result)
			}
		})
	}
}
