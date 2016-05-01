package libstorage

import (
	"github.com/emccode/libstorage/api/types"
	lstypes "github.com/emccode/libstorage/drivers/storage/libstorage/types"
)

func (d *driver) Name() string {
	return Name
}

func (d *driver) API() lstypes.Client {
	return &d.client
}

func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {

	si, err := d.getServiceInfo(ctx.ServiceName())
	if err != nil {
		return nil, err
	}

	return si.Driver.NextDevice, nil
}

func (d *driver) Type(ctx types.Context) (types.StorageType, error) {

	si, err := d.getServiceInfo(ctx.ServiceName())
	if err != nil {
		return "", err
	}

	return si.Driver.Type, nil
}

func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {
	return d.client.InstanceInspect(ctx, ctx.ServiceName())
}

func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	ctx = d.withContext(ctx)

	objMap, err := d.client.VolumesByService(
		ctx, ctx.ServiceName(), opts.Attachments)

	if err != nil {
		return nil, err
	}

	objs := []*types.Volume{}
	for _, o := range objMap {
		objs = append(objs, o)
	}

	return objs, nil
}

func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	ctx = d.withContext(ctx)

	return d.client.VolumeInspect(
		ctx, ctx.ServiceName(), volumeID, opts.Attachments)
}

func (d *driver) VolumeCreate(
	ctx types.Context,
	name string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	ctx = d.withContext(ctx)

	req := &types.VolumeCreateRequest{
		Name:             name,
		AvailabilityZone: opts.AvailabilityZone,
		IOPS:             opts.IOPS,
		Size:             opts.Size,
		Type:             opts.Type,
		Opts:             opts.Opts.Map(),
	}

	return d.client.VolumeCreate(ctx, ctx.ServiceName(), req)
}

func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	ctx = d.withContext(ctx)

	req := &types.VolumeCreateRequest{
		Name:             volumeName,
		AvailabilityZone: opts.AvailabilityZone,
		IOPS:             opts.IOPS,
		Size:             opts.Size,
		Type:             opts.Type,
		Opts:             opts.Opts.Map(),
	}

	return d.client.VolumeCreateFromSnapshot(
		ctx, ctx.ServiceName(), snapshotID, req)
}

func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {

	ctx = d.withContext(ctx)

	req := &types.VolumeCopyRequest{
		VolumeName: volumeName,
		Opts:       opts.Map(),
	}

	return d.client.VolumeCopy(
		ctx, ctx.ServiceName(), volumeID, req)
}

func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {

	ctx = d.withContext(ctx)

	req := &types.VolumeSnapshotRequest{
		SnapshotName: snapshotName,
		Opts:         opts.Map(),
	}

	return d.client.VolumeSnapshot(
		ctx, ctx.ServiceName(), volumeID, req)
}

func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {

	ctx = d.withContext(ctx)

	return d.client.VolumeRemove(
		ctx, ctx.ServiceName(), volumeID)
}

func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, error) {

	ctx = d.withContext(ctx)

	req := &types.VolumeAttachRequest{
		NextDeviceName: opts.NextDevice,
		Force:          opts.Force,
		Opts:           opts.Opts.Map(),
	}

	return d.client.VolumeAttach(
		ctx, ctx.ServiceName(), volumeID, req)
}

func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	ctx = d.withContext(ctx)

	req := &types.VolumeDetachRequest{
		Force: opts.Force,
		Opts:  opts.Opts.Map(),
	}

	return d.client.VolumeDetach(
		ctx, ctx.ServiceName(), volumeID, req)
}

func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {

	ctx = d.withContext(ctx)

	objMap, err := d.client.SnapshotsByService(
		ctx, ctx.ServiceName())

	if err != nil {
		return nil, err
	}

	objs := []*types.Snapshot{}
	for _, o := range objMap {
		objs = append(objs, o)
	}

	return objs, nil
}

func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {

	ctx = d.withContext(ctx)

	return d.client.SnapshotInspect(
		ctx, ctx.ServiceName(), snapshotID)
}

func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {

	ctx = d.withContext(ctx)

	req := &types.SnapshotCopyRequest{
		SnapshotName:  snapshotName,
		DestinationID: destinationID,
		Opts:          opts.Map(),
	}

	return d.client.SnapshotCopy(
		ctx, ctx.ServiceName(), snapshotID, req)
}

func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {

	ctx = d.withContext(ctx)

	return d.client.SnapshotRemove(
		ctx, ctx.ServiceName(), snapshotID)
}
