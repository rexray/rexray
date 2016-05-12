package executors

import (
	// load the storage executors
	//_ "github.com/emccode/libstorage/drivers/storage/ec2/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/gce/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/isilon/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/openstack/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/rackspace/executor"
	_ "github.com/emccode/libstorage/drivers/storage/scaleio/executor"
	_ "github.com/emccode/libstorage/drivers/storage/vbox/executor"
	_ "github.com/emccode/libstorage/drivers/storage/vfs/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/vmax/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/xtremio/executor"
)
