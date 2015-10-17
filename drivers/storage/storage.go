package storage

import (
	// loads the storage drivers
	_ "github.com/emccode/rexray/drivers/storage/ec2"
	_ "github.com/emccode/rexray/drivers/storage/openstack"
	_ "github.com/emccode/rexray/drivers/storage/rackspace"
	_ "github.com/emccode/rexray/drivers/storage/scaleio"
	_ "github.com/emccode/rexray/drivers/storage/xtremio"
)
