// +build !client
// +build !controller

package cli

import "github.com/AVENTER-UG/rexray/agent"

func init() {
	startFuncs = append(startFuncs, agent.Start)
}
