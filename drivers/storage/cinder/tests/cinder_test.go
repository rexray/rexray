package cinder

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/stretchr/testify/assert"

	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/server"
	apitests "github.com/codedellemc/libstorage/api/tests"
	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/api/utils"

	"github.com/codedellemc/libstorage/drivers/storage/cinder"
	// load and register the driver
	_ "github.com/codedellemc/libstorage/drivers/storage/cinder/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/cinder/storage"
)

// configYAML is an embedded config file that specifies the bare minimum
// configuration for creating an openstack volume.
var configYAML []byte

var volumeName string
var volumeName2 string

func skipTests() bool {
	travis, _ := strconv.ParseBool(os.Getenv("TRAVIS"))
	noTest, _ := strconv.ParseBool(os.Getenv("TEST_SKIP_CINDER"))
	return travis || noTest
}

func init() {
	// set the volumeName to a random name
	uuid, _ := types.NewUUID()
	uuids := strings.Split(uuid.String(), "-")
	volumeName = uuids[0]

	// set the second volumeName to a random name
	uuid, _ = types.NewUUID()
	uuids = strings.Split(uuid.String(), "-")
	volumeName2 = uuids[0]

	tenantName := os.Getenv("OS_TENANT_NAME")
	if tenantName == "" {
		tenantName = os.Getenv("OS_PROJECT_NAME")
	}

	configYAML = []byte(`
cinder:
  authURL: ` + os.Getenv("OS_AUTH_URL") + `
  username: ` + os.Getenv("OS_USERNAME") + `
  tenantName: ` + tenantName + `
  domainName: ` + os.Getenv("OS_USER_DOMAIN_NAME") + `
  password: ` + os.Getenv("OS_PASSWORD") + `
  availabilityZoneName: ` + os.Getenv("OS_AVAILABILITY_ZONE") + `
  regionName: ` + os.Getenv("OS_REGION_NAME") + `
`)
}

func TestMain(m *testing.M) {
	server.CloseOnAbort()
	ec := m.Run()
	os.Exit(ec)
}

func TestInstanceID(t *testing.T) { //PASSES lowercase hidden for testing other stuff
	if skipTests() {
		t.SkipNow()
	}

	sd, err := registry.NewStorageDriver(cinder.Name)
	if err != nil {
		t.Fatal(err)
	}
	se, err := registry.NewStorageExecutor(cinder.Name)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	configR := registry.NewConfig()
	if err := configR.ReadConfig(bytes.NewReader(configYAML)); err != nil {
		panic(err)
	}

	if err := sd.Init(ctx, configR); err != nil {
		t.Fatal(err)
	}
	if err := se.Init(ctx, configR); err != nil {
		t.Fatal(err)
	}

	iid, err := se.InstanceID(ctx, utils.NewStore())
	if err != nil {
		t.Fatal(err)
	}

	ctx = ctx.WithValue(context.InstanceIDKey, iid)
	i, err := sd.InstanceInspect(ctx, utils.NewStore())
	if err != nil {
		t.Fatal(err)
	}

	iid = i.InstanceID

	apitests.Run(
		t, cinder.Name, configYAML,
		(&apitests.InstanceIDTest{
			Driver:   cinder.Name,
			Expected: iid,
		}).Test)
}

func TestServices(t *testing.T) { //PASSES lowercase hidden for testing other stuff
	if skipTests() {
		t.SkipNow()
	}

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().Services(nil)
		assert.NoError(t, err)
		assert.Equal(t, len(reply), 1)

		_, ok := reply[cinder.Name]
		assert.True(t, ok)
	}
	apitests.Run(t, cinder.Name, configYAML, tf)
}

func volumeByName(
	t *testing.T, client types.Client, volumeName string) *types.Volume {

	log.WithField("volumeName", volumeName).Info("get volume")
	vols, err := client.API().Volumes(nil, 0)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	assert.Contains(t, vols, cinder.Name)
	for _, vol := range vols[cinder.Name] {
		if vol.Name == volumeName {
			return vol
		}
	}
	t.FailNow()
	t.Error("failed volumeByName")
	return nil
}

func TestVolumeCreateRemove(t *testing.T) { //PASSES lowercase hidden for testing other stuff
	if skipTests() {
		t.SkipNow()
	}
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		vol := volumeCreate(t, client, volumeName)
		volumeRemove(t, client, vol.ID)
	}
	apitests.Run(t, cinder.Name, configYAML, tf)
}

