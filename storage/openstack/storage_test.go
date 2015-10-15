package openstack

import (
	"flag"
	"os"
	"testing"

	"github.com/emccode/rexray/config"
	"github.com/emccode/rexray/storage"
)

var (
	c        *config.Config
	driver   storage.Driver
	runTests bool
)

func TestMain(m *testing.M) {
	flag.BoolVar(&runTests, "rackspace", false, "")
	flag.Parse()
	beforeTests()
	os.Exit(m.Run())
}

func beforeTests() {
	if !runTests {
		return
	}
	c = config.New()
	var err error
	driver, err = Init(c)
	if err != nil {
		panic(err)
	}
}

func TestGetInstance(t *testing.T) {
	if !runTests {
		return
	}
	instance, err := driver.GetInstance()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", instance)
}

func TestGetVolume(t *testing.T) {
	if !runTests {
		return
	}
	volume, err := driver.GetVolume("ccde08e3-d21b-467a-a7d3-bc92ffe0a14f", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", volume)
}

func TestGetVolumeByName(t *testing.T) {
	if !runTests {
		return
	}
	volumes, err := driver.GetVolume("", "Volume-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, volume := range volumes {
		t.Logf("%+v", volume)
	}
}

func TestGetVolumeAttach(t *testing.T) {
	if !runTests {
		return
	}
	volume, err := driver.GetVolumeAttach("12b64bd3-2c34-4fe1-b389-5cf8df668ef5", "5ad7727c-aa5a-43e4-8ab7-a499295032d7")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", volume)
}

func TestGetSnapshotFromVolumeID(t *testing.T) {
	if !runTests {
		return
	}
	snapshots, err := driver.GetSnapshot("738ea6b9-8c49-416c-97b7-a5264a799eb6", "", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", snapshots)
}

func TestGetSnapshotBySnapshotName(t *testing.T) {
	if !runTests {
		return
	}
	snapshots, err := driver.GetSnapshot("", "", "Volume-1-1")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", snapshots)
}

func TestGetSnapshotFromSnapshotID(t *testing.T) {
	if !runTests {
		return
	}
	snapshots, err := driver.GetSnapshot("", "83743ccc-200f-45bb-8144-e802ceb4b555", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", snapshots)
}

func TestCreateSnapshot(t *testing.T) {
	if !runTests {
		return
	}
	snapshot, err := driver.CreateSnapshot(false, "testing", "87ef25ed-9c5f-4030-ada7-eeaf4cba0814", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", snapshot)
}

func TestRemoveSnapshot(t *testing.T) {
	if !runTests {
		return
	}
	err := driver.RemoveSnapshot("ea14a2f0-16b2-47e9-b7ba-01d812f65205")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateVolume(t *testing.T) {
	if !runTests {
		return
	}
	volume, err := driver.CreateVolume(false, "testing", "", "", "", 0, 75, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", volume)
}

func TestRemoveVolume(t *testing.T) {
	if !runTests {
		return
	}
	err := driver.RemoveVolume("743e9de5-8de4-4f09-8249-0238849a3a29")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetDeviceNextAvailable(t *testing.T) {
	if !runTests {
		return
	}
	deviceName, err := driver.GetDeviceNextAvailable()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(deviceName)
}

func TestAttachVolume(t *testing.T) {
	if !runTests {
		return
	}
	volumeAttachment, err := driver.AttachVolume(false, "94e02a4a-71dc-4026-b561-1cd0cad37bce", "5ad7727c-aa5a-43e4-8ab7-a499295032d7")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", volumeAttachment)
	for _, i := range volumeAttachment {
		t.Logf("%+v", i)
	}
}

func TestDetachVolume(t *testing.T) {
	if !runTests {
		return
	}
	err := driver.DetachVolume(false, "94e02a4a-71dc-4026-b561-1cd0cad37bce", "5ad7727c-aa5a-43e4-8ab7-a499295032d7")
	if err != nil {
		t.Fatal(err)
	}
}
