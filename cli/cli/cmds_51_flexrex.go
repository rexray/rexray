// +build !rexray_build_type_agent
// +build !rexray_build_type_controller

package cli

import (
	"os"
	"path"

	"github.com/akutz/gotil"
	"github.com/codedellemc/rexray/util"
	"github.com/spf13/cobra"
)

func init() {
	initCmdFuncs = append(initCmdFuncs, func(c *CLI) {
		c.initFlexRexCmds()
		c.initFlexRexFlags()
	})
}

func (c *CLI) initFlexRexCmds() {
	c.flexRexCmd = &cobra.Command{
		Use:   "flexrex",
		Short: "The FlexVol REX-Ray plug-in manager",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
	c.c.AddCommand(c.flexRexCmd)

	c.flexRexInstallCmd = &cobra.Command{
		Use:   "install",
		Short: "Install the FlexVol REX-Ray plug-in",
		Run: func(cmd *cobra.Command, args []string) {
			fp := util.ScriptFilePath("flexrex")
			if !gotil.FileExists(fp) {
				if _, err := c.installScripts("flexrex"); err != nil {
					c.ctx.Fatal(err)
				}
			}
			err := os.MkdirAll(path.Dir(c.scriptPath), os.FileMode(0755))
			if err != nil {
				c.ctx.Fatal(err)
			}
			if err := os.Symlink(fp, c.scriptPath); err != nil {
				c.ctx.Fatal(err)
			}
			c.mustMarshalOutput(&scriptInfo{
				Name:      "flexrex",
				Path:      c.scriptPath,
				Installed: true,
				Modified:  false,
			}, nil)
		},
	}
	c.flexRexCmd.AddCommand(c.flexRexInstallCmd)

	c.flexRexUninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the FlexVol REX-Ray plug-in",
		Run: func(cmd *cobra.Command, args []string) {
			c.mustMarshalOutput(
				[]string{c.scriptPath},
				os.RemoveAll(c.scriptPath))
		},
	}
	c.flexRexCmd.AddCommand(c.flexRexUninstallCmd)
}

const (
	defaultFlexRexPath = "/usr/libexec/kubernetes/kubelet-plugins" +
		"/volume/exec/rexray~flexrex/flexrex"
)

func (c *CLI) initFlexRexFlags() {
	c.flexRexInstallCmd.Flags().StringVar(
		&c.scriptPath,
		"path",
		defaultFlexRexPath,
		"The absolute path to which the script should be installed")
	c.flexRexUninstallCmd.Flags().StringVar(
		&c.scriptPath,
		"path",
		defaultFlexRexPath,
		"The absolute path of the script to uninstall")
}
