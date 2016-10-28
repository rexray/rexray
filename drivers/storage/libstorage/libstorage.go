package libstorage

import (
	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"
)

const (
	// Name is the name of the driver.
	Name = types.LibStorageDriverName
)

var (
	lsxMutex = types.Run.Join("lsx.lock")
)

func init() {
	registry.RegisterStorageDriver(Name, newDriver)
}
