package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	apitypes "github.com/codedellemc/libstorage/api/types"
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
		Aliases: []string{"l", "list", "get", "inspect"},
		Short:   "List volumes",
		Example: "rexray volume ls [OPTIONS] [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			c.mustMarshalOutput(c.lsVolumes(args, c.defaultAttachments()))
		},
	}
	c.volumeCmd.AddCommand(c.volumeListCmd)

	c.volumeCreateCmd = &cobra.Command{
		Use:     "create",
		Aliases: []string{"c", "n", "new"},
		Short:   "Create a new volume",
		Example: "rexray volume create [OPTIONS] VOLUME [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			checkVolumeArgs(cmd, args)

			var (
				iid             *apitypes.InstanceID
				attachOpts      *apitypes.VolumeAttachOpts
				mountOpts       *apitypes.VolumeMountOpts
				withAttachments []*apitypes.VolumeAttachment
				processed       interface{}
				processedVols   []*apitypes.Volume
				processedPVols  []*volumeWithPath
				opts            = &apitypes.VolumeCreateOpts{
					AvailabilityZone: &c.availabilityZone,
					Size:             &c.size,
					Type:             &c.volumeType,
					IOPS:             &c.iops,
					Opts:             store(),
				}
			)

			if c.amount {
				processedPVols = []*volumeWithPath{}
			} else {
				processedVols = []*apitypes.Volume{}
			}

			if c.attach || c.amount {
				attachOpts = &apitypes.VolumeAttachOpts{
					Force: c.force,
					Opts:  opts.Opts,
				}
				if c.amount {
					mountOpts = &apitypes.VolumeMountOpts{
						NewFSType:   c.fsType,
						OverwriteFS: c.overwriteFs,
					}
					iid2, err := c.r.Executor().InstanceID(
						c.ctx, opts.Opts)
					if err != nil {
						log.Fatal(err)
					}
					iid = iid2
					withAttachments = []*apitypes.VolumeAttachment{
						&apitypes.VolumeAttachment{InstanceID: iid},
					}
				}
			}

			for _, name := range args {
				if c.dryRun {
					dv := &apitypes.Volume{
						Name:             name,
						Size:             c.size,
						Type:             c.volumeType,
						IOPS:             c.iops,
						AvailabilityZone: c.availabilityZone,
					}
					if c.attach || c.amount {
						dv.Attachments = withAttachments
					}
					if c.amount {
						processedPVols = append(
							processedPVols, &volumeWithPath{dv, ""})
					} else {
						processedVols = append(processedVols, dv)
					}
					continue
				}
				v, err := c.r.Storage().VolumeCreate(c.ctx, name, opts)
				if err != nil {
					c.logVolumeLoopError(
						processed,
						name,
						"error creating volume",
						err)
					continue
				}
				if c.attach || c.amount {
					nv, _, err := c.r.Storage().VolumeAttach(
						c.ctx, v.ID, attachOpts)
					if err != nil {
						c.logVolumeLoopError(
							processed,
							name,
							"error attaching volume",
							err)
						continue
					}
					if c.amount {
						p, _, err := c.r.Integration().Mount(
							c.ctx, nv.ID, "", mountOpts)
						if err != nil {
							c.logVolumeLoopError(
								processed,
								name,
								"error mounting volume",
								err)
							continue
						}
						processedPVols = append(
							processedPVols, &volumeWithPath{nv, p})
					} else {
						processedVols = append(processedVols, nv)
					}
				} else {
					processedVols = append(processedVols, v)
				}
				if c.amount {
					processed = processedPVols
				} else {
					processed = processedVols
				}
			}
			if c.dryRun {
				if c.amount {
					processed = processedPVols
				} else {
					processed = processedVols
				}
			}
			c.mustMarshalOutput(processed, nil)
		},
	}
	c.volumeCmd.AddCommand(c.volumeCreateCmd)

	c.volumeRemoveCmd = &cobra.Command{
		Use:     "rm",
		Short:   "Remove a volume",
		Example: "rexray volume rm [OPTIONS] VOLUME [VOLUME...]",
		Aliases: []string{"r", "remove", "delete", "del"},
		Run: func(cmd *cobra.Command, args []string) {
			checkVolumeArgs(cmd, args)

			result, err := c.lsVolumes(args, 0)
			if err != nil {
				log.Fatal(err)
			}

			if result.exactMatches() == 0 && result.regexMatches() == 0 {
				log.Fatal("no volumes found")
			}

			var (
				opts      = store()
				processed = []string{}
			)

			for _, v := range result.vols {
				// only remove exact matches or regexp matches. partial matches
				// should be ignored when removing volumes as they can result
				// in data loss
				if !(result.isExactMatch(v) || result.isRegexpMatch(v)) {
					continue
				}

				if c.dryRun {
					processed = append(processed, result.matchedIDOrName(v))
					continue
				}
				err := c.r.Storage().VolumeRemove(c.ctx, v.ID, opts)
				if err != nil {
					c.logVolumeLoopError(
						processed,
						result.matchedIDOrName(v),
						"error removing volume",
						err)
					continue
				}
				processed = append(processed, result.matchedIDOrName(v))
			}
			c.mustMarshalOutput(processed, nil)
		},
	}
	c.volumeCmd.AddCommand(c.volumeRemoveCmd)

	c.volumeAttachCmd = &cobra.Command{
		Use:     "attach",
		Aliases: []string{"a"},
		Short:   "Attach a volume",
		Example: "rexray volume attach [OPTIONS] VOLUME [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			checkVolumeArgs(cmd, args)

			result, err := c.lsVolumes(
				args,
				apitypes.VolAttReqOnlyUnattachedVols)
			if err != nil {
				log.Fatal(err)
			}

			if result.exactMatches() == 0 &&
				result.regexMatches() == 0 &&
				result.uniquePartialMatches() == 0 {
				log.Fatal("no volumes found")
			}

			var (
				processed = []*apitypes.Volume{}
				opts      = &apitypes.VolumeAttachOpts{
					Force: c.force,
					Opts:  store(),
				}
			)

			for _, v := range result.vols {
				// if a partial match it must be unique
				if pmt := result.isPartialMatch(v); pmt > 0 && !pmt.isUnique() {
					continue
				}
				if c.dryRun {
					processed = append(processed, v)
					continue
				}
				nv, _, err := c.r.Storage().VolumeAttach(c.ctx, v.ID, opts)
				if err != nil {
					c.logVolumeLoopError(
						processed,
						result.matchedIDOrName(v),
						"error attaching volume",
						err)
					continue
				}
				processed = append(processed, nv)
			}

			c.mustMarshalOutput(processed, nil)
		},
	}
	c.volumeCmd.AddCommand(c.volumeAttachCmd)

	c.volumeDetachCmd = &cobra.Command{
		Use:     "detach",
		Aliases: []string{"d"},
		Short:   "Detach a volume",
		Example: "rexray volume detach [OPTIONS] VOLUME [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			checkVolumeArgs(cmd, args)

			result, err := c.lsVolumes(
				args,
				apitypes.VolAttReqOnlyAttachedVols)
			if err != nil {
				log.Fatal(err)
			}

			if result.exactMatches() == 0 &&
				result.regexMatches() == 0 &&
				result.uniquePartialMatches() == 0 {
				log.Fatal("no volumes found")
			}

			var (
				processed = []*apitypes.Volume{}
				opts      = &apitypes.VolumeDetachOpts{
					Force: c.force,
					Opts:  store(),
				}
			)

			for _, v := range result.vols {
				// if a partial match it must be unique
				if pmt := result.isPartialMatch(v); pmt > 0 && !pmt.isUnique() {
					continue
				}
				if c.dryRun {
					v.Attachments = nil
					processed = append(processed, v)
					continue
				}
				_, err := c.r.Storage().VolumeDetach(c.ctx, v.ID, opts)
				if err != nil {
					c.logVolumeLoopError(
						processed,
						result.matchedIDOrName(v),
						"error detaching volume",
						err)
					continue
				}
				v.Attachments = nil
				processed = append(processed, v)
			}
			c.mustMarshalOutput(processed, nil)
		},
	}
	c.volumeCmd.AddCommand(c.volumeDetachCmd)

	c.volumeMountCmd = &cobra.Command{
		Use:     "mount",
		Aliases: []string{"m"},
		Short:   "Mount a volume",
		Example: "rexray volume mount [OPTIONS] VOLUME [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			checkVolumeArgs(cmd, args)

			// volumes that are attached or can be attached should be valid
			// candidates for a mount operation
			result, err := c.lsVolumes(
				args,
				apitypes.VolAttReqOnlyVolsAttachedToInstanceOrUnattachedVols)
			if err != nil {
				log.Fatal(err)
			}

			if result.exactMatches() == 0 &&
				result.regexMatches() == 0 &&
				result.uniquePartialMatches() == 0 {
				log.Fatal("no volumes found")
			}

			var (
				processed = []*volumeWithPath{}
				opts      = &apitypes.VolumeMountOpts{
					NewFSType:   c.fsType,
					OverwriteFS: c.overwriteFs,
				}
				withAttachments = []*apitypes.VolumeAttachment{
					&apitypes.VolumeAttachment{InstanceID: result.iid},
				}
			)

			for _, v := range result.vols {
				// if a partial match it must be unique
				if pmt := result.isPartialMatch(v); pmt > 0 && !pmt.isUnique() {
					continue
				}
				if c.dryRun {
					processed = append(processed, &volumeWithPath{v, ""})
					continue
				}
				p, _, err := c.r.Integration().Mount(c.ctx, v.ID, "", opts)
				if err != nil {
					c.logVolumeLoopError(
						processed,
						result.matchedIDOrName(v),
						"error mounting volume",
						err)
					continue
				}
				nv := &volumeWithPath{v, p}
				nv.Attachments = withAttachments
				processed = append(processed, nv)
			}
			c.mustMarshalOutput(processed, nil)
		},
	}
	c.volumeCmd.AddCommand(c.volumeMountCmd)

	c.volumeUnmountCmd = &cobra.Command{
		Use:     "unmount",
		Short:   "Unmount a volume",
		Aliases: []string{"u", "umount"},
		Example: "rexray volume unmount [OPTIONS] VOLUME [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			checkVolumeArgs(cmd, args)

			result, err := c.lsVolumes(
				args,
				apitypes.VolAttReqOnlyVolsAttachedToInstance)
			if err != nil {
				log.Fatal(err)
			}

			if result.exactMatches() == 0 &&
				result.regexMatches() == 0 &&
				result.uniquePartialMatches() == 0 {
				log.Fatal("no volumes found")
			}

			var (
				opts      = store()
				processed = []*apitypes.Volume{}
			)

			for _, v := range result.vols {
				// if a partial match it must be unique
				if pmt := result.isPartialMatch(v); pmt > 0 && !pmt.isUnique() {
					continue
				}
				if c.dryRun {
					v.Attachments = nil
					processed = append(processed, v)
					continue
				}
				err := c.r.Integration().Unmount(c.ctx, v.ID, "", opts)
				if err != nil {
					c.logVolumeLoopError(
						processed,
						result.matchedIDOrName(v),
						"error unmounting volume",
						err)
					continue
				}
				v.Attachments = nil
				processed = append(processed, v)
			}
			c.mustMarshalOutput(processed, nil)
		},
	}
	c.volumeCmd.AddCommand(c.volumeUnmountCmd)

	c.volumePathCmd = &cobra.Command{
		Use:     "path",
		Aliases: []string{"p"},
		Short:   "Print the volume path",
		Example: "rexray volume path [OPTIONS] [VOLUME...]",
		Run: func(cmd *cobra.Command, args []string) {
			result, err := c.lsVolumes(
				args,
				apitypes.VolAttReqOnlyVolsAttachedToInstance)
			if err != nil {
				log.Fatal(err)
			}

			var (
				opts      = store()
				processed = []*volumeWithPath{}
			)

			for _, v := range result.vols {
				if c.dryRun {
					processed = append(processed, &volumeWithPath{v, ""})
					continue
				}
				p, err := c.r.Integration().Path(c.ctx, v.ID, "", opts)
				if err != nil {
					c.logVolumeLoopError(
						processed,
						result.matchedIDOrName(v),
						"error getting volume path",
						err)
					continue
				}
				processed = append(processed, &volumeWithPath{v, p})
			}
			c.mustMarshalOutput(processed, nil)
		},
	}
	c.volumeCmd.AddCommand(c.volumePathCmd)
}

