// +build mock

package mock

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	"github.com/emccode/libstorage/drivers/storage/mock/executor"
)

const (
	// Name is the name of the driver.
	Name = executor.Name
)

type driver struct {
	executor.Executor
	nextDeviceInfo *types.NextDeviceInfo
	volumes        []*types.Volume
	snapshots      []*types.Snapshot
	storageType    types.StorageType
}

func init() {
	registry.RegisterStorageDriver(Name, newDriver)
}

func newDriver() types.StorageDriver {

	d := &driver{Executor: *executor.NewExecutor()}

	d.nextDeviceInfo = &types.NextDeviceInfo{
		Prefix:  "xvd",
		Pattern: `\w`,
		Ignore:  true,
	}

	d.volumes = []*types.Volume{
		&types.Volume{
			Name:             "Volume 0",
			ID:               "vol-000",
			AvailabilityZone: "zone-000",
			Type:             "gold",
			Size:             10240,
		},
		&types.Volume{
			Name:             "Volume 1",
			ID:               "vol-001",
			AvailabilityZone: "zone-001",
			Type:             "gold",
			Size:             40960,
		},
		&types.Volume{
			Name:             "Volume 2",
			ID:               "vol-002",
			AvailabilityZone: "zone-002",
			Type:             "gold",
			Size:             163840,
		},
	}

	d.snapshots = []*types.Snapshot{
		&types.Snapshot{
			Name:       "Snapshot 0",
			ID:         "snap-000",
			VolumeID:   "vol-000",
			VolumeSize: 100,
		},
		&types.Snapshot{
			Name:       "Snapshot 1",
			ID:         "snap-001",
			VolumeID:   "vol-001",
			VolumeSize: 101,
		},
		&types.Snapshot{
			Name:       "Snapshot 2",
			ID:         "snap-002",
			VolumeID:   "vol-002",
			VolumeSize: 102,
		},
	}

	return d
}

func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Block, nil
}

func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return d.nextDeviceInfo, nil
}

func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {
	iid, _ := d.InstanceID(ctx, opts)
	return &types.Instance{Name: "mockInstance", InstanceID: iid}, nil
}

func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	xiid := executor.GetInstanceID()

	if serviceName, ok := context.ServiceName(ctx); ok && serviceName == Name {
		if ld, ok := context.LocalDevices(ctx); ok {
			ldm := ld.DeviceMap

			if opts.Attachments {

				iid := context.MustInstanceID(ctx)
				if iid.ID == xiid.ID {

					d.volumes[0].Attachments = []*types.VolumeAttachment{
						&types.VolumeAttachment{
							DeviceName: "/dev/xvda",
							MountPoint: ldm["/dev/xvda"],
							InstanceID: iid,
							Status:     "attached",
							VolumeID:   d.volumes[0].ID,
						},
						&types.VolumeAttachment{
							DeviceName: "/dev/xvdb",
							MountPoint: ldm["/dev/xvdb"],
							InstanceID: iid,
							Status:     "attached",
							VolumeID:   d.volumes[1].ID,
						},
						&types.VolumeAttachment{
							DeviceName: "/dev/xvdc",
							MountPoint: ldm["/dev/xvdc"],
							InstanceID: iid,
							Status:     "attached",
							VolumeID:   d.volumes[2].ID,
						},
					}
				}
			}

		}
	}

	return d.volumes, nil
}

func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	for _, v := range d.volumes {
		if strings.ToLower(v.ID) == strings.ToLower(volumeID) {
			return v, nil
		}
	}
	return nil, utils.NewNotFoundError(volumeID)
}

func (d *driver) VolumeCreate(
	ctx types.Context,
	name string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	if name == "Volume 010" {
		return nil, goof.WithFieldE(
			"iops", opts.IOPS,
			"iops required",
			goof.WithFieldE(
				"size", opts.Size,
				"size required",
				goof.New("bzzzzT BROKEN"),
			),
		)
	}
	lenVols := len(d.volumes)

	volume := &types.Volume{
		Name:   name,
		ID:     fmt.Sprintf("vol-%03d", lenVols+1),
		Fields: map[string]string{},
	}

	if opts.AvailabilityZone != nil {
		volume.AvailabilityZone = *opts.AvailabilityZone
	}
	if opts.Type != nil {
		volume.Type = *opts.Type
	}
	if opts.Size != nil {
		volume.Size = *opts.Size
	}
	if opts.IOPS != nil {
		volume.IOPS = *opts.IOPS
	}

	if customFields := opts.Opts.GetStore("opts"); customFields != nil {
		if customFields.IsSet("owner") {
			volume.Fields["owner"] = customFields.GetString("owner")
		}
		if customFields.IsSet("priority") {
			volume.Fields["priority"] = customFields.GetString("priority")
		}
	}

	d.volumes = append(d.volumes, volume)

	return volume, nil
}

