package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (c *CLI) initServiceCmdsAndFlags() {
	c.initServiceCmds()
	c.initServiceFlags()
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
			c.start()
		},
	}
	c.c.AddCommand(c.serviceStartCmd)
	c.serviceCmd.AddCommand(c.serviceStartCmd)

	c.serviceRestartCmd = &cobra.Command{
		Use:     "restart",
		Aliases: []string{"reload", "force-reload"},
		Short:   "Restart the service",
		Run: func(cmd *cobra.Command, args []string) {
			c.restart()
		},
	}
	c.c.AddCommand(c.serviceRestartCmd)
	c.serviceCmd.AddCommand(c.serviceRestartCmd)

	c.serviceStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop the service",
		Run: func(cmd *cobra.Command, args []string) {
			stop()
		},
	}
	c.c.AddCommand(c.serviceStopCmd)
	c.serviceCmd.AddCommand(c.serviceStopCmd)

	c.serviceStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Print the service status",
		Run: func(cmd *cobra.Command, args []string) {
			c.status()
		},
	}
	c.c.AddCommand(c.serviceStatusCmd)
	c.serviceCmd.AddCommand(c.serviceStatusCmd)

	c.serviceInitSysCmd = &cobra.Command{
		Use:   "initsys",
		Short: "Print the detected init system type",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("\nInit System: %s\n", getInitSystemCmd())
		},
	}
	c.serviceCmd.AddCommand(c.serviceInitSysCmd)
}

func (c *CLI) initServiceFlags() {
	c.serviceStartCmd.Flags().BoolVarP(&c.fg, "foreground", "f", false,
		"Starts the service in the foreground")
	c.serviceStartCmd.Flags().BoolVarP(&c.force, "force", "", false,
		"Forces the service to start, ignoring errors")
	c.serviceStartCmd.Flags().BoolVarP(&c.fork, "fork", "", false,
		"Indicates that the server is being forked.")
}
