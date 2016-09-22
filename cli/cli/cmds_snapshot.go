package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (c *CLI) initSnapshotCmdsAndFlags() {
	c.initSnapshotCmds()
	c.initSnapshotFlags()
}

func (c *CLI) initSnapshotCmds() {

	c.snapshotCmd = &cobra.Command{
		Use:              "snapshot",
		Short:            "The snapshot manager",
		PersistentPreRun: c.preRunActivateLibStorage,
		Run: func(cmd *cobra.Command, args []string) {
			if isHelpFlags(cmd) {
				cmd.Usage()
			} else {
				c.snapshotGetCmd.Run(c.snapshotGetCmd, args)
			}
		},
	}
	c.c.AddCommand(c.snapshotCmd)

	c.snapshotGetCmd = &cobra.Command{
		Use:     "get",
		Short:   "Get one or more snapshots",
		Aliases: []string{"ls", "list"},
		Run: func(cmd *cobra.Command, args []string) {

			allSnapshots, err := c.r.Storage().Snapshots(c.ctx, store())
			if err != nil {
				log.Fatal(err)
			}

			if len(allSnapshots) > 0 {
				out, err := c.marshalOutput(&allSnapshots)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(out)
			}
		},
	}
	c.snapshotCmd.AddCommand(c.snapshotGetCmd)

	c.snapshotCreateCmd = &cobra.Command{
		Use:     "create",
		Short:   "Create a new snapshot",
		Aliases: []string{"new"},
		Run: func(cmd *cobra.Command, args []string) {

			if c.volumeID == "" {
				log.Fatalf("missing --volumeid")
			}

			snapshot, err := c.r.Storage().VolumeSnapshot(
				c.ctx, c.volumeID, c.snapshotName, store())
			if err != nil {
				log.Fatal(err)
			}

			out, err := c.marshalOutput(&snapshot)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(out)

		},
	}
	c.snapshotCmd.AddCommand(c.snapshotCreateCmd)

	c.snapshotRemoveCmd = &cobra.Command{
		Use:     "remove",
		Short:   "Remove a snapshot",
		Aliases: []string{"rm"},
		Run: func(cmd *cobra.Command, args []string) {

			if c.snapshotID == "" {
				log.Fatalf("missing --snapshotid")
			}

			err := c.r.Storage().SnapshotRemove(c.ctx, c.snapshotID, store())
			if err != nil {
				log.Fatal(err)
			}

		},
	}
	c.snapshotCmd.AddCommand(c.snapshotRemoveCmd)

	c.snapshotCopyCmd = &cobra.Command{
		Use:   "copy",
		Short: "Copies a snapshot",
		Run: func(cmd *cobra.Command, args []string) {

			if c.snapshotID == "" && c.volumeID == "" && c.volumeName == "" {
				log.Fatalf("missing --volumeid or --snapshotid or --volumename")
			}

			snapshot, err := c.r.Storage().SnapshotCopy(
				c.ctx, c.snapshotID, c.snapshotName,
				c.destinationRegion, store())
			if err != nil {
				log.Fatal(err)
			}

			out, err := c.marshalOutput(&snapshot)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(out)
		},
	}
	c.snapshotCmd.AddCommand(c.snapshotCopyCmd)
}

func (c *CLI) initSnapshotFlags() {
	c.snapshotGetCmd.Flags().StringVar(&c.snapshotName, "snapshotname", "", "snapshotname")
	c.snapshotGetCmd.Flags().StringVar(&c.volumeID, "volumeid", "", "volumeid")
	c.snapshotGetCmd.Flags().StringVar(&c.snapshotID, "snapshotid", "", "snapshotid")
	c.snapshotCreateCmd.Flags().BoolVar(&c.runAsync, "runasync", false, "runasync")
	c.snapshotCreateCmd.Flags().StringVar(&c.snapshotName, "snapshotname", "", "snapshotname")
	c.snapshotCreateCmd.Flags().StringVar(&c.volumeID, "volumeid", "", "volumeid")
	c.snapshotCreateCmd.Flags().StringVar(&c.description, "description", "", "description")
	c.snapshotRemoveCmd.Flags().StringVar(&c.snapshotID, "snapshotid", "", "snapshotid")
	c.snapshotCopyCmd.Flags().BoolVar(&c.runAsync, "runasync", false, "runasync")
	c.snapshotCopyCmd.Flags().StringVar(&c.volumeID, "volumeid", "", "volumeid")
	c.snapshotCopyCmd.Flags().StringVar(&c.snapshotID, "snapshotid", "", "snapshotid")
	c.snapshotCopyCmd.Flags().StringVar(&c.snapshotName, "snapshotname", "", "snapshotname")
	c.snapshotCopyCmd.Flags().StringVar(&c.destinationSnapshotName, "destinationsnapshotname", "", "destinationsnapshotname")
	c.snapshotCopyCmd.Flags().StringVar(&c.destinationRegion, "destinationregion", "", "destinationregion")

	c.addOutputFormatFlag(c.snapshotCmd.Flags())
	c.addOutputFormatFlag(c.snapshotGetCmd.Flags())
	c.addOutputFormatFlag(c.snapshotCopyCmd.Flags())
	c.addOutputFormatFlag(c.snapshotCreateCmd.Flags())
}
