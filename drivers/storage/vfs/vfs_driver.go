package vfs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
	"github.com/emccode/libstorage/api/utils"
	"github.com/emccode/libstorage/drivers/storage/vfs/executor"
)

const (
	// Name is the name of the driver.
	Name = executor.Name
)

type driver struct {
	executor.Executor
	config gofig.Config

	volJSONGlobPatt string
	volCount        int64
}

func init() {
	registry.RegisterRemoteStorageDriver(executor.Name, newDriver)
}

func newDriver() drivers.RemoteStorageDriver {
	return &driver{Executor: executor.Executor{}}
}

func (d *driver) Init(config gofig.Config) error {
	if err := d.Executor.Init(config); err != nil {
		return err
	}

	d.config = config
	d.volJSONGlobPatt = fmt.Sprintf("%s/*.json", d.Executor.VolumesDirPath())

	volJSONPaths, err := d.getVolJSONs()
	if err != nil {
		return nil
	}
	d.volCount = int64(len(volJSONPaths)) - 1

	return nil
}

func (d *driver) Type() types.StorageType {
	return types.Object
}

func (d *driver) NextDeviceInfo() *types.NextDeviceInfo {
	return &types.NextDeviceInfo{
		Ignore: true,
	}
}

func (d *driver) InstanceInspect(
	ctx context.Context,
	opts types.Store) (*types.Instance, error) {
	return &types.Instance{InstanceID: ctx.InstanceID()}, nil
}

func (d *driver) Volumes(
	ctx context.Context,
	opts *drivers.VolumesOpts) ([]*types.Volume, error) {

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
		volumes = append(volumes, v)
	}

	return volumes, nil
}

func (d *driver) VolumeInspect(
	ctx context.Context,
	volumeID string,
	opts *drivers.VolumeInspectOpts) (*types.Volume, error) {

	return d.getVolumeByID(volumeID)
}

func (d *driver) VolumeCreate(
	ctx context.Context,
	name string,
	opts *drivers.VolumeCreateOpts) (*types.Volume, error) {

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

	return v, nil
}

func (d *driver) VolumeCreateFromSnapshot(
	ctx context.Context,
	snapshotID, volumeName string,
	opts *drivers.VolumeCreateOpts) (*types.Volume, error) {
	return nil, nil
}

func (d *driver) VolumeCopy(
	ctx context.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {

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
		Fields:           map[string]string{},
	}

	if ogVol.Fields != nil {
		for k, v := range ogVol.Fields {
			newVol.Fields[k] = v
		}
	}

	if err := d.writeVolume(newVol); err != nil {
		return nil, err
	}

	return newVol, nil
}

func (d *driver) VolumeSnapshot(
	ctx context.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) VolumeRemove(
	ctx context.Context,
	volumeID string,
	opts types.Store) error {
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
	return nil
}

func (d *driver) Snapshots(
	ctx context.Context,
	opts types.Store) ([]*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotInspect(
	ctx context.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotCopy(
	ctx context.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotRemove(
	ctx context.Context,
	snapshotID string,
	opts types.Store) error {
	return nil
}

func (d *driver) getVolumeByID(volumeID string) (*types.Volume, error) {
	volJSONPath :=
		fmt.Sprintf("%s/%s.json", d.Executor.VolumesDirPath(), volumeID)

	if !gotil.FileExists(volJSONPath) {
		return nil, utils.NewNotFoundError(volumeID)
	}

	return readVolume(volJSONPath)
}

func readVolume(path string) (*types.Volume, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	v := &types.Volume{}
	if err := json.NewDecoder(f).Decode(v); err != nil {
		return nil, err
	}

	return v, nil
}

func (d *driver) writeVolume(v *types.Volume) error {
	volJSONPath :=
		fmt.Sprintf("%s/vol-%s.json", d.Executor.VolumesDirPath(), v.ID)

	f, err := os.Create(volJSONPath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)

	if err := enc.Encode(v); err != nil {
		return err
	}

	return nil
}

func (d *driver) getVolJSONs() ([]string, error) {
	volPatt := fmt.Sprintf("%s/*.json", d.Executor.VolumesDirPath())
	return filepath.Glob(volPatt)
}

func (d *driver) newVolumeID() string {
	return fmt.Sprintf("vfs-%3d", atomic.AddInt64(&d.volCount, 1))
}
