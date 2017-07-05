// +build !rexray_build_type_client
// +build !rexray_build_type_controller
// +build csi

package daemon

import (
	// load the modules
	_ "github.com/codedellemc/rexray/daemon/module/csi"
)