func checkVolumeArgs(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		return
	}
	fmt.Fprintln(os.Stderr, cmd.UsageString())
	os.Exit(1)
}

func (c *CLI) logVolumeLoopError(
	processed interface{}, name, msg string, err error) {
	logEntry := log.WithField("volume", name).WithError(err)
	if c.continueOnError {
		logEntry.Error(msg)
	} else {
		c.mustMarshalOutput(processed, nil)
		logEntry.Fatal(msg)
	}
}

type volumeWithPath struct {
	*apitypes.Volume
	Path string
}

type matchTypes int

const (
	matchTypeGlob matchTypes = iota
	matchTypeExactID
	matchTypeExactIDIgnoreCase
	matchTypePartialID
	matchTypePartialIDIgnoreCase
	matchTypeExactName
	matchTypeExactNameIgnoreCase
	matchTypePartialName
	matchTypePartialNameIgnoreCase
	matchTypeRegexpID
	matchTypeRegexpIDIgnoreCase
	matchTypeRegexpName
	matchTypeRegexpNameIgnoreCase
)

type matchedVolume struct {
	*apitypes.Volume
	matchType matchTypes
}

type regexpPair struct {
	*regexp.Regexp
	ignoreCase *regexp.Regexp
}

type lsVolumesResult struct {
	vols             []*apitypes.Volume
	iid              *apitypes.InstanceID
	volMatchType     map[*apitypes.Volume]matchTypes
	volMatchPatt     map[*apitypes.Volume]string
	matchTypeCount   map[matchTypes]int
	matchPattCount   map[string]int
	uniqStrMatchID   map[string]bool
	uniqStrMatchName map[string]bool
	volMatchVals     map[*apitypes.Volume]string
}

