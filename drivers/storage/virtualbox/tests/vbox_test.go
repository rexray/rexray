package tests

import (
	"os"
	"path"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/context"
	apitests "github.com/emccode/libstorage/api/tests"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	"github.com/emccode/libstorage/drivers/storage/virtualbox"
	"github.com/emccode/libstorage/drivers/storage/virtualbox/executor"
)

var (
	lsxbin       string
	vboxsvr      *mockVboxServer
	diskByIDPath = "./dev/disk/by-id"
	diskID       = "scsi-SATA_VBOX_HARDDISK_VB32a50c6d-2c3f5f6a"
	diskPath     = path.Join(diskByIDPath, diskID)
)

func createMockDisk() {
	err := os.MkdirAll(diskByIDPath, 0744)
	if err != nil {
		logrus.Error(err)
	}
	file, err := os.Create(diskPath)
	if err != nil {
		logrus.Error(err)
	}
	if err = file.Close(); err != nil {
		logrus.Error(err)
	}
}

func deleteMockDisk() {
	if err := os.Remove(diskPath); err != nil {
		logrus.Error(err)
	}
	if err := os.RemoveAll("./dev"); err != nil {
		logrus.Error(err)
	}
}

func TestMain(m *testing.M) {
	vboxsvr = newMockVBoxServer()
	vboxsvr.start()
	createMockDisk()

	code := m.Run()

	vboxsvr.stop()
	deleteMockDisk()
	os.Exit(code)
}

func TestExecutorInstanceID(t *testing.T) {
	exec := &executor.Executor{}
	conf := gofig.New()
	conf.Set("virtualbox.endpoint", vboxsvr.url())
	if err := exec.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	id, err := exec.InstanceID(context.Background(), utils.NewStore())
	if err != nil {
		t.Fatal(err)
	}
	if id.ID != "9f49850d-f617-4b43-a46d-272c380e7cc6" {
		t.Fatal("Executor not getting expected machine ID")
	}
}

func TestExecutorNextDevice(t *testing.T) {
	exec := &executor.Executor{}
	conf := gofig.New()
	conf.Set("virtualbox.endpoint", vboxsvr.url())
	if err := exec.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	_, err := exec.NextDevice(context.Background(), utils.NewStore())
	if err != types.ErrNotImplemented {
		t.Fatal("Executor.NextDevice() should not be implemented")
	}
}

func TestExecutorLocalDevices(t *testing.T) {
	exec := &executor.Executor{}
	conf := gofig.New()
	conf.Set("virtualbox.endpoint", vboxsvr.url())
	if err := exec.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	conf.Set("virtualbox.diskIDPath", diskByIDPath)
	devices, err := exec.LocalDevices(context.Background(), utils.NewStore())
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 {
		t.Fatal("Expected 1 disk device, but got ", len(devices))
	}
}

func TestDriverInstanceID(t *testing.T) {
	driver := &virtualbox.Driver{}
	conf := gofig.New()
	conf.Set("virtualbox.endpoint", vboxsvr.url())
	if err := driver.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	id, err := driver.InstanceID(context.Background(), utils.NewStore())
	if err != nil {
		t.Fatal(err)
	}
	if id.ID != "9f49850d-f617-4b43-a46d-272c380e7cc6" {
		t.Fatal("Driver not getting expected machine ID")
	}
}

func TestDriverNextDevice(t *testing.T) {
	driver := &virtualbox.Driver{}
	conf := gofig.New()
	conf.Set("virtualbox.endpoint", vboxsvr.url())
	if err := driver.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	_, err := driver.NextDevice(context.Background(), utils.NewStore())
	if err != types.ErrNotImplemented {
		t.Fatal("Executor.NextDevice() should not be implemented")
	}
}

func TestDriverInstanceInspect(t *testing.T) {
	driver := &virtualbox.Driver{}
	conf := gofig.New()
	conf.Set("virtualbox.endpoint", vboxsvr.url())
	if err := driver.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	ins, err := driver.InstanceInspect(context.Background(), utils.NewStore())
	if err != nil {
		t.Fatal(err)
	}
	if ins == nil {
		t.Fatal("Unexpected result: instance is nil")
	}
	if ins.InstanceID.ID != "9f49850d-f617-4b43-a46d-272c380e7cc6" {
		t.Fatal("Unexpected InstanceID returned:", ins.InstanceID)
	}
}

