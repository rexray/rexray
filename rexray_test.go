package rexray

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/util"
)

const (
	mockOSDriverName   = "mockOSDriver"
	mockVolDriverName  = "mockVolumeDriver"
	mockStorDriverName = "mockStorageDriver"
)

func TestMain(m *testing.M) {
	core.RegisterDriver(mockOSDriverName, newOSDriver)
	core.RegisterDriver(mockVolDriverName, newVolDriver)
	core.RegisterDriver(mockStorDriverName, newStorDriver)
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	os.Setenv("REXRAY_OSDRIVERS", mockOSDriverName)
	os.Setenv("REXRAY_VOLUMEDRIVERS", mockVolDriverName)
	os.Setenv("REXRAY_STORAGEDRIVERS", mockStorDriverName)

	r, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if err := r.InitDrivers(); err != nil {
		t.Fatal(err)
	}

	assertDriverNames(t, r)
}

func TestNewWithEnv(t *testing.T) {
	r, err := NewWithEnv(map[string]string{
		"REXRAY_OSDRIVERS":      mockOSDriverName,
		"REXRAY_VOLUMEDRIVERS":  mockVolDriverName,
		"REXRAY_STORAGEDRIVERS": mockStorDriverName,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := r.InitDrivers(); err != nil {
		t.Fatal(err)
	}

	assertDriverNames(t, r)
}

func TestNewWithConfigFile(t *testing.T) {
	var err error
	var tmp *os.File
	if tmp, err = ioutil.TempFile("", "TestNewWithConfigFile"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(yamlConfig1); err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	r, err := NewWithConfigFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	if err = r.InitDrivers(); err != nil {
		t.Fatal(err)
	}

	assertDriverNames(t, r)
}

func TestNewWithConfigReader(t *testing.T) {
	r, err := NewWithConfigReader(bytes.NewReader(yamlConfig1))

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
		strings.ToLower(mockOSDriverName),
		strings.ToLower(mockVolDriverName),
		strings.ToLower(mockStorDriverName),
		"linux",
		"docker",
		"ec2",
		"openstack",
		"rackspace",
		"scaleio",
		"xtremio",
	}

	var regDriverNames []string
	for dn := range core.DriverNames() {
		regDriverNames = append(regDriverNames, strings.ToLower(dn))
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

func assertDriverNames(t *testing.T, r *core.RexRay) {
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

func newOSDriver() core.Driver {
	var d core.OSDriver = &mockOSDriver{mockOSDriverName}
	return d
}

func (m *mockOSDriver) Init(r *core.RexRay) error {
	return nil
}

func (m *mockOSDriver) Name() string {
	return m.name
}

func (m *mockOSDriver) GetMounts(string, string) (core.MountInfoArray, error) {
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

func newVolDriver() core.Driver {
	var d core.VolumeDriver = &mockVolDriver{mockVolDriverName}
	return d
}

func (m *mockVolDriver) Init(r *core.RexRay) error {
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

func (m *mockVolDriver) Create(volumeName string, opts core.VolumeOpts) error {
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

func newStorDriver() core.Driver {
	var d core.StorageDriver = &mockStorDriver{mockStorDriverName}
	return d
}

func (m *mockStorDriver) Init(r *core.RexRay) error {
	return nil
}

func (m *mockStorDriver) Name() string {
	return m.name
}

func (m *mockStorDriver) GetVolumeMapping() ([]*core.BlockDevice, error) {
	return nil, nil
}

func (m *mockStorDriver) GetInstance() (*core.Instance, error) {
	return nil, nil
}

func (m *mockStorDriver) GetVolume(
	volumeID, volumeName string) ([]*core.Volume, error) {
	return nil, nil
}

func (m *mockStorDriver) GetVolumeAttach(
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
	return nil, nil
}

func (m *mockStorDriver) CreateSnapshot(
	runAsync bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {
	return nil, nil
}

func (m *mockStorDriver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {
	return nil, nil
}

func (m *mockStorDriver) RemoveSnapshot(snapshotID string) error {
	return nil
}

func (m *mockStorDriver) CreateVolume(
	runAsync bool,
	volumeName, volumeID, snapshotID, volumeType string,
	IOPS, size int64,
	availabilityZone string) (*core.Volume, error) {
	return nil, nil
}

func (m *mockStorDriver) RemoveVolume(volumeID string) error {
	return nil
}

func (m *mockStorDriver) GetDeviceNextAvailable() (string, error) {
	return "", nil
}

func (m *mockStorDriver) AttachVolume(
	runAsync bool, volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
	return nil, nil
}

func (m *mockStorDriver) DetachVolume(
	runAsync bool, volumeID string, instanceID string) error {
	return nil
}

func (m *mockStorDriver) CopySnapshot(
	runAsync bool, volumeID, snapshotID, snapshotName,
	destinationSnapshotName, destinationRegion string) (*core.Snapshot, error) {
	return nil, nil
}

var yamlConfig1 = []byte(`logLevel: error
osDrivers:
- mockOSDriver
volumeDrivers:
- mockVolumeDriver
storageDrivers:
- mockStorageDriver`)
