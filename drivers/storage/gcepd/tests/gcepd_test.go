package gcepd

import (
	"os"
	"strconv"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/stretchr/testify/assert"

	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/server"
	apitests "github.com/codedellemc/libstorage/api/tests"
	"github.com/codedellemc/libstorage/api/types"

	// load the driver
	"github.com/codedellemc/libstorage/drivers/storage/gcepd"
	gceUtils "github.com/codedellemc/libstorage/drivers/storage/gcepd/utils"
)

// Put contents of sample config.yml here
var (
	configYAML = []byte(`
gcepd:
  keyfile: /tmp/gce_key.json`)
)

var volumeName string
var volumeName2 string

// Check environment vars to see whether or not to run this test
func skipTests() bool {
	travis, _ := strconv.ParseBool(os.Getenv("TRAVIS"))
	noTest, _ := strconv.ParseBool(os.Getenv("TEST_SKIP_GCE"))
	return travis || noTest
}

// Set volume names to first part of UUID before the -
func init() {
	uuid, _ := types.NewUUID()
	uuids := strings.Split(uuid.String(), "-")
	volumeName = `a` + uuids[0]
	uuid, _ = types.NewUUID()
	uuids = strings.Split(uuid.String(), "-")
	volumeName2 = `b` + uuids[0]
}

func TestMain(m *testing.M) {
	server.CloseOnAbort()
	ec := m.Run()
	os.Exit(ec)
}

func TestInstanceID(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	// Get Instance ID metadata
	ctx := context.Background()
	iid, err := gceUtils.InstanceID(ctx)
	assert.NoError(t, err)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEqual(t, iid.ID, "")

	// test resulting InstanceID
	apitests.Run(
		t, gcepd.Name, configYAML,
		(&apitests.InstanceIDTest{
			Driver:   gcepd.Name,
			Expected: iid,
		}).Test)
}

func TestServices(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().Services(nil)
		assert.NoError(t, err)
		assert.Equal(t, len(reply), 1)

		_, ok := reply[gcepd.Name]
		assert.True(t, ok)
	}
	apitests.Run(t, gcepd.Name, configYAML, tf)
}

func volumeCreate(
	t *testing.T,
	client types.Client,
	volumeName string,
	diskType *string) *types.Volume {

	log.WithField("volumeName", volumeName).Info("creating volume")
	size := int64(10)

	volumeCreateRequest := &types.VolumeCreateRequest{
		Name: volumeName,
		Size: &size,
		Opts: nil,
	}

	if diskType != nil {
		volumeCreateRequest.Type = diskType
	}

	reply, err := client.API().VolumeCreate(nil, gcepd.Name,
		volumeCreateRequest)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
		t.Error("failed volumeCreate")
	}
	apitests.LogAsJSON(reply, t)

	assert.Equal(t, volumeName, reply.Name)
	assert.Equal(t, size, reply.Size)
	assert.Equal(t, volumeName, reply.ID)
	return reply
}

func volumeRemove(t *testing.T, client types.Client, volumeID string) {
	log.WithField("volumeID", volumeID).Info("removing volume")
	err := client.API().VolumeRemove(
		nil, gcepd.Name, volumeID, false)
	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeRemove")
		t.FailNow()
	}
}

func TestVolumeCreateRemove(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		vol := volumeCreate(t, client, volumeName, nil)
		volumeRemove(t, client, vol.ID)
	}
	apitests.Run(t, gcepd.Name, configYAML, tf)
}

func TestVolumeCreateDiskTypes(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		vol := volumeCreate(t, client, volumeName, nil)
		// SSD should be the default
		assert.Equal(t, gcepd.DiskTypeSSD, vol.Type)

		standardType := gcepd.DiskTypeStandard
		vol2 := volumeCreate(t, client, volumeName2, &standardType)
		assert.Equal(t, standardType, vol2.Type)
		volumeRemove(t, client, vol.ID)
		volumeRemove(t, client, vol2.ID)
	}
	apitests.Run(t, gcepd.Name, configYAML, tf)

}

