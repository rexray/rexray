// +build !client
// +build !controller

package cli

import "github.com/rexray/rexray/agent"

func init() {
	startFuncs = append(startFuncs, agent.Start)
}
