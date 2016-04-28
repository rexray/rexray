package client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/akutz/gotil"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/api/utils"
	"github.com/emccode/libstorage/cli/executors"
)

func (c *lsc) getTXCTX() context.Context {
	txIDUUID, _ := utils.NewUUID()
	txID := txIDUUID.String()
	ctx := c.ctx.WithTransactionID(txID)
	ctx = ctx.WithContextID(context.ContextKeyTransactionID, txID)
	txCR := time.Now().UTC()
	ctx = ctx.WithTransactionCreated(txCR)
	ctx = ctx.WithContextID(
		context.ContextKeyTransactionCreated,
		fmt.Sprintf("%d", txCR.Unix()))
	return ctx
}

func (c *lsc) InstanceID(service string) (*types.InstanceID, error) {
	return c.instanceID(c.getTXCTX(), service)
}

func (c *lsc) instanceID(
	ctx context.Context, service string) (*types.InstanceID, error) {

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

func (c *lsc) LocalDevices(service string) (map[string]string, error) {
	return c.localDevices(c.getTXCTX(), service)
}

func (c *lsc) localDevices(
	ctx context.Context, service string) (map[string]string, error) {

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

func (c *lsc) NextDevice(service string) (string, error) {

	ctx := c.getTXCTX()

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

func (c *lsc) Services() (apihttp.ServicesMap, error) {
	return c.Client.Services(c.ctx)
}

func (c *lsc) ServiceInspect(name string) (*types.ServiceInfo, error) {
	return c.Client.ServiceInspect(c.ctx, name)
}

func (c *lsc) Volumes(attachments bool) (apihttp.ServiceVolumeMap, error) {
	return c.Client.Volumes(c.ctx, attachments)
}

func (c *lsc) VolumesByService(
	service string, attachments bool) (apihttp.VolumeMap, error) {
	return c.Client.VolumesByService(c.ctx, service, attachments)
}

func (c *lsc) VolumeInspect(
	service, volumeID string, attachments bool) (*types.Volume, error) {
	return c.Client.VolumeInspect(c.ctx, service, volumeID, attachments)
}

func (c *lsc) VolumeCreate(
	service string,
	request *apihttp.VolumeCreateRequest) (*types.Volume, error) {

	ctx := c.getTXCTX()
	lsd, _ := registry.NewLocalStorageDriver(service)
	if lsd != nil {
		if err := lsd.Init(c.config); err != nil {
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

func (c *lsc) VolumeCopy(
	service, volumeID string,
	request *apihttp.VolumeCopyRequest) (*types.Volume, error) {

	ctx := c.getTXCTX()
	lsd, _ := registry.NewLocalStorageDriver(service)
	if lsd != nil {
		if err := lsd.Init(c.config); err != nil {
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

func (c *lsc) VolumeCreateFromSnapshot(
	service, snapshotID string,
	request *apihttp.VolumeCreateRequest) (*types.Volume, error) {

	ctx := c.getTXCTX()
	lsd, _ := registry.NewLocalStorageDriver(service)
	if lsd != nil {
		if err := lsd.Init(c.config); err != nil {
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

func (c *lsc) VolumeRemove(service, volumeID string) error {

	ctx := c.getTXCTX()
	lsd, _ := registry.NewLocalStorageDriver(service)
	if lsd != nil {
		if err := lsd.Init(c.config); err != nil {
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

func (c *lsc) VolumeAttach(
	service string,
	volumeID string,
	request *apihttp.VolumeAttachRequest) (*types.Volume, error) {
	return c.Client.VolumeAttach(c.ctx, service, volumeID, request)
}

func (c *lsc) VolumeDetach(
	service string,
	volumeID string,
	request *apihttp.VolumeDetachRequest) (*types.Volume, error) {
	reply, err := c.Client.VolumeDetach(c.ctx, service, volumeID, request)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (c *lsc) VolumeDetachAllForService(
	service string,
	request *apihttp.VolumeDetachRequest) (apihttp.VolumeMap, error) {
	return c.Client.VolumeDetachAllForService(c.ctx, service, request)
}

func (c *lsc) VolumeDetachAll(
	request *apihttp.VolumeDetachRequest) (apihttp.ServiceVolumeMap, error) {
	return c.Client.VolumeDetachAll(c.ctx, request)
}

func (c *lsc) VolumeSnapshot(
	service, volumeID string,
	request *apihttp.VolumeSnapshotRequest) (*types.Snapshot, error) {
	return c.Client.VolumeSnapshot(c.ctx, service, volumeID, request)
}

func (c *lsc) Snapshots() (apihttp.ServiceSnapshotMap, error) {
	return c.Client.Snapshots(c.ctx)
}

func (c *lsc) SnapshotsByService(
	service string) (apihttp.SnapshotMap, error) {
	return c.Client.SnapshotsByService(c.ctx, service)
}

func (c *lsc) SnapshotInspect(
	service, snapshotID string) (*types.Snapshot, error) {
	return c.Client.SnapshotInspect(c.ctx, service, snapshotID)
}

func (c *lsc) SnapshotRemove(service, snapshotID string) error {
	return c.Client.SnapshotRemove(c.ctx, service, snapshotID)
}

func (c *lsc) SnapshotCopy(
	service, snapshotID string,
	request *apihttp.SnapshotCopyRequest) (*types.Snapshot, error) {
	return c.Client.SnapshotCopy(c.ctx, service, snapshotID, request)
}
