package storage

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	gofig "github.com/akutz/gofig/types"

	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/vfs"
)

const (
	minSizeGiB = 1
)

type driver struct {
	ctx    types.Context
	config gofig.Config

	volJSONGlobPatt  string
	snapJSONGlobPatt string
	volCount         int64
	snapCount        int64

	volPath  string
	snapPath string
}

func init() {
	registry.RegisterStorageDriver(vfs.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{}
}

func (d *driver) Name() string {
	return vfs.Name
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.ctx = ctx
	d.config = config

	d.volPath = vfs.VolumesDirPath(config)
	d.snapPath = vfs.SnapshotsDirPath(config)

	ctx.WithField("vfs.root.path", vfs.RootDir(config)).Info("vfs.root")

	os.MkdirAll(d.volPath, 0755)
	os.MkdirAll(d.snapPath, 0755)

	d.volJSONGlobPatt = fmt.Sprintf("%s/*.json", d.volPath)
	d.snapJSONGlobPatt = fmt.Sprintf("%s/*.json", d.snapPath)

	volJSONPaths, err := d.getVolJSONs()
	if err != nil {
		return nil
	}
	d.volCount = int64(len(volJSONPaths)) - 1

	snapJSONPaths, err := d.getSnapJSONs()
	if err != nil {
		return nil
	}
	d.snapCount = int64(len(snapJSONPaths)) - 1

	return nil
}

type session struct {
	id string
}

func (d *driver) Login(ctx types.Context) (interface{}, error) {
	sess := &session{}
	if iid, ok := context.InstanceID(ctx); ok {
		sess.id = iid.String()
	}
	ctx.Debugf("vfs login=%s", sess.id)
	return sess, nil
}

func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Object, nil
}

func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return &types.NextDeviceInfo{
		Ignore: true,
	}, nil
}

func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	var err error
	if ctx, err = context.WithStorageSession(ctx); err != nil {
		return nil, err
	}

	context.MustSession(ctx)

	iid := context.MustInstanceID(ctx)
	if iid.ID != "" {
		return &types.Instance{Name: "vfsInstance", InstanceID: iid}, nil
	}

	var hostname string
	if err := iid.UnmarshalMetadata(&hostname); err != nil {
		return nil, err
	}

	return &types.Instance{
		Name: "vfsInstance",
		InstanceID: &types.InstanceID{
			ID:     hostname,
			Driver: iid.Driver,
		},
	}, nil
}

func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	context.MustSession(ctx)

	iid, iidOK := context.InstanceID(ctx)
	if iidOK {
		if iid.ID == "" {
			return nil, goof.New("missing instance ID")
		}
	}

	volJSONPaths, err := d.getVolJSONs()
	if err != nil {
		return nil, err
	}

	volumes := []*types.Volume{}

	for _, volJSONPath := range volJSONPaths {
		v, err := readVolume(volJSONPath)
		if err != nil {
			return nil, err
		}
		if opts.Attachments > 0 {
			v.AttachmentState = 0
		}
		if opts.Attachments.Mine() {
			atts := []*types.VolumeAttachment{}
			for _, a := range v.Attachments {
				if a.InstanceID != nil && a.InstanceID.ID == iid.ID {
					atts = append(atts, a)
				}
			}
			v.Attachments = atts
		}
		volumes = append(volumes, v)
	}

	return utils.SortVolumeByID(volumes), nil
}

func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	context.MustSession(ctx)

	v, err := d.getVolumeByID(volumeID)
	if err != nil {
		return nil, err
	}
	if opts.Attachments > 0 {
		v.AttachmentState = 0
	}
	if opts.Attachments.Mine() {
		iid, iidOK := context.InstanceID(ctx)
		if iidOK {
			if iid.ID == "" {
				return nil, goof.New("missing instance ID")
			}
		}
		atts := []*types.VolumeAttachment{}
		for _, a := range v.Attachments {
			if a.InstanceID != nil && a.InstanceID.ID == iid.ID {
				atts = append(atts, a)
			}
		}
		v.Attachments = atts
	}
	return v, nil
}

