package executors

import (
	// load the storage executors
	_ "github.com/codedellemc/libstorage/drivers/storage/ebs/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/efs/executor"
	//_ "github.com/codedellemc/libstorage/drivers/storage/gce/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/isilon/executor"
	//_ "github.com/codedellemc/libstorage/drivers/storage/openstack/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/scaleio/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/vbox/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/vfs/executor"
	//_ "github.com/codedellemc/libstorage/drivers/storage/vmax/executor"
	//_ "github.com/codedellemc/libstorage/drivers/storage/xtremio/executor"
)
