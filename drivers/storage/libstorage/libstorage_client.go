package libstorage

import (
	"encoding/json"
	"io"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	apiclient "github.com/emccode/libstorage/api/client"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/cli/executors"
)

type client struct {
	apiclient.Client
	ctx                types.Context
	config             gofig.Config
	svcInfo            types.ServicesMap
	lsxInfo            types.ExecutorsMap
	lsxBinPath         string
	enableIIDHeader    bool
	enableLclDevHeader bool
	instanceIDStore    types.Store
	localDevicesStore  types.Store
}

func (c *client) EnableInstanceIDHeaders(enabled bool) {
	c.enableIIDHeader = enabled
}

func (c *client) EnableLocalDevicesHeaders(enabled bool) {
	c.enableLclDevHeader = enabled
}

func (c *client) ServerName() string {
	return c.Client.ServerName
}

func (c *client) InstanceID(
	ctx types.Context,
	service string) (*types.InstanceID, error) {

	ctx = c.withContext(ctx)

	si, err := c.getServiceInfo(service)
	if err != nil {
		return nil, err
	}

	out, err := c.runExecutor(ctx, si.Driver.Name, executors.InstanceID)
	if err != nil {
		return nil, err
	}

	iid := &types.InstanceID{}
	if err := json.Unmarshal(out, iid); err != nil {
		return nil, err
	}

	return iid, nil
}

func (c *client) LocalDevices(
	ctx types.Context,
	service string) (map[string]string, error) {

	ctx = c.withContext(ctx)

	si, err := c.getServiceInfo(service)
	if err != nil {
		return nil, err
	}

	out, err := c.runExecutor(ctx, si.Driver.Name, executors.LocalDevices)
	if err != nil {
		return nil, err
	}

	ldm := map[string]string{}
	if err := json.Unmarshal(out, &ldm); err != nil {
		return nil, err
	}

	return ldm, nil
}

func (c *client) NextDevice(
	ctx types.Context,
	service string) (string, error) {

	ctx = c.withContext(ctx)

	si, err := c.getServiceInfo(service)
	if err != nil {
		return "", err
	}

	out, err := c.runExecutor(ctx, si.Driver.Name, executors.NextDevice)
	if err != nil {
		return "", err
	}

	return gotil.Trim(string(out)), nil
}

func (c *client) NextDeviceInfo(ctx types.Context) *types.NextDeviceInfo {
	// TODO libstorage StorageDriver .NextDeviceInfo
	return nil
}

func (c *client) Type(ctx types.Context) types.StorageType {
	// TODO libstorage StorageDriver .Type
	return ""
}

func (c *client) Root(
	ctx types.Context) ([]string, error) {
	return c.Client.Root(c.withContext(ctx))
}

func (c *client) Services(
	ctx types.Context) (map[string]*types.ServiceInfo, error) {
	return c.Client.Services(c.withContext(ctx))
}

func (c *client) ServiceInspect(
	ctx types.Context, name string) (*types.ServiceInfo, error) {
	return c.Client.ServiceInspect(c.withContext(ctx), name)
}

func (c *client) Volumes(
	ctx types.Context,
	attachments bool) (types.ServiceVolumeMap, error) {
	return c.Client.Volumes(c.withContext(ctx), attachments)
}

func (c *client) VolumesByService(
	ctx types.Context,
	service string,
	attachments bool) (types.VolumeMap, error) {
	return c.Client.VolumesByService(c.withContext(ctx), service, attachments)
}

func (c *client) VolumeInspect(
	ctx types.Context,
	service, volumeID string,
	attachments bool) (*types.Volume, error) {
	return c.Client.VolumeInspect(
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

	vol, err := c.Client.VolumeCreate(ctx, service, request)
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

	vol, err := c.Client.VolumeCreateFromSnapshot(
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

	vol, err := c.Client.VolumeCopy(ctx, service, volumeID, request)
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

	err := c.Client.VolumeRemove(ctx, service, volumeID)
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
	return c.Client.VolumeAttach(c.withContext(ctx), service, volumeID, request)
}

func (c *client) VolumeDetach(
	ctx types.Context,
	service string,
	volumeID string,
	request *types.VolumeDetachRequest) (*types.Volume, error) {
	return c.Client.VolumeDetach(c.withContext(ctx), service, volumeID, request)
}

func (c *client) VolumeDetachAll(
	ctx types.Context,
	request *types.VolumeDetachRequest) (types.ServiceVolumeMap, error) {
	return c.Client.VolumeDetachAll(c.withContext(ctx), request)
}

func (c *client) VolumeDetachAllForService(
	ctx types.Context,
	service string,
	request *types.VolumeDetachRequest) (types.VolumeMap, error) {
	return c.Client.VolumeDetachAllForService(
		c.withContext(ctx), service, request)
}

func (c *client) VolumeSnapshot(
	ctx types.Context,
	service string,
	volumeID string,
	request *types.VolumeSnapshotRequest) (*types.Snapshot, error) {
	return c.Client.VolumeSnapshot(c.withContext(ctx), service, volumeID, request)
}

func (c *client) Snapshots(
	ctx types.Context) (types.ServiceSnapshotMap, error) {
	return c.Client.Snapshots(c.withContext(ctx))
}

func (c *client) SnapshotsByService(
	ctx types.Context, service string) (types.SnapshotMap, error) {
	return c.Client.SnapshotsByService(c.withContext(ctx), service)
}

func (c *client) SnapshotInspect(
	ctx types.Context,
	service, snapshotID string) (*types.Snapshot, error) {
	return c.Client.SnapshotInspect(c.withContext(ctx), service, snapshotID)
}

func (c *client) SnapshotRemove(
	ctx types.Context,
	service, snapshotID string) error {
	return c.Client.SnapshotRemove(c.withContext(ctx), service, snapshotID)
}

func (c *client) SnapshotCopy(
	ctx types.Context,
	service, snapshotID string,
	request *types.SnapshotCopyRequest) (*types.Snapshot, error) {
	return c.Client.SnapshotCopy(c.withContext(ctx), service, snapshotID, request)
}

func (c *client) Executors(
	ctx types.Context) (map[string]*types.ExecutorInfo, error) {
	return c.Client.Executors(c.withContext(ctx))
}

func (c *client) ExecutorHead(
	ctx types.Context,
	name string) (*types.ExecutorInfo, error) {
	return c.Client.ExecutorHead(c.withContext(ctx), name)
}

func (c *client) ExecutorGet(
	ctx types.Context, name string) (io.ReadCloser, error) {
	return c.Client.ExecutorGet(c.withContext(ctx), name)
}
