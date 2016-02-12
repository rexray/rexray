package docker

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/docker/docker/pkg/mount"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/util"
)

const (
	providerName            = "docker"
	defaultVolumeSize int64 = 16
)

type driver struct {
	r *core.RexRay
}

var (
	mountDirectoryPath string
)

func init() {
	core.RegisterDriver(providerName, newDriver)
	gofig.Register(configRegistration())
	mountDirectoryPath = util.LibFilePath("volumes")
	os.MkdirAll(mountDirectoryPath, 0755)
}

func newDriver() core.Driver {
	return &driver{}
}

func ef() goof.Fields {
	return goof.Fields{
		"provider": providerName,
	}
}

func eff(fields goof.Fields) map[string]interface{} {
	errFields := map[string]interface{}{
		"provider": providerName,
	}
	if fields != nil {
		for k, v := range fields {
			errFields[k] = v
		}
	}
	return errFields
}

func (d *driver) Init(r *core.RexRay) error {
	d.r = r

	fields := eff(map[string]interface{}{
		"volumeType":       d.volumeType(),
		"iops":             d.iops(),
		"size":             d.size(),
		"availabilityZone": d.availabilityZone(),
		"volumeRootPath":   d.volumeRootPath(),
		"provider":         providerName,
	})

	log.WithFields(fields).Info("docker volume driver initialized")

	return nil
}

func getVolumeMountPath(name string) (string, error) {
	if name == "" {
		return "", goof.New("Missing volume name")
	}

	return fmt.Sprintf("%s/%s", mountDirectoryPath, name), nil
}

// Name will return the name of the volume driver manager
func (d *driver) Name() string {
	return providerName
}

// Mount will perform the steps to get an existing Volume with or without a fileystem mounted to a guest
func (d *driver) Mount(volumeName, volumeID string, overwriteFs bool, newFsType string, preempt bool) (string, error) {
	log.WithFields(log.Fields{
		"moduleName":  d.r.Context,
		"volumeName":  volumeName,
		"volumeID":    volumeID,
		"overwriteFs": overwriteFs,
		"newFsType":   newFsType,
		"driverName":  d.Name()}).Info("mounting volume")

	var err error
	var vols []*core.Volume
	var volAttachments []*core.VolumeAttachment
	var instance *core.Instance

	if vols, volAttachments, instance, err = d.prefixToMountUnmount(
		volumeName, volumeID); err != nil {
		return "", err
	}

	if len(volAttachments) == 0 {
		mp, err := getVolumeMountPath(vols[0].Name)
		if err != nil {
			return "", err
		}

		log.Debug("performing precautionary unmount")
		_ = d.r.OS.Unmount(mp)

		volAttachments, err = d.r.Storage.AttachVolume(
			false, vols[0].VolumeID, instance.InstanceID, preempt)
		if err != nil {
			return "", err
		}
	}

	if len(volAttachments) == 0 {
		return "", goof.New("Volume did not attach")
	}

	if volAttachments[0].DeviceName == "" {
		return "", goof.New("no device name returned")
	}

	mounts, err := d.r.OS.GetMounts(volAttachments[0].DeviceName, "")
	if err != nil {
		return "", err
	}

	if len(mounts) > 0 {
		return d.volumeMountPath(mounts[0].Mountpoint), nil
	}

	switch {
	case d.volumeType() != "":
		newFsType = d.volumeType()
	case newFsType == "":
		newFsType = "ext4"
	}

	if err := d.r.OS.Format(
		volAttachments[0].DeviceName, newFsType, overwriteFs); err != nil {
		return "", err
	}

	mountPath, err := getVolumeMountPath(vols[0].Name)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(mountPath, 0755); err != nil {
		return "", err
	}

	if err := d.r.OS.Mount(
		volAttachments[0].DeviceName, mountPath, "", ""); err != nil {
		return "", err
	}

	return d.volumeMountPath(mountPath), nil
}

// Unmount will perform the steps to unmount and existing volume and detach
func (d *driver) Unmount(volumeName, volumeID string) error {

	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"volumeName": volumeName,
		"volumeID":   volumeID,
		"driverName": d.Name()}).Info("unmounting volume")

	var err error
	var vols []*core.Volume
	var volAttachments []*core.VolumeAttachment

	if vols, volAttachments, _, err = d.prefixToMountUnmount(
		volumeName, volumeID); err != nil {
		return err
	}

	if len(volAttachments) == 0 {
		return nil
	}

	mounts, err := d.r.OS.GetMounts(volAttachments[0].DeviceName, "")
	if err != nil {
		return err
	}

	if len(mounts) > 0 {
		err := d.r.OS.Unmount(mounts[0].Mountpoint)
		if err != nil {
			return err
		}
	}

	err = d.r.Storage.DetachVolume(false, vols[0].VolumeID, "", false)
	if err != nil {
		return err
	}
	return nil
}

