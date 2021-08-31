// +build !agent
// +build !controller

package cli

import (
	log "github.com/sirupsen/logrus"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/spf13/cobra"
)

func init() {
	initCmdFuncs = append(initCmdFuncs, func(c *CLI) {
		c.initDeviceCmds()
		c.initDeviceFlags()
	})
}

func (c *CLI) initDeviceCmds() {
	c.deviceCmd = &cobra.Command{
		Use:              "device",
		Short:            "The device manager",
		PersistentPreRun: c.preRunActivateLibStorage,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
	c.c.AddCommand(c.deviceCmd)

	c.deviceGetCmd = &cobra.Command{
		Use:     "ls",
		Short:   "Get a device's mount(s)",
		Aliases: []string{"get", "list"},
		Run: func(cmd *cobra.Command, args []string) {
			c.mustMarshalOutput(c.r.OS().Mounts(
				c.ctx, c.deviceName, c.mountPoint, store()))
		},
	}
	c.deviceCmd.AddCommand(c.deviceGetCmd)

	c.deviceMountCmd = &cobra.Command{
		Use:   "mount",
		Short: "Mount a device",
		Run: func(cmd *cobra.Command, args []string) {
			if c.deviceName == "" || c.mountPoint == "" {
				log.Fatal("Missing --devicename and --mountpoint")
			}
			if err := c.r.OS().Mount(
				c.ctx, c.deviceName, c.mountPoint,
				&apitypes.DeviceMountOpts{
					MountOptions: c.mountOptions,
					MountLabel:   c.mountLabel,
				}); err != nil {
				log.Fatal(err)
			}
		},
	}
	c.deviceCmd.AddCommand(c.deviceMountCmd)

	c.devuceUnmountCmd = &cobra.Command{
		Use:   "unmount",
		Short: "Unmount a device",
		Run: func(cmd *cobra.Command, args []string) {
			if c.mountPoint == "" {
				log.Fatal("Missing --mountpoint")
			}
			if err := c.r.OS().Unmount(
				c.ctx, c.mountPoint, store()); err != nil {
				log.Fatal(err)
			}
		},
	}
	c.deviceCmd.AddCommand(c.devuceUnmountCmd)

	c.deviceFormatCmd = &cobra.Command{
		Use:   "format",
		Short: "Format a device",
		Run: func(cmd *cobra.Command, args []string) {
			if c.deviceName == "" {
				log.Fatal("Missing --devicename")
			}
			if c.fsType == "" {
				c.fsType = "ext4"
			}
			if err := c.r.OS().Format(
				c.ctx, c.deviceName,
				&apitypes.DeviceFormatOpts{
					NewFSType:   c.fsType,
					OverwriteFS: c.overwriteFs,
				}); err != nil {
				log.Fatal(err)
			}
		},
	}
	c.deviceCmd.AddCommand(c.deviceFormatCmd)
}

func (c *CLI) initDeviceFlags() {
	c.deviceGetCmd.Flags().StringVar(&c.deviceName, "deviceName", "", "")
	c.deviceGetCmd.Flags().StringVar(&c.mountPoint, "mountPoint", "", "")
	c.deviceMountCmd.Flags().StringVar(&c.deviceName, "deviceName", "", "")
	c.deviceMountCmd.Flags().StringVar(&c.mountPoint, "mountPoint", "", "")
	c.deviceMountCmd.Flags().StringVar(&c.mountOptions, "mountOptions", "", "")
	c.deviceMountCmd.Flags().StringVar(&c.mountLabel, "mountLabel", "", "")
	c.devuceUnmountCmd.Flags().StringVar(&c.mountPoint, "mountPoint", "", "")
	c.deviceFormatCmd.Flags().StringVar(&c.deviceName, "deviceName", "", "")
	c.deviceFormatCmd.Flags().StringVar(&c.fsType, "fsType", "", "")
	c.deviceFormatCmd.Flags().BoolVar(&c.overwriteFs, "overwriteFS", false, "")

	c.addOutputFormatFlag(c.deviceCmd.Flags())
	c.addOutputFormatFlag(c.deviceGetCmd.Flags())
}
