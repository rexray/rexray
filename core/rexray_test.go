package core

import (
	"os"
	"testing"

	"github.com/emccode/rexray/core/config"
	"github.com/emccode/rexray/util"
)

const (
	mockOSDriverName   = "mockOSDriver"
	mockVolDriverName  = "mockVolumeDriver"
	mockStorDriverName = "mockStorageDriver"
)

func TestMain(m *testing.M) {
	RegisterDriver(mockOSDriverName, newOSDriver)
	RegisterDriver(mockVolDriverName, newVolDriver)
	RegisterDriver(mockStorDriverName, newStorDriver)
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	c := config.New()
	c.OSDrivers = []string{mockOSDriverName}
	c.VolumeDrivers = []string{mockVolDriverName}
	c.StorageDrivers = []string{mockStorDriverName}
	r, err := New(c)

	if err != nil {
		t.Fatal(err)
	}

	if err := r.InitDrivers(); err != nil {
		t.Fatal(err)
	}

	assertDriverNames(t, r)
}

func TestDriverNames(t *testing.T) {
	allDriverNames := []string{
		mockOSDriverName,
		mockVolDriverName,
		mockStorDriverName,
	}

	var regDriverNames []string
	for dn := range DriverNames() {
		regDriverNames = append(regDriverNames, dn)
	}

	for _, n := range allDriverNames {
		if !util.StringInSlice(n, regDriverNames) {
			t.Fail()
		}
	}

	for _, n := range regDriverNames {
		if !util.StringInSlice(n, allDriverNames) {
			t.Fail()
		}
	}
}

func assertDriverNames(t *testing.T, r *RexRay) {
	od := <-r.OS.Drivers()
	if od.Name() != mockOSDriverName {
		t.Fatalf("expected %s but was %s", mockOSDriverName, od.Name())
	}

	vd := <-r.Volume.Drivers()
	if vd.Name() != mockVolDriverName {
		t.Fatalf("expected %s but was %s", mockVolDriverName, vd.Name())
	}

	sd := <-r.Storage.Drivers()
	if sd.Name() != mockStorDriverName {
		t.Fatalf("expected %s but was %s", mockStorDriverName, sd.Name())
	}
}

type mockOSDriver struct {
	name string
}

func newOSDriver() Driver {
	var d OSDriver = &mockOSDriver{mockOSDriverName}
	return d
}

func (m *mockOSDriver) Init(r *RexRay) error {
	return nil
}

func (m *mockOSDriver) Name() string {
	return m.name
}

func (m *mockOSDriver) GetMounts(string, string) (MountInfoArray, error) {
	return nil, nil
}

func (m *mockOSDriver) Mounted(string) (bool, error) {
	return false, nil
}

func (m *mockOSDriver) Unmount(string) error {
	return nil
}

func (m *mockOSDriver) Mount(string, string, string, string) error {
	return nil
}

func (m *mockOSDriver) Format(string, string, bool) error {
	return nil
}

type mockVolDriver struct {
	name string
}

func newVolDriver() Driver {
	var d VolumeDriver = &mockVolDriver{mockVolDriverName}
	return d
}

func (m *mockVolDriver) Init(r *RexRay) error {
	return nil
}

func (m *mockVolDriver) Name() string {
	return m.name
}

func (m *mockVolDriver) Mount(
	volumeName, volumeID string,
	overwriteFs bool, newFsType string) (string, error) {
	return "", nil
}

func (m *mockVolDriver) Unmount(volumeName, volumeID string) error {
	return nil
}

func (m *mockVolDriver) Path(volumeName, volumeID string) (string, error) {
	return "", nil
}

func (m *mockVolDriver) Create(volumeName string, opts VolumeOpts) error {
	return nil
}

func (m *mockVolDriver) Remove(volumeName string) error {
	return nil
}

func (m *mockVolDriver) Attach(volumeName, instanceID string) (string, error) {
	return "", nil
}

func (m *mockVolDriver) Detach(volumeName, instanceID string) error {
	return nil
}

func (m *mockVolDriver) NetworkName(
	volumeName, instanceID string) (string, error) {
	return "", nil
}

type mockStorDriver struct {
	name string
}

func newStorDriver() Driver {
	var d StorageDriver = &mockStorDriver{mockStorDriverName}
	return d
}

func (m *mockStorDriver) Init(r *RexRay) error {
	return nil
}

func (m *mockStorDriver) Name() string {
	return m.name
}

func (m *mockStorDriver) GetVolumeMapping() ([]*BlockDevice, error) {
	return nil, nil
}

func (m *mockStorDriver) GetInstance() (*Instance, error) {
	return nil, nil
}

func (m *mockStorDriver) GetVolume(
	volumeID, volumeName string) ([]*Volume, error) {
	return nil, nil
}

func (m *mockStorDriver) GetVolumeAttach(
	volumeID, instanceID string) ([]*VolumeAttachment, error) {
	return nil, nil
}

func (m *mockStorDriver) CreateSnapshot(
	runAsync bool,
	snapshotName, volumeID, description string) ([]*Snapshot, error) {
	return nil, nil
}

func (m *mockStorDriver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*Snapshot, error) {
	return nil, nil
}

func (m *mockStorDriver) RemoveSnapshot(snapshotID string) error {
	return nil
}

func (m *mockStorDriver) CreateVolume(
	runAsync bool,
	volumeName, volumeID, snapshotID, volumeType string,
	IOPS, size int64,
	availabilityZone string) (*Volume, error) {
	return nil, nil
}

func (m *mockStorDriver) RemoveVolume(volumeID string) error {
	return nil
}

func (m *mockStorDriver) GetDeviceNextAvailable() (string, error) {
	return "", nil
}

func (m *mockStorDriver) AttachVolume(
	runAsync bool, volumeID, instanceID string) ([]*VolumeAttachment, error) {
	return nil, nil
}

func (m *mockStorDriver) DetachVolume(
	runAsync bool, volumeID string, instanceID string) error {
	return nil
}

func (m *mockStorDriver) CopySnapshot(
	runAsync bool, volumeID, snapshotID, snapshotName,
	destinationSnapshotName, destinationRegion string) (*Snapshot, error) {
	return nil, nil
}
