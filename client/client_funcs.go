package client

import (
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	libstor "github.com/emccode/libstorage/drivers/storage/libstorage"
)

func (c *client) API() libstor.Client {
	return c.lsc
}

func (c *client) Driver() types.StorageDriver {
	return c.sd
}

func (c *client) List(service string) ([]*types.Volume, error) {
	return c.id.Volumes(c.ctx.WithServiceName(service), utils.NewStore())
}

func (c *client) Inspect(service, volumeName string) (*types.Volume, error) {
	return c.id.Inspect(
		c.ctx.WithServiceName(service), volumeName, utils.NewStore())
}

func (c *client) Mount(
	service, volumeID, volumeName string,
	opts *types.VolumeMountOpts) (string, *types.Volume, error) {
	return c.id.Mount(
		c.ctx.WithServiceName(service), volumeID, volumeName, opts)
}

func (c *client) Unmount(service, volumeID, volumeName string) error {
	return c.id.Unmount(
		c.ctx.WithServiceName(service), volumeID, volumeName, utils.NewStore())
}

func (c *client) Path(service, volumeID, volumeName string) (string, error) {
	return c.id.Path(
		c.ctx.WithServiceName(service), volumeID, volumeName, utils.NewStore())
}

func (c *client) Create(
	service, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	return c.id.Create(c.ctx.WithServiceName(service), volumeName, opts)
}

func (c *client) Remove(service, volumeName string) error {
	return c.id.Remove(
		c.ctx.WithServiceName(service), volumeName, utils.NewStore())
}

func (c *client) Attach(
	service, volumeName string,
	opts *types.VolumeAttachOpts) (string, error) {
	return c.id.Attach(c.ctx.WithServiceName(service), volumeName, opts)
}

func (c *client) Detach(
	service, volumeName string,
	opts *types.VolumeDetachOpts) error {
	return c.id.Detach(c.ctx.WithServiceName(service), volumeName, opts)
}

func (c *client) NetworkName(service, volumeName string) (string, error) {
	return c.id.NetworkName(
		c.ctx.WithServiceName(service), volumeName, utils.NewStore())
}