func (d *driver) VolumeCreate(
	ctx types.Context,
	name string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	fields := map[string]interface{}{"name": name}
	if opts != nil {
		if opts.AvailabilityZone != nil {
			fields["availabilityZone"] = *opts.AvailabilityZone
		}
		if opts.Encrypted != nil {
			fields["encrypted"] = *opts.Encrypted
		}
		if opts.EncryptionKey != nil {
			fields["encryptionKey"] = *opts.EncryptionKey
		}
		if opts.IOPS != nil {
			fields["iops"] = *opts.IOPS
		}
		if opts.Size == nil {
			size := int64(minSizeGiB)
			opts.Size = &size
		}
		fields["size"] = *opts.Size
		if opts.Type != nil {
			fields["type"] = *opts.Type
		}
		//if opts.Opts != nil {
		//	fields["opts"] = opts.Opts
		//}
	}
	ctx.WithFields(fields).Debug("creating volume")

	context.MustSession(ctx)

	v := &types.Volume{
		ID:     d.newVolumeID(),
		Name:   name,
		Fields: map[string]string{},
	}

	if opts.AvailabilityZone != nil {
		v.AvailabilityZone = *opts.AvailabilityZone
	}
	if opts.IOPS != nil {
		v.IOPS = *opts.IOPS
	}
	v.Size = *opts.Size

	if opts.Type != nil {
		v.Type = *opts.Type
	}
	if opts.Encrypted != nil {
		v.Encrypted = *opts.Encrypted
	}
	if customFields := opts.Opts.GetStore("opts"); customFields != nil {
		for _, k := range customFields.Keys() {
			v.Fields[k] = customFields.GetString(k)
		}
	}

	if err := d.writeVolume(v); err != nil {
		return nil, err
	}

	volDir := path.Join(vfs.VolumesDirPath(d.config), v.ID)
	os.MkdirAll(volDir, 0755)

	return v, nil
}

func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	context.MustSession(ctx)

	snap, err := d.getSnapshotByID(snapshotID)
	if err != nil {
		return nil, err
	}

	ogVol, err := d.getVolumeByID(snap.VolumeID)
	if err != nil {
		return nil, err
	}

	v := &types.Volume{
		ID:               d.newVolumeID(),
		Name:             volumeName,
		Fields:           ogVol.Fields,
		AvailabilityZone: ogVol.AvailabilityZone,
		IOPS:             ogVol.IOPS,
		Size:             ogVol.Size,
		Type:             ogVol.Type,
	}

	if opts.AvailabilityZone != nil {
		v.AvailabilityZone = *opts.AvailabilityZone
	}
	if opts.IOPS != nil {
		v.IOPS = *opts.IOPS
	}
	if opts.Size != nil {
		v.Size = *opts.Size
	}
	if opts.Type != nil {
		v.Type = *opts.Type
	}
	if customFields := opts.Opts.GetStore("opts"); customFields != nil {
		for _, k := range customFields.Keys() {
			v.Fields[k] = customFields.GetString(k)
		}
	}

	if err := d.writeVolume(v); err != nil {
		return nil, err
	}

	volDir := path.Join(vfs.VolumesDirPath(d.config), v.ID)
	os.MkdirAll(volDir, 0755)

	return v, nil
}

func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {

	context.MustSession(ctx)

	ogVol, err := d.getVolumeByID(volumeID)
	if err != nil {
		return nil, err
	}

	newVol := &types.Volume{
		ID:               d.newVolumeID(),
		Name:             volumeName,
		AvailabilityZone: ogVol.AvailabilityZone,
		IOPS:             ogVol.IOPS,
		Size:             ogVol.Size,
		Type:             ogVol.Type,
		Fields:           ogVol.Fields,
	}

	if customFields := opts.GetStore("opts"); customFields != nil {
		for _, k := range customFields.Keys() {
			newVol.Fields[k] = customFields.GetString(k)
		}
	}

	if err := d.writeVolume(newVol); err != nil {
		return nil, err
	}

	volDir := path.Join(vfs.VolumesDirPath(d.config), newVol.ID)
	os.MkdirAll(volDir, 0755)

	return newVol, nil
}

