package storage

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/digitalocean/godo"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	apiUtils "github.com/AVENTER-UG/rexray/libstorage/api/utils"

	do "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/dobs"
)

const (
	minSizeGiB = 1
)

type driver struct {
	name           string
	config         gofig.Config
	client         *godo.Client
	maxAttempts    int
	statusDelay    int64
	statusTimeout  time.Duration
	defaultRegion  *string
	convUnderscore bool
}

func init() {
	registry.RegisterStorageDriver(do.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{name: do.Name}
}

func (d *driver) Name() string {
	return do.Name
}

func (d *driver) Init(
	ctx types.Context,
	config gofig.Config) error {

	d.config = config
	token := d.config.GetString(do.ConfigToken)
	d.maxAttempts = d.config.GetInt(do.ConfigStatusMaxAttempts)

	statusDelayStr := d.config.GetString(do.ConfigStatusInitDelay)
	statusDelay, err := time.ParseDuration(statusDelayStr)
	if err != nil {
		return err
	}
	d.statusDelay = statusDelay.Nanoseconds()

	statusTimeoutStr := d.config.GetString(do.ConfigStatusTimeout)
	d.statusTimeout, err = time.ParseDuration(statusTimeoutStr)
	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"maxStatusAttempts": d.maxAttempts,
		"statusDelay": fmt.Sprintf(
			"%v", time.Duration(d.statusDelay)*time.Nanosecond),
		"statusTimeout": d.statusTimeout,
	}

	if region := d.config.GetString(do.ConfigRegion); region != "" {
		d.defaultRegion = &region
		fields["region"] = region
	}

	if token == "" {
		fields["token"] = ""
	} else {
		fields["token"] = "******"
	}

	client, err := Client(token)
	if err != nil {
		return err
	}
	d.client = client

	d.convUnderscore = d.config.GetBool(do.ConfigConvertUnderscores)

	ctx.WithFields(fields).Info("storage driver initialized")

	return nil
}

func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Block, nil
}

// DigitalOcean volumes are are found using device-by-id, ex:
// /dev/disk/by-id/scsi-0DO_Volume_volume-nyc1-01 See
// https://www.digitalocean.com/community/tutorials/how-to-use-block-storage-on-digitalocean#preparing-volumes-for-use-in-linux
func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {

	return nil, nil
}

func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	iid := context.MustInstanceID(ctx)
	return &types.Instance{
		InstanceID:   iid,
		Region:       iid.Fields[do.InstanceIDFieldRegion],
		Name:         iid.Fields[do.InstanceIDFieldName],
		ProviderName: iid.Driver,
	}, nil
}

func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	region := d.mustRegion(ctx)
	if region == nil || *region == "" {
		return nil, goof.New("No region provided or configured")
	}

	listOpts := &godo.ListVolumeParams{
		ListOptions: &godo.ListOptions{PerPage: 200},
		Region:      *region,
	}

	var volumes []*types.Volume
	for {
		doVolumes, resp, err := d.client.Storage.ListVolumes(ctx, listOpts)
		if err != nil {
			return nil, err
		}

		for _, vol := range doVolumes {
			volume, err := d.toTypesVolume(ctx, &vol, opts.Attachments)
			if err != nil {
				return nil, goof.New("error converting to types.Volume")
			}
			volumes = append(volumes, volume)
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		listOpts.ListOptions.Page = page + 1
	}

	return volumes, nil
}

func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	doVolume, _, err := d.client.Storage.GetVolume(ctx, volumeID)
	if err != nil {
		return nil, err
	}

	volume, err := d.toTypesVolume(ctx, doVolume, opts.Attachments)
	if err != nil {
		return nil, goof.New("error converting to types.Volume")
	}
	return volume, nil
}

