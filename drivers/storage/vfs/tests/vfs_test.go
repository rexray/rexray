package vfs

import (
	"testing"

	"github.com/akutz/gofig"
	"github.com/stretchr/testify/assert"

	apitests "github.com/emccode/libstorage/api/tests"
	"github.com/emccode/libstorage/client"

	// load the driver
	"github.com/emccode/libstorage/drivers/storage/vfs"
	vfsx "github.com/emccode/libstorage/drivers/storage/vfs/executor"
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

func TestInstanceID(t *testing.T) {
	iid, err := vfsx.GetInstanceID()
	assert.NoError(t, err)
	apitests.RunGroup(
		t, vfs.Name, nil,
		(&apitests.InstanceIDTest{
			Driver:   vfs.Name,
			Expected: iid,
		}).Test)
}

func TestNextDevice(t *testing.T) {
	apitests.RunGroup(
		t, vfs.Name, nil,
		(&apitests.NextDeviceTest{
			Driver:   vfs.Name,
			Expected: "/dev/xvda",
		}).Test)
}

func TestLocalDevices(t *testing.T) {
	apitests.RunGroup(
		t, vfs.Name, nil,
		(&apitests.LocalDevicesTest{
			Driver: vfs.Name,
			Expected: map[string]string{
				"/dev/xvda": "",
				"/dev/xvdb": "",
				"/dev/xvdc": "",
				"/dev/xvdd": "",
			},
		}).Test)
}
