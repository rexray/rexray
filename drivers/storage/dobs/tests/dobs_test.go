package dobs

import (
	"os"
	"strconv"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/server"
	apitests "github.com/codedellemc/libstorage/api/tests"
	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/api/utils"
	do "github.com/codedellemc/libstorage/drivers/storage/dobs"
	doUtils "github.com/codedellemc/libstorage/drivers/storage/dobs/utils"
	"github.com/stretchr/testify/assert"
)

var (
	configYAML = []byte(`
dobs:
  token: 12345
  region: sfo2`)
)

func skipTests() bool {
	travis, _ := strconv.ParseBool(os.Getenv("TRAVIS"))
	noTest, _ := strconv.ParseBool(os.Getenv("TEST_SKIP_DO"))
	return travis || noTest
}

var volumeName string
var volumeName2 string

func init() {
	uuid, _ := types.NewUUID()
	uuids := strings.Split(uuid.String(), "-")
	volumeName = uuids[0]
	if _, err := strconv.Atoi(string(volumeName[0])); err == nil {
		// TODO randomly select a-z here
		volumeName = strings.Join([]string{"a", volumeName[1:]}, "")
	}

	uuid, _ = types.NewUUID()
	uuids = strings.Split(uuid.String(), "-")
	volumeName2 = uuids[0]
	if _, err := strconv.Atoi(string(volumeName2[0])); err == nil {
		// TODO randomly select a-z here
		volumeName2 = strings.Join([]string{"a", volumeName2[1:]}, "")
	}
}

func TestMain(m *testing.M) {
	server.CloseOnAbort()
	ec := m.Run()
	os.Exit(ec)
}

func TestConfig(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	tfDO := func(config gofig.Config, client types.Client, t *testing.T) {
		assert.NotEqual(t, config.GetString(do.ConfigRegion), "")
		assert.NotEqual(t, config.GetString(do.ConfigToken), "")
	}

	apitests.Run(t, do.Name, configYAML, tfDO)
}

func TestInstanceID(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	sd, err := registry.NewStorageDriver(do.Name)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := sd.Init(ctx, gofigCore.New()); err != nil {
		t.Fatal(err)
	}

	iid, err := doUtils.InstanceID(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}

	ctx = ctx.WithValue(context.InstanceIDKey, iid)
	i, err := sd.InstanceInspect(ctx, utils.NewStore())
	if err != nil {
		t.Fatal(err)
	}

	iid = i.InstanceID
	apitests.Run(
		t, do.Name, nil,
		(&apitests.InstanceIDTest{
			Driver:   do.Name,
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

		_, ok := reply[do.Name]
		assert.True(t, ok)
	}

	apitests.Run(t, do.Name, configYAML, tf)
}

func TestVolumeAttach(t *testing.T) {
	if skipTests() {
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
	apitests.Run(t, do.Name, configYAML, tf)
}

func TestVolumeCreateRemove(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		vol := volumeCreate(t, client, volumeName)
		volumeRemove(t, client, vol.ID)
	}

	apitests.Run(t, do.Name, configYAML, tf)
}

func TestVolumes(t *testing.T) {
	if skipTests() {
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

	apitests.Run(t, do.Name, configYAML, tf)
}

// Test implementation functions

func volumeCreate(t *testing.T, client types.Client,
	volumeName string) *types.Volume {
	log.WithField("volumeName", volumeName).Info("creating volume")

	size := int64(10)
	opts := map[string]interface{}{
		"priority": 2,
		"owner":    "libstorage@example.com",
	}

	volumeCreateRequest := &types.VolumeCreateRequest{
		Name: volumeName,
		Size: &size,
		Opts: opts,
	}

	// Send request and retrieve created libStorage types.Volume
	reply, err := client.API().VolumeCreate(
		nil, do.Name, volumeCreateRequest)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
		t.Error("failed volumeCreate")
	}
	apitests.LogAsJSON(reply, t)

	// Check if name and size are same
	assert.Equal(t, volumeName, reply.Name)
	assert.Equal(t, size, reply.Size)
	return reply
}

func volumeByName(
	t *testing.T, client types.Client, volumeName string) *types.Volume {
	log.WithField("volumeName", volumeName).Info(
		"get volume by digitalocean.Name")
	vols, err := client.API().Volumes(nil, 0)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	assert.Contains(t, vols, do.Name)
	for _, vol := range vols[do.Name] {
		if vol.Name == volumeName {
			return vol
		}
	}
	// No matching volumes found
	t.FailNow()
	t.Error("failed volumeByName")
	return nil
}

func volumeByID(
	t *testing.T, client types.Client, volumeID string) *types.Volume {
	log.WithField("volumeID", volumeID).Info(
		"get volume by digitalocean.Name using ID")
	// Retrieve all volumes
	vols, err := client.API().Volumes(nil, 0)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	// Filter volumes to those under the digitalocean service,
	// and find a volume matching inputted volume ID
	assert.Contains(t, vols, do.Name)
	for _, vol := range vols[do.Name] {
		if vol.ID == volumeID {
			return vol
		}
	}
	// No matching volumes found
	t.FailNow()
	t.Error("failed volumeByID")
	return nil
}

func volumeRemove(t *testing.T, client types.Client, volumeID string) {
	log.WithField("volumeID", volumeID).Info("removing volume")
	err := client.API().VolumeRemove(nil, do.Name, volumeID, false)

	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeRemove")
		t.FailNow()
	}
}

func volumeAttach(
	t *testing.T, client types.Client, volumeID string) *types.Volume {
	log.WithField("volumeID", volumeID).Info("attaching volume")

	reply, token, err := client.API().VolumeAttach(
		nil, do.Name, volumeID, &types.VolumeAttachRequest{})

	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeAttach")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.NotEqual(t, token, "")

	return reply
}

func volumeInspectAttached(
	t *testing.T, client types.Client, volumeID string) *types.Volume {
	log.WithField("volumeID", volumeID).Info("inspecting volume")

	reply, err := client.API().VolumeInspect(
		nil, do.Name, volumeID,
		types.VolAttReqTrue)
	assert.NoError(t, err)

	if err != nil {
		t.Error("failed volumeInspectAttached")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 1)
	return reply
}

func volumeInspectDetached(
	t *testing.T, client types.Client, volumeID string) *types.Volume {
	log.WithField("volumeID", volumeID).Info("inspecting volume")

	reply, err := client.API().VolumeInspect(
		nil, do.Name, volumeID,
		types.VolAttReq)

	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeInspectDetached")
		t.FailNow()
	}

	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 0)
	return reply
}

func volumeInspectDetachedFail(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(nil, do.Name, volumeID, 0)
	assert.NoError(t, err)

	if err != nil {
		t.Error("failed volumeInspectDetachedFail")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 0)
	return reply
}

func volumeDetach(
	t *testing.T, client types.Client, volumeID string) *types.Volume {
	log.WithField("volumeID", volumeID).Info("detaching volume")

	reply, err := client.API().VolumeDetach(
		nil, do.Name, volumeID, &types.VolumeDetachRequest{})

	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeDetach")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 0)
	return reply
}
