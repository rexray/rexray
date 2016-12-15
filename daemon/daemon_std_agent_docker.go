// +build !rexray_build_type_client
// +build !rexray_build_type_controller
// +build libstorage_integration_driver_docker

package daemon

import (
	// load the modules
	_ "github.com/codedellemc/rexray/daemon/module/docker/volumedriver"
)
