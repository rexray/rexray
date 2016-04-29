package client

import (
	"github.com/emccode/libstorage/api/types"
	libstor "github.com/emccode/libstorage/drivers/storage/libstorage"
)

func (c *client) API() libstor.Client {
	return c.lsc
}

func (c *client) OS() types.OSDriver {
	return c.od
}

func (c *client) Storage() types.StorageDriver {
	return c.sd
}

func (c *client) Integration() types.IntegrationDriver {
	return c.id
}
