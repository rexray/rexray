// +build !client
// +build !controller

package main

import (
	// Load the agent's modules
	_ "github.com/codedellemc/rexray/agent/csi"

	// Load the in-tree CSI plug-ins
	_ "github.com/codedellemc/rexray/agent/csi/libstorage"

	// Load vendored CSI plug-ins
	_ "github.com/codedellemc/csi-blockdevices/provider"
	_ "github.com/codedellemc/csi-nfs/provider"
	_ "github.com/thecodeteam/csi-vfs/provider"
)
