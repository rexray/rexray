package client

import (
	"github.com/emccode/libstorage/api/types"
	lstypes "github.com/emccode/libstorage/drivers/storage/libstorage/types"
)

func (c *client) API() lstypes.Client {
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
