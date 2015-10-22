package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (c *CLI) initAdapterCmdsAndFlags() {
	c.initAdapterCmds()
	c.initAdapterFlags()
}

func (c *CLI) initAdapterCmds() {
	c.adapterCmd = &cobra.Command{
		Use:   "adapter",
		Short: "The adapter manager",
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
		Short:   "List the available adapter types",
		Aliases: []string{"ls", "list"},
		Run: func(cmd *cobra.Command, args []string) {
			for n := range c.r.DriverNames() {
				fmt.Printf("Storage Driver: %v\n", n)
			}
		},
	}
	c.adapterCmd.AddCommand(c.adapterGetTypesCmd)

	c.adapterGetInstancesCmd = &cobra.Command{
		Use:   "instances",
		Short: "List the configured adapter instances",
		Run: func(cmd *cobra.Command, args []string) {

			allInstances, err := c.r.Storage.GetInstances()
			if err != nil {
				panic(err)
			}

			if len(allInstances) > 0 {
				out, err := c.marshalOutput(&allInstances)
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
	c.addOutputFormatFlag(c.adapterGetInstancesCmd.Flags())
}
