package cli

import (
	"regexp"
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
		Use:              "volume",
		Short:            "The volume manager",
		PersistentPreRun: c.preRunActivateLibStorage,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
	c.c.AddCommand(c.volumeCmd)

	c.volumeListCmd = &cobra.Command{
		Use:     "ls",
		Short:   "List volumes",
		Aliases: []string{"list", "get", "inspect"},
		Example: "rexray volume ls [OPTIONS] VOLUME [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			c.mustMarshalOutput3(c.lsVolumes(args))
		},
	}
	c.volumeCmd.AddCommand(c.volumeListCmd)

	c.volumeCreateCmd = &cobra.Command{
		Use:     "create",
		Short:   "Create a new volume",
		Aliases: []string{"new"},
		Run: func(cmd *cobra.Command, args []string) {
			opts := &apitypes.VolumeCreateOpts{
				AvailabilityZone: &c.availabilityZone,
				Size:             &c.size,
				Type:             &c.volumeType,
				IOPS:             &c.iops,
				Opts:             store(),
			}
			if c.volumeID != "" && c.volumeName != "" {
				c.mustMarshalOutput(c.r.Storage().VolumeCopy(
					c.ctx, c.volumeID, c.volumeName, opts.Opts))
			} else if c.snapshotID != "" && c.volumeName != "" {
				c.mustMarshalOutput(c.r.Storage().VolumeCreateFromSnapshot(
					c.ctx, c.snapshotID, c.volumeName, opts))
			} else {
				c.mustMarshalOutput(c.r.Storage().VolumeCreate(
					c.ctx, c.volumeName, opts))
			}
		},
	}
	c.volumeCmd.AddCommand(c.volumeCreateCmd)

	c.volumeRemoveCmd = &cobra.Command{
		Use:     "rm",
		Short:   "Remove a volume",
		Example: "rexray volume rm [OPTIONS] VOLUME [VOLUME...]",
		Aliases: []string{"remove"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatal("no volumes specified")
			}
			vols, uniq, err := c.lsVolumes(args)
			if err != nil {
				log.Fatal(err)
			}
			if !uniq {
				log.Fatal("unique match not found")
			}
			opts := store()
			results := []*apitypes.Volume{}
			for _, v := range vols {
				if c.dryRun {
					results = append(results, v)
					continue
				}
				err := c.r.Storage().VolumeRemove(c.ctx, v.ID, opts)
				if err != nil {
					log.WithFields(log.Fields{
						"id":   v.ID,
						"name": v.Name,
					}).WithError(err).Fatal("error removing volume")
				}
			}
			if c.dryRun {
				c.mustMarshalOutput(results, nil)
			}
		},
	}
	c.volumeCmd.AddCommand(c.volumeRemoveCmd)

	c.volumeAttachCmd = &cobra.Command{
		Use:     "attach",
		Short:   "Attach a volume",
		Example: "rexray volume attach [OPTIONS] VOLUME [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatal("no volumes specified")
			}
			vols, uniq, err := c.lsVolumes(args)
			if err != nil {
				log.Fatal(err)
			}
			if !uniq {
				log.Fatal("unique match not found")
			}
			opts := &apitypes.VolumeAttachOpts{
				Force: c.force,
				Opts:  store(),
			}
			results := []*apitypes.Volume{}
			for _, v := range vols {
				if c.dryRun {
					results = append(results, v)
					continue
				}
				nv, _, err := c.r.Storage().VolumeAttach(c.ctx, v.ID, opts)
				if err != nil {
					log.WithFields(log.Fields{
						"id":   v.ID,
						"name": v.Name,
					}).WithError(err).Fatal("error attaching volume")
				}
				results = append(results, nv)
			}
			c.mustMarshalOutput(results, nil)
		},
	}
	c.volumeCmd.AddCommand(c.volumeAttachCmd)

	c.volumeDetachCmd = &cobra.Command{
		Use:     "detach",
		Short:   "Detach a volume",
		Example: "rexray volume detach [OPTIONS] VOLUME [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatal("no volumes specified")
			}
			vols, uniq, err := c.lsVolumes(args)
			if err != nil {
				log.Fatal(err)
			}
			if !uniq {
				log.Fatal("unique match not found")
			}
			opts := &apitypes.VolumeDetachOpts{
				Force: c.force,
				Opts:  store(),
			}
			results := []*apitypes.Volume{}
			for _, v := range vols {
				if c.dryRun {
					results = append(results, v)
					continue
				}
				_, err := c.r.Storage().VolumeDetach(c.ctx, v.ID, opts)
				if err != nil {
					log.WithFields(log.Fields{
						"id":   v.ID,
						"name": v.Name,
					}).WithError(err).Fatal("error detaching volume")
				}
			}
			if c.dryRun {
				c.mustMarshalOutput(results, nil)
			}
		},
	}
	c.volumeCmd.AddCommand(c.volumeDetachCmd)

	c.volumeMountCmd = &cobra.Command{
		Use:     "mount",
		Short:   "Mount a volume",
		Example: "rexray volume mount [OPTIONS] VOLUME [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatal("no volumes specified")
			}
			vols, uniq, err := c.lsVolumes(args)
			if err != nil {
				log.Fatal(err)
			}
			if !uniq {
				log.Fatal("unique match not found")
			}
			opts := &apitypes.VolumeMountOpts{
				NewFSType:   c.fsType,
				OverwriteFS: c.overwriteFs,
			}
			results := []*volumeWithPath{}
			iid, err := c.r.Executor().InstanceID(c.ctx, voluemStatusStore)
			if err != nil {
				log.Fatal(err)
			}
			withAttachments := []*apitypes.VolumeAttachment{
				&apitypes.VolumeAttachment{InstanceID: iid},
			}
			for _, v := range vols {
				if c.dryRun {
					v.Attachments = withAttachments
					results = append(results, &volumeWithPath{v, ""})
					continue
				}
				p, _, err := c.r.Integration().Mount(c.ctx, v.ID, "", opts)
				if err != nil {
					log.WithFields(log.Fields{
						"id":   v.ID,
						"name": v.Name,
					}).WithError(err).Fatal("error mounting volume")
				}
				nv := &volumeWithPath{v, p}
				nv.Attachments = withAttachments
				results = append(results, nv)
			}
			c.mustMarshalOutput(results, nil)
		},
	}
	c.volumeCmd.AddCommand(c.volumeMountCmd)

	c.volumeUnmountCmd = &cobra.Command{
		Use:     "unmount",
		Short:   "Unmount a volume",
		Example: "rexray volume unmount [OPTIONS] VOLUME [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatal("no volumes specified")
			}
			vols, uniq, err := c.lsVolumes(args)
			if err != nil {
				log.Fatal(err)
			}
			if !uniq {
				log.Fatal("unique match not found")
			}
			opts := store()
			results := []*apitypes.Volume{}
			noAttachments := []*apitypes.VolumeAttachment{}
			for _, v := range vols {
				if c.dryRun {
					v.Attachments = noAttachments
					results = append(results, v)
					continue
				}
				err := c.r.Integration().Unmount(c.ctx, v.ID, "", opts)
				if err != nil {
					log.WithFields(log.Fields{
						"id":   v.ID,
						"name": v.Name,
					}).WithError(err).Fatal("error unmounting volume")
				}
			}
			if c.dryRun {
				c.mustMarshalOutput(results, nil)
			}
		},
	}
	c.volumeCmd.AddCommand(c.volumeUnmountCmd)

	c.volumePathCmd = &cobra.Command{
		Use:     "path",
		Short:   "Print the volume path",
		Example: "rexray volume path [OPTIONS] VOLUME [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				args = []string{"/.*/"}
			}
			vols, _, err := c.lsVolumes(args)
			if err != nil {
				log.Fatal(err)
			}
			opts := store()
			results := []*volumeWithPath{}
			for _, v := range vols {
				if c.dryRun {
					results = append(results, &volumeWithPath{v, ""})
					continue
				}
				p, err := c.r.Integration().Path(c.ctx, v.ID, "", opts)
				if err != nil {
					log.WithFields(log.Fields{
						"id":   v.ID,
						"name": v.Name,
					}).WithError(err).Fatal("error getting volume path")
				}
				results = append(results, &volumeWithPath{v, p})
			}
			c.mustMarshalOutput(results, nil)
		},
	}
	c.volumeCmd.AddCommand(c.volumePathCmd)
}

