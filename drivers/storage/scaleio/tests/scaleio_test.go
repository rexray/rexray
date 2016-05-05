package scaleio

import (
	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/emccode/libstorage/api/server"
	"github.com/emccode/libstorage/api/server/executors"
	apitests "github.com/emccode/libstorage/api/tests"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"strings"
	"testing"

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

var volumeName string
var volumeName2 string

func init() {
	uuid, _ := utils.NewUUID()
	uuids := strings.Split(uuid.String(), "-")
	volumeName = uuids[0]
	uuid, _ = utils.NewUUID()
	uuids = strings.Split(uuid.String(), "-")
	volumeName2 = uuids[0]
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
		t.Error("failed TestInstanceID")
		t.FailNow()
	}
	assert.NotEqual(t, iid, "")

	apitests.Run(
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
	log.WithField("volumeName", volumeName).Info("creating volume")
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

	reply, err := client.API().VolumeCreate(nil, name, volumeCreateRequest)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
		t.Error("failed volumeCreate")
	}
	apitests.LogAsJSON(reply, t)

	assert.Equal(t, volumeName, reply.Name)
	assert.Equal(t, size, reply.Size)
	return reply
}

func volumeByName(t *testing.T, client types.Client, volumeName string) *types.Volume {
	log.WithField("volumeName", volumeName).Info("get volume by name")
	vols, err := client.API().Volumes(nil, false)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	assert.Contains(t, vols, name)
	for _, vol := range vols[name] {
		if vol.Name == volumeName {
			return vol
		}
	}
	t.FailNow()
	t.Error("failed volumeByName")
	return nil
}

func TestVolumeCreateRemove(t *testing.T) {
	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
		t.SkipNow()
	}

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		vol := volumeCreate(t, client, volumeName)
		volumeRemove(t, client, vol.ID)
	}
	apitests.Run(t, name, configYAML, tf)
}

func volumeRemove(t *testing.T, client types.Client, volumeID string) {
	log.WithField("volumeID", volumeID).Info("removing volume")
	err := client.API().VolumeRemove(
		nil, name, volumeID)
	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeRemove")
		t.FailNow()
	}
}

func TestVolumes(t *testing.T) {
	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
		t.SkipNow()
	}

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		_ = volumeCreate(t, client, volumeName)
		_ = volumeCreate(t, client, volumeName2)

		vol1 := volumeByName(t, client, volumeName)
		vol2 := volumeByName(t, client, volumeName2)

		volumeRemove(t, client, vol1.ID)
		volumeRemove(t, client, vol2.ID)
	}
	apitests.Run(t, name, configYAML, tf)
}

func volumeAttach(t *testing.T, client types.Client, volumeID string) *types.Volume {
	log.WithField("volumeID", volumeID).Info("attaching volume")
	reply, token, err := client.API().VolumeAttach(
		nil, name, volumeID, &types.VolumeAttachRequest{})

	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeAttach")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.NotEqual(t, token, "")

	return reply
}

func volumeInspect(t *testing.T, client types.Client, volumeID string) *types.Volume {
	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(nil, name, volumeID, false)
	assert.NoError(t, err)

	if err != nil {
		t.Error("failed volumeInspect")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	return reply
}

func volumeInspectAttached(t *testing.T, client types.Client, volumeID string) *types.Volume {
	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(nil, name, volumeID, true)
	assert.NoError(t, err)

	if err != nil {
		t.Error("failed volumeInspectAttached")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 1)
	return reply
}

func volumeInspectAttachedFail(t *testing.T, client types.Client, volumeID string) *types.Volume {
	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(nil, name, volumeID, true)
	assert.NoError(t, err)

	if err != nil {
		t.Error("failed volumeInspectAttachedFail")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 0)
	return reply
}

func volumeInspectDetached(t *testing.T, client types.Client, volumeID string) *types.Volume {
	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(nil, name, volumeID, true)
	assert.NoError(t, err)

	if err != nil {
		t.Error("failed volumeInspectDetached")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 0)
	apitests.LogAsJSON(reply, t)
	return reply
}

func volumeInspectDetachedFail(t *testing.T, client types.Client, volumeID string) *types.Volume {
	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(nil, name, volumeID, false)
	assert.NoError(t, err)

	if err != nil {
		t.Error("failed volumeInspectDetachedFail")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 1)
	apitests.LogAsJSON(reply, t)
	return reply
}

func volumeDetach(t *testing.T, client types.Client, volumeID string) *types.Volume {
	log.WithField("volumeID", volumeID).Info("detaching volume")
	reply, err := client.API().VolumeDetach(
		nil, name, volumeID, &types.VolumeDetachRequest{})
	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeDetach")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 0)
	return reply
}

func TestVolumeAttach(t *testing.T) {
	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
		t.SkipNow()
	}
	var vol *types.Volume
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		vol = volumeCreate(t, client, volumeName)
		_ = volumeAttach(t, client, vol.ID)
		_ = volumeInspectAttached(t, client, vol.ID)
		_ = volumeInspectDetachedFail(t, client, vol.ID)
		_ = volumeDetach(t, client, vol.ID)
		_ = volumeInspectDetached(t, client, vol.ID)
		volumeRemove(t, client, vol.ID)
	}
	apitests.Run(t, name, configYAML, tf)
}