func (r *lsVolumesResult) matchedIDOrName(v *apitypes.Volume) string {
	return r.volMatchVals[v]
}

func (r *lsVolumesResult) add(v *apitypes.Volume, mt matchTypes, mp string) {

	r.vols = append(r.vols, v)
	r.volMatchType[v] = mt
	r.volMatchPatt[v] = mp

	switch mt {
	case matchTypeExactID, matchTypeExactIDIgnoreCase,
		matchTypePartialID, matchTypePartialIDIgnoreCase,
		matchTypeRegexpID, matchTypeRegexpIDIgnoreCase:
		if v.ID == "" {
			r.volMatchVals[v] = mp
		} else {
			r.volMatchVals[v] = v.ID
		}
	case matchTypeExactName, matchTypeExactNameIgnoreCase,
		matchTypePartialName, matchTypePartialNameIgnoreCase,
		matchTypeRegexpName, matchTypeRegexpNameIgnoreCase:
		if v.Name == "" {
			if v.ID == "" {
				r.volMatchVals[v] = mp
			} else {
				r.volMatchVals[v] = v.ID
			}
		} else {
			r.volMatchVals[v] = v.Name
		}
	}

	switch mt {
	case matchTypeExactID, matchTypeExactIDIgnoreCase,
		matchTypePartialID, matchTypePartialIDIgnoreCase:
		if _, ok := r.uniqStrMatchID[mp]; !ok {
			r.uniqStrMatchID[mp] = true
		} else {
			r.uniqStrMatchID[mp] = false
		}
	case matchTypeExactName, matchTypeExactNameIgnoreCase,
		matchTypePartialName, matchTypePartialNameIgnoreCase:
		if _, ok := r.uniqStrMatchName[mp]; !ok {
			r.uniqStrMatchName[mp] = true
		} else {
			r.uniqStrMatchName[mp] = false
		}
	}

	if i, ok := r.matchTypeCount[mt]; ok {
		r.matchTypeCount[mt] = i + 1
	} else {
		r.matchTypeCount[mt] = 1
	}

	if i, ok := r.matchPattCount[mp]; ok {
		r.matchPattCount[mp] = i + 1
	} else {
		r.matchPattCount[mp] = 1
	}
}

