// +build !rexray_build_type_client
// +build !rexray_build_type_agent

package rexray

import (
	// load the libstorage packages
	_ "github.com/codedellemc/rexray/libstorage/imports/storage"
)
