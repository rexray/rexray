package docker

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
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
	mountDirectoryPath = util.LibFilePath("volumes")
	os.MkdirAll(mountDirectoryPath, 0755)
}

func newDriver() core.Driver {
	return &driver{}
}

func (d *driver) Init(r *core.RexRay) error {
	d.r = r
	log.WithField("provider", providerName).Debug("volume driver initialized")
	return nil
}

func getVolumeMountPath(name string) (string, error) {
	if name == "" {
		return "", errors.New("Missing volume name")
	}

	return fmt.Sprintf("%s/%s", mountDirectoryPath, name), nil
}

// Name will return the name of the volume driver manager
func (d *driver) Name() string {
	return providerName
}

// Mount will perform the steps to get an existing Volume with or without a fileystem mounted to a guest
func (d *driver) Mount(volumeName, volumeID string, overwriteFs bool, newFsType string) (string, error) {
	log.WithFields(log.Fields{
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
		volAttachments, err = d.r.Storage.AttachVolume(
			false, vols[0].VolumeID, instance.InstanceID)
		if err != nil {
			return "", err
		}
	}

	if len(volAttachments) == 0 {
		return "", errors.New("Volume did not attach")
	}

	mounts, err := d.r.OS.GetMounts(volAttachments[0].DeviceName, "")
	if err != nil {
		return "", err
	}

	if len(mounts) > 0 {
		return mounts[0].Mountpoint, nil
	}

	switch {
	case os.Getenv("REXRAY_DOCKER_VOLUMETYPE") != "":
		newFsType = os.Getenv("REXRAY_DOCKER_VOLUMETYPE")
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

	return mountPath, nil
}

// Unmount will perform the steps to unmount and existing volume and detach
func (d *driver) Unmount(volumeName, volumeID string) error {

	log.WithFields(log.Fields{
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

	err = d.r.Storage.DetachVolume(false, vols[0].VolumeID, "")
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
		return nil, errors.New("No instances")
	case len(instances) > 1:
		return nil,
			errors.New("Too many instances returned, limit the storagedrivers")
	}

	return instances[0], nil
}

func (d *driver) prefixToMountUnmount(
	volumeName,
	volumeID string) ([]*core.Volume, []*core.VolumeAttachment, *core.Instance, error) {
	if volumeName == "" && volumeID == "" {
		return nil, nil, nil, errors.New("Missing volume name or ID")
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
		return nil, nil, nil, errors.New("No volumes returned by name")
	case len(vols) > 1:
		return nil, nil, nil, errors.New("Multiple volumes returned by name")
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
		"volumeName": volumeName,
		"volumeID":   volumeID,
		"driverName": d.Name()}).Info("getting path to volume")
	if volumeName == "" && volumeID == "" {
		return "", errors.New("Missing volume name or ID")
	}

	instances, err := d.r.Storage.GetInstances()
	if err != nil {
		return "", err
	}

	switch {
	case len(instances) == 0:
		return "", errors.New("No instances")
	case len(instances) > 1:
		return "", errors.New("Too many instances returned, limit the storagedrivers")
	}

	volumes, err := d.r.Storage.GetVolume(volumeID, volumeName)
	if err != nil {
		return "", err
	}

	switch {
	case len(volumes) == 0:
		return "", errors.New("No volumes returned by name")
	case len(volumes) > 1:
		return "", errors.New("Multiple volumes returned by name")
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

	return mounts[0].Mountpoint, nil
}

// Create will create a remote volume
func (d *driver) Create(volumeName string, volumeOpts core.VolumeOpts) error {
	log.WithFields(log.Fields{
		"volumeName": volumeName,
		"volumeOpts": volumeOpts,
		"driverName": d.Name()}).Info("creating volume")

	if volumeName == "" {
		return errors.New("Missing volume name")
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

	volumeType := createInitVolumeType(volumeOpts)
	IOPS := createInitIOPS(volumeOpts)
	size := createInitSize(volumeOpts)
	availabilityZone := createInitAvailabilityZone(volumeOpts)

	var snapshotID string
	if snapshotID, err = d.createGetSnapshotID(volumeOpts); err != nil {
		return err
	}

	var volumeID string
	if volumeID, err = d.createInitVolumeID(
		snapshotID, volumeName, volumeOpts); err != nil {
		return err
	}

	if len(volumes) == 0 {
		if _, err = d.r.Storage.CreateVolume(
			false, volumeName, volumeID, snapshotID,
			volumeType, IOPS, size, availabilityZone); err != nil {
			return err
		}
	}

	if newFsType != "" || overwriteFs {
		_, err = d.Mount(volumeName, "", overwriteFs, newFsType)
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

func (d *driver) createInitVolumeID(
	snapshotID, volumeName string,
	volumeOpts core.VolumeOpts) (string, error) {

	var ok bool
	var optVolumeName string

	if optVolumeName, ok = volumeOpts["volumename"]; !ok {
		return volumeOpts["volumeid"], nil
	}

	var err error
	var volumes []*core.Volume
	if volumes, err = d.r.Storage.GetVolume("", optVolumeName); err != nil {
		return "", err
	}

	switch {
	case len(volumes) == 0:
		return "", errors.WithField(
			"optVolumeName", optVolumeName, "No volumes returned")
	case len(volumes) > 1:
		return "", errors.WithField(
			"optVolumeName", optVolumeName, "Too many volumes returned")
	}

	return volumes[0].VolumeID, nil
}

func (d *driver) createGetSnapshotID(
	volumeOpts core.VolumeOpts) (string, error) {

	var ok bool
	var optSnapshotName string

	if optSnapshotName, ok = volumeOpts["snapshotname"]; !ok {
		return volumeOpts["snapshotid"], nil
	}

	var err error
	var snapshots []*core.Snapshot

	if snapshots, err = d.r.Storage.GetSnapshot(
		"", "", optSnapshotName); err != nil {
		return "", err
	}

	switch {
	case len(snapshots) == 0:
		return "", errors.WithField(
			"optSnapshotName", optSnapshotName, "No snapshots returned")
	case len(snapshots) > 1:
		return "", errors.WithField(
			"optSnapshotName", optSnapshotName, "Too many snapshots returned")
	}

	return snapshots[0].SnapshotID, nil
}

func (d *driver) createGetInstance() error {
	var err error
	var instances []*core.Instance

	if instances, err = d.r.Storage.GetInstances(); err != nil {
		return err
	}

	switch {
	case len(instances) == 0:
		return errors.New("No instances")
	case len(instances) > 1:
		return errors.New(
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
		return nil, overwriteFs, errors.WithField(
			"volumeName", volumeName, "Too many volumes returned")
	}

	return volumes, overwriteFs, nil
}

func createInitVolumeType(volumeOpts core.VolumeOpts) string {
	var ok bool
	var volumeType string
	if volumeType, ok = volumeOpts["volumetype"]; ok {
		return volumeType
	}
	return os.Getenv("REXRAY_DOCKER_VOLUMETYPE")
}

func createInitIOPS(volumeOpts core.VolumeOpts) int64 {
	return createInitInt64("iops", "REXRAY_DOCKER_IOPS", volumeOpts)
}

func createInitSize(volumeOpts core.VolumeOpts) int64 {
	return createInitInt64("size", "REXRAY_DOCKER_SIZE", volumeOpts)
}

func createInitInt64(
	optKey, envVar string,
	volumeOpts core.VolumeOpts) int64 {
	var k bool
	var i int
	var s string
	if s, k = volumeOpts[optKey]; k {
		i, _ = strconv.Atoi(s)
	} else {
		i, _ = strconv.Atoi(os.Getenv(envVar))
	}
	return int64(i)
}

func createInitAvailabilityZone(volumeOpts core.VolumeOpts) string {
	return createInitString(
		"availabilityzone", "REXRAY_DOCKER_AVAILABILITYZONE", volumeOpts)
}

func createInitString(optKey, envVar string,
	volumeOpts core.VolumeOpts) string {
	var ok bool
	var s string
	if s, ok = volumeOpts[optKey]; ok {
		return s
	}
	return os.Getenv(envVar)
}

// Remove will remove a remote volume
func (d *driver) Remove(volumeName string) error {
	log.WithFields(log.Fields{
		"volumeName": volumeName,
		"driverName": d.Name()}).Info("removing volume")

	if volumeName == "" {
		return errors.New("Missing volume name")
	}

	instances, err := d.r.Storage.GetInstances()
	if err != nil {
		return err
	}

	switch {
	case len(instances) == 0:
		return errors.New("No instances")
	case len(instances) > 1:
		return errors.New("Too many instances returned, limit the storagedrivers")
	}

	volumes, err := d.r.Storage.GetVolume("", volumeName)
	if err != nil {
		return err
	}

	switch {
	case len(volumes) == 0:
		return errors.New("No volumes returned by name")
	case len(volumes) > 1:
		return errors.New("Multiple volumes returned by name")
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

// Attach will attach a volume to an instance
func (d *driver) Attach(volumeName, instanceID string) (string, error) {
	log.WithFields(log.Fields{
		"volumeName": volumeName,
		"instanceID": instanceID,
		"driverName": d.Name()}).Info("attaching volume")

	volumes, err := d.r.Storage.GetVolume("", volumeName)
	if err != nil {
		return "", err
	}

	switch {
	case len(volumes) == 0:
		return "", errors.New("No volumes returned by name")
	case len(volumes) > 1:
		return "", errors.New("Multiple volumes returned by name")
	}

	_, err = d.r.Storage.AttachVolume(true, volumes[0].VolumeID, instanceID)
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
func (d *driver) Detach(volumeName, instanceID string) error {
	log.WithFields(log.Fields{
		"volumeName": volumeName,
		"instanceID": instanceID,
		"driverName": d.Name()}).Info("detaching volume")

	volume, err := d.r.Storage.GetVolume("", volumeName)
	if err != nil {
		return err
	}

	return d.r.Storage.DetachVolume(true, volume[0].VolumeID, instanceID)
}

// NetworkName will return relevant information about how a volume can be discovered on an OS
func (d *driver) NetworkName(volumeName, instanceID string) (string, error) {
	log.WithFields(log.Fields{
		"volumeName": volumeName,
		"instanceID": instanceID,
		"driverName": d.Name()}).Info("returning network name")

	volumes, err := d.r.Storage.GetVolume("", volumeName)
	if err != nil {
		return "", err
	}

	switch {
	case len(volumes) == 0:
		return "", errors.New("No volumes returned by name")
	case len(volumes) > 1:
		return "", errors.New("Multiple volumes returned by name")
	}

	volumeAttachment, err := d.r.Storage.GetVolumeAttach(
		volumes[0].VolumeID, instanceID)
	if err != nil {
		return "", err
	}

	if len(volumeAttachment) == 0 {
		return "", errors.New("Volume not attached")
	}

	volumes, err = d.r.Storage.GetVolume("", volumeName)
	if err != nil {
		return "", err
	}

	return volumes[0].NetworkName, nil
}