func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {

	context.MustSession(ctx)

	v, err := d.getVolumeByID(volumeID)
	if err != nil {
		return nil, err
	}

	s := &types.Snapshot{
		ID:         d.newSnapshotID(v.ID),
		VolumeID:   v.ID,
		VolumeSize: v.Size,
		Name:       snapshotName,
		Status:     "online",
		StartTime:  time.Now().Unix(),
		Fields:     v.Fields,
	}

	if customFields := opts.GetStore("opts"); customFields != nil {
		for _, k := range customFields.Keys() {
			s.Fields[k] = customFields.GetString(k)
		}
	}

	if err := d.writeSnapshot(s); err != nil {
		return nil, err
	}

	return s, nil
}

func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeRemoveOpts) error {

	context.MustSession(ctx)

	volJSONPath := d.getVolPath(volumeID)
	if !gotil.FileExists(volJSONPath) {
		return utils.NewNotFoundError(volumeID)
	}
	os.Remove(volJSONPath)

	volDir := path.Join(vfs.VolumesDirPath(d.config), volumeID)
	os.RemoveAll(volDir)
	return nil
}

func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	switch os.Getenv("VFS_STDLIB_GOOF_ERR") {
	case "1":
		return nil, "", goof.WithFieldE(
			"path", "/tmp", "this error has an inner error",
			goof.New("a bug squashing expedition all died"))
	case "2":
		return nil, "", goof.WithFieldE(
			"path", "/tmp", "this error has an inner error",
			errors.New("a bug squashing expedition all died"))
	}

	context.MustSession(ctx)

	vol, err := d.getVolumeByID(volumeID)
	if err != nil {
		return nil, "", err
	}

	nextDevice := ""
	if opts.NextDevice != nil {
		nextDevice = *opts.NextDevice
	}

	att := &types.VolumeAttachment{
		VolumeID:   vol.ID,
		InstanceID: context.MustInstanceID(ctx),
		DeviceName: nextDevice,
		Status:     "attached",
	}

	vol.Attachments = append(vol.Attachments, att)
	if err := d.writeVolume(vol); err != nil {
		return nil, "", err
	}

	vol.Attachments = []*types.VolumeAttachment{att}

	return vol, nextDevice, nil
}

func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	context.MustSession(ctx)

	vol, err := d.getVolumeByID(volumeID)
	if err != nil {
		return nil, err
	}

	iid := context.MustInstanceID(ctx)

	y := -1
	for x, att := range vol.Attachments {
		if att.InstanceID.ID == iid.ID {
			y = x
			break
		}
	}

	if y > -1 {
		vol.Attachments = append(vol.Attachments[:y], vol.Attachments[y+1:]...)
		if err := d.writeVolume(vol); err != nil {
			return nil, err
		}
	}

	vol.Attachments = nil
	return vol, nil
}

func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {

	context.MustSession(ctx)

	snapJSONPaths, err := d.getSnapJSONs()
	if err != nil {
		return nil, err
	}

	snapshots := []*types.Snapshot{}

	for _, snapJSONPath := range snapJSONPaths {
		s, err := readSnapshot(snapJSONPath)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}

	return snapshots, nil
}

func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {

	context.MustSession(ctx)

	snap, err := d.getSnapshotByID(snapshotID)
	if err != nil {
		return nil, err
	}
	return snap, nil
}

func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {

	context.MustSession(ctx)

	ogSnap, err := d.getSnapshotByID(snapshotID)
	if err != nil {
		return nil, err
	}

	newSnap := &types.Snapshot{
		ID:         d.newSnapshotID(ogSnap.VolumeID),
		VolumeID:   ogSnap.VolumeID,
		VolumeSize: ogSnap.VolumeSize,
		Name:       snapshotName,
		Status:     "online",
		StartTime:  time.Now().Unix(),
		Fields:     ogSnap.Fields,
	}

	if customFields := opts.GetStore("opts"); customFields != nil {
		for _, k := range customFields.Keys() {
			newSnap.Fields[k] = customFields.GetString(k)
		}
	}

	if err := d.writeSnapshot(newSnap); err != nil {
		return nil, err
	}

	return newSnap, nil
}

func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {

	context.MustSession(ctx)

	snapJSONPath := d.getSnapPath(snapshotID)
	if !gotil.FileExists(snapJSONPath) {
		return utils.NewNotFoundError(snapshotID)
	}
	os.Remove(snapJSONPath)
	return nil
}