type volumeWithPath struct {
	*apitypes.Volume
	Path string
}

func (c *CLI) lsVolumes(args []string) ([]*apitypes.Volume, bool, error) {
	opts := &apitypes.VolumesOpts{Attachments: c.volumeAttached}
	vols, err := c.r.Storage().Volumes(c.ctx, opts)
	if err != nil {
		return nil, false, err
	}
	if len(args) == 0 {
		return vols, false, nil
	}
	isRX := regexp.MustCompile(`^/(.*)/$`)
	patts := make([]interface{}, len(args))

	for i, a := range args {
		if m := isRX.FindStringSubmatch(a); len(m) > 0 {
			rx, err := regexp.Compile(m[1])
			if err != nil {
				return nil, false, err
			}
			patts[i] = rx
		} else {
			patts[i] = a
		}
	}
	uniqPrefixMatches := true
	prefixes := map[string]bool{}
	results := []*apitypes.Volume{}
	for _, vol := range vols {
		for _, p := range patts {
			switch tp := p.(type) {
			case string:
				if strings.HasPrefix(vol.ID, tp) ||
					strings.HasPrefix(vol.Name, tp) {
					results = append(results, vol)
					if uniqPrefixMatches {
						if _, exists := prefixes[tp]; !exists {
							prefixes[tp] = true
						} else {
							uniqPrefixMatches = false
						}
					}
				}
			case *regexp.Regexp:
				if tp.MatchString(vol.ID) || tp.MatchString(vol.Name) {
					results = append(results, vol)
				}
			}
		}
	}

	return results, uniqPrefixMatches, nil
}

