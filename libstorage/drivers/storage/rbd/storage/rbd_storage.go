package storage

import (
	"regexp"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	apiUtils "github.com/AVENTER-UG/rexray/libstorage/api/utils"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/rbd"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/rbd/utils"
)

const (
	rbdDefaultOrder = 22
	bytesPerGiB     = 1024 * 1024 * 1024
	validNameRX     = `[0-9A-Za-z_\-]+`
	minSizeGiB      = 1
)

var (
	featureLayering   = "layering"
	defaultObjectSize = "4M"
)

type driver struct {
	config      gofig.Config
	defaultPool string
	multiPool   bool
}

func init() {
	registry.RegisterStorageDriver(rbd.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{}
}

func (d *driver) Name() string {
	return rbd.Name
}

// Init initializes the driver.
func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	d.defaultPool = d.config.GetString(rbd.ConfigDefaultPool)
	cephArgs := d.config.GetString(rbd.ConfigCephArgs)
	if cephArgs == "" || !utils.StrContainsClient(cephArgs) {
		d.multiPool = true
	}
	ctx.Info("storage driver initialized")
	return nil
}

func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Block, nil
}

func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return nil, nil
}

func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	iid := context.MustInstanceID(ctx)
	return &types.Instance{
		InstanceID: iid,
	}, nil
}

func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	var (
		pools   []string
		err     error
		volumes []*types.Volume
	)

	if d.multiPool {
		// Get all Volumes in all pools
		pools, err = utils.GetRadosPools(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		ctx.WithField("pool", d.defaultPool).Info(
			"Only using default pool since cephArgs is set")
		pools = append(pools, d.defaultPool)
	}

	for _, pool := range pools {
		images, err := utils.GetRBDImages(ctx, pool)
		if err != nil {
			return nil, err
		}

		vols, err := d.toTypeVolumes(
			ctx, images, opts.Attachments)
		if err != nil {
			/* Should we try to continue instead? */
			return nil, err
		}
		volumes = append(volumes, vols...)
	}

	return volumes, nil
}

func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	pool, image, err := d.parseVolumeID(&volumeID)
	if err != nil {
		return nil, err
	}

	info, err := utils.GetRBDInfo(ctx, pool, image)
	if err != nil {
		return nil, err
	}

	// no volume returned
	if info == nil {
		return nil, apiUtils.NewNotFoundError(volumeID)
	}

	/* GetRBDInfo returns more details about an image than what we get back
	   from GetRBDImages. We could just use that and then grab the image we
	   want, but using GetRBDInfo() instead in case we ever want to send
	   back  more detaild information from VolumeInspect() than Volumes() */
	images := []*utils.RBDImage{
		&utils.RBDImage{
			Name: info.Name,
			Size: info.Size,
			Pool: info.Pool,
		},
	}

	vols, err := d.toTypeVolumes(ctx, images, opts.Attachments)
	if err != nil {
		return nil, err
	}

	return vols[0], nil
}

func (d *driver) VolumeInspectByName(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	// volumeName and volumeID are the same for RBD
	return d.VolumeInspect(
		ctx,
		volumeName,
		opts,
	)

}

func (d *driver) VolumeCreate(ctx types.Context, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	fields := map[string]interface{}{
		"driverName": d.Name(),
		"volumeName": volumeName,
	}

	ctx.WithFields(fields).Debug("creating volume")

	pool, imageName, err := d.parseVolumeID(&volumeName)
	if err != nil {
		return nil, err
	}

	info, err := utils.GetRBDInfo(ctx, pool, imageName)
	if err != nil {
		return nil, err
	}

	// volume already exists
	if info != nil {
		return nil, goof.New("Volume already exists")
	}

	//TODO: config options for order and features?

	features := []*string{&featureLayering}

	if opts.Size == nil {
		size := int64(minSizeGiB)
		opts.Size = &size
	}

	fields["opts.Size"] = *opts.Size
	if *opts.Size < minSizeGiB {
		fields["minSize"] = minSizeGiB
		return nil, goof.WithFields(fields, "volume size too small")
	}

	err = utils.RBDCreate(
		ctx,
		pool,
		imageName,
		opts.Size,
		&defaultObjectSize,
		features,
	)
	if err != nil {
		return nil, goof.WithFieldsE(fields,
			"Failed to create new volume", err)
	}

	volumeID := utils.GetVolumeID(pool, imageName)
	return d.VolumeInspect(ctx, *volumeID,
		&types.VolumeInspectOpts{
			Attachments: types.VolAttNone,
		},
	)
}

func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	return nil, types.ErrNotImplemented
}

func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {
	return nil, types.ErrNotImplemented
}

func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeRemoveOpts) error {

	fields := map[string]interface{}{
		"driverName": d.Name(),
		"volumeID":   volumeID,
	}

	ctx.WithFields(fields).Debug("deleting volume")

	pool, imageName, err := d.parseVolumeID(&volumeID)
	if err != nil {
		return goof.WithError("Unable to set image name", err)
	}

	err = utils.RBDRemove(ctx, pool, imageName)
	if err != nil {
		return goof.WithError("Error while deleting RBD image", err)
	}
	ctx.WithFields(fields).Debug("removed volume")

	return nil
}

