package vfs

import (
	"testing"

	"github.com/akutz/gofig"

	apitests "github.com/emccode/libstorage/api/tests"
	"github.com/emccode/libstorage/client"

	// load the driver
	"github.com/emccode/libstorage/drivers/storage/vfs"
)

func TestVolumes(t *testing.T) {

	tf := func(config gofig.Config, client client.Client, t *testing.T) {
		reply, err := client.Volumes()
		if err != nil {
			t.Fatal(err)
		}
		apitests.LogAsJSON(reply, t)
	}

	apitests.Run(t, vfs.Name, nil, tf)
}
