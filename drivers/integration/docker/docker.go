package docker

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

const (
	providerName            = "docker"
	defaultVolumeSize int64 = 16
)

type driver struct {
	config       gofig.Config
	mountDirPath string
}

var (
	mountDirectoryPath string
)

type volumeMapping struct {
	Name             string `json:"Name"`
	VolumeMountPoint string `json:"Mountpoint"`
}

func (v *volumeMapping) VolumeName() string {
	return v.Name
}

func (v *volumeMapping) MountPoint() string {
	return v.VolumeMountPoint
}

func init() {
	registry.RegisterIntegrationDriver(providerName, newDriver)
	gofig.Register(configRegistration())
}

func newDriver() types.IntegrationDriver {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	return nil
}

func (d *driver) Name() string {
	return providerName
}

// List returns all available volume mappings.
func (d *driver) List(
	ctx types.Context,
	opts types.Store) ([]types.VolumeMapping, error) {

	vols, err := ctx.Client().Storage().Volumes(
		ctx,
		&types.VolumesOpts{
			Attachments: true,
			Opts:        opts,
		},
	)

	if err != nil {
		return nil, err
	}

	volMaps := []types.VolumeMapping{}
	for _, v := range vols {
		volMaps = append(volMaps, &volumeMapping{
			Name:             v.Name,
			VolumeMountPoint: v.MountPoint(),
		})
	}

	return volMaps, nil
}

// Inspect returns a specific volume as identified by the provided
// volume name.
func (d *driver) Inspect(
	ctx types.Context,
	volumeName string,
	opts types.Store) (types.VolumeMapping, error) {

	objs, err := d.List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var obj types.VolumeMapping
	for _, o := range objs {
		if strings.ToLower(volumeName) == strings.ToLower(o.VolumeName()) {
			obj = o
			break
		}
	}

	if obj == nil {
		return nil, utils.NewNotFoundError(volumeName)
	}

	return obj, nil
}

// Mount will return a mount point path when specifying either a volumeName
// or volumeID.  If a overwriteFs boolean is specified it will overwrite
// the FS based on newFsType if it is detected that there is no FS present.
func (d *driver) Mount(
	ctx types.Context,
	volumeID, volumeName string,
	opts *types.VolumeMountOpts) (string, *types.Volume, error) {

	ctx.WithFields(log.Fields{
		"volumeName":  volumeName,
		"volumeID":    volumeID,
		"overwriteFS": opts.OverwriteFS,
		"newFSType":   opts.NewFSType,
		"driverName":  d.Name()}).Info("mounting volume")

	vol, err := d.inspectByIDOrName(
		ctx, volumeID, volumeName, opts.Opts)
	if err != nil {
		return "", nil, err
	}

	if len(vol.Attachments) == 0 {
		mp, err := d.getVolumeMountPath(vol.Name)
		if err != nil {
			return "", nil, err
		}

		ctx.Debug("performing precautionary unmount")
		_ = ctx.Client().OS().Unmount(ctx, mp, opts.Opts)

		vol, err = ctx.Client().Storage().VolumeAttach(
			ctx, vol.ID, &types.VolumeAttachOpts{Force: opts.Preempt})
		if err != nil {
			return "", nil, err
		}
	}

	if len(vol.Attachments) == 0 {
		return "", nil, goof.New("volume did not attach")
	}

	if vol.Attachments[0].DeviceName == "" {
		return "", nil, goof.New("no device name returned")
	}

	mounts, err := ctx.Client().OS().Mounts(
		ctx, vol.Attachments[0].DeviceName, "", opts.Opts)
	if err != nil {
		return "", nil, err
	}

	if len(mounts) > 0 {
		return d.volumeMountPath(mounts[0].MountPoint), vol, nil
	}

	if opts.NewFSType == "" {
		opts.NewFSType = d.fsType()
	}

	if err := ctx.Client().OS().Format(
		ctx,
		vol.Attachments[0].DeviceName,
		&types.DeviceFormatOpts{
			NewFSType:   opts.NewFSType,
			OverwriteFS: opts.OverwriteFS,
		}); err != nil {
		return "", nil, err
	}

	mountPath, err := d.getVolumeMountPath(vol.Name)
	if err != nil {
		return "", nil, err
	}

	if err := os.MkdirAll(mountPath, 0755); err != nil {
		return "", nil, err
	}

	if err := ctx.Client().OS().Mount(
		ctx,
		vol.Attachments[0].DeviceName,
		mountPath,
		&types.DeviceMountOpts{}); err != nil {
		return "", nil, err
	}

	return d.volumeMountPath(mountPath), vol, nil
}

// Unmount will unmount the specified volume by volumeName or volumeID.
func (d *driver) Unmount(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) error {
	return nil
}

// Path will return the mounted path of the volumeName or volumeID.
func (d *driver) Path(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (string, error) {
	return "", nil
}

// Create will create a new volume with the volumeName and opts.
func (d *driver) Create(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	return nil, nil
}

// Remove will remove a volume of volumeName.
func (d *driver) Remove(
	ctx types.Context,
	volumeName string,
	opts types.Store) error {
	return nil
}

// Attach will attach a volume based on volumeName to the instance of
// instanceID.
func (d *driver) Attach(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeAttachOpts) (string, error) {
	return "", nil
}

// Detach will detach a volume based on volumeName to the instance of
// instanceID.
func (d *driver) Detach(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeDetachOpts) error {
	return nil
}

// NetworkName will return an identifier of a volume that is relevant when
// corelating a local device to a device that is the volumeName to the
// local instanceID.
func (d *driver) NetworkName(
	ctx types.Context,
	volumeName string,
	opts types.Store) (string, error) {
	return "", nil
}

func (d *driver) volumeRootPath() string {
	return d.config.GetString("linux.volume.rootPath")
}

func (d *driver) volumeType() string {
	return d.config.GetString("docker.volumeType")
}

func (d *driver) iops() string {
	return d.config.GetString("docker.iops")
}

func (d *driver) size() string {
	return d.config.GetString("docker.size")
}

func (d *driver) availabilityZone() string {
	return d.config.GetString("docker.availabilityZone")
}

func (d *driver) fsType() string {
	return d.config.GetString("docker.fsType")
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("Docker")
	r.Key(gofig.String, "", "ext4", "", "docker.fsType")
	r.Key(gofig.String, "", "", "", "docker.volumeType")
	r.Key(gofig.String, "", "", "", "docker.iops")
	r.Key(gofig.String, "", "", "", "docker.size")
	r.Key(gofig.String, "", "", "", "docker.availabilityZone")
	r.Key(gofig.String, "", "/data", "", "linux.volume.rootpath")
	return r
}
