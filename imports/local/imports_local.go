package local

import (
	// load the config
	_ "github.com/codedellemc/libstorage/imports/config"

	// load the libStorage storage driver
	_ "github.com/codedellemc/libstorage/drivers/storage/libstorage"
)
