// +build !client
// +build !controller
// +build libstorage_integration_driver_linux

package daemon

import (
	// load the modules
	_ "github.com/codedellemc/rexray/daemon/module/docker/volumedriver"
)