func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	s, err := d.SnapshotInspect(ctx, snapshotID, nil)
	if err != nil {
		return nil, err
	}

	v, err := d.VolumeInspect(ctx, s.VolumeID, nil)
	if err != nil {
		return nil, err
	}

	lenVols := len(d.volumes)

	volume := &types.Volume{
		Name:   volumeName,
		ID:     fmt.Sprintf("vol-%03d", lenVols+1),
		Fields: map[string]string{},
	}

	if opts.AvailabilityZone != nil {
		volume.AvailabilityZone = v.AvailabilityZone
	}
	if opts.Type != nil {
		volume.Type = v.Type
	}
	if opts.Size != nil {
		volume.Size = v.Size
	}
	if opts.IOPS != nil {
		volume.IOPS = v.IOPS
	}

	if opts.Opts.IsSet("owner") {
		volume.Fields["owner"] = opts.Opts.GetString("owner")
	}
	if opts.Opts.IsSet("priority") {
		volume.Fields["priority"] = opts.Opts.GetString("priority")
	}
	return volume, nil
}

func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {

	ctx.WithFields(log.Fields{
		"volumeID":   volumeID,
		"volumeName": volumeName,
	}).Debug("mockDriver.VolumeCopy")

	lenVols := len(d.volumes)

	var ogvol *types.Volume
	for _, v := range d.volumes {
		if strings.ToLower(v.ID) == strings.ToLower(volumeID) {
			ogvol = v
			break
		}
	}

	volume := &types.Volume{
		Name:             volumeName,
		ID:               fmt.Sprintf("vol-%03d", lenVols+1),
		AvailabilityZone: ogvol.AvailabilityZone,
		Type:             ogvol.Type,
		Size:             ogvol.Size,
		Fields:           map[string]string{},
	}

	for k, v := range ogvol.Fields {
		volume.Fields[k] = v
	}

	d.volumes = append(d.volumes, volume)

	return volume, nil

}

func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {

	ctx.WithFields(log.Fields{
		"volumeID":     volumeID,
		"snapshotName": snapshotName,
	}).Debug("mockDriver.VolumeSnapshot")

	lenSnaps := len(d.snapshots)

	snapshot := &types.Snapshot{
		Name:     snapshotName,
		ID:       fmt.Sprintf("snap-%03d", lenSnaps+1),
		VolumeID: volumeID,
		Fields:   map[string]string{},
	}

	d.snapshots = append(d.snapshots, snapshot)

	return snapshot, nil
}

func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {

	ctx.WithFields(log.Fields{
		"volumeID": volumeID,
	}).Debug("mockDriver.VolumeRemove")

	var xToRemove int
	var volume *types.Volume
	for x, v := range d.volumes {
		if strings.ToLower(v.ID) == strings.ToLower(volumeID) {
			volume = v
			xToRemove = x
			break
		}
	}

	if volume == nil {
		return utils.NewNotFoundError(volumeID)
	}

	d.volumes = append(d.volumes[:xToRemove], d.volumes[xToRemove+1:]...)

	return nil
}

func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	var modVol *types.Volume
	for _, vol := range d.volumes {
		if vol.ID == volumeID {
			modVol = vol
			break
		}
	}

	modVol.Attachments = []*types.VolumeAttachment{
		&types.VolumeAttachment{
			DeviceName: *opts.NextDevice,
			MountPoint: "",
			InstanceID: context.MustInstanceID(ctx),
			Status:     "attached",
			VolumeID:   modVol.ID,
		},
	}

	return modVol, "1234", nil
}

func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	var modVol *types.Volume
	for _, vol := range d.volumes {
		if vol.ID == volumeID {
			modVol = vol
			break
		}
	}

	modVol.Attachments = nil

	return modVol, nil
}

func (d *driver) VolumeDetachAll(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {

	for _, vol := range d.volumes {
		vol.Attachments = nil
	}

	return nil
}

func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {

	return d.snapshots, nil
}

func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {

	for _, v := range d.snapshots {
		if strings.ToLower(v.ID) == strings.ToLower(snapshotID) {
			return v, nil
		}
	}
	return nil, goof.New("invalid snapshot id")
}

func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {

	ctx.WithFields(log.Fields{
		"snapshotID":    snapshotID,
		"snapshotName":  snapshotName,
		"destinationID": destinationID,
	}).Debug("mockDriver.SnapshotCopy")

	lenSnaps := len(d.snapshots)

	var ogsnap *types.Snapshot
	for _, s := range d.snapshots {
		if strings.ToLower(s.ID) == strings.ToLower(snapshotID) {
			ogsnap = s
			break
		}
	}

	snapshot := &types.Snapshot{
		Name:     snapshotName,
		ID:       fmt.Sprintf("snap-%03d", lenSnaps+1),
		VolumeID: ogsnap.VolumeID,
		Fields:   map[string]string{},
	}

	for k, s := range ogsnap.Fields {
		snapshot.Fields[k] = s
	}

	d.snapshots = append(d.snapshots, snapshot)

	return snapshot, nil
}

func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {

	ctx.WithFields(log.Fields{
		"snapshotID": snapshotID,
	}).Debug("mockDriver.SnapshotRemove")

	var xToRemove int
	var snapshot *types.Snapshot
	for x, s := range d.snapshots {
		if strings.ToLower(s.ID) == strings.ToLower(snapshotID) {
			snapshot = s
			xToRemove = x
			break
		}
	}

	if snapshot == nil {
		return utils.NewNotFoundError(snapshotID)
	}

	d.snapshots = append(d.snapshots[:xToRemove], d.snapshots[xToRemove+1:]...)

	return nil
}