// func TestVolumes(t *testing.T) {
// 	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
// 		t.SkipNow()
// 	}
//
// 	volumeName := "Volume-010"
//
// 	TestVolumeRemove(t)
//
// 	var vol *types.Volume
// 	tf := func(config gofig.Config, client types.Client, t *testing.T) {
// 		vol = volumeCreate(t, client, volumeName)
// 		if vol == nil {
// 			t.FailNow()
// 		}
// 	}
// 	apitests.Run(t, name, configYAML, tf)
//
// 	if vol == nil {
// 		t.FailNow()
// 	}
//
// 	tf = func(config gofig.Config, client types.Client, t *testing.T) {
// 		reply, err := client.API().VolumeInspect(
// 			nil, name, vol.ID, false)
// 		assert.NoError(t, err)
// 		apitests.LogAsJSON(reply, t)
// 		assert.Equal(t, volumeName, reply.Name)
// 	}
// 	apitests.Run(t, name, configYAML, tf)
// }

// if len(vol.Attachments) > 0 {
// 	_, err := client.API().VolumeDetach(
// 		nil, name, vol.ID, &types.VolumeDetachRequest{})
// 	assert.NoError(t, err)
// 	if err != nil {
// 		t.FailNow()
// 	}
// }

//
// func TestVolumeRemoveIfPresent(t *testing.T) {
// 	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
// 		t.SkipNow()
// 	}
//
// 	tf := func(config gofig.Config, client types.Client, t *testing.T) {
// 		vols, err := client.API().Volumes(nil, false)
// 		assert.NoError(t, err)
// 		if err != nil {
// 			t.FailNow()
// 		}
// 		assert.Contains(t, vols, name)
// 		for _, vol := range vols[name] {
// 			if vol.Name == volumeName {
// 				if len(vol.Attachments) > 0 {
// 					_, err := client.API().VolumeDetach(
// 						nil, name, vol.ID, &types.VolumeDetachRequest{})
// 					assert.NoError(t, err)
// 					if err != nil {
// 						t.FailNow()
// 					}
// 				}
// 				err = client.API().VolumeRemove(
// 					nil, name, vol.ID)
// 				assert.NoError(t, err)
// 				if err != nil {
// 					t.FailNow()
// 				}
// 				break
// 			}
// 		}
// 	}
// 	apitests.Run(t, name, configYAML, tf)
// }
//
// func TestVolumes(t *testing.T) {
// 	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
// 		t.SkipNow()
// 	}
//
// 	volumeName := "Volume-010"
//
// 	TestVolumeRemove(t)
//
// 	var vol *types.Volume
// 	tf := func(config gofig.Config, client types.Client, t *testing.T) {
// 		vol = volumeCreate(t, client, volumeName)
// 		if vol == nil {
// 			t.FailNow()
// 		}
// 	}
// 	apitests.Run(t, name, configYAML, tf)
//
// 	if vol == nil {
// 		t.FailNow()
// 	}
//
// 	tf = func(config gofig.Config, client types.Client, t *testing.T) {
// 		reply, err := client.API().VolumeInspect(
// 			nil, name, vol.ID, false)
// 		assert.NoError(t, err)
// 		apitests.LogAsJSON(reply, t)
// 		assert.Equal(t, volumeName, reply.Name)
// 	}
// 	apitests.Run(t, name, configYAML, tf)
// }
//
// func TestVolumeAttach(t *testing.T) {
// 	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); travis {
// 		t.SkipNow()
// 	}
//
// 	volumeName := "Volume-007"
//
// 	tf := func(config gofig.Config, client types.Client, t *testing.T) {
// 		vols, err := client.API().Volumes(
// 			nil, false)
// 		assert.NoError(t, err)
// 		if err != nil {
// 			t.FailNow()
// 		}
// 		assert.Contains(t, vols, name)
//
// 		for _, vol := range vols[name] {
// 			if vol.Name == volumeName {
// 				if len(vol.Attachments) > 0 {
// 					_, err := client.API().VolumeDetach(
// 						nil, name, vol.ID, &types.VolumeDetachRequest{})
// 					assert.NoError(t, err)
// 				}
// 				err = client.API().VolumeRemove(
// 					nil, name, vol.ID)
// 				assert.NoError(t, err)
// 				if err != nil {
// 					t.FailNow()
// 				}
// 				break
// 			}
// 		}
//
// 		vol := volumeCreate(t, client, volumeName)
// 		if vol == nil {
// 			t.FailNow()
// 		}
//
// 		reply, token, err := client.API().VolumeAttach(
// 			nil, name, vol.ID, &types.VolumeAttachRequest{})
//
// 		assert.NoError(t, err)
// 		if err != nil {
// 			t.FailNow()
// 		}
// 		assert.NotEqual(t, token, "")
// 		apitests.LogAsJSON(reply, t)
// 		reply, err = client.API().VolumeInspect(nil, name, vol.ID, true)
// 		assert.NoError(t, err)
// 		apitests.LogAsJSON(reply, t)
// 		assert.Len(t, reply.Attachments, 1)
//
// 		reply, err = client.API().VolumeDetach(nil,
// 			name, vol.ID, &types.VolumeDetachRequest{})
//
// 		assert.NoError(t, err)
// 		if err != nil {
// 			t.FailNow()
// 		}
// 		apitests.LogAsJSON(reply, t)
// 		reply, err = client.API().VolumeInspect(nil, name, vol.ID, true)
// 		assert.NoError(t, err)
// 		apitests.LogAsJSON(reply, t)
// 		assert.Len(t, reply.Attachments, 0)
//
// 	}
// 	apitests.Run(t, name, configYAML, tf)
//
// }
