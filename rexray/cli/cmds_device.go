package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (c *CLI) initDeviceCmdsAndFlags() {
	c.initDeviceCmds()
	c.initDeviceFlags()
}

func (c *CLI) initDeviceCmds() {
	c.deviceCmd = &cobra.Command{
		Use:   "device",
		Short: "The device manager",
		Run: func(cmd *cobra.Command, args []string) {
			if isHelpFlags(cmd) {
				cmd.Usage()
			} else {
				c.deviceGetCmd.Run(c.deviceGetCmd, args)
			}
		},
	}
	c.c.AddCommand(c.deviceCmd)

	c.deviceGetCmd = &cobra.Command{
		Use:     "get",
		Short:   "Get a device's mount(s)",
		Aliases: []string{"ls", "list"},
		Run: func(cmd *cobra.Command, args []string) {

			mounts, err := c.r.OS.GetMounts(c.deviceName, c.mountPoint)
			if err != nil {
				log.Fatal(err)
			}

			out, err := c.marshalOutput(&mounts)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(out)
		},
	}

	c.deviceMountCmd = &cobra.Command{
		Use:   "mount",
		Short: "Mount a device",
		Run: func(cmd *cobra.Command, args []string) {

			if c.deviceName == "" || c.mountPoint == "" {
				log.Fatal("Missing --devicename and --mountpoint")
			}

			// mountOptions = fmt.Sprintf("val,%s", mountOptions)
			err := c.r.OS.Mount(
				c.deviceName, c.mountPoint, c.mountOptions, c.mountLabel)
			if err != nil {
				log.Fatal(err)
			}

		},
	}

	c.devuceUnmountCmd = &cobra.Command{
		Use:   "unmount",
		Short: "Unmount a device",
		Run: func(cmd *cobra.Command, args []string) {

			if c.mountPoint == "" {
				log.Fatal("Missing --mountpoint")
			}

			err := c.r.OS.Unmount(c.mountPoint)
			if err != nil {
				log.Fatal(err)
			}

		},
	}

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

			err := c.r.OS.Format(c.deviceName, c.fsType, c.overwriteFs)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
}

func (c *CLI) initDeviceFlags() {
	c.deviceGetCmd.Flags().StringVar(&c.deviceName, "devicename", "", "devicename")
	c.deviceGetCmd.Flags().StringVar(&c.mountPoint, "mountpoint", "", "mountpoint")
	c.deviceMountCmd.Flags().StringVar(&c.deviceName, "devicename", "", "devicename")
	c.deviceMountCmd.Flags().StringVar(&c.mountPoint, "mountpoint", "", "mountpoint")
	c.deviceMountCmd.Flags().StringVar(&c.mountOptions, "mountoptions", "", "mountoptions")
	c.deviceMountCmd.Flags().StringVar(&c.mountLabel, "mountlabel", "", "mountlabel")
	c.devuceUnmountCmd.Flags().StringVar(&c.mountPoint, "mountpoint", "", "mountpoint")
	c.deviceFormatCmd.Flags().StringVar(&c.deviceName, "devicename", "", "devicename")
	c.deviceFormatCmd.Flags().StringVar(&c.fsType, "fstype", "", "fstype")
	c.deviceFormatCmd.Flags().BoolVar(&c.overwriteFs, "overwritefs", false, "overwritefs")

	c.addOutputFormatFlag(c.deviceCmd.Flags())
	c.addOutputFormatFlag(c.deviceGetCmd.Flags())
}
