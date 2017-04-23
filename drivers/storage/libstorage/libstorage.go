package libstorage

import (
	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"
)

const (
	// Name is the name of the driver.
	Name = types.LibStorageDriverName
)

func init() {
	registry.RegisterStorageDriver(Name, newDriver)
}
