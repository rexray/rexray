package mock

import (
	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

type mockStorDriver struct {
	name string
}

type badMockStorDriver struct {
	mockStorDriver
}

func newStorDriver() core.Driver {
	var d core.StorageDriver = &mockStorDriver{MockStorDriverName}
	return d
}

func newBadStorDriver() core.Driver {
	var d core.StorageDriver = &badMockStorDriver{
		mockStorDriver{BadMockStorDriverName}}
	return d
}

func (m *mockStorDriver) Init(r *core.RexRay) error {
	return nil
}

func (m *badMockStorDriver) Init(r *core.RexRay) error {
	return errors.New("init error")
}

func (m *mockStorDriver) Name() string {
	return m.name
}

func (m *mockStorDriver) GetVolumeMapping() ([]*core.BlockDevice, error) {
	return []*core.BlockDevice{&core.BlockDevice{
		DeviceName:   "test",
		ProviderName: m.name,
		InstanceID:   "test",
		Region:       "test",
	}}, nil
}

func (m *mockStorDriver) GetInstance() (*core.Instance, error) {
	return &core.Instance{
		Name:         "test",
		InstanceID:   "test",
		ProviderName: m.name,
		Region:       "test"}, nil
}

func (m *mockStorDriver) GetVolume(
	volumeID, volumeName string) ([]*core.Volume, error) {
	return []*core.Volume{&core.Volume{
		Name:             "test",
		VolumeID:         "test",
		AvailabilityZone: "test",
	}}, nil
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
