// +build !client
// +build !controller

package main

import (
	// Load the agent's modules
	_ "github.com/AVENTER-UG/rexray/agent/csi"

	// Load the in-tree CSI plug-ins
	_ "github.com/AVENTER-UG/rexray/agent/csi/libstorage"

	// Load vendored CSI plug-ins
	_ "github.com/thecodeteam/csi-blockdevices/provider"
	_ "github.com/thecodeteam/csi-nfs/provider"
	_ "github.com/thecodeteam/csi-vfs/provider"
)
