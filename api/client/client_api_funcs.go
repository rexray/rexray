package client

import (
	"encoding/base64"
	"fmt"
	"io"
	"strconv"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
)

// Root returns a list of root resources.
func (c *Client) Root(ctx context.Context) (apihttp.RootResources, error) {
	cctx(&ctx)
	reply := apihttp.RootResources{}
	if _, err := c.httpGet(ctx, "/", &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// Services returns a map of the configured Services.
func (c *Client) Services(ctx context.Context) (apihttp.ServicesMap, error) {
	cctx(&ctx)
	reply := apihttp.ServicesMap{}
	if _, err := c.httpGet(ctx, "/services", &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// ServiceInspect returns information about a service.
func (c *Client) ServiceInspect(
	ctx context.Context, name string) (*types.ServiceInfo, error) {
	cctx(&ctx)
	reply := &types.ServiceInfo{}
	if _, err := c.httpGet(ctx,
		fmt.Sprintf("/services/%s", name), &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// Volumes returns a list of all Volumes for all Services.
func (c *Client) Volumes(
	ctx context.Context,
	attachments bool) (apihttp.ServiceVolumeMap, error) {
	cctx(&ctx)
	reply := apihttp.ServiceVolumeMap{}
	url := fmt.Sprintf("/volumes?attachments=%v", attachments)
	if _, err := c.httpGet(ctx, url, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// VolumesByService returns a list of all Volumes for a service.
func (c *Client) VolumesByService(
	ctx context.Context,
	service string,
	attachments bool) (apihttp.VolumeMap, error) {
	cctx(&ctx)
	reply := apihttp.VolumeMap{}
	url := fmt.Sprintf("/volumes/%s?attachments=%v", service, attachments)
	if _, err := c.httpGet(ctx, url, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// VolumeInspect gets information about a single volume.
func (c *Client) VolumeInspect(
	ctx context.Context,
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
func (c *Client) VolumeCreate(
	ctx context.Context,
	service string,
	request *apihttp.VolumeCreateRequest) (*types.Volume, error) {
	cctx(&ctx)
	reply := types.Volume{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/volumes/%s", service), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// VolumeCreateFromSnapshot creates a single volume from a snapshot.
func (c *Client) VolumeCreateFromSnapshot(
	ctx context.Context,
	service, snapshotID string,
	request *apihttp.VolumeCreateRequest) (*types.Volume, error) {
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
func (c *Client) VolumeCopy(
	ctx context.Context,
	service, volumeID string,
	request *apihttp.VolumeCopyRequest) (*types.Volume, error) {
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
func (c *Client) VolumeRemove(
	ctx context.Context,
	service, volumeID string) error {
	cctx(&ctx)
	if _, err := c.httpDelete(ctx,
		fmt.Sprintf("/volumes/%s/%s", service, volumeID), nil); err != nil {
		return err
	}
	return nil
}

// VolumeAttach attaches a single volume.
func (c *Client) VolumeAttach(
	ctx context.Context,
	service string,
	volumeID string,
	request *apihttp.VolumeAttachRequest) (*types.Volume, error) {
	cctx(&ctx)
	reply := types.Volume{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/volumes/%s/%s?attach",
			service, volumeID), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// VolumeDetach attaches a single volume.
func (c *Client) VolumeDetach(
	ctx context.Context,
	service string,
	volumeID string,
	request *apihttp.VolumeDetachRequest) (*types.Volume, error) {
	cctx(&ctx)
	reply := types.Volume{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/volumes/%s/%s?detach",
			service, volumeID), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// VolumeDetachAll attaches all volumes from all services.
func (c *Client) VolumeDetachAll(
	ctx context.Context,
	request *apihttp.VolumeDetachRequest) (apihttp.ServiceVolumeMap, error) {
	cctx(&ctx)
	reply := apihttp.ServiceVolumeMap{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/volumes?detach"), request, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// VolumeDetachAllForService detaches all volumes from a service.
func (c *Client) VolumeDetachAllForService(
	ctx context.Context,
	service string,
	request *apihttp.VolumeDetachRequest) (apihttp.VolumeMap, error) {
	cctx(&ctx)
	reply := apihttp.VolumeMap{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf(
			"/volumes/%s?detach", service), request, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// VolumeSnapshot creates a single snapshot.
func (c *Client) VolumeSnapshot(
	ctx context.Context,
	service string,
	volumeID string,
	request *apihttp.VolumeSnapshotRequest) (*types.Snapshot, error) {
	cctx(&ctx)
	reply := types.Snapshot{}
	if _, err := c.httpPost(ctx,
		fmt.Sprintf("/volumes/%s/%s?snapshot",
			service, volumeID), request, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// Snapshots returns a list of all Snapshots for all services.
func (c *Client) Snapshots(
	ctx context.Context) (apihttp.ServiceSnapshotMap, error) {
	cctx(&ctx)
	reply := apihttp.ServiceSnapshotMap{}
	if _, err := c.httpGet(ctx, "/snapshots", &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// SnapshotsByService returns a list of all Snapshots for a single service.
func (c *Client) SnapshotsByService(
	ctx context.Context, service string) (apihttp.SnapshotMap, error) {
	cctx(&ctx)
	reply := apihttp.SnapshotMap{}
	if _, err := c.httpGet(ctx,
		fmt.Sprintf("/snapshots/%s", service), &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// SnapshotInspect gets information about a single snapshot.
func (c *Client) SnapshotInspect(
	ctx context.Context,
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
func (c *Client) SnapshotRemove(
	ctx context.Context,
	service, snapshotID string) error {
	cctx(&ctx)
	if _, err := c.httpDelete(ctx,
		fmt.Sprintf("/snapshots/%s/%s", service, snapshotID), nil); err != nil {
		return err
	}
	return nil
}

// SnapshotCopy copies a snapshot to a new snapshot.
func (c *Client) SnapshotCopy(
	ctx context.Context,
	service, snapshotID string,
	request *apihttp.SnapshotCopyRequest) (*types.Snapshot, error) {
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
func (c *Client) Executors(
	ctx context.Context) (apihttp.ExecutorsMap, error) {
	cctx(&ctx)
	reply := apihttp.ExecutorsMap{}
	if _, err := c.httpGet(ctx, "/executors", &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// ExecutorHead returns information about an executor.
func (c *Client) ExecutorHead(
	ctx context.Context,
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
func (c *Client) ExecutorGet(
	ctx context.Context, name string) (io.ReadCloser, error) {
	cctx(&ctx)
	res, err := c.httpGet(ctx, fmt.Sprintf("/executors/%s", name), nil)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func cctx(ctx *context.Context) {
	if *ctx == nil {
		*ctx = context.Background()
	}
}
