package libstorage

import (
	"github.com/rexray/rexray/libstorage/api/registry"
	"github.com/rexray/rexray/libstorage/api/types"
)

const (
	// Name is the name of the driver.
	Name = types.LibStorageDriverName
)

func init() {
	registry.RegisterStorageDriver(Name, newDriver)
}