func (d *driver) VolumeInspectByName(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {
	region := d.mustRegion(ctx)
	if region == nil || *region == "" {
		return nil, goof.New("No region provided or configured")
	}

	volumeName = d.convUnderscores(volumeName)

	doVolumes, _, err := d.client.Storage.ListVolumes(
		ctx,
		&godo.ListVolumeParams{
			Region: *region,
			Name:   volumeName,
		},
	)
	if err != nil {
		return nil, err
	}
	if len(doVolumes) == 0 {
		return nil, apiUtils.NewNotFoundError(volumeName)
	}
	if len(doVolumes) > 1 {
		return nil, goof.New("too many volumes returned")
	}

	volume, err := d.toTypesVolume(ctx, &doVolumes[0], opts.Attachments)
	if err != nil {
		return nil, goof.New("error converting to types.Volume")
	}
	return volume, nil
}

func (d *driver) VolumeCreate(
	ctx types.Context,
	name string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	name = d.convUnderscores(name)
	fields := map[string]interface{}{
		"volumeName": name,
	}

	if opts.AvailabilityZone == nil || *opts.AvailabilityZone == "" {
		opts.AvailabilityZone = d.mustRegion(ctx)
		if opts.AvailabilityZone == nil || *opts.AvailabilityZone == "" {
			return nil, goof.WithFields(fields,
				"No region for volume create")
		}
	}
	fields["region"] = *opts.AvailabilityZone

	if opts.Size == nil {
		size := int64(minSizeGiB)
		opts.Size = &size
	}

	fields["size"] = *opts.Size

	if *opts.Size < minSizeGiB {
		fields["minSize"] = minSizeGiB
		return nil, goof.WithFields(fields, "volume size too small")
	}

	volumeReq := &godo.VolumeCreateRequest{
		Region:        *opts.AvailabilityZone,
		Name:          name,
		SizeGigaBytes: *opts.Size,
	}

	volume, _, err := d.client.Storage.CreateVolume(ctx, volumeReq)
	if err != nil {
		ctx.WithFields(fields).WithError(err).Error(
			"error returned from create volume")
		return nil, err
	}

	return d.VolumeInspect(ctx, volume.ID,
		&types.VolumeInspectOpts{
			Attachments: types.VolAttReqTrue,
		},
	)
}

func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context, snapshotID string, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	return nil, types.ErrNotImplemented
}

func (d *driver) VolumeCopy(
	ctx types.Context, volumeID string, volumeName string,
	opts types.Store) (*types.Volume, error) {
	return nil, types.ErrNotImplemented
}

func (d *driver) VolumeSnapshot(
	ctx types.Context, volumeID string, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeRemoveOpts) error {

	volume, _, err := d.client.Storage.GetVolume(ctx, volumeID)
	if err != nil {
		return err
	}

	if len(volume.DropletIDs) > 0 {
		if !opts.Force {
			return goof.New("volume currently attached")
		}

		err = d.volumeDetach(ctx, volume)
		if err != nil {
			return err
		}
	}

	_, err = d.client.Storage.DeleteVolume(ctx, volumeID)
	if err != nil {
		return err
	}

	return nil
}

