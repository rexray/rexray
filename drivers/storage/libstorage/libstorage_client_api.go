package libstorage

import (
	"io"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
)

func (c *client) Instances(
	ctx types.Context) (map[string]*types.Instance, error) {

	imap, err := c.APIClient.Instances(c.withContext(ctx))
	if err != nil {
		return nil, err
	}
	for service, i := range imap {
		i.InstanceID.Formatted = true
		c.instanceIDCache.Set(service, i)
	}
	return imap, nil
}

func (c *client) InstanceInspect(
	ctx types.Context, service string) (*types.Instance, error) {

	i, err := c.APIClient.InstanceInspect(c.withContext(ctx), service)
	if err != nil {
		return nil, err
	}
	i.InstanceID.Formatted = true
	c.instanceIDCache.Set(service, i)
	return i, nil
}

func (c *client) Root(
	ctx types.Context) ([]string, error) {
	return c.APIClient.Root(c.withContext(ctx))
}

func (c *client) Services(
	ctx types.Context) (map[string]*types.ServiceInfo, error) {
	svcInfo, err := c.APIClient.Services(c.withContext(ctx))
	if err != nil {
		return nil, err
	}
	for k, v := range svcInfo {
		c.svcAndLSXCache.Set(k, v)
	}
	return svcInfo, err
}

func (c *client) ServiceInspect(
	ctx types.Context, name string) (*types.ServiceInfo, error) {
	return c.APIClient.ServiceInspect(c.withContext(ctx), name)
}

func (c *client) Volumes(
	ctx types.Context,
	attachments bool) (types.ServiceVolumeMap, error) {
	return c.APIClient.Volumes(c.withContext(ctx), attachments)
}

func (c *client) VolumesByService(
	ctx types.Context,
	service string,
	attachments bool) (types.VolumeMap, error) {
	return c.APIClient.VolumesByService(c.withContext(ctx), service, attachments)
}

func (c *client) VolumeInspect(
	ctx types.Context,
	service, volumeID string,
	attachments bool) (*types.Volume, error) {
	return c.APIClient.VolumeInspect(
		c.withContext(ctx), service, volumeID, attachments)
}

func (c *client) VolumeCreate(
	ctx types.Context,
	service string,
	request *types.VolumeCreateRequest) (*types.Volume, error) {

	ctx = c.withContext(ctx)

	lsd, _ := registry.NewClientDriver(service)
	if lsd != nil {
		if err := lsd.Init(ctx, c.config); err != nil {
			return nil, err
		}

		if err := lsd.VolumeCreateBefore(
			&ctx, service, request); err != nil {
			return nil, err
		}
	}

	vol, err := c.APIClient.VolumeCreate(ctx, service, request)
	if err != nil {
		return nil, err
	}

	if lsd != nil {
		lsd.VolumeCreateAfter(ctx, vol)
	}

	return vol, nil
}

func (c *client) VolumeCreateFromSnapshot(
	ctx types.Context,
	service, snapshotID string,
	request *types.VolumeCreateRequest) (*types.Volume, error) {

	ctx = c.withContext(ctx)

	lsd, _ := registry.NewClientDriver(service)
	if lsd != nil {
		if err := lsd.Init(ctx, c.config); err != nil {
			return nil, err
		}

		if err := lsd.VolumeCreateFromSnapshotBefore(
			&ctx, service, snapshotID, request); err != nil {
			return nil, err
		}
	}

	vol, err := c.APIClient.VolumeCreateFromSnapshot(
		ctx, service, snapshotID, request)
	if err != nil {
		return nil, err
	}

	if lsd != nil {
		lsd.VolumeCreateFromSnapshotAfter(ctx, vol)
	}

	return vol, nil
}

func (c *client) VolumeCopy(
	ctx types.Context,
	service, volumeID string,
	request *types.VolumeCopyRequest) (*types.Volume, error) {

	ctx = c.withContext(ctx)

	lsd, _ := registry.NewClientDriver(service)
	if lsd != nil {
		if err := lsd.Init(ctx, c.config); err != nil {
			return nil, err
		}

		if err := lsd.VolumeCopyBefore(
			&ctx, service, volumeID, request); err != nil {
			return nil, err
		}
	}

	vol, err := c.APIClient.VolumeCopy(ctx, service, volumeID, request)
	if err != nil {
		return nil, err
	}

	if lsd != nil {
		lsd.VolumeCopyAfter(ctx, vol)
	}

	return vol, nil
}