func volumeCreate(
	t *testing.T, client types.Client, volumeName string) *types.Volume {
	log.WithField("volumeName", volumeName).Info("creating volume")
	size := int64(1)

	opts := map[string]interface{}{
		"priority": 2,
		"owner":    "root@example.com",
	}

	volumeCreateRequest := &types.VolumeCreateRequest{
		Name: volumeName,
		Size: &size,
		Opts: opts,
	}

	reply, err := client.API().VolumeCreate(nil, cinder.Name, volumeCreateRequest)

	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
		t.Error("failed volumeCreate")
	}

	apitests.LogAsJSON(reply, t)
	assert.Equal(t, volumeName, reply.Name)
	assert.Equal(t, int64(1), reply.Size)
	return reply
}

func volumeRemove(t *testing.T, client types.Client, volumeID string) {
	log.WithField("volumeID", volumeID).Info("removing volume")
	err := client.API().VolumeRemove(
		nil, cinder.Name, volumeID, false)
	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeRemove")
		t.FailNow()
	}
}

func TestVolumes(t *testing.T) { //PASSES lowercase hidden for testing other stuff
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
	apitests.Run(t, cinder.Name, configYAML, tf)
}

func volumeAttach(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	log.WithField("volumeID", volumeID).Info("attaching volume")
	reply, token, err := client.API().VolumeAttach(
		nil, cinder.Name, volumeID, &types.VolumeAttachRequest{})

	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeAttach")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.NotEqual(t, token, "")

	return reply
}

func volumeInspect(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(nil, cinder.Name, volumeID, 0)
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

	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(
		nil, cinder.Name, volumeID, types.VolAttReqTrue)
	assert.NoError(t, err)

	if err != nil {
		t.Error("failed volumeInspectAttached")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 1)
	return reply
}

func volumeInspectAttachedFail(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(
		nil, cinder.Name, volumeID, types.VolAttReqTrue)
	assert.NoError(t, err)

	if err != nil {
		t.Error("failed volumeInspectAttachedFail")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 0)
	return reply
}

func volumeInspectDetached(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(
		nil, cinder.Name, volumeID, types.VolAttReqOnlyUnattachedVols)

	if err != nil {
		t.Error("failed volumeInspectDetached")
		t.FailNow()
	}
	assert.NoError(t, err)

	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 0)

	return reply
}

func volumeInspectDetachedFail(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(nil, cinder.Name, volumeID, 0)
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

func volumeDetach(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	log.WithField("volumeID", volumeID).Info("detaching volume")
	reply, err := client.API().VolumeDetach(
		nil, cinder.Name, volumeID, &types.VolumeDetachRequest{})
	assert.NoError(t, err)
	if err != nil {
		t.Error("failed volumeDetach")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	assert.Len(t, reply.Attachments, 0)
	return reply
}

func TestVolumeAttach(t *testing.T) { //PASSES lowercase hidden to test other stuff
	if skipTests() {
		t.SkipNow()
	}
	var vol *types.Volume
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		vol = volumeCreate(t, client, volumeName)
		_ = volumeAttach(t, client, vol.ID)
		_ = volumeInspectAttached(t, client, vol.ID)
		//_ = volumeInspectDetachedFail(t, client, vol.ID)
		_ = volumeDetach(t, client, vol.ID)
		_ = volumeInspectDetached(t, client, vol.ID)
		volumeRemove(t, client, vol.ID)
	}
	apitests.Run(t, cinder.Name, configYAML, tf)
}

func TestSnapshots(t *testing.T) {
	if skipTests() {
		t.SkipNow()
	}
	var vol *types.Volume
	var cvol *types.Volume
	var snap *types.Snapshot
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		vol = volumeCreate(t, client, volumeName)
		snap = volumeSnapshot(t, client, vol.ID, "libstorage test snapshot")
		_ = snapshotInspect(t, client, snap.ID)
		_ = snapshotByName(t, client, "libstorage test snapshot")
		cvol = volumeCreateFromSnapshot(t, client, snap.ID, volumeName2)
		_ = volumeInspectDetached(t, client, cvol.ID)
		volumeRemove(t, client, cvol.ID)
		snapshotRemove(t, client, snap.ID)
		volumeRemove(t, client, vol.ID)
	}
	apitests.Run(t, cinder.Name, configYAML, tf)
}