func (c *CLI) initVolumeFlags() {
	c.volumeListCmd.Flags().BoolVar(&c.volumeAttached, "attached", false,
		"Set to \"true\" to obtain volume-device mappings for volumes "+
			"attached to this host")
	c.volumeCreateCmd.Flags().StringVar(&c.volumeName, "volumeName", "", "")
	c.volumeCreateCmd.Flags().StringVar(&c.volumeType, "volumeType", "", "")
	c.volumeCreateCmd.Flags().StringVar(&c.volumeID, "volumeID", "", "")
	c.volumeCreateCmd.Flags().StringVar(&c.snapshotID, "snapshotID", "", "")
	c.volumeCreateCmd.Flags().Int64Var(&c.iops, "iops", 0, "")
	c.volumeCreateCmd.Flags().Int64Var(&c.size, "size", 0, "")
	c.volumeCreateCmd.Flags().StringVar(
		&c.availabilityZone, "availabilityZone", "", "")
	c.volumeAttachCmd.Flags().BoolVar(&c.force, "force", false, "force")
	c.volumeDetachCmd.Flags().BoolVar(&c.force, "force", false, "")
	c.volumeMountCmd.Flags().BoolVar(&c.overwriteFs, "overwritefs", false, "")
	c.volumeMountCmd.Flags().StringVar(&c.fsType, "fsType", "", "")

	c.addDryRunFlag(c.volumeRemoveCmd.Flags())
	c.addDryRunFlag(c.volumeAttachCmd.Flags())
	c.addDryRunFlag(c.volumeDetachCmd.Flags())
	c.addDryRunFlag(c.volumeMountCmd.Flags())
	c.addDryRunFlag(c.volumeUnmountCmd.Flags())

	c.addOutputFormatFlag(c.volumeListCmd.Flags())
	c.addOutputFormatFlag(c.volumeCreateCmd.Flags())
	c.addOutputFormatFlag(c.volumeAttachCmd.Flags())
	c.addOutputFormatFlag(c.volumeMountCmd.Flags())
	c.addOutputFormatFlag(c.volumeUnmountCmd.Flags())
	c.addOutputFormatFlag(c.volumePathCmd.Flags())
}
