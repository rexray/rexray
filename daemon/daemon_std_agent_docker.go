// +build !client
// +build !controller

package daemon

import (
	// load the modules
	_ "github.com/codedellemc/rexray/daemon/module/docker/volumedriver"
)
