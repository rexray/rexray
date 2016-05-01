package scaleio

import (
	"github.com/akutz/gofig"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/server"
	"github.com/emccode/libstorage/api/server/executors"
	apitests "github.com/emccode/libstorage/api/tests"
	"github.com/emccode/libstorage/api/types"

	// load the  driver
	// "github.com/emccode/libstorage/drivers/storage/libstorage"
	_ "github.com/emccode/libstorage/drivers/storage/scaleio"
	scaleiox "github.com/emccode/libstorage/drivers/storage/scaleio/executor"
)

const name = "scaleio"

var (
	lsxbin string

	lsxLinuxInfo, _  = executors.ExecutorInfoInspect("lsx-linux", false)
	lsxDarwinInfo, _ = executors.ExecutorInfoInspect("lsx-darwin", false)
	//lsxWindowsInfo, _ = executors.ExecutorInfoInspect("lsx-windows.exe", false)

	configYAML = []byte(`
scaleio:
  endpoint: https://192.168.50.12/api
  insecure: true
  userName: admin
  password: Scaleio123
  systemName: cluster1
  protectionDomainName: pdomain
  storagePoolName: pool1
`)
)

func init() {
}

func TestMain(m *testing.M) {
	server.CloseOnAbort()
	ec := m.Run()
	os.Exit(ec)
}

func TestInstanceID(t *testing.T) {
	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
		t.SkipNow()
	}

	iid, err := scaleiox.GetInstanceID()
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	assert.NotEqual(t, iid, "")

	apitests.RunGroup(
		t, name, configYAML,
		(&apitests.InstanceIDTest{
			Driver:   name,
			Expected: iid,
		}).Test)
}

func TestServices(t *testing.T) {
	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
		t.SkipNow()
	}

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().Services(nil)
		assert.NoError(t, err)
		assert.Equal(t, len(reply), 1)

		_, ok := reply[name]
		assert.True(t, ok)
	}
	apitests.Run(t, name, configYAML, tf)
}

func volumeCreate(t *testing.T, client types.Client, volumeName string) *types.Volume {
	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
		t.SkipNow()
	}

	size := int64(8)

	opts := map[string]interface{}{
		"priority": 2,
		"owner":    "root@example.com",
	}

	volumeCreateRequest := &types.VolumeCreateRequest{
		Name: volumeName,
		Size: &size,
		Opts: opts,
	}

	reply, err := client.API().VolumeCreate(
		context.Background().WithServiceName(scaleiox.Name),
		name, volumeCreateRequest)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)

	assert.Equal(t, volumeName, reply.Name)
	assert.Equal(t, size, reply.Size)
	return reply
}

func TestVolumeCreate(t *testing.T) {
	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
		t.SkipNow()
	}

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		_ = volumeCreate(t, client, "Volume-001")
	}
	apitests.Run(t, name, configYAML, tf)
}

func TestVolumes(t *testing.T) {
	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
		t.SkipNow()
	}

	volumeName := "Volume-010"

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		vols, err := client.API().Volumes(nil, false)
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}
		assert.Contains(t, vols, name)
		for _, vol := range vols[name] {
			if vol.Name == volumeName {
				err := client.API().VolumeRemove(nil, name, vol.ID)
				assert.NoError(t, err)
				if err != nil {
					t.FailNow()
				}
			}
		}
	}
	apitests.Run(t, name, configYAML, tf)

	var vol *types.Volume
	tf = func(config gofig.Config, client types.Client, t *testing.T) {
		vol = volumeCreate(t, client, volumeName)
		if vol == nil {
			t.FailNow()
		}
	}
	apitests.Run(t, name, configYAML, tf)

	if vol == nil {
		t.FailNow()
	}

	tf = func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().VolumeInspect(
			nil, name, vol.ID, false)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)
		assert.Equal(t, volumeName, reply.Name)
	}
	apitests.Run(t, name, configYAML, tf)
}

func TestVolumeAttach(t *testing.T) {
	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
		t.SkipNow()
	}

	volumeName := "Volume-007"

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		vols, err := client.API().Volumes(
			nil, true)
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}
		assert.Contains(t, vols, name)

		for _, vol := range vols[name] {
			if vol.Name == volumeName {
				if len(vol.Attachments) > 0 {
					_, err := client.API().VolumeDetach(
						nil, name, vol.ID, &types.VolumeDetachRequest{})
					assert.NoError(t, err)
				}
				err = client.API().VolumeRemove(
					nil, name, vol.ID)
				assert.NoError(t, err)
				if err != nil {
					t.FailNow()
				}
				break
			}
		}

		vol := volumeCreate(t, client, volumeName)
		if vol == nil {
			t.FailNow()
		}

		reply, token, err := client.API().VolumeAttach(
			nil, name, vol.ID, &types.VolumeAttachRequest{})

		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}
		assert.NotEqual(t, token, "")
		apitests.LogAsJSON(reply, t)
		reply, err = client.API().VolumeInspect(nil, name, vol.ID, true)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)
		assert.Len(t, reply.Attachments, 1)

		reply, err = client.API().VolumeDetach(nil,
			name, vol.ID, &types.VolumeDetachRequest{})

		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}
		apitests.LogAsJSON(reply, t)
		reply, err = client.API().VolumeInspect(nil, name, vol.ID, true)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)
		assert.Len(t, reply.Attachments, 0)

	}
	apitests.Run(t, name, configYAML, tf)

}
