package mock

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
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

func newDriver() drivers.StorageDriver {

	d := &driver{Executor: *executor.NewExecutor()}

	d.nextDeviceInfo = &types.NextDeviceInfo{
		Prefix:  "xvd",
		Pattern: `\w`,
		Ignore:  false,
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
			Name:     "Snapshot 0",
			ID:       "snap-000",
			VolumeID: "vol-000",
		},
		&types.Snapshot{
			Name:     "Snapshot 1",
			ID:       "snap-001",
			VolumeID: "vol-001",
		},
		&types.Snapshot{
			Name:     "Snapshot 2",
			ID:       "snap-002",
			VolumeID: "vol-002",
		},
	}

	return d
}

func (d *driver) Type() types.StorageType {
	return types.Block
}

func (d *driver) NextDeviceInfo() *types.NextDeviceInfo {
	return d.nextDeviceInfo
}

func (d *driver) InstanceInspect(
	ctx context.Context,
	opts types.Store) (*types.Instance, error) {
	iid, _ := d.InstanceID(ctx, opts)
	return &types.Instance{InstanceID: iid}, nil
}

func (d *driver) Volumes(
	ctx context.Context,
	opts *drivers.VolumesOpts) ([]*types.Volume, error) {
	return d.volumes, nil
}

func (d *driver) VolumeInspect(
	ctx context.Context,
	volumeID string,
	opts *drivers.VolumeInspectOpts) (*types.Volume, error) {

	for _, v := range d.volumes {
		if strings.ToLower(v.ID) == strings.ToLower(volumeID) {
			return v, nil
		}
	}
	return nil, utils.NewNotFoundError(volumeID)
}

func (d *driver) VolumeCreate(
	ctx context.Context,
	name string,
	opts *drivers.VolumeCreateOpts) (*types.Volume, error) {

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

	if opts.Opts.IsSet("owner") {
		volume.Fields["owner"] = opts.Opts.GetString("owner")
	}
	if opts.Opts.IsSet("priority") {
		volume.Fields["priority"] = opts.Opts.GetString("priority")
	}

	d.volumes = append(d.volumes, volume)

	return volume, nil
}

func (d *driver) VolumeCreateFromSnapshot(
	ctx context.Context,
	snapshotID, volumeName string,
	opts types.Store) (*types.Volume, error) {

	return nil, nil
}

func (d *driver) VolumeCopy(
	ctx context.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {

	ctx.Log().WithFields(log.Fields{
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
	ctx context.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {

	ctx.Log().WithFields(log.Fields{
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
	ctx context.Context,
	volumeID string,
	opts types.Store) error {

	ctx.Log().WithFields(log.Fields{
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
	ctx context.Context,
	volumeID string,
	opts *drivers.VolumeAttachByIDOpts) (*types.Volume, error) {

	return nil, nil
}

func (d *driver) VolumeDetach(
	ctx context.Context,
	volumeID string,
	opts types.Store) error {

	if strings.ToLower(ctx.Value("serviceID").(string)) == "testservice2" &&
		strings.ToLower(volumeID) == "mockdriver2-vol-001" {
		return goof.New("volume detach error")
	}

	return nil
}

func (d *driver) Snapshots(
	ctx context.Context,
	opts types.Store) ([]*types.Snapshot, error) {

	return d.snapshots, nil
}

func (d *driver) SnapshotInspect(
	ctx context.Context,
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
	ctx context.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {

	ctx.Log().WithFields(log.Fields{
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
	ctx context.Context,
	snapshotID string,
	opts types.Store) error {

	ctx.Log().WithFields(log.Fields{
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
