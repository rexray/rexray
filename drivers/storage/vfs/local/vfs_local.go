package local

import (
	"os"
	"path"

	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
	apihttp "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/drivers/storage/vfs"
)

type driver struct {
	config gofig.Config
}

func init() {
	registry.RegisterLocalStorageDriver(vfs.Name, newDriver)
}

func newDriver() drivers.LocalStorageDriver {
	return &driver{}
}

func (d *driver) Name() string {
	return vfs.Name
}

func (d *driver) Init(config gofig.Config) error {
	d.config = config
	os.MkdirAll(vfs.VolumesDirPath(config), 0755)
	return nil
}

func (d *driver) InstanceInspectBefore(ctx *context.Context) error {
	return nil
}

func (d *driver) InstanceInspectAfter(
	ctx context.Context, result *types.Instance) {
}

func (d *driver) VolumesBefore(ctx *context.Context) error {
	return nil
}

func (d *driver) VolumesAfter(
	ctx context.Context, result *apihttp.ServiceVolumeMap) {
}

func (d *driver) VolumesByServiceBefore(
	ctx *context.Context, service string) error {
	return nil
}

func (d *driver) VolumesByServiceAfter(
	ctx context.Context, service string, result *apihttp.VolumeMap) {
}

func (d *driver) VolumeInspectBefore(
	ctx *context.Context,
	service, volumeID string, attachments bool) error {
	return nil
}

func (d *driver) VolumeInspectAfter(
	ctx context.Context,
	result *types.Volume) {
}

func (d *driver) VolumeCreateBefore(
	ctx *context.Context,
	service string, request *apihttp.VolumeCreateRequest) error {
	return nil
}

func (d *driver) VolumeCreateAfter(
	ctx context.Context,
	result *types.Volume) {
	volDir := path.Join(vfs.VolumesDirPath(d.config), result.ID)
	os.MkdirAll(volDir, 0755)
}

func (d *driver) VolumeCreateFromSnapshotBefore(
	ctx *context.Context,
	service, snapshotID string,
	request *apihttp.VolumeCreateRequest) error {
	return nil
}

func (d *driver) VolumeCreateFromSnapshotAfter(
	ctx context.Context, result *types.Volume) {
	volDir := path.Join(vfs.VolumesDirPath(d.config), result.ID)
	os.MkdirAll(volDir, 0755)
}

func (d *driver) VolumeCopyBefore(
	ctx *context.Context,
	service, volumeID string, request *apihttp.VolumeCopyRequest) error {
	return nil
}

func (d *driver) VolumeCopyAfter(
	ctx context.Context,
	result *types.Volume) {
	volDir := path.Join(vfs.VolumesDirPath(d.config), result.ID)
	os.MkdirAll(volDir, 0755)
}

func (d *driver) VolumeRemoveBefore(
	ctx *context.Context, service, volumeID string) error {
	return nil
}

func (d *driver) VolumeRemoveAfter(
	ctx context.Context, service, volumeID string) {
	volDir := path.Join(vfs.VolumesDirPath(d.config), volumeID)
	os.RemoveAll(volDir)
}

func (d *driver) VolumeSnapshotBefore(
	ctx *context.Context,
	service, volumeID string,
	request *apihttp.VolumeSnapshotRequest) error {
	return nil
}

func (d *driver) VolumeSnapshotAfter(
	ctx context.Context, result *types.Snapshot) {
}

func (d *driver) VolumeAttachBefore(
	ctx *context.Context,
	service, volumeID string,
	request *apihttp.VolumeAttachRequest) error {
	return nil
}

func (d *driver) VolumeAttachAfter(ctx context.Context, result *types.Volume) {
}

func (d *driver) VolumeDetachBefore(
	ctx *context.Context,
	service, volumeID string,
	request *apihttp.VolumeDetachRequest) error {
	return nil
}

func (d *driver) VolumeDetachAfter(
	ctx context.Context, result *types.Volume) {
}

func (d *driver) SnapshotsBefore(ctx *context.Context) error {
	return nil
}

func (d *driver) SnapshotsAfter(
	ctx context.Context, result *apihttp.ServiceSnapshotMap) {
}

func (d *driver) SnapshotsByServiceBefore(
	ctx *context.Context, service string) error {
	return nil
}

func (d *driver) SnapshotsByServiceAfter(
	ctx context.Context, service string, result *apihttp.SnapshotMap) {
}

func (d *driver) SnapshotInspectBefore(
	ctx *context.Context,
	service, snapshotID string) error {
	return nil
}

func (d *driver) SnapshotInspectAfter(
	ctx context.Context, result *types.Volume) {
}

func (d *driver) SnapshotCopyBefore(
	ctx *context.Context,
	service, snapshotID, string,
	request *apihttp.SnapshotCopyRequest) error {
	return nil
}

func (d *driver) SnapshotCopyAfter(
	ctx context.Context, result *types.Snapshot) {
}

func (d *driver) SnapshotRemoveBefore(
	ctx *context.Context, service, snapshotID string) error {
	return nil
}

func (d *driver) SnapshotRemoveAfter(ctx context.Context, snapshotID string) {
}
