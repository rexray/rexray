package registry

import (
	"github.com/emccode/libstorage/api/types"
	lstypes "github.com/emccode/libstorage/drivers/storage/libstorage/types"
)

type sdm struct {
	types.StorageDriver
	types.Context
}

// NewStorageDriverManager returns a new storage driver manager.
func NewStorageDriverManager(
	d types.StorageDriver) types.StorageDriver {
	return &sdm{StorageDriver: d}
}

func (d *sdm) API() lstypes.Client {
	if sd, ok := d.StorageDriver.(lstypes.Driver); ok {
		return sd.API()
	}
	return nil
}
