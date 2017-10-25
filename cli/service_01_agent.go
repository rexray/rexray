// +build !client
// +build !controller

package cli

import "github.com/thecodeteam/rexray/agent"

func init() {
	startFuncs = append(startFuncs, agent.Start)
}