func (r *lsVolumesResult) isExactMatch(v *apitypes.Volume) bool {
	return r.isExactIDMatch(v) || r.isExactNameMatch(v)
}

func (r *lsVolumesResult) isExactIDMatch(v *apitypes.Volume) bool {
	if t, ok := r.volMatchType[v]; ok && t == matchTypeExactID {
		return true
	}
	if t, ok := r.volMatchType[v]; ok && t == matchTypeExactIDIgnoreCase {
		return true
	}
	return false
}

func (r *lsVolumesResult) isExactNameMatch(v *apitypes.Volume) bool {
	if t, ok := r.volMatchType[v]; ok && t == matchTypeExactName {
		return true
	}
	if t, ok := r.volMatchType[v]; ok && t == matchTypeExactNameIgnoreCase {
		return true
	}
	return false
}

type partialMatchType int

const (
	_                               = iota
	partialMatchID partialMatchType = 1 << iota
	partialMatchName
	partialMatchUniqueID
	partialMatchUniqueName
)

func (p partialMatchType) isUnique() bool {
	return p&partialMatchUniqueID == partialMatchUniqueID ||
		p&partialMatchUniqueName == partialMatchUniqueName
}

func (r *lsVolumesResult) isPartialMatch(v *apitypes.Volume) partialMatchType {
	return r.isPartialIDMatch(v) | r.isPartialNameMatch(v)
}

