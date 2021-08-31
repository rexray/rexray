package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/util"
)

var (
	installFunc   func(apitypes.Context)
	uninstallFunc func(apitypes.Context, bool)
)

func init() {
	initCmdFuncs = append(initCmdFuncs, func(c *CLI) {
		c.initOtherCmds()
		c.initOtherFlags()
	})
}

func (c *CLI) initOtherCmds() {
	c.versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			util.PrintVersion(os.Stdout)
		},
	}
	c.c.AddCommand(c.versionCmd)

	c.envCmd = &cobra.Command{
		Use:   "env",
		Short: "Print the REX-Ray environment",
		Run: func(cmd *cobra.Command, args []string) {
			evs := c.config.EnvVars()
			for _, ev := range evs {
				fmt.Println(ev)
			}
		},
	}
	c.c.AddCommand(c.envCmd)

	c.installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install REX-Ray",
		Run: func(cmd *cobra.Command, args []string) {
			if installFunc != nil {
				if err := installSelfCert(c.ctx, c.config); err != nil {
					fmt.Fprintf(os.Stderr, "install failed: %v\n", err)
					os.Exit(1)
				}
				installFunc(c.ctx)
			}
		},
	}
	c.c.AddCommand(c.installCmd)

	c.uninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall REX-Ray",
		Run: func(cmd *cobra.Command, args []string) {
			if uninstallFunc != nil {
				pkgManager, _ := cmd.Flags().GetBool("package")
				uninstallFunc(c.ctx, pkgManager)
			}
		},
	}
	c.c.AddCommand(c.uninstallCmd)
}

func (c *CLI) initOtherFlags() {
	c.c.PersistentFlags().BoolP("help", "?", false,
		"Help about the current command")

	c.c.PersistentFlags().StringVarP(&c.cfgFile, "config", "c", "",
		"The path to a custom REX-Ray configuration file")
	c.c.PersistentFlags().BoolP(
		"verbose", "v", false, "Print verbose help information")
	//c.c.PersistentFlags().StringVarP(&c.service, "service", "s", "",
	//	"The name of the libStorage service")

	for _, fs := range c.config.FlagSets() {
		c.c.PersistentFlags().AddFlagSet(fs)
	}

	c.uninstallCmd.Flags().Bool("package", false,
		"A flag indicating a package manager is performing the uninstallation")
}
