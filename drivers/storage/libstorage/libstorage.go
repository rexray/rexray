package libstorage

import (
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils/paths"
)

const (
	// Name is the name of the driver.
	Name = types.LibStorageDriverName
)

var (
	lsxMutex = paths.Run.Join("lsx.lock")
)

func init() {
	registry.RegisterStorageDriver(Name, newDriver)
}