func (d *driver) getInstance() (*core.Instance, error) {
	instances, err := d.r.Storage.GetInstances()
	if err != nil {
		return nil, err
	}

	switch {
	case len(instances) == 0:
		return nil, goof.New("No instances")
	case len(instances) > 1:
		return nil,
			goof.New("Too many instances returned, limit the storagedrivers")
	}

	return instances[0], nil
}

func (d *driver) prefixToMountUnmount(
	volumeName,
	volumeID string) ([]*core.Volume, []*core.VolumeAttachment, *core.Instance, error) {
	if volumeName == "" && volumeID == "" {
		return nil, nil, nil, goof.New("Missing volume name or ID")
	}

	var instance *core.Instance
	var err error
	if instance, err = d.getInstance(); err != nil {
		return nil, nil, nil, err
	}

	var vols []*core.Volume
	if vols, err = d.r.Storage.GetVolume(volumeID, volumeName); err != nil {
		return nil, nil, nil, err
	}

	switch {
	case len(vols) == 0:
		return nil, nil, nil, goof.New("No volumes returned by name")
	case len(vols) > 1:
		return nil, nil, nil, goof.New("Multiple volumes returned by name")
	}

	var volAttachments []*core.VolumeAttachment
	if volAttachments, err = d.r.Storage.GetVolumeAttach(
		vols[0].VolumeID, instance.InstanceID); err != nil {
		return nil, nil, nil, err
	}

	return vols, volAttachments, instance, nil
}

// Path returns the mounted path of the volume
func (d *driver) Path(volumeName, volumeID string) (string, error) {
	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"volumeName": volumeName,
		"volumeID":   volumeID,
		"driverName": d.Name()}).Info("getting path to volume")
	if volumeName == "" && volumeID == "" {
		return "", goof.New("Missing volume name or ID")
	}

	instances, err := d.r.Storage.GetInstances()
	if err != nil {
		return "", err
	}

	switch {
	case len(instances) == 0:
		return "", goof.New("No instances")
	case len(instances) > 1:
		return "", goof.New("Too many instances returned, limit the storagedrivers")
	}

	volumes, err := d.r.Storage.GetVolume(volumeID, volumeName)
	if err != nil {
		return "", err
	}

	switch {
	case len(volumes) == 0:
		return "", goof.New("No volumes returned by name")
	case len(volumes) > 1:
		return "", goof.New("Multiple volumes returned by name")
	}

	volumeAttachment, err := d.r.Storage.GetVolumeAttach(volumes[0].VolumeID, instances[0].InstanceID)
	if err != nil {
		return "", err
	}

	if len(volumeAttachment) == 0 {
		return "", nil
	}

	mounts, err := d.r.OS.GetMounts(volumeAttachment[0].DeviceName, "")
	if err != nil {
		return "", err
	}

	if len(mounts) == 0 {
		return "", nil
	}

	return d.volumeMountPath(mounts[0].Mountpoint), nil
}

