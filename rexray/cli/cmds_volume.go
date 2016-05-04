package cli

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	apitypes "github.com/emccode/libstorage/api/types"
)

func (c *CLI) initVolumeCmdsAndFlags() {
	c.initVolumeCmds()
	c.initVolumeFlags()
}

func (c *CLI) initVolumeCmds() {

	c.volumeCmd = &cobra.Command{
		Use:   "volume",
		Short: "The volume manager",
		Run: func(cmd *cobra.Command, args []string) {
			if isHelpFlags(cmd) {
				cmd.Usage()
			} else {
				c.volumeGetCmd.Run(c.volumeGetCmd, args)
			}
		},
	}
	c.c.AddCommand(c.volumeCmd)

	c.volumeMapCmd = &cobra.Command{
		Use:   "map",
		Short: "Print the volume mapping(s)",
		Run: func(cmd *cobra.Command, args []string) {

			allBlockDevices, err := c.r.Storage().Volumes(
				c.ctx, &apitypes.VolumesOpts{Attachments: true})
			if err != nil {
				log.Fatalf("Error: %s", err)
			}

			if len(allBlockDevices) > 0 {
				out, err := c.marshalOutput(&allBlockDevices)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(out)
			}
		},
	}
	c.volumeCmd.AddCommand(c.volumeMapCmd)

	c.volumeGetCmd = &cobra.Command{
		Use:     "get",
		Short:   "Get one or more volumes",
		Aliases: []string{"ls", "list"},
		Run: func(cmd *cobra.Command, args []string) {

			vols, err := c.r.Storage().Volumes(
				c.ctx, &apitypes.VolumesOpts{Attachments: false})
			if err != nil {
				log.Fatal(err)
			}
			if c.volumeID != "" || c.volumeName != "" {
				for _, v := range vols {
					if strings.ToLower(v.ID) == strings.ToLower(c.volumeID) ||
						strings.ToLower(v.Name) == strings.ToLower(c.volumeName) {
						out, err := c.marshalOutput(v)
						if err != nil {
							log.Fatal(err)
						}
						fmt.Println(out)
						return
					}
				}
				return
			}

			if len(vols) > 0 {
				out, err := c.marshalOutput(vols)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(out)
			}
		},
	}
	c.volumeCmd.AddCommand(c.volumeGetCmd)

	c.volumeCreateCmd = &cobra.Command{
		Use:     "create",
		Short:   "Create a new volume",
		Aliases: []string{"new"},
		Run: func(cmd *cobra.Command, args []string) {

			if c.size == 0 && c.snapshotID == "" && c.volumeID == "" {
				log.Fatalf("missing --size")
			}

			opts := &apitypes.VolumeCreateOpts{
				AvailabilityZone: &c.availabilityZone,
				Size:             &c.size,
				Type:             &c.volumeType,
				IOPS:             &c.iops,
				Opts:             store(),
			}

			var (
				err    error
				volume *apitypes.Volume
			)

			if c.volumeID != "" && c.volumeName != "" {
				volume, err = c.r.Storage().VolumeCopy(
					c.ctx, c.volumeID, c.volumeName, opts.Opts)
			} else if c.snapshotID != "" && c.volumeName != "" {
				volume, err = c.r.Storage().VolumeCreateFromSnapshot(
					c.ctx, c.snapshotID, c.volumeName, opts)
			} else if c.volumeName != "" {
				volume, err = c.r.Storage().VolumeCreate(
					c.ctx, c.volumeName, opts)
			}
			// TODO Get All Volumes
			if err != nil {
				log.Fatal(err)
			}

			out, err := c.marshalOutput(&volume)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(out)

		},
	}
	c.volumeCmd.AddCommand(c.volumeCreateCmd)

	c.volumeRemoveCmd = &cobra.Command{
		Use:     "remove",
		Short:   "Remove a volume",
		Aliases: []string{"rm"},
		Run: func(cmd *cobra.Command, args []string) {

			if c.volumeID == "" {
				log.Fatalf("missing --volumeid")
			}

			err := c.r.Storage().VolumeRemove(c.ctx, c.volumeID, store())
			if err != nil {
				log.Fatal(err)
			}

		},
	}
	c.volumeCmd.AddCommand(c.volumeRemoveCmd)

	c.volumeAttachCmd = &cobra.Command{
		Use:   "attach",
		Short: "Attach a volume",
		Run: func(cmd *cobra.Command, args []string) {

			if c.volumeID == "" {
				log.Fatalf("missing --volumeid")
			}

			vol, _, err := c.r.Storage().VolumeAttach(
				c.ctx, c.volumeID,
				&apitypes.VolumeAttachOpts{Force: c.force})

			if err != nil {
				log.Fatal(err)
			}

			out, err := c.marshalOutput(vol)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(out)

		},
	}
	c.volumeCmd.AddCommand(c.volumeAttachCmd)

	c.volumeDetachCmd = &cobra.Command{
		Use:   "detach",
		Short: "Detach a volume",
		Run: func(cmd *cobra.Command, args []string) {

			if c.volumeID == "" {
				log.Fatalf("missing --volumeid")
			}

			_, err := c.r.Storage().VolumeDetach(
				c.ctx, c.volumeID, &apitypes.VolumeDetachOpts{Force: c.force})
			if err != nil {
				log.Fatal(err)
			}

		},
	}
	c.volumeCmd.AddCommand(c.volumeDetachCmd)

	c.volumeMountCmd = &cobra.Command{
		Use:   "mount",
		Short: "Mount a volume",
		Run: func(cmd *cobra.Command, args []string) {
			if c.volumeName == "" && c.volumeID == "" {
				log.Fatal("Missing --volumename or --volumeid")
			}

			mountPath, _, err := c.r.Integration().Mount(
				c.ctx, c.volumeID, c.volumeName,
				&apitypes.VolumeMountOpts{
					NewFSType:   c.fsType,
					OverwriteFS: c.overwriteFs,
				})
			if err != nil {
				log.Fatal(err)
			}

			out, err := c.marshalOutput(&mountPath)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(out)

		},
	}
	c.volumeCmd.AddCommand(c.volumeMountCmd)

	c.volumeUnmountCmd = &cobra.Command{
		Use:   "unmount",
		Short: "Unmount a volume",
		Run: func(cmd *cobra.Command, args []string) {

			if c.volumeName == "" && c.volumeID == "" {
				log.Fatal("Missing --volumename or --volumeid")
			}

			err := c.r.Integration().Unmount(
				c.ctx, c.volumeID, c.volumeName, store())
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	c.volumeCmd.AddCommand(c.volumeUnmountCmd)

	c.volumePathCmd = &cobra.Command{
		Use:   "path",
		Short: "Print the volume path",
		Run: func(cmd *cobra.Command, args []string) {

			if c.volumeName == "" && c.volumeID == "" {
				log.Fatal("Missing --volumename or --volumeid")
			}

			mountPath, err := c.r.Integration().Path(
				c.ctx, c.volumeID, c.volumeName, store())
			if err != nil {
				log.Fatal(err)
			}

			if mountPath != "" {
				out, err := c.marshalOutput(&mountPath)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(out)
			}
		},
	}
	c.volumeCmd.AddCommand(c.volumePathCmd)
}

func (c *CLI) initVolumeFlags() {
	c.volumeGetCmd.Flags().StringVar(&c.volumeName, "volumename", "", "volumename")
	c.volumeGetCmd.Flags().StringVar(&c.volumeID, "volumeid", "", "volumeid")
	c.volumeCreateCmd.Flags().BoolVar(&c.runAsync, "runasync", false, "runasync")
	c.volumeCreateCmd.Flags().StringVar(&c.volumeName, "volumename", "", "volumename")
	c.volumeCreateCmd.Flags().StringVar(&c.volumeType, "volumetype", "", "volumetype")
	c.volumeCreateCmd.Flags().StringVar(&c.volumeID, "volumeid", "", "volumeid")
	c.volumeCreateCmd.Flags().StringVar(&c.snapshotID, "snapshotid", "", "snapshotid")
	c.volumeCreateCmd.Flags().Int64Var(&c.iops, "iops", 0, "IOPS")
	c.volumeCreateCmd.Flags().Int64Var(&c.size, "size", 0, "size")
	c.volumeCreateCmd.Flags().StringVar(&c.availabilityZone, "availabilityzone", "", "availabilityzone")
	c.volumeRemoveCmd.Flags().StringVar(&c.volumeID, "volumeid", "", "volumeid")
	c.volumeAttachCmd.Flags().BoolVar(&c.runAsync, "runasync", false, "runasync")
	c.volumeAttachCmd.Flags().StringVar(&c.volumeID, "volumeid", "", "volumeid")
	c.volumeAttachCmd.Flags().StringVar(&c.instanceID, "instanceid", "", "instanceid")
	c.volumeAttachCmd.Flags().BoolVar(&c.force, "force", false, "force")
	c.volumeDetachCmd.Flags().BoolVar(&c.runAsync, "runasync", false, "runasync")
	c.volumeDetachCmd.Flags().StringVar(&c.volumeID, "volumeid", "", "volumeid")
	c.volumeDetachCmd.Flags().StringVar(&c.instanceID, "instanceid", "", "instanceid")
	c.volumeDetachCmd.Flags().BoolVar(&c.force, "force", false, "force")
	c.volumeMountCmd.Flags().StringVar(&c.volumeID, "volumeid", "", "volumeid")
	c.volumeMountCmd.Flags().StringVar(&c.volumeName, "volumename", "", "volumename")
	c.volumeMountCmd.Flags().BoolVar(&c.overwriteFs, "overwritefs", false, "overwritefs")
	c.volumeMountCmd.Flags().StringVar(&c.fsType, "fstype", "", "fstype")
	c.volumeUnmountCmd.Flags().StringVar(&c.volumeID, "volumeid", "", "volumeid")
	c.volumeUnmountCmd.Flags().StringVar(&c.volumeName, "volumename", "", "volumename")
	c.volumePathCmd.Flags().StringVar(&c.volumeID, "volumeid", "", "volumeid")
	c.volumePathCmd.Flags().StringVar(&c.volumeName, "volumename", "", "volumename")

	c.addOutputFormatFlag(c.volumeCmd.Flags())
	c.addOutputFormatFlag(c.volumeGetCmd.Flags())
	c.addOutputFormatFlag(c.volumeCreateCmd.Flags())
	c.addOutputFormatFlag(c.volumeAttachCmd.Flags())
	c.addOutputFormatFlag(c.volumeMountCmd.Flags())
	c.addOutputFormatFlag(c.volumePathCmd.Flags())
	c.addOutputFormatFlag(c.volumeMapCmd.Flags())
}
