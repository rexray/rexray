package client

import (
	"encoding/base64"
	"fmt"
	"io"
	"strconv"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/types"
)

// Root returns a list of root resources.
func (c *client) Root(ctx types.Context) ([]string, error) {
	cctx(&ctx)
	reply := []string{}
	if _, err := c.httpGet(ctx, "/", &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

const (
	ctxInstanceForSvc = 1000 + iota
)

type ctxInstanceForSvcT struct{}

// Instances returns a list of instances.
func (c *client) Instances(
	ctx types.Context) (map[string]*types.Instance, error) {
	sis, err := c.Services(
		ctx.WithValue(ctxInstanceForSvc, &ctxInstanceForSvcT{}))
	if err != nil {
		return nil, err
	}
	instances := map[string]*types.Instance{}
	for service, si := range sis {
		instances[service] = si.Instance
	}
	return instances, nil
}

// InstanceInspect inspects an instance.
func (c *client) InstanceInspect(
	ctx types.Context, service string) (*types.Instance, error) {
	si, err := c.ServiceInspect(
		ctx.WithValue(ctxInstanceForSvc, &ctxInstanceForSvcT{}), service)
	if err != nil {
		return nil, err
	}
	return si.Instance, nil
}

// Services returns a map of the configured Services.
func (c *client) Services(
	ctx types.Context) (map[string]*types.ServiceInfo, error) {

	cctx(&ctx)
	reply := map[string]*types.ServiceInfo{}

	url := "/services"
	if ctx.Value(ctxInstanceForSvc) != nil {
		url = "/services?instance"
	}

	if _, err := c.httpGet(ctx, url, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// ServiceInspect returns information about a service.
func (c *client) ServiceInspect(
	ctx types.Context, name string) (*types.ServiceInfo, error) {

	cctx(&ctx)
	reply := &types.ServiceInfo{}

	url := fmt.Sprintf("/services/%s", name)
	if ctx.Value(ctxInstanceForSvc) != nil {
		url = fmt.Sprintf("/services/%s?instance", name)
	}

	if _, err := c.httpGet(ctx, url, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// Volumes returns a list of all Volumes for all Services.
func (c *client) Volumes(
	ctx types.Context,
	attachments bool) (types.ServiceVolumeMap, error) {
	cctx(&ctx)
	reply := types.ServiceVolumeMap{}
	url := fmt.Sprintf("/volumes?attachments=%v", attachments)
	if _, err := c.httpGet(ctx, url, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// VolumesByService returns a list of all Volumes for a service.
func (c *client) VolumesByService(
	ctx types.Context,
	service string,
	attachments bool) (types.VolumeMap, error) {
	cctx(&ctx)
	reply := types.VolumeMap{}
	url := fmt.Sprintf("/volumes/%s?attachments=%v", service, attachments)
	if _, err := c.httpGet(ctx, url, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// VolumeInspect gets information about a single volume.
func (c *client) VolumeInspect(
	ctx types.Context,
	service, volumeID string,
	attachments bool) (*types.Volume, error) {
	cctx(&ctx)
	reply := types.Volume{}
	url := fmt.Sprintf(
		"/volumes/%s/%s?attachments=%v", service, volumeID, attachments)
	if _, err := c.httpGet(ctx, url, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// VolumeCreate creates a single volume.
func (c *client) VolumeCreate(
	ctx types.Context,
	service string,
	request *types.VolumeCreateRequest) (*types.Volume, error) {
	cctx(&ctx)
	reply := types.Volume{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/volumes/%s", service), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// VolumeCreateFromSnapshot creates a single volume from a snapshot.
func (c *client) VolumeCreateFromSnapshot(
	ctx types.Context,
	service, snapshotID string,
	request *types.VolumeCreateRequest) (*types.Volume, error) {
	cctx(&ctx)
	reply := types.Volume{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/snapshots/%s/%s?create",
			service, snapshotID), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// VolumeCopy copies a single volume.
func (c *client) VolumeCopy(
	ctx types.Context,
	service, volumeID string,
	request *types.VolumeCopyRequest) (*types.Volume, error) {
	cctx(&ctx)
	reply := types.Volume{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/volumes/%s/%s?copy", service, volumeID),
		request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// VolumeRemove removes a single volume.
func (c *client) VolumeRemove(
	ctx types.Context,
	service, volumeID string) error {
	cctx(&ctx)
	if _, err := c.httpDelete(ctx,
		fmt.Sprintf("/volumes/%s/%s", service, volumeID), nil); err != nil {
		return err
	}
	return nil
}

// VolumeAttach attaches a single volume.
func (c *client) VolumeAttach(
	ctx types.Context,
	service string,
	volumeID string,
	request *types.VolumeAttachRequest) (*types.Volume, string, error) {
	cctx(&ctx)
	reply := types.VolumeAttachResponse{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/volumes/%s/%s?attach",
			service, volumeID), request, &reply); err != nil {
		return nil, "", err
	}
	return reply.Volume, reply.AttachToken, nil
}

// VolumeDetach attaches a single volume.
func (c *client) VolumeDetach(
	ctx types.Context,
	service string,
	volumeID string,
	request *types.VolumeDetachRequest) (*types.Volume, error) {
	cctx(&ctx)
	reply := types.Volume{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/volumes/%s/%s?detach",
			service, volumeID), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// VolumeDetachAll attaches all volumes from all types.
func (c *client) VolumeDetachAll(
	ctx types.Context,
	request *types.VolumeDetachRequest) (types.ServiceVolumeMap, error) {
	cctx(&ctx)
	reply := types.ServiceVolumeMap{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/volumes?detach"), request, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// VolumeDetachAllForService detaches all volumes from a service.
func (c *client) VolumeDetachAllForService(
	ctx types.Context,
	service string,
	request *types.VolumeDetachRequest) (types.VolumeMap, error) {
	cctx(&ctx)
	reply := types.VolumeMap{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf(
			"/volumes/%s?detach", service), request, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// VolumeSnapshot creates a single snapshot.
func (c *client) VolumeSnapshot(
	ctx types.Context,
	service string,
	volumeID string,
	request *types.VolumeSnapshotRequest) (*types.Snapshot, error) {
	cctx(&ctx)
	reply := types.Snapshot{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/volumes/%s/%s?snapshot",
			service, volumeID), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// Snapshots returns a list of all Snapshots for all types.
func (c *client) Snapshots(
	ctx types.Context) (types.ServiceSnapshotMap, error) {
	cctx(&ctx)
	reply := types.ServiceSnapshotMap{}
	if _, err := c.httpGet(ctx, "/snapshots", &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// SnapshotsByService returns a list of all Snapshots for a single service.
func (c *client) SnapshotsByService(
	ctx types.Context, service string) (types.SnapshotMap, error) {
	cctx(&ctx)
	reply := types.SnapshotMap{}
	if _, err := c.httpGet(ctx,
		fmt.Sprintf("/snapshots/%s", service), &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// SnapshotInspect gets information about a single snapshot.
func (c *client) SnapshotInspect(
	ctx types.Context,
	service, snapshotID string) (*types.Snapshot, error) {
	cctx(&ctx)
	reply := types.Snapshot{}
	if _, err := c.httpGet(ctx,
		fmt.Sprintf(
			"/snapshots/%s/%s", service, snapshotID), &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// SnapshotRemove removes a single snapshot.
func (c *client) SnapshotRemove(
	ctx types.Context,
	service, snapshotID string) error {
	cctx(&ctx)
	if _, err := c.httpDelete(ctx,
		fmt.Sprintf("/snapshots/%s/%s", service, snapshotID), nil); err != nil {
		return err
	}
	return nil
}

// SnapshotCopy copies a snapshot to a new snapshot.
func (c *client) SnapshotCopy(
	ctx types.Context,
	service, snapshotID string,
	request *types.SnapshotCopyRequest) (*types.Snapshot, error) {
	cctx(&ctx)
	reply := types.Snapshot{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/snapshots/%s/%s?copy",
			service, snapshotID), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// Executors returns information about the executors.
func (c *client) Executors(
	ctx types.Context) (map[string]*types.ExecutorInfo, error) {
	cctx(&ctx)
	reply := map[string]*types.ExecutorInfo{}
	if _, err := c.httpGet(ctx, "/executors", &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// ExecutorHead returns information about an executor.
func (c *client) ExecutorHead(
	ctx types.Context,
	name string) (*types.ExecutorInfo, error) {
	cctx(&ctx)

	res, err := c.httpHead(ctx, fmt.Sprintf("/executors/%s", name))
	if err != nil {
		return nil, err
	}

	size, err := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return nil, err
	}

	buf, err := base64.StdEncoding.DecodeString(res.Header.Get("Content-MD5"))
	if err != nil {
		return nil, err
	}

	return &types.ExecutorInfo{
		Name:        name,
		Size:        size,
		MD5Checksum: fmt.Sprintf("%x", buf),
	}, nil
}

// ExecutorGet downloads an executor.
func (c *client) ExecutorGet(
	ctx types.Context, name string) (io.ReadCloser, error) {
	cctx(&ctx)
	res, err := c.httpGet(ctx, fmt.Sprintf("/executors/%s", name), nil)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func cctx(ctx *types.Context) {
	if *ctx == nil {
		*ctx = context.Background()
	}
}