// Create will create a remote volume
func (d *driver) Create(volumeName string, volumeOpts core.VolumeOpts) error {
	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"volumeName": volumeName,
		"volumeOpts": volumeOpts,
		"driverName": d.Name()}).Info("creating volume")

	if volumeName == "" {
		return goof.New("Missing volume name")
	}

	var err error

	if err = d.createGetInstance(); err != nil {
		return err
	}

	for k, v := range volumeOpts {
		volumeOpts[strings.ToLower(k)] = v
	}
	newFsType := volumeOpts["newfstype"]

	var overwriteFs bool
	var volumes []*core.Volume

	volumes, overwriteFs, err = d.createGetVolumes(volumeName, volumeOpts)
	if err != nil {
		return err
	}

	if len(volumes) > 0 {
		return nil
	}

	var volFrom *core.Volume
	var volumeID string
	if volFrom, err = d.createInitVolume(
		volumeName, volumeOpts); err != nil {
		return err
	} else if volFrom != nil {
		volumeID = volFrom.VolumeID
	}

	var snapFrom *core.Snapshot
	var snapshotID string
	if snapFrom, err = d.createGetSnapshot(volumeOpts); err != nil {
		return err
	} else if snapFrom != nil {
		snapshotID = snapFrom.SnapshotID
	}

	volumeType := d.createInitVolumeType(volumeOpts, volFrom)
	IOPS := d.createInitIOPS(volumeOpts, volFrom)
	size := d.createInitSize(volumeOpts, volFrom, snapFrom)
	availabilityZone := d.createInitAvailabilityZone(volumeOpts)

	if len(volumes) == 0 {
		if _, err = d.r.Storage.CreateVolume(
			false, volumeName, volumeID, snapshotID,
			volumeType, IOPS, size, availabilityZone); err != nil {
			return err
		}
	}

	if newFsType != "" || overwriteFs {
		_, err = d.Mount(volumeName, "", overwriteFs, newFsType, false)
		if err != nil {
			log.WithFields(log.Fields{
				"volumeName":  volumeName,
				"overwriteFs": overwriteFs,
				"newFsType":   newFsType,
				"driverName":  d.Name()}).Error("Failed to create or mount file system")
		}
		err = d.Unmount(volumeName, "")
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *driver) createInitVolume(
	volumeName string,
	volumeOpts core.VolumeOpts) (*core.Volume, error) {

	var optVolumeName string
	var optVolumeID string

	optVolumeName, _ = volumeOpts["volumename"]
	optVolumeID, _ = volumeOpts["volumeid"]

	if optVolumeName == "" && optVolumeID == "" {
		return nil, nil
	}

	var err error
	var volumes []*core.Volume
	if volumes, err = d.r.Storage.GetVolume(optVolumeID, optVolumeName); err != nil {
		return nil, err
	}

	switch {
	case len(volumes) == 0:
		return nil, goof.WithField(
			"optVolumeName", optVolumeName, "No volumes returned")
	case len(volumes) > 1:
		return nil, goof.WithField(
			"optVolumeName", optVolumeName, "Too many volumes returned")
	}

	return volumes[0], nil
}

func (d *driver) createGetSnapshot(
	volumeOpts core.VolumeOpts) (*core.Snapshot, error) {

	var optSnapshotName string
	var optSnapshotID string

	optSnapshotName, _ = volumeOpts["snapshotname"]
	optSnapshotID, _ = volumeOpts["snapshotid"]

	if optSnapshotName == "" && optSnapshotID == "" {
		return nil, nil
	}

	var err error
	var snapshots []*core.Snapshot

	if snapshots, err = d.r.Storage.GetSnapshot(
		"", optSnapshotID, optSnapshotName); err != nil {
		return nil, err
	}

	switch {
	case len(snapshots) == 0:
		return nil, goof.WithField(
			"optSnapshotName", optSnapshotName, "No snapshots returned")
	case len(snapshots) > 1:
		return nil, goof.WithField(
			"optSnapshotName", optSnapshotName, "Too many snapshots returned")
	}

	return snapshots[0], nil
}

func (d *driver) createGetInstance() error {
	var err error
	var instances []*core.Instance

	if instances, err = d.r.Storage.GetInstances(); err != nil {
		return err
	}

	switch {
	case len(instances) == 0:
		return goof.New("No instances")
	case len(instances) > 1:
		return goof.New(
			"Too many instances returned, limit the storagedrivers")
	}

	return nil
}

func (d *driver) createGetVolumes(
	volumeName string,
	volumeOpts core.VolumeOpts) ([]*core.Volume, bool, error) {
	var err error
	var volumes []*core.Volume

	if volumes, err = d.r.Storage.GetVolume("", volumeName); err != nil {
		return nil, false, err
	}

	overwriteFs, _ := strconv.ParseBool(volumeOpts["overwritefs"])

	switch {
	case len(volumes) == 1 && !overwriteFs:
		return volumes, overwriteFs, nil
	case len(volumes) > 1:
		return nil, overwriteFs, goof.WithField(
			"volumeName", volumeName, "Too many volumes returned")
	}

	return volumes, overwriteFs, nil
}

func (d *driver) createInitVolumeType(volumeOpts core.VolumeOpts, volume *core.Volume) string {
	var ok bool
	var volumeType string
	if volumeType, ok = volumeOpts["volumetype"]; ok {
		return volumeType
	} else if volume != nil {
		return volume.VolumeType
	} else if volumeType, ok = d.createInitEnv(d.volumeType()); ok {
		return volumeType
	}
	return ""
}

