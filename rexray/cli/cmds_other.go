package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/emccode/rexray/util"
)

func (c *CLI) initOtherCmdsAndFlags() {
	c.initOtherCmds()
	c.initOtherFlags()
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

	c.installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install REX-Ray",
		Run: func(cmd *cobra.Command, args []string) {
			install()
		},
	}
	c.c.AddCommand(c.installCmd)

	c.uninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall REX-Ray",
		Run: func(cmd *cobra.Command, args []string) {
			pkgManager, _ := cmd.Flags().GetBool("package")
			uninstall(pkgManager)
		},
	}
	c.c.AddCommand(c.uninstallCmd)
}

func (c *CLI) initOtherFlags() {
	cobra.HelpFlagShorthand = "?"
	cobra.HelpFlagUsageFormatString = "Help for %s"

	c.c.PersistentFlags().StringVarP(&c.cfgFile, "config", "c",
		fmt.Sprintf("%s/.rexray/config.yml", util.HomeDir()),
		"The REX-Ray configuration file")
	c.c.PersistentFlags().BoolP(
		"verbose", "v", false, "Print verbose help information")

	c.c.PersistentFlags().AddFlagSet(c.r.Config.GlobalFlags)
	c.c.PersistentFlags().AddFlagSet(c.r.Config.AdditionalFlags)

	c.uninstallCmd.Flags().Bool("package", false,
		"A flag indicating a package manager is performing the uninstallation")
}
