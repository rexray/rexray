package imports

import (
	// import the storage sub-system
	_ "github.com/emccode/rexray/storage"
	_ "github.com/emccode/rexray/storage/ec2"
	_ "github.com/emccode/rexray/storage/openstack"
	_ "github.com/emccode/rexray/storage/rackspace"
	_ "github.com/emccode/rexray/storage/scaleio"
	_ "github.com/emccode/rexray/storage/xtremio"
)