func (r *lsVolumesResult) isPartialIDMatch(
	v *apitypes.Volume) partialMatchType {

	if t, ok := r.volMatchType[v]; ok && t == matchTypePartialID {
		pm := partialMatchID
		if mp, ok := r.volMatchPatt[v]; ok {
			if uniq, ok := r.uniqStrMatchID[mp]; ok && uniq {
				pm = pm | partialMatchUniqueID
			}
		}
		return pm
	}

	if t, ok := r.volMatchType[v]; ok && t == matchTypePartialIDIgnoreCase {
		pm := partialMatchID
		if mp, ok := r.volMatchPatt[v]; ok {
			if uniq, ok := r.uniqStrMatchID[mp]; ok && uniq {
				pm = pm | partialMatchUniqueID
			}
		}
		return pm
	}

	return 0
}

func (r *lsVolumesResult) isPartialNameMatch(
	v *apitypes.Volume) partialMatchType {

	if t, ok := r.volMatchType[v]; ok && t == matchTypePartialName {
		pm := partialMatchName
		if mp, ok := r.volMatchPatt[v]; ok {
			if uniq, ok := r.uniqStrMatchName[mp]; ok && uniq {
				pm = pm | partialMatchUniqueName
			}
		}
		return pm
	}

	if t, ok := r.volMatchType[v]; ok && t == matchTypePartialNameIgnoreCase {
		pm := partialMatchName
		if mp, ok := r.volMatchPatt[v]; ok {
			if uniq, ok := r.uniqStrMatchName[mp]; ok && uniq {
				pm = pm | partialMatchUniqueName
			}
		}
		return pm
	}

	return 0
}

func (r *lsVolumesResult) isRegexpMatch(v *apitypes.Volume) bool {
	if t, ok := r.volMatchType[v]; ok && t >= matchTypeRegexpID {
		return true
	}
	return false
}

func (r *lsVolumesResult) exactMatches() int {
	return r.exactIDMatches() + r.exactNameMatches()
}

func (r *lsVolumesResult) exactIDMatches() int {
	total := 0
	if i, ok := r.matchTypeCount[matchTypeExactID]; ok {
		total = total + i
	}
	if i, ok := r.matchTypeCount[matchTypeExactIDIgnoreCase]; ok {
		total = total + i
	}
	return total
}

func (r *lsVolumesResult) exactNameMatches() int {
	total := 0
	if i, ok := r.matchTypeCount[matchTypeExactName]; ok {
		total = total + i
	}
	if i, ok := r.matchTypeCount[matchTypeExactNameIgnoreCase]; ok {
		total = total + i
	}
	return total
}

func (r *lsVolumesResult) partialMatches() int {
	return r.partialIDMatches() + r.partialNameMatches()
}

func (r *lsVolumesResult) partialIDMatches() int {
	total := 0
	if i, ok := r.matchTypeCount[matchTypePartialID]; ok {
		total = total + i
	}
	if i, ok := r.matchTypeCount[matchTypePartialIDIgnoreCase]; ok {
		total = total + i
	}
	return total
}

func (r *lsVolumesResult) partialNameMatches() int {
	total := 0
	if i, ok := r.matchTypeCount[matchTypePartialName]; ok {
		total = total + i
	}
	if i, ok := r.matchTypeCount[matchTypePartialNameIgnoreCase]; ok {
		total = total + i
	}
	return total
}

