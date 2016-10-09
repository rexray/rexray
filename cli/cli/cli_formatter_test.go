package cli

import (
	"os"
	"testing"

	apitypes "github.com/emccode/libstorage/api/types"
)

func TestFormatVolume(t *testing.T) {
	fmtOutput(os.Stdout, "tmpl", &apitypes.Volume{
		ID:     "vol-1234",
		Name:   "abuilds",
		Size:   10240,
		Status: "attached",
	})
}

func TestFormatVolumes(t *testing.T) {
	fmtOutput(os.Stdout, "tmpl", []*apitypes.Volume{
		&apitypes.Volume{
			ID:     "vol-5",
			Name:   "bbuilds",
			Size:   10240,
			Status: "attached",
		},
		&apitypes.Volume{
			ID:   "vol-1234",
			Name: "abuilds",
			Size: 10240,
		},
	})
}
