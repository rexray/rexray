package vfs

import (
	"testing"

	"github.com/akutz/gofig"

	apiclient "github.com/emccode/libstorage/api/client"
	apitests "github.com/emccode/libstorage/api/tests"

	// load the driver
	"github.com/emccode/libstorage/drivers/storage/vfs"
)

func TestRoot(t *testing.T) {

	tf := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.Root()
		if err != nil {
			t.Fatal(err)
		}
		apitests.LogAsJSON(reply, t)
	}

	apitests.Run(t, vfs.Name, nil, tf)
}

func TestVolumes(t *testing.T) {

	tf := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.Volumes()
		if err != nil {
			t.Fatal(err)
		}
		apitests.LogAsJSON(reply, t)
	}

	apitests.Run(t, vfs.Name, nil, tf)
}
