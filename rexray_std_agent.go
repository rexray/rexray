// +build !client
// +build !controller

package main

import (
	// Load the agent's modules
	_ "github.com/codedellemc/rexray/agent/docker"
)
