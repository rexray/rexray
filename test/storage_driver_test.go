package test

import (
	"testing"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

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

func TestStorageDriverName(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if d.Name() != mockStorDriverName {
		t.Fatalf("driver name != %s, == %s", mockStorDriverName, d.Name())
	}
}

func TestStorageDriverManagerName(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if r.Storage.Name() != mockStorDriverName {
		t.Fatalf("driver name != %s, == %s", mockStorDriverName, r.Storage.Name())
	}
}

func TestStorageDriverManagerNameNoDrivers(t *testing.T) {
	r, _ := getRexRayNoDrivers()
	if r.Storage.Name() != "" {
		t.Fatal("name not empty")
	}
}

func (m *mockStorDriver) GetVolumeMapping() ([]*core.BlockDevice, error) {
	return []*core.BlockDevice{&core.BlockDevice{
		DeviceName:   "test",
		ProviderName: mockStorDriverName,
		InstanceID:   "test",
		Region:       "test",
	}}, nil
}

func TestStorageDriverGetVolumeMapping(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.GetVolumeMapping(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetVolumeMapping(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetVolumeMapping(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetVolumeMappingNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetVolumeMapping(); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func (m *mockStorDriver) GetInstance() (*core.Instance, error) {
	return &core.Instance{
		Name:         "test",
		InstanceID:   "test",
		ProviderName: mockStorDriverName,
		Region:       "test"}, nil
}

func TestStorageDriverGetInstance(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.GetInstance(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetInstance(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetInstance(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetInstanceNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetInstance(); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestGetInstances(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetInstances(); err != nil {
		t.Fatal(err)
	}
}

func TestGetInstancesNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetInstances(); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func (m *mockStorDriver) GetVolume(
	volumeID, volumeName string) ([]*core.Volume, error) {
	return []*core.Volume{&core.Volume{
		Name:             "test",
		VolumeID:         "test",
		AvailabilityZone: "test",
	}}, nil
}

func TestStorageDriverGetVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.GetVolume(
		"", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetVolume(
		"", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetVolumeNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetVolume(
		"", ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func (m *mockStorDriver) GetVolumeAttach(
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
	return nil, nil
}

func TestStorageDriverGetVolumeAttach(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.GetVolumeAttach(
		"", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetVolumeAttach(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetVolumeAttach(
		"", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetVolumeAttachNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetVolumeAttach(
		"", ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func (m *mockStorDriver) CreateSnapshot(
	runAsync bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {
	return nil, nil
}

func TestStorageDriverCreateSnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.CreateSnapshot(
		false, "", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerCreateSnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.CreateSnapshot(
		false, "", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerCreateSnapshotNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.CreateSnapshot(
		false, "", "", ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func (m *mockStorDriver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {
	return nil, nil
}

func TestStorageDriverGetSnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.GetSnapshot(
		"", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetSnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetSnapshot(
		"", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetSnapshotNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetSnapshot(
		"", "", ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func (m *mockStorDriver) RemoveSnapshot(snapshotID string) error {
	return nil
}

func TestStorageDriverRemoveSnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if err := d.RemoveSnapshot(""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerRemoveSnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Storage.RemoveSnapshot(""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerRemoveSnapshotNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Storage.RemoveSnapshot(""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func (m *mockStorDriver) CreateVolume(
	runAsync bool,
	volumeName, volumeID, snapshotID, volumeType string,
	IOPS, size int64,
	availabilityZone string) (*core.Volume, error) {
	return nil, nil
}

func TestStorageDriverCreateVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.CreateVolume(
		false, "", "", "", "", 0, 0, ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerCreateVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.CreateVolume(
		false, "", "", "", "", 0, 0, ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerCreateVolumeNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.CreateVolume(
		false, "", "", "", "", 0, 0, ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func (m *mockStorDriver) RemoveVolume(volumeID string) error {
	return nil
}

func TestStorageDriverRemoveVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if err := d.RemoveVolume(""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerRemoveVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Storage.RemoveVolume(""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerRemoveVolumeNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Storage.RemoveVolume(""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func (m *mockStorDriver) GetDeviceNextAvailable() (string, error) {
	return "", nil
}

func TestStorageDriverGetDeviceNextAvailable(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.GetDeviceNextAvailable(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetDeviceNextAvailable(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetDeviceNextAvailable(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetDeviceNextAvailableNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetDeviceNextAvailable(); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func (m *mockStorDriver) AttachVolume(
	runAsync bool, volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
	return nil, nil
}

func TestStorageDriverAttachVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.AttachVolume(
		false, "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerAttachVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.AttachVolume(
		false, "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerAttachVolumeNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.AttachVolume(
		false, "", ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func (m *mockStorDriver) DetachVolume(
	runAsync bool, volumeID string, instanceID string) error {
	return nil
}

func TestStorageDriverDetachVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if err := d.DetachVolume(
		false, "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerDetachVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Storage.DetachVolume(
		false, "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerDetachVolumeNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Storage.DetachVolume(
		false, "", ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func (m *mockStorDriver) CopySnapshot(
	runAsync bool, volumeID, snapshotID, snapshotName,
	destinationSnapshotName, destinationRegion string) (*core.Snapshot, error) {
	return nil, nil
}

func TestStorageDriverCopySnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.CopySnapshot(false, "", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerCopySnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.CopySnapshot(false, "", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerCopySnapshotNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.CopySnapshot(
		false, "", "", "", "", ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}