func (d *driver) createInitIOPS(volumeOpts core.VolumeOpts, volume *core.Volume) int64 {
	if ok, i := d.createInitInt64("iops", "", volumeOpts); ok {
		return i
	} else if volume != nil {
		return volume.IOPS
	} else if ok, i := d.createInitInt64("", d.iops(), volumeOpts); ok {
		return i
	}
	return 0
}

func (d *driver) createInitSize(volumeOpts core.VolumeOpts, volume *core.Volume, snapshot *core.Snapshot) int64 {
	if ok, i := d.createInitInt64("size", "", volumeOpts); ok {
		return i
	} else if volume != nil {
		sizei, _ := strconv.Atoi(volume.Size)
		return int64(sizei)
	} else if snapshot != nil {
		sizei, _ := strconv.Atoi(snapshot.VolumeSize)
		return int64(sizei)
	} else if ok, i := d.createInitInt64("", d.size(), volumeOpts); ok {
		return i
	}

	return defaultVolumeSize
}

func (d *driver) createInitInt64(
	optKey, envVar string,
	volumeOpts core.VolumeOpts) (bool, int64) {

	atoi := func(v string) int64 {
		sizei, _ := strconv.Atoi(v)
		return int64(sizei)
	}

	if envVar != "" {
		return true, atoi(envVar)
	}

	if s, k := volumeOpts[optKey]; k {
		return true, atoi(s)
	}

	return false, 0
}

func (d *driver) createInitEnv(envVal string) (string, bool) {
	if envVal == "" {
		return "", false
	}
	return envVal, true
}

func (d *driver) createInitAvailabilityZone(volumeOpts core.VolumeOpts) string {
	return d.createInitString(
		"availabilityzone", d.availabilityZone(), volumeOpts)
}

func (d *driver) createInitString(optKey, envVar string,
	volumeOpts core.VolumeOpts) string {
	var ok bool
	var s string
	if s, ok = volumeOpts[optKey]; ok {
		return s
	}
	return d.availabilityZone()
}

// Remove will remove a remote volume
func (d *driver) Remove(volumeName string) error {
	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"volumeName": volumeName,
		"driverName": d.Name()}).Info("removing volume")

	if volumeName == "" {
		return goof.New("Missing volume name")
	}

	instances, err := d.r.Storage.GetInstances()
	if err != nil {
		return err
	}

	switch {
	case len(instances) == 0:
		return goof.New("No instances")
	case len(instances) > 1:
		return goof.New("Too many instances returned, limit the storagedrivers")
	}

	volumes, err := d.r.Storage.GetVolume("", volumeName)
	if err != nil {
		return err
	}

	switch {
	case len(volumes) == 0:
		return goof.New("No volumes returned by name")
	case len(volumes) > 1:
		return goof.New("Multiple volumes returned by name")
	}

	err = d.Unmount("", volumes[0].VolumeID)
	if err != nil {
		return err
	}

	err = d.r.Storage.RemoveVolume(volumes[0].VolumeID)
	if err != nil {
		return err
	}

	return nil
}

func (d *driver) Get(volumeName string) (core.VolumeMap, error) {
	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"volumeName": volumeName,
		"driverName": d.Name()}).Info("getting volume")

	if volumeName == "" {
		return nil, goof.New("Missing volume name")
	}

	volumes, err := d.get(volumeName)
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, goof.New("no volumes returned")
	}

	return volumes[0], nil
}