func (r *lsVolumesResult) uniquePartialMatches() int {
	return r.uniquePartialIDMatches() + r.uniquePartialNameMatches()
}

func (r *lsVolumesResult) uniquePartialIDMatches() int {
	total := 0
	for _, v := range r.uniqStrMatchID {
		if v {
			total = total + 1
		}
	}
	return total
}

func (r *lsVolumesResult) uniquePartialNameMatches() int {
	total := 0
	for _, v := range r.uniqStrMatchName {
		if v {
			total = total + 1
		}
	}
	return total
}

func (r *lsVolumesResult) regexMatches() int {
	total := 0
	if i, ok := r.matchTypeCount[matchTypeRegexpID]; ok {
		total = total + i
	}
	if i, ok := r.matchTypeCount[matchTypeRegexpName]; ok {
		total = total + i
	}
	if i, ok := r.matchTypeCount[matchTypeRegexpIDIgnoreCase]; ok {
		total = total + i
	}
	if i, ok := r.matchTypeCount[matchTypeRegexpNameIgnoreCase]; ok {
		total = total + i
	}
	return total
}

func (c *CLI) defaultAttachments() apitypes.VolumeAttachmentsTypes {
	if c.volumeUnattached {
		return apitypes.VolAttReqOnlyUnattachedVols
	}
	if c.volumeAttachedToMe {
		return apitypes.VolAttReqWithDevMapOnlyVolsAttachedToInstance
	}
	if c.volumeAttached {
		return apitypes.VolAttReqOnlyAttachedVols |
			apitypes.VolumeAttachmentsDevices
	}
	return apitypes.VolAttReq |
		apitypes.VolumeAttachmentsDevices
}

func (c *CLI) lsVolumes(
	args []string,
	attachments apitypes.VolumeAttachmentsTypes) (*lsVolumesResult, error) {

	//opts :=
	vols, err := c.r.Storage().Volumes(c.ctx,
		&apitypes.VolumesOpts{Attachments: attachments})
	if err != nil {
		return nil, err
	}

	result := &lsVolumesResult{}

	if len(args) == 0 {
		result.vols = vols
		return result, nil
	}

	var (
		isRX  = regexp.MustCompile(`^/(.*)/$`)
		patts = make([]interface{}, len(args))
	)

	for i, a := range args {
		if m := isRX.FindStringSubmatch(a); len(m) > 0 {
			rx, err := regexp.Compile(m[1])
			if err != nil {
				return nil, err
			}
			rxp := &regexpPair{rx, nil}

			rx, err = regexp.Compile(fmt.Sprintf(`(?i)%s`, m[1]))
			if err != nil {
				return nil, err
			}
			rxp.ignoreCase = rx

			patts[i] = rxp
		} else {
			patts[i] = a
		}
	}

	result.matchPattCount = map[string]int{}
	result.matchTypeCount = map[matchTypes]int{}
	result.volMatchPatt = map[*apitypes.Volume]string{}
	result.volMatchType = map[*apitypes.Volume]matchTypes{}
	result.uniqStrMatchID = map[string]bool{}
	result.uniqStrMatchName = map[string]bool{}
	result.volMatchVals = map[*apitypes.Volume]string{}

NextVol:
	for _, vol := range vols {
		for _, p := range patts {
			switch tp := p.(type) {
			case string:
				if mt, mp, ok := stringMatchVolIDOrName(
					vol, tp, vol.ID, matchTypeExactID); ok {
					result.add(vol, mt, mp)
					continue NextVol
				}
				if mt, mp, ok := stringMatchVolIDOrName(
					vol, tp, vol.Name, matchTypeExactName); ok {
					result.add(vol, mt, mp)
					continue NextVol
				}
			case *regexpPair:
				if mt, mp, ok := regexpMatchVolIDOrName(
					vol, tp, vol.ID, matchTypeRegexpID); ok {
					result.add(vol, mt, mp)
					continue NextVol
				}
				if mt, mp, ok := regexpMatchVolIDOrName(
					vol, tp, vol.Name, matchTypeRegexpName); ok {
					result.add(vol, mt, mp)
					continue NextVol
				}
			}
		}
	}

	return result, nil
}

