package docker

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/emccode/rexray/drivers/volume"
	rros "github.com/emccode/rexray/os"
	"github.com/emccode/rexray/storage"
)

var (
	providerName string
)

type Driver struct{}

func init() {
	providerName = "docker"
	volumedriver.Register("docker", Init)
}

func Init() (volumedriver.Driver, error) {
	driver := &Driver{}

	if os.Getenv("REXRAY_DEBUG") == "true" {
		log.Println("Volume Manager Driver Initialized: " + providerName)
	}

	return driver, nil
}

func getVolumeMountPath(name string) (string, error) {
	if name == "" {
		return "", errors.New("Missing volume name")
	}

	return fmt.Sprintf("/var/lib/docker/volumes/%s", name), nil
}

// MountVolume will perform the steps to get an existing Volume with or without a fileystem mounted to a guest
func (driver *Driver) MountVolume(volumeName, volumeID string, overwriteFs bool, newFsType string) (string, error) {

	instances, err := storage.GetInstance()
	if err != nil {
		return "", err
	}

	switch {
	case len(instances) == 0:
		return "", errors.New("No instances")
	case len(instances) > 1:
		return "", errors.New("Too many instances returned, limit the storagedrivers")
	}

	volumes, err := storage.GetVolume(volumeID, volumeName)
	if err != nil {
		return "", err
	}

	switch {
	case len(volumes) == 0:
		return "", errors.New("No volumes returned by name")
	case len(volumes) > 1:
		return "", errors.New("Multiple volumes returned by name")
	}

	volumeAttachment, err := storage.GetVolumeAttach(volumes[0].VolumeID, instances[0].InstanceID)
	if err != nil {
		return "", err
	}

	if len(volumeAttachment) == 0 {
		volumeAttachment, err = storage.AttachVolume(false, volumes[0].VolumeID, instances[0].InstanceID)
		if err != nil {
			return "", err
		}
	}

	mounts, err := rros.GetMounts(volumeAttachment[0].DeviceName, "")
	if err != nil {
		return "", err
	}

	if len(mounts) > 0 {
		return mounts[0].Mountpoint, nil
	}

	if newFsType == "" {
		newFsType = "ext4"
	}

	if err := rros.Format(volumeAttachment[0].DeviceName, newFsType, overwriteFs); err != nil {
		return "", err
	}

	mountPath, err := getVolumeMountPath(volumes[0].Name)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(mountPath, 0755); err != nil {
		return "", err
	}

	if err := rros.Mount(volumeAttachment[0].DeviceName, mountPath, "", ""); err != nil {
		return "", err
	}

	return mountPath, nil
}

// UnmountVolume will perform the steps to unmount and existing volume and detach
func (driver *Driver) UnmountVolume(volumeName, volumeID string) error {

	instances, err := storage.GetInstance()
	if err != nil {
		return err
	}

	switch {
	case len(instances) == 0:
		return errors.New("No instances")
	case len(instances) > 1:
		return errors.New("Too many instances returned, limit the storagedrivers")
	}

	volumes, err := storage.GetVolume(volumeID, volumeName)
	if err != nil {
		return err
	}

	switch {
	case len(volumes) == 0:
		return errors.New("No volumes returned by name")
	case len(volumes) > 1:
		return errors.New("Multiple volumes returned by name")
	}

	volumeAttachment, err := storage.GetVolumeAttach(volumes[0].VolumeID, instances[0].InstanceID)
	if err != nil {
		return err
	}

	mounts, err := rros.GetMounts(volumeAttachment[0].DeviceName, "")
	if err != nil {
		return err
	}

	err = rros.Unmount(mounts[0].Mountpoint)
	if err != nil {
		return err
	}

	if len(volumeAttachment) == 0 {
		return nil
	}

	err = storage.DetachVolume(false, volumes[0].VolumeID, "")
	if err != nil {
		return err
	}
	return nil

}
