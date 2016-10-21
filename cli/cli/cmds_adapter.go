package cli

import (
	"github.com/spf13/cobra"
)

func (c *CLI) initAdapterCmdsAndFlags() {
	c.initAdapterCmds()
	c.initAdapterFlags()
}

func (c *CLI) initAdapterCmds() {
	c.adapterCmd = &cobra.Command{
		Use:              "adapter",
		Short:            "The adapter manager",
		PersistentPreRun: c.preRunActivateLibStorage,
		Run: func(cmd *cobra.Command, args []string) {
			if isHelpFlags(cmd) {
				cmd.Usage()
			} else {
				c.adapterGetTypesCmd.Run(c.adapterGetTypesCmd, args)
			}
		},
	}
	c.c.AddCommand(c.adapterCmd)

	c.adapterGetTypesCmd = &cobra.Command{
		Use:     "types",
		Short:   "List the configured services",
		Aliases: []string{"ls", "list"},
		Run: func(cmd *cobra.Command, args []string) {
			c.mustMarshalOutput(c.r.API().Services(c.ctx))
		},
	}
	c.adapterCmd.AddCommand(c.adapterGetTypesCmd)

	c.adapterGetInstancesCmd = &cobra.Command{
		Use:   "instances",
		Short: "List the configured adapter instances",
		Run: func(cmd *cobra.Command, args []string) {
			c.mustMarshalOutput(c.r.API().Instances(c.ctx))
		},
	}
	c.adapterCmd.AddCommand(c.adapterGetInstancesCmd)
}

func (c *CLI) initAdapterFlags() {
	c.addOutputFormatFlag(c.adapterGetTypesCmd.Flags())
	c.addOutputFormatFlag(c.adapterGetInstancesCmd.Flags())
}
