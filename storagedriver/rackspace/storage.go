package rackspace

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/emccode/rexray/storagedriver"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/snapshots"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/volumes"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

var (
	providerName string
)

var (
	ErrMissingVolumeID = errors.New("Missing VolumeID")
)

type Driver struct {
	Provider           *gophercloud.ProviderClient
	Client             *gophercloud.ServiceClient
	ClientBlockStorage *gophercloud.ServiceClient
	Region             string
	InstanceID         string
}

func init() {
	storagedriver.Register("rackspace", Init)
	providerName = "RackSpace"
}

func getInstanceID() (string, error) {
	cmdOut, err := exec.Command("/usr/bin/xenstore-read", "name").Output()
	if err != nil {
		return "", fmt.Errorf("%s: %s", storagedriver.ErrDriverInstanceDiscovery, err)
	}

	instanceID := strings.Replace(string(cmdOut), "\n", "", -1)

	validInstanceID := regexp.MustCompile(`^instance-`)
	valid := validInstanceID.MatchString(instanceID)
	if !valid {
		return "", storagedriver.ErrDriverInstanceDiscovery
	}

	instanceID = strings.Replace(instanceID, "instance-", "", 1)
	return instanceID, nil
}

func Init() (storagedriver.Driver, error) {

	instanceID, err := getInstanceID()
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	region, err := getInstanceRegion()
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	region = strings.ToUpper(region)
	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	clientBlockStorage, err := openstack.NewBlockStorageV1(provider, gophercloud.EndpointOpts{
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	driver := &Driver{
		Provider:           provider,
		Client:             client,
		ClientBlockStorage: clientBlockStorage,
		Region:             region,
		InstanceID:         instanceID,
	}

	return driver, nil
}

func (driver *Driver) getInstance() (*servers.Server, error) {
	server, err := servers.Get(driver.Client, driver.InstanceID).Extract()
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (driver *Driver) GetInstance() (interface{}, error) {
	server, err := driver.getInstance()
	if err != nil {
		return nil, err
	}

	instance := &storagedriver.Instance{
		ProviderName: providerName,
		InstanceID:   driver.InstanceID,
		Region:       driver.Region,
		Name:         server.Name,
	}

	return instance, nil
}

func (driver *Driver) GetBlockDeviceMapping() (interface{}, error) {
	blockDevices, err := driver.getBlockDevices(driver.InstanceID)
	if err != nil {
		return nil, err
	}

	var BlockDevices []*storagedriver.BlockDevice
	for _, blockDevice := range blockDevices {
		sdBlockDevice := &storagedriver.BlockDevice{
			ProviderName: providerName,
			InstanceID:   driver.InstanceID,
			VolumeID:     blockDevice.VolumeID,
			DeviceName:   blockDevice.Device,
			Region:       driver.Region,
			Status:       "",
		}
		BlockDevices = append(BlockDevices, sdBlockDevice)
	}

	return BlockDevices, nil

}

func (driver *Driver) getBlockDevices(instanceID string) ([]volumeattach.VolumeAttachment, error) {
	// volumes := volumeattach.Get(driver.Client, driver.InstanceID, "")
	allPages, err := volumeattach.List(driver.Client, driver.InstanceID).AllPages()

	// volumeAttachments, err := volumes.VolumeAttachmentResult.ExtractAll()
	volumeAttachments, err := volumeattach.ExtractVolumeAttachments(allPages)
	if err != nil {
		return []volumeattach.VolumeAttachment{}, fmt.Errorf("Error: %v", err)
	}

	return volumeAttachments, nil

}

func getInstanceRegion() (string, error) {
	cmdOut, err := exec.Command("/usr/bin/xenstore-read", "vm-data/provider_data/region").Output()
	if err != nil {
		return "", fmt.Errorf("Error: %v", err)
	}

	region := strings.Replace(string(cmdOut), "\n", "", -1)
	return region, nil
}

func (driver *Driver) getVolume(volumeID, volumeName string) (volumesRet []volumes.Volume, err error) {
	if volumeID != "" {
		volume, err := volumes.Get(driver.ClientBlockStorage, volumeID).Extract()
		if err != nil {
			return []volumes.Volume{}, err
		}
		volumesRet = append(volumesRet, *volume)
	} else {
		listOpts := &volumes.ListOpts{
			Name: volumeName,
		}

		allPages, err := volumes.List(driver.ClientBlockStorage, listOpts).AllPages()
		if err != nil {
			return []volumes.Volume{}, err
		}
		volumesRet, err = volumes.ExtractVolumes(allPages)
		if err != nil {
			return []volumes.Volume{}, err
		}

	}

	return volumesRet, nil
}

func (driver *Driver) GetVolume(volumeID, volumeName string) (interface{}, error) {
	volumesRet, err := driver.getVolume(volumeID, volumeName)
	if err != nil {
		return []*storagedriver.Volume{}, err
	}

	var volumesSD []*storagedriver.Volume
	for _, volume := range volumesRet {
		var attachmentsSD []*storagedriver.VolumeAttachment
		for _, attachment := range volume.Attachments {
			attachmentSD := &storagedriver.VolumeAttachment{
				VolumeID:   attachment["volume_id"].(string),
				InstanceID: attachment["server_id"].(string),
				DeviceName: attachment["device"].(string),
				Status:     "",
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		volumeSD := &storagedriver.Volume{
			Name:             volume.Name,
			VolumeID:         volume.ID,
			AvailabilityZone: volume.AvailabilityZone,
			Status:           volume.Status,
			VolumeType:       volume.VolumeType,
			IOPS:             0,
			Size:             strconv.Itoa(volume.Size),
			Attachments:      attachmentsSD,
		}
		volumesSD = append(volumesSD, volumeSD)
	}

	return volumesSD, nil
}

func (driver *Driver) GetVolumeAttach(volumeID, instanceID string) (interface{}, error) {
	if volumeID == "" {
		return []*storagedriver.VolumeAttachment{}, ErrMissingVolumeID
	}
	volume, err := driver.GetVolume(volumeID, "")
	if err != nil {
		return []*storagedriver.VolumeAttachment{}, err
	}

	if instanceID != "" {
		var attached bool
		for _, volumeAttachment := range volume.([]*storagedriver.Volume)[0].Attachments {
			if volumeAttachment.InstanceID == instanceID {
				return volume.([]*storagedriver.Volume)[0].Attachments, nil
			}
		}
		if !attached {
			return []*storagedriver.VolumeAttachment{}, nil
		}
	}
	return volume.([]*storagedriver.Volume)[0].Attachments, nil
}

func (driver *Driver) getSnapshot(volumeID, snapshotID, snapshotName string) (allSnapshots []snapshots.Snapshot, err error) {

	if snapshotID != "" {
		snapshot, err := snapshots.Get(driver.ClientBlockStorage, snapshotID).Extract()
		if err != nil {
			return []snapshots.Snapshot{}, err
		}

		allSnapshots = append(allSnapshots, *snapshot)
	} else {
		opts := snapshots.ListOpts{
			VolumeID: volumeID,
			Name:     snapshotName,
		}

		allPages, err := snapshots.List(driver.ClientBlockStorage, opts).AllPages()
		if err != nil {
			return []snapshots.Snapshot{}, err
		}

		allSnapshots, err = snapshots.ExtractSnapshots(allPages)
		if err != nil {
			return []snapshots.Snapshot{}, fmt.Errorf("Failed to extract snapshots: %v", err)
		}
	}

	return allSnapshots, nil
}

func (driver *Driver) GetSnapshot(volumeID, snapshotID, snapshotName string) (interface{}, error) {
	snapshots, err := driver.getSnapshot(volumeID, snapshotID, snapshotName)
	if err != nil {
		return []*storagedriver.Snapshot{}, err
	}

	var snapshotsInt []*storagedriver.Snapshot
	for _, snapshot := range snapshots {
		snapshotSD := &storagedriver.Snapshot{
			Name:        snapshot.Name,
			VolumeID:    snapshot.VolumeID,
			SnapshotID:  snapshot.ID,
			VolumeSize:  strconv.Itoa(snapshot.Size),
			StartTime:   snapshot.CreatedAt,
			Description: snapshot.Description,
			Status:      snapshot.Status,
		}
		snapshotsInt = append(snapshotsInt, snapshotSD)
	}

	return snapshotsInt, nil
}

func (driver *Driver) CreateSnapshot(runAsync bool, snapshotName, volumeID, description string) (interface{}, error) {
	opts := snapshots.CreateOpts{
		Name:        snapshotName,
		VolumeID:    volumeID,
		Description: description,
		Force:       true,
	}

	resp, err := snapshots.Create(driver.ClientBlockStorage, opts).Extract()
	if err != nil {
		return storagedriver.Snapshot{}, err
	}

	if !runAsync {
		log.Println("Waiting for snapshot to complete")
		err = snapshots.WaitForStatus(driver.ClientBlockStorage, resp.ID, "available", 120)
		if err != nil {
			return storagedriver.Snapshot{}, err
		}
	}

	snapshot, err := driver.GetSnapshot("", resp.ID, "")
	if err != nil {
		return storagedriver.Snapshot{}, err
	}

	log.Println(fmt.Sprintf("Created Snapshot: %v", snapshot.([]*storagedriver.Snapshot)))
	return snapshot.([]*storagedriver.Snapshot), nil

}

func (driver *Driver) RemoveSnapshot(snapshotID string) error {
	resp := snapshots.Delete(driver.ClientBlockStorage, snapshotID)
	if resp.Err != nil {
		return resp.Err
	}

	log.Println("Removed Snapshot: " + snapshotID)
	return nil
}

func (driver *Driver) CreateVolume(runAsync bool, volumeName string, volumeID string, snapshotID string, volumeType string, IOPS int64, size int64) (interface{}, error) {
	if snapshotID != "" {
		snapshot, err := driver.GetSnapshot("", snapshotID, "")
		if err != nil {
			return "", err
		}

		sizeInt, err := strconv.Atoi(snapshot.([]*storagedriver.Snapshot)[0].VolumeSize)
		if err != nil {
			return "", err
		}
		size = int64(sizeInt)
	}

	if volumeID != "" {
		volume, err := driver.GetVolume(volumeID, "")
		if err != nil {
			return "", err
		}

		sizeInt, err := strconv.Atoi(volume.([]*storagedriver.Volume)[0].Size)
		if err != nil {
			return "", err
		}
		size = int64(sizeInt)
	}

	options := &volumes.CreateOpts{
		Name:       volumeName,
		Size:       int(size),
		SnapshotID: snapshotID,
		VolumeType: volumeType,
	}
	resp, err := volumes.Create(driver.ClientBlockStorage, options).Extract()
	if err != nil {
		return storagedriver.Volume{}, err
	}

	// return resp, nil

	if !runAsync {
		log.Println("Waiting for volume creation to complete")
		err = volumes.WaitForStatus(driver.ClientBlockStorage, resp.ID, "available", 120)
		if err != nil {
			return storagedriver.Volume{}, err
		}
	}

	volume, err := driver.GetVolume(resp.ID, "")
	if err != nil {
		return storagedriver.Volume{}, err
	}

	log.Println(fmt.Sprintf("Created volume: %+v", volume.([]*storagedriver.Volume)[0]))
	return volume.([]*storagedriver.Volume)[0], nil
}

func (driver *Driver) RemoveVolume(volumeID string) error {
	if volumeID == "" {
		return ErrMissingVolumeID
	}
	res := volumes.Delete(driver.ClientBlockStorage, volumeID)
	if res.Err != nil {
		return res.Err
	}

	log.Println("Deleted Volume: " + volumeID)
	return nil
}

func (driver *Driver) GetDeviceNextAvailable() (string, error) {
	letters := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"}
	blockDeviceNames := make(map[string]bool)

	blockDeviceMapping, err := driver.GetBlockDeviceMapping()
	if err != nil {
		return "", err
	}

	for _, blockDevice := range blockDeviceMapping.([]*storagedriver.BlockDevice) {
		re, _ := regexp.Compile(`^/dev/xvd([a-z])`)
		res := re.FindStringSubmatch(blockDevice.DeviceName)
		if len(res) > 0 {
			blockDeviceNames[res[1]] = true
		}
	}

	localDevices, err := getLocalDevices()
	if err != nil {
		return "", err
	}

	for _, localDevice := range localDevices {
		re, _ := regexp.Compile(`^xvd([a-z])`)
		res := re.FindStringSubmatch(localDevice)
		if len(res) > 0 {
			blockDeviceNames[res[1]] = true
		}
	}

	for _, letter := range letters {
		if !blockDeviceNames[letter] {
			nextDeviceName := "/dev/xvd" + letter
			log.Println("Got next device name: " + nextDeviceName)
			return nextDeviceName, nil
		}
	}
	return "", errors.New("No available device")
}

func getLocalDevices() (deviceNames []string, err error) {
	file := "/proc/partitions"
	contentBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return []string{}, errors.New(fmt.Sprintf("Couldn't read %s: %v", file, err))
	}

	content := string(contentBytes)

	lines := strings.Split(content, "\n")
	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) == 4 {
			deviceNames = append(deviceNames, fields[3])
		}
	}

	return deviceNames, nil
}

func (driver *Driver) AttachVolume(runAsync bool, volumeID, instanceID string) (interface{}, error) {
	nextDeviceName, err := driver.GetDeviceNextAvailable()
	if err != nil {
		return storagedriver.VolumeAttachment{}, err
	}

	options := &volumeattach.CreateOpts{
		Device:   nextDeviceName,
		VolumeID: volumeID,
	}

	_, err = volumeattach.Create(driver.Client, instanceID, options).Extract()
	if err != nil {
		return storagedriver.VolumeAttachment{}, err
	}

	if !runAsync {
		log.Println("Waiting for volume attachment to complete")
		err = driver.waitVolumeAttach(volumeID)
		if err != nil {
			return storagedriver.VolumeAttachment{}, err
		}
	}

	volumeAttachment, err := driver.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return storagedriver.VolumeAttachment{}, err
	}

	log.Println(fmt.Sprintf("Attached volume %s to instance %s", volumeID, instanceID))
	return volumeAttachment, nil
}

func (driver *Driver) DetachVolume(runAsync bool, volumeID, instanceID string) error {
	if volumeID == "" {
		return ErrMissingVolumeID
	}
	volume, err := driver.GetVolume(volumeID, "")
	if err != nil {
		return err
	}

	resp := volumeattach.Delete(driver.Client, volume.([]*storagedriver.Volume)[0].Attachments[0].InstanceID, volumeID)
	if resp.Err != nil {
		return resp.Err
	}

	if !runAsync {
		log.Println("Waiting for volume detachment to complete")
		err = driver.waitVolumeDetach(volumeID)
		if err != nil {
			return err
		}
	}

	log.Println("Detached volume", volumeID)
	return nil
}

func (driver *Driver) waitVolumeAttach(volumeID string) error {
	if volumeID == "" {
		return ErrMissingVolumeID
	}
	for {
		volume, err := driver.GetVolume(volumeID, "")
		if err != nil {
			return err
		}
		if volume.([]*storagedriver.Volume)[0].Status == "attached" {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (driver *Driver) waitVolumeDetach(volumeID string) error {
	if volumeID == "" {
		return ErrMissingVolumeID
	}
	for {
		volume, err := driver.GetVolume(volumeID, "")
		if err != nil {
			return err
		}
		if len(volume.([]*storagedriver.Volume)[0].Attachments) == 0 {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}
