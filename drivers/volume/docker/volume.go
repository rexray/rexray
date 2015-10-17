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

	volumeAttachment, err := d.r.Storage.GetVolumeAttach(
		volumes[0].VolumeID, instances[0].InstanceID)
	if err != nil {
		return "", err
	}

	if len(volumeAttachment) == 0 {
		volumeAttachment, err = d.r.Storage.AttachVolume(
			false, volumes[0].VolumeID, instances[0].InstanceID)
		if err != nil {
			return "", err
		}
	}

	if len(volumeAttachment) == 0 {
		return "", errors.New("Volume did not attach")
	}

	mounts, err := d.r.OS.GetMounts(volumeAttachment[0].DeviceName, "")
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

	if err := d.r.OS.Format(volumeAttachment[0].DeviceName, newFsType, overwriteFs); err != nil {
		return "", err
	}

	mountPath, err := getVolumeMountPath(volumes[0].Name)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(mountPath, 0755); err != nil {
		return "", err
	}

	if err := d.r.OS.Mount(volumeAttachment[0].DeviceName, mountPath, "", ""); err != nil {
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
	if volumeName == "" && volumeID == "" {
		return errors.New("Missing volume name or ID")
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

	volumes, err := d.r.Storage.GetVolume(volumeID, volumeName)
	if err != nil {
		return err
	}

	switch {
	case len(volumes) == 0:
		return errors.New("No volumes returned by name")
	case len(volumes) > 1:
		return errors.New("Multiple volumes returned by name")
	}

	volumeAttachment, err := d.r.Storage.GetVolumeAttach(volumes[0].VolumeID, instances[0].InstanceID)
	if err != nil {
		return err
	}

	if len(volumeAttachment) == 0 {
		return nil
	}

	mounts, err := d.r.OS.GetMounts(volumeAttachment[0].DeviceName, "")
	if err != nil {
		return err
	}

	if len(mounts) > 0 {
		err := d.r.OS.Unmount(mounts[0].Mountpoint)
		if err != nil {
			return err
		}
	}

	err = d.r.Storage.DetachVolume(false, volumes[0].VolumeID, "")
	if err != nil {
		return err
	}
	return nil

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

	for k, v := range volumeOpts {
		volumeOpts[strings.ToLower(k)] = v
	}

	newFsType := volumeOpts["newfstype"]
	overwriteFs, _ := strconv.ParseBool(volumeOpts["overwritefs"])

	switch {
	case len(volumes) == 1 && !overwriteFs:
		return nil
	case len(volumes) > 1:
		return errors.WithField("volumeName", volumeName, "Too many volumes returned")
	}

	var (
		ok               bool
		volumeType       string
		IOPSi            int
		sizei            int
		availabilityZone string
		optVolumeID      string
		optSnapshotID    string
	)

	if volumeType, ok = volumeOpts["volumetype"]; !ok {
		volumeType = os.Getenv("REXRAY_DOCKER_VOLUMETYPE")
	}

	if IOPSs, ok := volumeOpts["iops"]; ok {
		IOPSi, _ = strconv.Atoi(IOPSs)
	} else {
		IOPSi, _ = strconv.Atoi(os.Getenv("REXRAY_DOCKER_IOPS"))
	}
	IOPS := int64(IOPSi)

	if sizes, ok := volumeOpts["size"]; ok {
		sizei, _ = strconv.Atoi(sizes)
	} else {
		sizei, _ = strconv.Atoi(os.Getenv("REXRAY_DOCKER_SIZE"))
	}
	size := int64(sizei)
	if size == 0 {
		size = defaultVolumeSize
	}

	if availabilityZone, ok = volumeOpts["availabilityzone"]; !ok {
		availabilityZone = os.Getenv("REXRAY_DOCKER_AVAILABILITYZONE")
	}

	if optSnapshotName, ok := volumeOpts["snapshotname"]; !ok {
		optSnapshotID = volumeOpts["snapshotid"]
	} else {
		snapshots, err := d.r.Storage.GetSnapshot("", "", optSnapshotName)
		if err != nil {
			return err
		}

		switch {
		case len(snapshots) == 0:
			return errors.WithField("optSnapshotName", optSnapshotName, "No snapshots returned")
		case len(snapshots) > 1:
			return errors.WithField("optSnapshotName", optSnapshotName, "Too many snapshots returned")
		}

		optSnapshotID = snapshots[0].SnapshotID
	}

	if optVolumeName, ok := volumeOpts["volumename"]; !ok {
		optVolumeID = volumeOpts["volumeid"]
	} else {
		volumes, err := d.r.Storage.GetVolume("", optVolumeName)
		if err != nil {
			return err
		}

		switch {
		case len(volumes) == 0:
			return errors.WithField("optVolumeName", optVolumeName, "No volumes returned")
		case len(volumes) > 1:
			return errors.WithField("optVolumeName", optVolumeName, "Too many volumes returned")
		}

		optVolumeID = volumes[0].VolumeID
	}

	if len(volumes) == 0 {
		_, err := d.r.Storage.CreateVolume(
			false, volumeName, optVolumeID, optSnapshotID, volumeType, IOPS, size, availabilityZone)
		if err != nil {
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
