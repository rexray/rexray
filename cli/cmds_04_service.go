// +build !client

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	initCmdFuncs = append(initCmdFuncs, func(c *CLI) {
		c.initServiceCmds()
		c.initServiceFlags()
	})
}

func (c *CLI) initServiceCmds() {
	c.serviceCmd = &cobra.Command{
		Use:   "service",
		Short: "The service controller",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
	c.c.AddCommand(c.serviceCmd)

	c.serviceStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the service",
		Run: func(cmd *cobra.Command, args []string) {
			serviceStart(c.ctx, c.config, c.nopid)
		},
	}
	c.c.AddCommand(c.serviceStartCmd)
	c.serviceCmd.AddCommand(c.serviceStartCmd)

	c.serviceRestartCmd = &cobra.Command{
		Use:     "restart",
		Aliases: []string{"reload", "force-reload"},
		Short:   "Restart the service",
		Run: func(cmd *cobra.Command, args []string) {
			serviceRestart(c.ctx, c.config, c.nopid)
		},
	}
	c.c.AddCommand(c.serviceRestartCmd)
	c.serviceCmd.AddCommand(c.serviceRestartCmd)

	c.serviceStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop the service",
		Run: func(cmd *cobra.Command, args []string) {
			serviceStop(c.ctx)
		},
	}
	c.c.AddCommand(c.serviceStopCmd)
	c.serviceCmd.AddCommand(c.serviceStopCmd)

	c.serviceStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Print the service status",
		Run: func(cmd *cobra.Command, args []string) {
			serviceStatus(c.ctx)
		},
	}
	c.c.AddCommand(c.serviceStatusCmd)
	c.serviceCmd.AddCommand(c.serviceStatusCmd)

	c.serviceInitSysCmd = &cobra.Command{
		Use:   "initsys",
		Short: "Print the detected init system type",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(getInitSystemCmd())
		},
	}
	c.c.AddCommand(c.serviceInitSysCmd)
}

func (c *CLI) initServiceFlags() {
	c.serviceStartCmd.Flags().BoolVarP(&c.nopid, "nopid", "", false,
		"Disable PID file checking")
	c.serviceStartCmd.Flags().BoolVarP(&c.force, "force", "", false,
		"Forces the service to start, ignoring errors")
}
