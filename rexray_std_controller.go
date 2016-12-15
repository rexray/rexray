// +build !rexray_build_type_client
// +build !rexray_build_type_agent

package rexray

import (
	// load the libstorage packages
	_ "github.com/codedellemc/libstorage/imports/executors"
	_ "github.com/codedellemc/libstorage/imports/remote"
	_ "github.com/codedellemc/libstorage/imports/routers"
)
