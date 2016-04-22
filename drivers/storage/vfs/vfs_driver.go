package vfs

import (
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
	"github.com/emccode/libstorage/drivers/storage/vfs/executor"
)

const (
	// Name is the name of the driver.
	Name = executor.Name
)

type driver struct {
	executor.Executor
}

func init() {
	registry.RegisterStorageDriver(executor.Name, newDriver)
}

func newDriver() drivers.StorageDriver {
	return &driver{executor.Executor{}}
}

func (d *driver) Type() types.StorageType {
	return types.Object
}

func (d *driver) NextDeviceInfo() *types.NextDeviceInfo {
	return nil
}

func (d *driver) InstanceInspect(
	ctx context.Context,
	opts types.Store) (*types.Instance, error) {
	return &types.Instance{InstanceID: ctx.InstanceID()}, nil
}

func (d *driver) Volumes(
	ctx context.Context,
	opts *drivers.VolumesOpts) ([]*types.Volume, error) {
	return nil, nil
}

func (d *driver) VolumeInspect(
	ctx context.Context,
	volumeID string,
	opts *drivers.VolumeInspectOpts) (*types.Volume, error) {
	return nil, nil
}

func (d *driver) VolumeCreate(
	ctx context.Context,
	name string,
	opts *drivers.VolumeCreateOpts) (*types.Volume, error) {
	return nil, nil
}

func (d *driver) VolumeCreateFromSnapshot(
	ctx context.Context,
	snapshotID, volumeName string,
	opts *drivers.VolumeCreateOpts) (*types.Volume, error) {
	return nil, nil
}

func (d *driver) VolumeCopy(
	ctx context.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {
	return nil, nil
}

func (d *driver) VolumeSnapshot(
	ctx context.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) VolumeRemove(
	ctx context.Context,
	volumeID string,
	opts types.Store) error {
	return nil
}

func (d *driver) VolumeAttach(
	ctx context.Context,
	volumeID string,
	opts *drivers.VolumeAttachByIDOpts) (*types.Volume, error) {
	return nil, nil
}

func (d *driver) VolumeDetach(
	ctx context.Context,
	volumeID string,
	opts types.Store) error {
	return nil
}

func (d *driver) Snapshots(
	ctx context.Context,
	opts types.Store) ([]*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotInspect(
	ctx context.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotCopy(
	ctx context.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotRemove(
	ctx context.Context,
	snapshotID string,
	opts types.Store) error {
	return nil
}
