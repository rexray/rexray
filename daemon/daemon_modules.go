// +build !exclude_admin

package daemon

import (
	// load the modules
	_ "github.com/emccode/rexray/daemon/module/admin"
	_ "github.com/emccode/rexray/daemon/module/docker/remotevolumedriver"
	_ "github.com/emccode/rexray/daemon/module/docker/volumedriver"
)