func (d *driver) volumeDetach(
	ctx types.Context,
	volume *godo.Volume) error {

	for _, dropletID := range volume.DropletIDs {
		action, _, err := d.client.StorageActions.DetachByDropletID(
			ctx, volume.ID, dropletID)
		if err != nil {
			return err
		}

		err = d.waitForAction(ctx, volume.ID, action)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) VolumeAttach(
	ctx types.Context, volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {
	vol, _, err := d.client.Storage.GetVolume(ctx, volumeID)
	if err != nil {
		return nil, "", goof.WithError("error retrieving volume", err)
	}

	if len(vol.DropletIDs) > 0 {
		if !opts.Force {
			return nil, "", goof.New("volume already attached")
		}

		err = d.volumeDetach(ctx, vol)
		if err != nil {
			return nil, "", err
		}
	}

	dropletID := mustInstanceIDID(ctx)
	dropletIDI, err := strconv.Atoi(*dropletID)
	if err != nil {
		return nil, "", err
	}

	action, _, err := d.client.StorageActions.Attach(
		ctx, volumeID, dropletIDI)
	if err != nil {
		return nil, "", err
	}

	err = d.waitForAction(ctx, volumeID, action)
	if err != nil {
		return nil, "", err
	}

	attachedVol, err := d.VolumeInspect(ctx, volumeID,
		&types.VolumeInspectOpts{
			Attachments: types.VolAttReqTrue},
	)
	if err != nil {
		return nil, "", goof.WithError("error retrieving volume", err)
	}

	return attachedVol, attachedVol.Name, nil
}

func (d *driver) VolumeDetach(
	ctx types.Context, volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {
	vol, _, err := d.client.Storage.GetVolume(ctx, volumeID)
	if err != nil {
		return nil, goof.WithError("error getting volume", err)
	}

	if len(vol.DropletIDs) == 0 {
		return nil, goof.WithError("volume already detached", err)
	}

	err = d.volumeDetach(ctx, vol)
	if err != nil {
		return nil, err
	}

	ctx.WithField("volumeID", volumeID).Info("detached volume")

	detachedVol, err := d.VolumeInspect(ctx, volumeID,
		&types.VolumeInspectOpts{
			Attachments: types.VolAttReqTrue,
			Opts:        opts.Opts,
		},
	)
	if err != nil {
		return nil, goof.WithError("error getting volume information",
			err)
	}

	return detachedVol, nil
}

func (d *driver) Snapshots(
	ctx types.Context, opts types.Store) ([]*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

func (d *driver) SnapshotInspect(
	ctx types.Context, snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

func (d *driver) SnapshotCopy(
	ctx types.Context, snapshotID string, snapshotName string, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

func (d *driver) SnapshotRemove(
	ctx types.Context, snapshotID string, opts types.Store) error {
	return types.ErrNotImplemented
}

func mustInstanceIDID(ctx types.Context) *string {
	return &context.MustInstanceID(ctx).ID
}

func (d *driver) toTypesVolume(
	ctx types.Context,
	volume *godo.Volume,
	attachments types.VolumeAttachmentsTypes) (*types.Volume, error) {

	var (
		ld    *types.LocalDevices
		ldOK  bool
		iidID string
	)

	if attachments.Devices() {
		// Get local devices map from context
		if ld, ldOK = context.LocalDevices(ctx); !ldOK {
			return nil, goof.New(
				"error getting local devices from context")
		}

		// We require the InstanceID to match the VM
		iidID = context.MustInstanceID(ctx).ID
	}

	vol := &types.Volume{
		Name:             volume.Name,
		ID:               volume.ID,
		Encrypted:        false,
		Size:             volume.SizeGigaBytes,
		AvailabilityZone: volume.Region.Slug,
	}

	// Collect attachment info for the volume
	if attachments.Requested() {
		var atts []*types.VolumeAttachment
		for _, id := range volume.DropletIDs {
			strDropletID := strconv.Itoa(id)
			attachment := &types.VolumeAttachment{
				VolumeID: volume.ID,
				InstanceID: &types.InstanceID{
					ID:     strDropletID,
					Driver: d.Name(),
				},
			}

			if attachments.Devices() {
				if strDropletID == iidID {
					if dev, ok := ld.DeviceMap[vol.Name]; ok {
						attachment.DeviceName = dev
					}
				}
			}
			atts = append(atts, attachment)
		}
		if len(atts) > 0 {
			vol.Attachments = atts
		}
	}

	return vol, nil
}

func (d *driver) waitForAction(
	ctx types.Context,
	volumeID string,
	action *godo.Action) error {

	f := func() (interface{}, error) {
		duration := d.statusDelay
		for i := 1; i <= d.maxAttempts; i++ {
			action, _, err := d.client.StorageActions.Get(
				ctx, volumeID, action.ID)
			if err != nil {
				return nil, err
			}
			if action.Status == godo.ActionCompleted {
				return nil, nil
			}
			ctx.WithField("status", action.Status).Debug(
				"still waiting for action",
			)
			time.Sleep(time.Duration(duration) * time.Nanosecond)
			duration = int64(2) * duration
		}
		return nil, goof.WithField("maxAttempts", d.maxAttempts,
			"Status attempts exhausted")
	}

	_, ok, err := apiUtils.WaitFor(f, d.statusTimeout)
	if !ok {
		return goof.WithFields(goof.Fields{
			"volumeID":      volumeID,
			"statusTimeout": d.statusTimeout},
			"Timeout occured waiting for storage action")
	}
	if err != nil {
		return goof.WithFieldE("volumeID", volumeID,
			"Error while waiting for storage action to finish", err)
	}
	return nil
}

func (d *driver) mustRegion(ctx types.Context) *string {
	if iid, ok := context.InstanceID(ctx); ok {
		if v, ok := iid.Fields[do.InstanceIDFieldRegion]; ok && v != "" {
			return &v
		}
	}
	return d.defaultRegion
}

func (d *driver) convUnderscores(name string) string {
	if d.convUnderscore {
		name = strings.Replace(name, "_", "-", -1)
	}
	return name
}
