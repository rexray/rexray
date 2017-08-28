// +build !client
// +build !controller

package cli

import "github.com/codedellemc/rexray/agent"

func init() {
	startFuncs = append(startFuncs, agent.Start)
}
