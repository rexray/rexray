package linux

import (
	"os"
	"path"

	"fmt"

	log "github.com/sirupsen/logrus"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	apiconfig "github.com/AVENTER-UG/rexray/libstorage/api/utils/config"
)

const (
	providerName            = "linux"
	defaultVolumeSize int64 = 16
)

var ctxExactMountKey = interface{}("exactmount")

type driver struct {
	config gofig.Config
}

type volumeMapping struct {
	Name             string                 `json:"Name"`
	VolumeMountPoint string                 `json:"Mountpoint"`
	VolumeStatus     map[string]interface{} `json:"Status"`
}

func (v *volumeMapping) VolumeName() string {
	return v.Name
}

func (v *volumeMapping) MountPoint() string {
	return v.VolumeMountPoint
}

func (v *volumeMapping) Status() map[string]interface{} {
	return v.VolumeStatus
}

func init() {
	registry.RegisterIntegrationDriver(providerName, newDriver)
}

func newDriver() types.IntegrationDriver {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config

	ctx.WithFields(log.Fields{
		types.ConfigIgVolOpsMountRootPath:       d.volumeRootPath(),
		types.ConfigIgVolOpsCreateDefaultType:   d.volumeType(),
		types.ConfigIgVolOpsCreateDefaultIOPS:   d.iops(),
		types.ConfigIgVolOpsCreateDefaultSize:   d.size(),
		types.ConfigIgVolOpsCreateDefaultAZ:     d.availabilityZone(),
		types.ConfigIgVolOpsCreateDefaultFsType: d.fsType(),
		types.ConfigIgVolOpsMountPath:           d.mountDirPath(),
		types.ConfigIgVolOpsCreateImplicit:      d.volumeCreateImplicit(),
	}).Info("linux integration driver successfully initialized")

	return nil
}

func (d *driver) Name() string {
	return providerName
}

func buildVolumeStatus(v *types.Volume, service string) map[string]interface{} {
	vs := make(map[string]interface{})
	vs["name"] = v.Name
	vs["size"] = v.Size
	vs["iops"] = v.IOPS
	vs["type"] = v.Type
	vs["availabilityZone"] = v.AvailabilityZone
	vs["fields"] = v.Fields
	vs["service"] = service
	vs["server"] = service

	return vs
}

