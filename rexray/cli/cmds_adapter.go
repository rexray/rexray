package cli

import (
	"fmt"
	"log"

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
			services, err := c.r.API().Services(c.ctx)
			if err != nil {
				log.Fatalf("Error: %s", err)
			}
			if len(services) > 0 {
				out, err := c.marshalOutput(&services)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(out)
			}
		},
	}
	c.adapterCmd.AddCommand(c.adapterGetTypesCmd)

	c.adapterGetInstancesCmd = &cobra.Command{
		Use:   "instances",
		Short: "List the configured adapter instances",
		Run: func(cmd *cobra.Command, args []string) {

			instances, err := c.r.API().Instances(c.ctx)
			if err != nil {
				log.Fatalf("Error: %s", err)
			}

			if len(instances) > 0 {
				out, err := c.marshalOutput(&instances)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(out)
			}
		},
	}
	c.adapterCmd.AddCommand(c.adapterGetInstancesCmd)
}

func (c *CLI) initAdapterFlags() {
	c.addOutputFormatFlag(c.adapterGetTypesCmd.Flags())
	c.addOutputFormatFlag(c.adapterGetInstancesCmd.Flags())
}
