package test

import (
	"testing"

	"github.com/emccode/rexray/core"
)

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
