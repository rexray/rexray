package rackspace

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/stretchr/testify/assert"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/server"
	apitests "github.com/emccode/libstorage/api/tests"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"

	// load the  driver
	rackspace "github.com/emccode/libstorage/drivers/storage/rackspace"
	rackspacex "github.com/emccode/libstorage/drivers/storage/rackspace/executor"
	_ "github.com/emccode/libstorage/drivers/storage/rackspace/storage"
)

var (
	configYAML = []byte(`
rackspace:
  authURL: https://identity.api.rackspacecloud.com/v2.0
  username: ` + os.Getenv("RACKSPACE_USERNAME") + `
  password: ` + os.Getenv("RACKSPACE_PASSWORD") + `
  availabilityZoneName: ` + os.Getenv("RACKSPACE_AVAILABILITY_ZONE") + `
  regionName: ` + os.Getenv("RACKSPACE_REGION") + `
`)
)

var volumeName string
var volumeName2 string

func skipTests() bool {
	travis, _ := strconv.ParseBool(os.Getenv("TRAVIS"))
	noTest, _ := strconv.ParseBool(os.Getenv("TEST_SKIP_RACKSPACE"))
	return travis || noTest
}

func init() {
	uuid, _ := types.NewUUID()
	uuids := strings.Split(uuid.String(), "-")
	volumeName = uuids[0]
	uuid, _ = types.NewUUID()
	uuids = strings.Split(uuid.String(), "-")
	volumeName2 = uuids[0]
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

	sd, err := registry.NewStorageDriver(rackspace.Name)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	configR := gofig.New()
	if err := configR.ReadConfig(bytes.NewReader(configYAML)); err != nil {
		panic(err)
	}
	if err := sd.Init(ctx, configR); err != nil {
		t.Fatal(err)
	}
	iid, err := rackspacex.InstanceID(configR)
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
		t, rackspace.Name, configYAML,
		(&apitests.InstanceIDTest{
			Driver:   rackspace.Name,
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

		_, ok := reply[rackspace.Name]
		assert.True(t, ok)
	}
	apitests.Run(t, rackspace.Name, configYAML, tf)
}

func volumeByName(
	t *testing.T, client types.Client, volumeName string) *types.Volume {

	log.WithField("volumeName", volumeName).Info("get volume byrackspace.Name")
	vols, err := client.API().Volumes(nil, false)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	assert.Contains(t, vols, rackspace.Name)
	for _, vol := range vols[rackspace.Name] {
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
	apitests.Run(t, rackspace.Name, configYAML, tf)
}

func volumeCreate(
	t *testing.T, client types.Client, volumeName string) *types.Volume {
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

	reply, err := client.API().VolumeCreate(nil, rackspace.Name, volumeCreateRequest)

	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
		t.Error("failed volumeCreate")
	}

	apitests.LogAsJSON(reply, t)
	assert.Equal(t, volumeName, reply.Name)
	assert.Equal(t, int64(75), reply.Size)
	return reply
}

func volumeRemove(t *testing.T, client types.Client, volumeID string) {
	log.WithField("volumeID", volumeID).Info("removing volume")
	err := client.API().VolumeRemove(
		nil, rackspace.Name, volumeID)
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
	apitests.Run(t, rackspace.Name, configYAML, tf)
}

func volumeAttach(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	log.WithField("volumeID", volumeID).Info("attaching volume")
	reply, token, err := client.API().VolumeAttach(
		nil, rackspace.Name, volumeID, &types.VolumeAttachRequest{})

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
	reply, err := client.API().VolumeInspect(nil, rackspace.Name, volumeID, false)
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
	reply, err := client.API().VolumeInspect(nil, rackspace.Name, volumeID, true)
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
	reply, err := client.API().VolumeInspect(nil, rackspace.Name, volumeID, true)
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
	reply, err := client.API().VolumeInspect(nil, rackspace.Name, volumeID, true)
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

func volumeInspectDetachedFail(
	t *testing.T, client types.Client, volumeID string) *types.Volume {

	log.WithField("volumeID", volumeID).Info("inspecting volume")
	reply, err := client.API().VolumeInspect(nil, rackspace.Name, volumeID, false)
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
		nil, rackspace.Name, volumeID, &types.VolumeDetachRequest{})
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
		_ = volumeInspectDetachedFail(t, client, vol.ID)
		_ = volumeDetach(t, client, vol.ID)
		_ = volumeInspectDetached(t, client, vol.ID)
		volumeRemove(t, client, vol.ID)
	}
	apitests.Run(t, rackspace.Name, configYAML, tf)
}

func testSnapshots(t *testing.T) { //currently fails due to timeout lowercase privated for now
	if skipTests() {
		t.SkipNow()
	}
	var vol *types.Volume
	var cvol *types.Volume
	var snap *types.Snapshot
	var csnap *types.Snapshot
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		vol = volumeCreate(t, client, volumeName)
		snap = volumeSnapshot(t, client, vol.ID, "libstorage test snapshot")
		_ = snapshotInspect(t, client, snap.ID)
		_ = snapshotByName(t, client, "libstorage test snapshot")
		csnap = snapshotCopy(t, client, snap.ID, "listorage snapshot copy", "DestinationID not used")
		_ = snapshotInspect(t, client, csnap.ID)
		cvol = volumeCreateFromSnapshot(t, client, snap.ID, volumeName2)
		_ = volumeInspectDetached(t, client, cvol.ID)
		volumeRemove(t, client, cvol.ID)
		snapshotRemove(t, client, snap.ID)
		snapshotRemove(t, client, csnap.ID)
		volumeRemove(t, client, vol.ID)
	}
	apitests.Run(t, rackspace.Name, configYAML, tf)
}

func snapshotInspect(
	t *testing.T, client types.Client, snapshotID string) *types.Snapshot {
	log.WithField("snapshotID", snapshotID).Info("inspecting snapshot")
	reply, err := client.API().SnapshotInspect(nil, rackspace.Name, snapshotID)
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
	log.WithField("snapshotName", snapshotName).Info("get snapshot by rackspace.Name")
	snapshots, err := client.API().Snapshots(nil)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	assert.Contains(t, snapshots, rackspace.Name)
	for _, vol := range snapshots[rackspace.Name] {
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

	reply, err := client.API().VolumeSnapshot(nil, rackspace.Name,
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

	reply, err := client.API().SnapshotCopy(nil, rackspace.Name,
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
		nil, rackspace.Name, snapshotID)
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
		rackspace.Name, snapshotID, volumeCreateRequest)
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
		rackspace.Name, volumeID, volumeCopyRequest)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
		t.Error("failed volumeCopy")
	}
	apitests.LogAsJSON(reply, t)

	assert.Equal(t, volumeName, reply.Name)

	return reply
}