func TestDriverVolumes(t *testing.T) {
	driver := &virtualbox.Driver{}
	conf := gofig.New()
	conf.Set("virtualbox.endpoint", vboxsvr.url())
	conf.Set("virtualbox.diskIDPath", diskByIDPath)
	if err := driver.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	vols, err := driver.Volumes(context.Background(), &types.VolumesOpts{})
	if err != nil {
		t.Fatal("Volumes method failed: ", err)
	}
	if vols == nil {
		t.Fatal("Unexpected result: volume nil")
	}
	if len(vols) != 2 {
		t.Fatal("Expecting 2 volumes for virtualbox, got ", len(vols))
	}
}

func TestDriverVolumeCreate(t *testing.T) {
	driver := &virtualbox.Driver{}
	conf := gofig.New()
	conf.Set("virtualbox.endpoint", vboxsvr.url())
	conf.Set("virtualbox.diskIDPath", diskByIDPath)
	if err := driver.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	size := int64(2000000)
	iops := int64(10)
	vol, err := driver.VolumeCreate(
		context.Background(),
		"default-vol",
		&types.VolumeCreateOpts{Size: &size, IOPS: &iops})
	if err != nil {
		t.Fatal("Volumes method failed: ", err)
	}
	if vol == nil {
		t.Fatal("Unexpected result: volume nil")
	}
}

func TestDriverVolumeRemove(t *testing.T) {
	driver := &virtualbox.Driver{}
	conf := gofig.New()
	conf.Set("virtualbox.endpoint", vboxsvr.url())
	conf.Set("virtualbox.diskIDPath", diskByIDPath)
	if err := driver.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	err := driver.VolumeRemove(
		context.Background(), "32a50c6d-ddcc-4e0a-a3c6-5c126a5f3f2c",
		utils.NewStore())

	if err != nil {
		t.Fatal("Volumes method failed: ", err)
	}
}

func TestDriverVolumeAttach(t *testing.T) {
	driver := &virtualbox.Driver{}
	conf := gofig.New()
	conf.Set("virtualbox.endpoint", vboxsvr.url())
	conf.Set("virtualbox.diskIDPath", diskByIDPath)
	if err := driver.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	_, err := driver.VolumeAttach(
		context.Background(), "32a50c6d-ddcc-4e0a-a3c6-5c126a5f3f2c",
		&types.VolumeAttachOpts{})

	if err == nil {
		t.Fatal("Expected failure, volume alread attached")
	}

}

func TestDriverVolumeAttachWithForce(t *testing.T) {
	driver := &virtualbox.Driver{}
	conf := gofig.New()
	conf.Set("virtualbox.endpoint", vboxsvr.url())
	conf.Set("virtualbox.diskIDPath", diskByIDPath)
	if err := driver.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	opts := &types.VolumeAttachOpts{Force: true}
	_, err := driver.VolumeAttach(
		context.Background(), "32a50c6d-ddcc-4e0a-a3c6-5c126a5f3f2c", opts)

	if err == nil {
		t.Fatal("Expected failure.")
	}

}

func TestDriverVolumeDetach(t *testing.T) {
	driver := &virtualbox.Driver{}
	conf := gofig.New()
	conf.Set("virtualbox.endpoint", vboxsvr.url())
	conf.Set("virtualbox.diskIDPath", diskByIDPath)
	if err := driver.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	_, err := driver.VolumeDetach(
		context.Background(), "32a50c6d-ddcc-4e0a-a3c6-5c126a5f3f2c",
		&types.VolumeDetachOpts{})

	if err == nil {
		t.Fatal("Expected error here")
	}

}

// *****************************************************************************
// server, driver, and executor integration tests
// *****************************************************************************
func TestDriverIntegrationInstanceID(t *testing.T) {
	// skipping until tests.RunGroup works.
	t.Skip("Test skipped")
	// override default server
	os.Setenv("VIRTUALBOX_ENDPOINT", vboxsvr.url())

	driver := &virtualbox.Driver{}
	conf := gofig.New()
	if err := driver.Init(context.Background(), conf); err != nil {
		t.Fatal(err)
	}
	id, err := driver.InstanceID(context.Background(), utils.NewStore())

	if err != nil {
		t.Fatal(err)
	}

	apitests.RunGroup(
		t, driver.Name(), nil,
		(&apitests.InstanceIDTest{
			Driver:   driver.Name(),
			Expected: id,
		}).Test)
}