func stringMatchVolIDOrName(
	vol *apitypes.Volume,
	toMatch, toBeMatched string,
	firstMatchType matchTypes) (matchTypes, string, bool) {

	// matchTypeExactXYZ
	if toMatch == toBeMatched {
		return firstMatchType, toMatch, true
	}

	// matchTypeExactXYZIgnoreCase
	if strings.EqualFold(toMatch, toBeMatched) {
		return firstMatchType + 1, toMatch, true
	}

	// matchTypePartialXYZ
	if strings.HasPrefix(toBeMatched, toMatch) {
		return firstMatchType + 2, toMatch, true
	}

	// matchTypePartialXYZIgnoreCase
	if strings.HasPrefix(
		strings.ToLower(toBeMatched), strings.ToLower(toMatch)) {
		return firstMatchType + 3, toMatch, true
	}

	return 0, "", false
}

func regexpMatchVolIDOrName(
	vol *apitypes.Volume,
	toMatch *regexpPair, toBeMatched string,
	firstMatchType matchTypes) (matchTypes, string, bool) {

	// matchTypeRegexpXYZ
	if toMatch.MatchString(toBeMatched) {
		return firstMatchType, toMatch.String(), true
	}

	// matchTypeRegexpXYZIgnoreCase
	if toMatch.ignoreCase.MatchString(toBeMatched) {
		return firstMatchType + 1, toMatch.ignoreCase.String(), true
	}

	return 0, "", false
}

const (
	volumeStatusAttached    = "attached"
	volumeStatusAvailable   = "available"
	volumeStatusUnavailable = "unavailable"
	volumeStatusUnknown     = "unknown"
	volumeStatusError       = "error"
)

func (c *CLI) initVolumeFlags() {
	c.volumeListCmd.Flags().BoolVar(&c.volumeAttached, "attached", false,
		"A flag that indicates only attached volumes should be returned")
	c.volumeListCmd.Flags().BoolVar(&c.volumeAttachedToMe, "attachedToMe",
		false,
		"A flag that indicates only volumes attached to this host should "+
			"be returned")
	c.volumeListCmd.Flags().BoolVar(&c.volumeUnattached, "available",
		false,
		"A flag that indicates only available volumes should be returned")

	c.volumeAttachCmd.Flags().BoolVar(&c.force, "force", false, "")
	c.volumeDetachCmd.Flags().BoolVar(&c.force, "force", false, "")
	c.volumeMountCmd.Flags().BoolVar(&c.overwriteFs, "overwriteFS", false, "")
	c.volumeMountCmd.Flags().StringVar(&c.fsType, "fsType", "", "")

	c.volumeCreateCmd.Flags().StringVar(&c.volumeType, "volumeType", "", "")
	c.volumeCreateCmd.Flags().StringVar(&c.snapshotID, "snapshotID", "", "")
	c.volumeCreateCmd.Flags().Int64Var(&c.iops, "iops", 0, "")
	c.volumeCreateCmd.Flags().Int64Var(&c.size, "size", 0, "")
	c.volumeCreateCmd.Flags().StringVar(
		&c.availabilityZone, "availabilityZone", "", "")
	c.volumeCreateCmd.Flags().BoolVar(&c.attach, "attach", false,
		"Attach the new volume")
	c.volumeCreateCmd.Flags().BoolVar(&c.amount, "amount", false,
		"Attach and mount the new volume")
	c.volumeCreateCmd.Flags().BoolVar(&c.force, "force", false, "")
	c.volumeCreateCmd.Flags().BoolVar(&c.overwriteFs, "overwriteFS", false, "")
	c.volumeCreateCmd.Flags().StringVar(&c.fsType, "fsType", "", "")

	c.addQuietFlag(c.volumeCmd.PersistentFlags())
	c.addOutputFormatFlag(c.volumeCmd.PersistentFlags())
	c.addDryRunFlag(c.volumeCmd.PersistentFlags())

	c.addContinueOnErrorFlag(c.volumeCreateCmd.Flags())
	c.addContinueOnErrorFlag(c.volumeRemoveCmd.Flags())
	c.addContinueOnErrorFlag(c.volumeAttachCmd.Flags())
	c.addContinueOnErrorFlag(c.volumeDetachCmd.Flags())
	c.addContinueOnErrorFlag(c.volumeMountCmd.Flags())
	c.addContinueOnErrorFlag(c.volumeUnmountCmd.Flags())
	c.addContinueOnErrorFlag(c.volumePathCmd.Flags())
}