func (d *driver) get(volumeName string) ([]core.VolumeMap, error) {
	instances, err := d.r.Storage.GetInstances()
	if err != nil {
		return nil, err
	}

	switch {
	case len(instances) == 0:
		return nil, goof.New("No instances")
	case len(instances) > 1:
		return nil, goof.New("Too many instances returned, limit the storagedrivers")
	}

	volumes, err := d.r.Storage.GetVolume("", volumeName)
	if err != nil {
		return nil, err
	}

	switch {
	case len(volumes) == 0:
		return nil, goof.New("No volumes returned")
	case len(volumes) > 1 && volumeName != "":
		return nil, goof.New("too many volumes returned")
	}

	var volList []core.VolumeMap

Volumes:
	for _, volume := range volumes {
		if len(volume.Attachments) == 0 {
			volList = append(volList, core.VolumeMap{
				"Name": volume.Name,
			})
			continue Volumes
		}

		var attachment *core.VolumeAttachment
	Attachments:
		for _, att := range volume.Attachments {
			if att.InstanceID == instances[0].InstanceID {
				attachment = att
				break Attachments
			}
		}

		if attachment == nil {
			volList = append(volList, core.VolumeMap{
				"Name": volume.Name,
			})
			continue Volumes
		}

		allMounts, err := d.r.OS.GetMounts("", "")
		if err != nil {
			return nil, err
		}

		var mounts []*mount.Info
		for _, mount := range allMounts {
			if mount.Source == attachment.DeviceName {
				mounts = append(mounts, mount)
			}
		}

		if len(mounts) == 0 {
			volList = append(volList, core.VolumeMap{
				"Name": volume.Name,
			})
			continue Volumes
		}

		volList = append(volList, core.VolumeMap{
			"Name":       volume.Name,
			"Mountpoint": d.volumeMountPath(mounts[0].Mountpoint),
		})

	}
	return volList, nil
}

// List will list all volumes
func (d *driver) List() ([]core.VolumeMap, error) {
	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"driverName": d.Name()}).Info("listing volumes")

	return d.get("")
}

// Attach will attach a volume to an instance
func (d *driver) Attach(volumeName, instanceID string, force bool) (string, error) {
	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"volumeName": volumeName,
		"instanceID": instanceID,
		"driverName": d.Name()}).Info("attaching volume")

	volumes, err := d.r.Storage.GetVolume("", volumeName)
	if err != nil {
		return "", err
	}

	switch {
	case len(volumes) == 0:
		return "", goof.New("No volumes returned by name")
	case len(volumes) > 1:
		return "", goof.New("Multiple volumes returned by name")
	}

	_, err = d.r.Storage.AttachVolume(true, volumes[0].VolumeID, instanceID, force)
	if err != nil {
		return "", err
	}

	volumes, err = d.r.Storage.GetVolume("", volumeName)
	if err != nil {
		return "", err
	}

	return volumes[0].NetworkName, nil
}

// Remove will remove a remote volume
func (d *driver) Detach(volumeName, instanceID string, force bool) error {
	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"volumeName": volumeName,
		"instanceID": instanceID,
		"driverName": d.Name()}).Info("detaching volume")

	volume, err := d.r.Storage.GetVolume("", volumeName)
	if err != nil {
		return err
	}

	return d.r.Storage.DetachVolume(true, volume[0].VolumeID, instanceID, force)
}

// NetworkName will return relevant information about how a volume can be discovered on an OS
func (d *driver) NetworkName(volumeName, instanceID string) (string, error) {
	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"volumeName": volumeName,
		"instanceID": instanceID,
		"driverName": d.Name()}).Info("returning network name")

	volumes, err := d.r.Storage.GetVolume("", volumeName)
	if err != nil {
		return "", err
	}

	switch {
	case len(volumes) == 0:
		return "", goof.New("No volumes returned by name")
	case len(volumes) > 1:
		return "", goof.New("Multiple volumes returned by name")
	}

	volumeAttachment, err := d.r.Storage.GetVolumeAttach(
		volumes[0].VolumeID, instanceID)
	if err != nil {
		return "", err
	}

	if len(volumeAttachment) == 0 {
		return "", goof.New("Volume not attached")
	}

	volumes, err = d.r.Storage.GetVolume("", volumeName)
	if err != nil {
		return "", err
	}

	return volumes[0].NetworkName, nil
}

func (d *driver) volumeMountPath(target string) string {
	return fmt.Sprintf("%s%s", target, d.volumeRootPath())
}

func (d *driver) volumeRootPath() string {
	return d.r.Config.GetString("linux.volume.rootPath")
}

func (d *driver) volumeType() string {
	return d.r.Config.GetString("docker.volumeType")
}

func (d *driver) iops() string {
	return d.r.Config.GetString("docker.iops")
}

func (d *driver) size() string {
	return d.r.Config.GetString("docker.size")
}

func (d *driver) availabilityZone() string {
	return d.r.Config.GetString("docker.availabilityZone")
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("Docker")
	r.Key(gofig.String, "", "", "", "docker.volumeType")
	r.Key(gofig.String, "", "", "", "docker.iops")
	r.Key(gofig.String, "", "", "", "docker.size")
	r.Key(gofig.String, "", "", "", "docker.availabilityZone")
	r.Key(gofig.String, "", "/data", "", "linux.volume.rootpath")
	return r
}