func volumeByName(
	t *testing.T,
	client types.Client,
	volumeName string) *types.Volume {

	log.WithField("volumeName", volumeName).Info("get volume name")
	vols, err := client.API().Volumes(nil, 0)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	assert.Contains(t, vols, gcepd.Name)
	for _, vol := range vols[gcepd.Name] {
		if vol.Name == volumeName && vol.ID == volumeName {
			return vol
		}
	}
	t.FailNow()
	t.Error("failed volumeByName")
	return nil
}

func TestVolumes(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		_ = volumeCreate(t, client, volumeName, nil)
		_ = volumeCreate(t, client, volumeName2, nil)

		vol1 := volumeByName(t, client, volumeName)
		vol2 := volumeByName(t, client, volumeName2)

		volumeRemove(t, client, vol1.ID)
		volumeRemove(t, client, vol2.ID)
	}
	apitests.Run(t, gcepd.Name, configYAML, tf)
}

func volumeAttach(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	log.WithField("volumeID", volumeID).Info("attaching volume")
	reply, token, err := client.API().VolumeAttach(
		nil, gcepd.Name, volumeID, &types.VolumeAttachRequest{})

	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeAttach")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.NotEqual(t, token, "")

	return reply
}

func volumeDetach(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	log.WithField("volumeID", volumeID).Info("detaching volume")
	reply, err := client.API().VolumeDetach(
		nil, gcepd.Name, volumeID, &types.VolumeDetachRequest{})
	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeDetach")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 0)
	return reply
}

func volumeInspect(
	t *testing.T,
	client types.Client,
	volumeID string,
	attachFlag types.VolumeAttachmentsTypes) *types.Volume {

	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(
		nil, gcepd.Name, volumeID, attachFlag)
	assert.NoError(t, err)

	if err != nil {
		t.Error("failed volumeInspect")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	return reply
}

func volumeInspectAttached(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	reply := volumeInspect(t, client, volumeID, types.VolAttReqForInstance)
	assert.Len(t, reply.Attachments, 1)
	assert.Equal(t, "", reply.Attachments[0].DeviceName)
	return reply
}

func volumeInspectNoAttachments(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	reply := volumeInspect(t, client, volumeID, types.VolAttFalse)
	assert.Len(t, reply.Attachments, 0)
	return reply
}

func volumeInspectAttachedDevices(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	reply := volumeInspect(t, client, volumeID,
		types.VolAttReqWithDevMapForInstance)
	assert.Len(t, reply.Attachments, 1)
	assert.NotEqual(t, "", reply.Attachments[0].DeviceName)
	return reply
}

func volumeInspectDetached(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	reply := volumeInspect(t, client, volumeID, types.VolAttReq)
	assert.Len(t, reply.Attachments, 0)
	return reply
}

func volumeInspectAvailable(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	reply := volumeInspect(t, client, volumeID,
		types.VolAttReqOnlyUnattachedVols)
	assert.Len(t, reply.Attachments, 0)
	return reply
}

func volumeAttachFail(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	log.WithField("volumeID", volumeID).Info("attaching volume")
	reply, _, err := client.API().VolumeAttach(
		nil, gcepd.Name, volumeID, &types.VolumeAttachRequest{})

	assert.Error(t, err)
	if err == nil {
		t.Error("volumeAttach succeeded when it should have failed")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)

	return reply
}

func TestVolumeAttach(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}
	var vol *types.Volume
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		vol = volumeCreate(t, client, volumeName, nil)
		_ = volumeAttach(t, client, vol.ID)
		_ = volumeInspectAttached(t, client, vol.ID)
		_ = volumeInspectAttachedDevices(t, client, vol.ID)
		_ = volumeInspectNoAttachments(t, client, vol.ID)
		_ = volumeAttachFail(t, client, vol.ID)
		_ = volumeDetach(t, client, vol.ID)
		_ = volumeInspectDetached(t, client, vol.ID)
		_ = volumeInspectAvailable(t, client, vol.ID)
		volumeRemove(t, client, vol.ID)
	}
	apitests.Run(t, gcepd.Name, configYAML, tf)
}
