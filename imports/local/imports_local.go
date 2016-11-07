package local

import (
	// load the config
	_ "github.com/codedellemc/libstorage/imports/config"

	// load the libStorage storage driver
	_ "github.com/codedellemc/libstorage/drivers/storage/libstorage"

	// load the os drivers
	_ "github.com/codedellemc/libstorage/drivers/os/darwin"
	_ "github.com/codedellemc/libstorage/drivers/os/linux"

	// load the integration drivers
	_ "github.com/codedellemc/libstorage/drivers/integration/docker"
)