func (c *client) VolumeRemove(
	ctx types.Context,
	service, volumeID string) error {

	ctx = c.withContext(ctx)

	lsd, _ := registry.NewClientDriver(service)
	if lsd != nil {
		if err := lsd.Init(ctx, c.config); err != nil {
			return err
		}

		if err := lsd.VolumeRemoveBefore(
			&ctx, service, volumeID); err != nil {
			return err
		}
	}

	err := c.APIClient.VolumeRemove(ctx, service, volumeID)
	if err != nil {
		return err
	}

	if lsd != nil {
		lsd.VolumeRemoveAfter(ctx, service, volumeID)
	}

	return nil
}

func (c *client) VolumeAttach(
	ctx types.Context,
	service string,
	volumeID string,
	request *types.VolumeAttachRequest) (*types.Volume, error) {
	return c.APIClient.VolumeAttach(c.withContext(ctx), service, volumeID, request)
}

func (c *client) VolumeDetach(
	ctx types.Context,
	service string,
	volumeID string,
	request *types.VolumeDetachRequest) (*types.Volume, error) {
	return c.APIClient.VolumeDetach(c.withContext(ctx), service, volumeID, request)
}

func (c *client) VolumeDetachAll(
	ctx types.Context,
	request *types.VolumeDetachRequest) (types.ServiceVolumeMap, error) {
	return c.APIClient.VolumeDetachAll(c.withContext(ctx), request)
}

func (c *client) VolumeDetachAllForService(
	ctx types.Context,
	service string,
	request *types.VolumeDetachRequest) (types.VolumeMap, error) {
	return c.APIClient.VolumeDetachAllForService(
		c.withContext(ctx), service, request)
}

func (c *client) VolumeSnapshot(
	ctx types.Context,
	service string,
	volumeID string,
	request *types.VolumeSnapshotRequest) (*types.Snapshot, error) {
	return c.APIClient.VolumeSnapshot(c.withContext(ctx), service, volumeID, request)
}

func (c *client) Snapshots(
	ctx types.Context) (types.ServiceSnapshotMap, error) {
	return c.APIClient.Snapshots(c.withContext(ctx))
}

func (c *client) SnapshotsByService(
	ctx types.Context, service string) (types.SnapshotMap, error) {
	return c.APIClient.SnapshotsByService(c.withContext(ctx), service)
}

func (c *client) SnapshotInspect(
	ctx types.Context,
	service, snapshotID string) (*types.Snapshot, error) {
	return c.APIClient.SnapshotInspect(c.withContext(ctx), service, snapshotID)
}

func (c *client) SnapshotRemove(
	ctx types.Context,
	service, snapshotID string) error {
	return c.APIClient.SnapshotRemove(c.withContext(ctx), service, snapshotID)
}

func (c *client) SnapshotCopy(
	ctx types.Context,
	service, snapshotID string,
	request *types.SnapshotCopyRequest) (*types.Snapshot, error) {
	return c.APIClient.SnapshotCopy(c.withContext(ctx), service, snapshotID, request)
}

func (c *client) Executors(
	ctx types.Context) (map[string]*types.ExecutorInfo, error) {

	lsxInfo, err := c.APIClient.Executors(c.withContext(ctx))
	if err != nil {
		return nil, err
	}
	for k, v := range lsxInfo {
		c.svcAndLSXCache.Set(k, v)
	}
	return lsxInfo, nil
}

func (c *client) ExecutorHead(
	ctx types.Context,
	name string) (*types.ExecutorInfo, error) {
	return c.APIClient.ExecutorHead(c.withContext(ctx), name)
}

func (c *client) ExecutorGet(
	ctx types.Context, name string) (io.ReadCloser, error) {
	return c.APIClient.ExecutorGet(c.withContext(ctx), name)
}
