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

	// load the client drivers
	//_ "github.com/codedellemc/libstorage/drivers/storage/ec2/client"
	//_ "github.com/codedellemc/libstorage/drivers/storage/gce/client"
	//_ "github.com/codedellemc/libstorage/drivers/storage/isilon/client"
	// _ "github.com/codedellemc/libstorage/drivers/storage/mock/client"
	//_ "github.com/codedellemc/libstorage/drivers/storage/openstack/client"
	//_ "github.com/codedellemc/libstorage/drivers/storage/rackspace/client"
	// _ "github.com/codedellemc/libstorage/drivers/storage/scaleio"
	//_ "github.com/codedellemc/libstorage/drivers/storage/vbox/client"
	//_ "github.com/codedellemc/libstorage/drivers/storage/scaleio/client"
	_ "github.com/codedellemc/libstorage/drivers/storage/vfs/client"
	//_ "github.com/codedellemc/libstorage/drivers/storage/virtualbox"
	//_ "github.com/codedellemc/libstorage/drivers/storage/vmax/client"
	//_ "github.com/codedellemc/libstorage/drivers/storage/xtremio/client"
)