// List returns all available volume mappings.
func (d *driver) List(
	ctx types.Context,
	opts types.Store) ([]types.VolumeMapping, error) {

	client := context.MustClient(ctx)
	vols, err := client.Storage().Volumes(
		ctx,
		&types.VolumesOpts{
			Attachments: opts.GetAttachments(),
			Opts:        opts,
		},
	)
	if err != nil {
		return nil, err
	}

	serviceName, serviceNameOK := context.ServiceName(ctx)
	if !serviceNameOK {
		return nil, goof.New("service name is missing")
	}

	volMaps := []types.VolumeMapping{}
	for _, v := range vols {
		vs := buildVolumeStatus(v, serviceName)
		volMaps = append(volMaps, &volumeMapping{
			Name:             v.Name,
			VolumeMountPoint: v.MountPoint(),
			VolumeStatus:     vs,
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

	fields := log.Fields{
		"volumeName": volumeName,
		"opts":       opts}
	ctx.WithFields(fields).Info("inspecting volume")

	serviceName, serviceNameOK := context.ServiceName(ctx)
	if !serviceNameOK {
		return nil, goof.New("service name is missing")
	}

	vol, err := d.volumeInspectByName(
		ctx,
		volumeName,
		opts.GetAttachments(),
		opts,
	)
	if err != nil {
		return nil, err
	}
	if vol == nil {
		return nil, utils.NewNotFoundError(volumeName)
	}
	vs := buildVolumeStatus(vol, serviceName)
	obj := &volumeMapping{
		Name:             vol.Name,
		VolumeMountPoint: vol.MountPoint(),
		VolumeStatus:     vs,
	}

	fields = log.Fields{
		"volumeName": volumeName,
		"volume":     obj}
	ctx.WithFields(fields).Info("volume inspected")

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
		"volumeName": volumeName,
		"volumeID":   volumeID,
		"opts":       opts}).Info("mounting volume")

	lsAtt := types.VolAttReqWithDevMapOnlyVolsAttachedToInstanceOrUnattachedVols
	if opts.Preempt {
		lsAtt = types.VolAttReqWithDevMapForInstance
	}

	vol, err := d.volumeInspectByIDOrName(
		ctx, volumeID, volumeName, lsAtt, opts.Opts)
	if isErrNotFound(err) && d.volumeCreateImplicit() {
		if vol, err = d.Create(ctx, volumeName, &types.VolumeCreateOpts{
			Opts: utils.NewStore(),
		}); err != nil {
			return "", nil, goof.WithError(
				"problem creating volume implicitly", err)
		}
	} else if err != nil {
		return "", nil, err
	}

	if vol == nil {
		return "", nil, goof.New("no volume returned or created")
	}

	client := context.MustClient(ctx)
	if vol.AttachmentState == types.VolumeAvailable ||
		(opts.Preempt && vol.AttachmentState != types.VolumeAttached) {

		mp, err := d.getVolumeMountPath(vol.Name)
		if err != nil {
			return "", nil, err
		}

		ctx.Debug("performing precautionary unmount")
		_ = client.OS().Unmount(ctx, mp, opts.Opts)

		var token string
		vol, token, err = client.Storage().VolumeAttach(
			ctx, vol.ID, &types.VolumeAttachOpts{
				Force: opts.Preempt,
				Opts:  utils.NewStore(),
			})
		if err != nil {
			return "", nil, err
		}

		if token != "" {
			opts := &types.WaitForDeviceOpts{
				LocalDevicesOpts: types.LocalDevicesOpts{
					ScanType: apiconfig.DeviceScanType(d.config),
					Opts:     opts.Opts,
				},
				Token:   token,
				Timeout: apiconfig.DeviceAttachTimeout(d.config),
			}

			_, _, err = client.Executor().WaitForDevice(ctx, opts)
			if err != nil {
				return "", nil, goof.WithError(
					"problem with device discovery", err)
			}
		}

		vol, err = d.volumeInspectByIDOrName(
			ctx, vol.ID, "", types.VolAttReqTrue, opts.Opts)
		if err != nil {
			return "", nil, err
		}

	}

	if len(vol.Attachments) == 0 {
		return "", nil, goof.New("volume did not attach")
	}

	inst, err := client.Storage().InstanceInspect(ctx, utils.NewStore())
	if err != nil {
		return "", nil, goof.New("problem getting instance ID")
	}
	var ma *types.VolumeAttachment
	for _, att := range vol.Attachments {
		if att.InstanceID.ID == inst.InstanceID.ID {
			ma = att
			break
		}
	}

	if ma == nil {
		return "", nil, goof.New("no local attachment found")
	}

	if ma.DeviceName == "" {
		return "", nil, goof.New("no device name returned")
	}

	mountPath, err := d.getVolumeMountPath(vol.Name)
	if err != nil {
		return "", nil, err
	}

	mounts, err := client.OS().Mounts(
		ctx, ma.DeviceName, "", opts.Opts)
	if err != nil {
		return "", nil, err
	}

	if len(mounts) > 0 {
		if _, ok := ctx.Value(ctxExactMountKey).(interface{}); ok {
			for _, m := range mounts {
				if m.MountPoint == mountPath {
					ctx.Debug("returning existing mount")
					return d.volumeMountPath(
						m.MountPoint), vol, nil
				}
			}
			return "", nil, goof.New("device is already mounted")
		}
		return d.volumeMountPath(mounts[0].MountPoint), vol, nil
	}

	if opts.NewFSType == "" {
		opts.NewFSType = d.fsType()
	}
	if err := client.OS().Format(
		ctx,
		ma.DeviceName,
		&types.DeviceFormatOpts{
			NewFSType:   opts.NewFSType,
			OverwriteFS: opts.OverwriteFS,
		}); err != nil {
		return "", nil, err
	}

	if err := os.MkdirAll(mountPath, 0755); err != nil {
		return "", nil, err
	}

	if err := client.OS().Mount(
		ctx,
		ma.DeviceName,
		mountPath,
		&types.DeviceMountOpts{}); err != nil {
		return "", nil, err
	}

	mntPath := d.volumeMountPath(mountPath)

	fields := log.Fields{
		"vol":     vol,
		"mntPath": mntPath,
	}
	ctx.WithFields(fields).Info("volume mounted")

	return mntPath, vol, nil
}

// Unmount will unmount the specified volume by volumeName or volumeID.
func (d *driver) Unmount(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {

	ctx.WithFields(log.Fields{
		"volumeName": volumeName,
		"volumeID":   volumeID,
		"opts":       opts}).Info("unmounting volume")

	if volumeName == "" && volumeID == "" {
		return nil, goof.New("missing volume name or ID")
	}

	vol, err := d.volumeInspectByIDOrName(
		ctx, volumeID, volumeName,
		types.VolAttReqWithDevMapOnlyVolsAttachedToInstance, opts)
	if err != nil {
		return nil, err
	}

	if len(vol.Attachments) == 0 {
		return nil, nil
	}

	client := context.MustClient(ctx)

	inst, err := client.Storage().InstanceInspect(ctx, utils.NewStore())
	if err != nil {
		return nil, goof.New("problem getting instance ID")
	}
	var ma *types.VolumeAttachment
	for _, att := range vol.Attachments {
		if att.InstanceID.ID == inst.InstanceID.ID {
			ma = att
			break
		}
	}

	if ma == nil {
		return nil, goof.New("no attachment found for instance")
	}

	if ma.DeviceName == "" {
		return nil, goof.New("no device name found for attachment")
	}

	mounts, err := client.OS().Mounts(
		ctx, ma.DeviceName, "", opts)
	if err != nil {
		return nil, err
	}

	for _, mount := range mounts {
		ctx.WithField("mount", mount).Debug("retrieved mount")
	}

	if len(mounts) > 0 {
		for _, mount := range mounts {
			ctx.WithField("mount", mount).Debug("unmounting mount point")
			err = client.OS().Unmount(ctx, mount.MountPoint, opts)
			if err != nil {
				return nil, err
			}
		}
	}

	vol, err = client.Storage().VolumeDetach(ctx, vol.ID,
		&types.VolumeDetachOpts{
			Force: opts.GetBool("force"),
			Opts:  utils.NewStore(),
		})
	if err != nil {
		return nil, err
	}

	ctx.WithFields(log.Fields{
		"vol": vol}).Info("unmounted and detached volume")

	return vol, nil
}

// Path will return the mounted path of the volumeName or volumeID.
func (d *driver) Path(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (string, error) {

	ctx.WithFields(log.Fields{
		"volumeName": volumeName,
		"volumeID":   volumeID,
		"opts":       opts}).Info("getting path to volume")

	vol, err := d.volumeInspectByIDOrName(
		ctx, volumeID, volumeName, types.VolAttReqTrue, opts)
	if err != nil {
		return "", err
	} else if vol == nil {
		return "", utils.NewNotFoundError(
			fmt.Sprintf("volumeID=%s,volumeName=%s", volumeID, volumeName))
	}

	if len(vol.Attachments) == 0 {
		return "", nil
	}

	client := context.MustClient(ctx)

	mounts, err := client.OS().Mounts(
		ctx, vol.Attachments[0].DeviceName, "", opts)
	if err != nil {
		return "", err
	}

	if len(mounts) == 0 {
		return "", nil
	}

	volPath := d.volumeMountPath(mounts[0].MountPoint)

	ctx.WithFields(log.Fields{
		"volPath": volPath,
		"vol":     vol}).Info("returning path to volume")

	return volPath, nil
}

// Create will create a new volume with the volumeName and opts.
func (d *driver) Create(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	if volumeName == "" {
		return nil, goof.New("missing volume name or ID")
	}

	optsNew := &types.VolumeCreateOpts{}
	az := d.availabilityZone()
	optsNew.AvailabilityZone = &az
	size := d.size()
	optsNew.Size = &size
	volumeType := d.volumeType()
	optsNew.Type = &volumeType
	iops := d.iops()
	optsNew.IOPS = &iops
	optsNew.Encrypted = opts.Encrypted
	optsNew.EncryptionKey = opts.EncryptionKey

	if opts.Opts.IsSet("availabilityZone") {
		az = opts.Opts.GetString("availabilityZone")
	}
	if opts.Opts.IsSet("size") {
		size = opts.Opts.GetInt64("size")
	}
	if opts.Opts.IsSet("volumeType") {
		volumeType = opts.Opts.GetString("volumeType")
	}
	if opts.Opts.IsSet("type") {
		volumeType = opts.Opts.GetString("type")
	}
	if opts.Opts.IsSet("iops") {
		iops = opts.Opts.GetInt64("iops")
	}

	optsNew.Opts = opts.Opts

	ctx.WithFields(log.Fields{
		"volumeName":       volumeName,
		"availabilityZone": az,
		"size":             size,
		"volumeType":       volumeType,
		"IOPS":             iops,
		"encrypted":        optsNew.Encrypted,
		"encryptionKey":    optsNew.EncryptionKey,
		"opts":             opts}).Info("creating volume")

	client := context.MustClient(ctx)
	vol, err := client.Storage().VolumeCreate(ctx, volumeName, optsNew)
	if err != nil {
		return nil, err
	}

	ctx.WithFields(log.Fields{
		"volumeName": volumeName,
		"vol":        vol}).Info("volume created")

	return vol, nil
}

// Remove will remove a volume of volumeName.
func (d *driver) Remove(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeRemoveOpts) error {

	if volumeName == "" {
		return goof.New("missing volume name or ID")
	}

	vol, err := d.volumeInspectByIDOrName(
		ctx, "", volumeName, 0, opts.Opts)
	if err != nil {
		return err
	}

	if vol == nil {
		return goof.New("volume not found")
	}

	client := context.MustClient(ctx)

	// check to see if there is a config override for force remove
	if !opts.Force {
		opts.Force = d.volumeRemoveForce()
	}

	return client.Storage().VolumeRemove(ctx, vol.ID, opts)
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
	return d.config.GetString(types.ConfigIgVolOpsMountRootPath)
}

func (d *driver) volumeType() string {
	return d.config.GetString(types.ConfigIgVolOpsCreateDefaultType)
}

func (d *driver) iops() int64 {
	return int64(d.config.GetInt(types.ConfigIgVolOpsCreateDefaultIOPS))
}

func (d *driver) size() int64 {
	return int64(d.config.GetInt(types.ConfigIgVolOpsCreateDefaultSize))
}

func (d *driver) availabilityZone() string {
	return d.config.GetString(types.ConfigIgVolOpsCreateDefaultAZ)
}

func (d *driver) fsType() string {
	return d.config.GetString(types.ConfigIgVolOpsCreateDefaultFsType)
}

func (d *driver) mountDirPath() string {
	return d.config.GetString(types.ConfigIgVolOpsMountPath)
}

func (d *driver) volumeCreateImplicit() bool {
	return d.config.GetBool(types.ConfigIgVolOpsCreateImplicit)
}

func (d *driver) volumeRemoveForce() bool {
	return d.config.GetBool(types.ConfigIgVolOpsRemoveForce)
}

// register the gofig configuration
func init() {
	registry.RegisterConfigReg(
		"Integration",
		func(ctx types.Context, r gofig.ConfigRegistration) {

			r.Key(
				gofig.String,
				"", "ext4", "",
				types.ConfigIgVolOpsCreateDefaultFsType)

			r.Key(
				gofig.String,
				"", "", "", types.ConfigIgVolOpsCreateDefaultType)

			r.Key(
				gofig.String,
				"", "", "",
				types.ConfigIgVolOpsCreateDefaultIOPS)

			r.Key(
				gofig.String,
				"", "16", "",
				types.ConfigIgVolOpsCreateDefaultSize)

			r.Key(
				gofig.String,
				"", "", "",
				types.ConfigIgVolOpsCreateDefaultAZ)

			r.Key(
				gofig.String,
				"",
				path.Join(context.MustPathConfig(ctx).Lib, "volumes"),
				"",
				types.ConfigIgVolOpsMountPath)

			r.Key(
				gofig.String,
				"", "/data", "",
				types.ConfigIgVolOpsMountRootPath)

			r.Key(
				gofig.Bool,
				"", true, "",
				types.ConfigIgVolOpsCreateImplicit)

			r.Key(
				gofig.Bool,
				"", false, "",
				types.ConfigIgVolOpsMountPreempt)
		})
}