func snapshotInspect(
	t *testing.T, client types.Client, snapshotID string) *types.Snapshot {
	log.WithField("snapshotID", snapshotID).Info("inspecting snapshot")
	reply, err := client.API().SnapshotInspect(nil, cinder.Name, snapshotID)
	assert.NoError(t, err)

	if err != nil {
		t.Error("failed snapshotInspect")
		t.FailNow()
	}
	apitests.LogAsJSON(reply, t)
	return reply
}

func snapshotByName(
	t *testing.T, client types.Client, snapshotName string) *types.Snapshot {
	log.WithField("snapshotName", snapshotName).Info("get snapshotByName")
	snapshots, err := client.API().Snapshots(nil)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	assert.Contains(t, snapshots, cinder.Name)
	for _, vol := range snapshots[cinder.Name] {
		if vol.Name == snapshotName {
			return vol
		}
	}
	t.FailNow()
	t.Error("failed snapshotByName")
	return nil
}

func volumeSnapshot(
	t *testing.T, client types.Client,
	volumeID, snapshotName string) *types.Snapshot {
	log.WithField("snapshotName", snapshotName).Info("creating snapshot")

	/*
		opts := map[string]interface{}{
			"priority": 2,
			"owner":    "root@example.com",
		}*/

	volumeSnapshotRequest := &types.VolumeSnapshotRequest{
		SnapshotName: snapshotName,
		//	Opts: opts,
	}

	reply, err := client.API().VolumeSnapshot(nil, cinder.Name,
		volumeID, volumeSnapshotRequest)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
		t.Error("failed snapshotCreate")
	}
	apitests.LogAsJSON(reply, t)

	assert.Equal(t, snapshotName, reply.Name)
	assert.Equal(t, volumeID, reply.VolumeID)
	return reply
}

func snapshotCopy(
	t *testing.T, client types.Client,
	snapshotID, snapshotName, destinationID string) *types.Snapshot {
	log.WithField("snapshotName", snapshotName).Info("copying snapshot")

	snapshotCopyRequest := &types.SnapshotCopyRequest{
		SnapshotName: snapshotName,
	}

	reply, err := client.API().SnapshotCopy(nil, cinder.Name,
		snapshotID, snapshotCopyRequest)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
		t.Error("failed snapshotCopy")
	}
	apitests.LogAsJSON(reply, t)

	assert.Equal(t, snapshotName, reply.Name)
	return reply
}

func snapshotRemove(t *testing.T, client types.Client, snapshotID string) {
	log.WithField("snapshotID", snapshotID).Info("removing snapshot")
	err := client.API().SnapshotRemove(
		nil, cinder.Name, snapshotID)
	assert.NoError(t, err)

	if err != nil {
		t.Error("failed snapshotRemove")
		t.FailNow()
	}
}

func volumeCreateFromSnapshot(
	t *testing.T, client types.Client,
	snapshotID, volumeName string) *types.Volume {
	fields := map[string]interface{}{
		"snapshotID": snapshotID,
		"volumeName": volumeName,
	}
	log.WithFields(fields).Info("creating volume from snapshot")
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

	reply, err := client.API().VolumeCreateFromSnapshot(nil,
		cinder.Name, snapshotID, volumeCreateRequest)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
		t.Error("failed volumeCreateFromSnapshot")
	}
	apitests.LogAsJSON(reply, t)

	assert.Equal(t, volumeName, reply.Name)
	assert.Equal(t, size, reply.Size)
	assert.Equal(t, opts["priority"], 2)
	assert.Equal(t, opts["owner"], "root@example.com")

	return reply
}

func volumeCopy(
	t *testing.T, client types.Client,
	volumeID, volumeName string) *types.Volume {
	fields := map[string]interface{}{
		"volumeID":   volumeID,
		"volumeName": volumeName,
	}
	log.WithFields(fields).Info("copying volume")

	volumeCopyRequest := &types.VolumeCopyRequest{
		VolumeName: volumeName,
	}

	reply, err := client.API().VolumeCopy(nil,
		cinder.Name, volumeID, volumeCopyRequest)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
		t.Error("failed volumeCopy")
	}
	apitests.LogAsJSON(reply, t)

	assert.Equal(t, volumeName, reply.Name)

	return reply
}
