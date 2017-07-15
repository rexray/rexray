// +build !libstorage_storage_driver libstorage_storage_driver_cinder
// +build darwin

package executor

import (
	// load the packages
	_ "github.com/codedellemc/libstorage/drivers/os/darwin"
)
