package mount

import (
	"strings"
	"testing"
)

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