func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	fields := map[string]interface{}{
		"driverName": d.Name(),
		"volumeID":   volumeID,
	}

	ctx.WithFields(fields).Debug("attaching volume")

	pool, imageName, err := d.parseVolumeID(&volumeID)
	if err != nil {
		return nil, "", goof.WithError("Unable to set image name", err)
	}

	vol, err := d.VolumeInspect(
		ctx, volumeID, &types.VolumeInspectOpts{
			Attachments: types.VolAttReq,
		})
	if err != nil {
		if _, ok := err.(*types.ErrNotFound); ok {
			return nil, "", err
		}
		return nil, "", goof.WithError("error getting volume", err)
	}

	if vol.AttachmentState != types.VolumeAvailable {
		if !opts.Force {
			return nil, "",
				goof.New("volume in wrong state for attach")
		}
	}

	_, err = utils.RBDMap(ctx, pool, imageName)
	if err != nil {
		return nil, "", err
	}

	vol, err = d.VolumeInspect(ctx, volumeID,
		&types.VolumeInspectOpts{
			Attachments: types.VolAttReqTrue,
		},
	)
	if err != nil {
		return nil, "", err
	}

	return vol, volumeID, nil
}

func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	fields := map[string]interface{}{
		"driverName": d.Name(),
		"volumeID":   volumeID,
	}

	ctx.WithFields(fields).Debug("detaching volume")

	// Can't rely on local devices header, so get local attachments
	localAttachMap, err := utils.GetMappedRBDs(ctx)
	if err != nil {
		return nil, err
	}

	dev, found := localAttachMap[volumeID]
	if !found {
		return nil, goof.New("Volume not attached")
	}

	err = utils.RBDUnmap(ctx, &dev)
	if err != nil {
		return nil, goof.WithError("Unable to detach volume", err)
	}

	return d.VolumeInspect(
		ctx, volumeID, &types.VolumeInspectOpts{
			Attachments: types.VolAttReqTrue,
		},
	)
}

func (d *driver) VolumeDetachAll(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {
	return types.ErrNotImplemented
}

func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {
	return types.ErrNotImplemented
}

func (d *driver) toTypeVolumes(
	ctx types.Context,
	images []*utils.RBDImage,
	getAttachments types.VolumeAttachmentsTypes) ([]*types.Volume, error) {

	lsVolumes := make([]*types.Volume, len(images))

	var localAttachMap map[string]string

	// Even though this will be the same as LocalDevices header, we can't
	// rely on that being present unless getAttachments.Devices is set
	if getAttachments.Requested() {
		var err error
		localAttachMap, err = utils.GetMappedRBDs(ctx)
		if err != nil {
			return nil, err
		}
	}

	for i, image := range images {
		rbdID := utils.GetVolumeID(&image.Pool, &image.Name)
		lsVolume := &types.Volume{
			Name: image.Name,
			ID:   *rbdID,
			Type: image.Pool,
			Size: int64(image.Size / bytesPerGiB),
		}

		if getAttachments.Requested() && localAttachMap != nil {
			// Set volumeAttachmentState to Unknown, because this
			// driver (currently) has no way of knowing if an image
			// is attached anywhere else but to the caller
			lsVolume.AttachmentState = types.VolumeAttachmentStateUnknown
			var attachments []*types.VolumeAttachment
			if _, found := localAttachMap[*rbdID]; found {
				lsVolume.AttachmentState = types.VolumeAttached
				attachment := &types.VolumeAttachment{
					VolumeID:   *rbdID,
					InstanceID: context.MustInstanceID(ctx),
				}
				if getAttachments.Devices() {
					ld, ok := context.LocalDevices(ctx)
					if ok {
						attachment.DeviceName = ld.DeviceMap[*rbdID]
					} else {
						ctx.Warnf("Unable to get local device map for volume %s", *rbdID)
					}
				}
				attachments = append(attachments, attachment)
				lsVolume.Attachments = attachments
			} else {
				//Check if RBD has watchers to infer attachment
				//to a different host
				b, err := utils.RBDHasWatchers(
					ctx, &image.Pool, &image.Name,
				)
				if err != nil {
					ctx.Warnf("Unable to determine attachment state: %v", err)
				} else {
					if b {
						lsVolume.AttachmentState = types.VolumeUnavailable
					} else {
						lsVolume.AttachmentState = types.VolumeAvailable
					}
				}
			}
		}
		lsVolumes[i] = lsVolume
	}

	return lsVolumes, nil
}

func (d *driver) parseVolumeID(name *string) (*string, *string, error) {

	// Look for <pool>.<name>
	re, _ := regexp.Compile(
		`^(` + validNameRX + `)\.(` + validNameRX + `)$`)
	res := re.FindStringSubmatch(*name)
	if len(res) == 3 {
		// Name includes pool already
		return &res[1], &res[2], nil
	}

	// make sure <name> is valid
	re, _ = regexp.Compile(`^` + validNameRX + `$`)
	if !re.MatchString(*name) {
		return nil, nil, goof.New(
			"Invalid character(s) found in volume name")
	}

	pool := d.defaultPool
	return &pool, name, nil
}
