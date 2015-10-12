package rackspace

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/emccode/rexray/config"
	"github.com/emccode/rexray/errors"
	"github.com/emccode/rexray/storage"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/snapshots"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/volumes"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

const ProviderName = "Rackspace"

const (
	minSize = 75 //rackspace is 75
)

type Driver struct {
	Provider           *gophercloud.ProviderClient
	Client             *gophercloud.ServiceClient
	ClientBlockStorage *gophercloud.ServiceClient
	Region             string
	InstanceID         string
	Config             *config.Config
}

func ef() errors.Fields {
	return errors.Fields{
		"provider": ProviderName,
	}
}

func eff(fields errors.Fields) map[string]interface{} {
	errFields := map[string]interface{}{
		"provider": ProviderName,
	}
	if fields != nil {
		for k, v := range fields {
			errFields[k] = v
		}
	}
	return errFields
}

func init() {
	storage.Register(ProviderName, Init)
}

func (d *Driver) newCmd(name string, args ...string) *exec.Cmd {
	return newCmd(d.Config, name, args...)
}

func newCmd(cfg *config.Config, name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Env = cfg.EnvVars()
	return c
}

func getInstanceID(cfg *config.Config) (string, error) {

	cmd := newCmd(cfg, "/usr/bin/xenstore-read", "name")
	cmdOut, err := cmd.Output()

	if err != nil {
		return "",
			errors.WithFields(eff(errors.Fields{
				"cmd.Path": cmd.Path,
				"cmd.Args": cmd.Args,
				"cmd.Out":  cmdOut,
			}), "error getting instance id")
	}

	instanceID := strings.Replace(string(cmdOut), "\n", "", -1)

	validInstanceID := regexp.MustCompile(`^instance-`)
	valid := validInstanceID.MatchString(instanceID)
	if !valid {
		return "", errors.WithFields(eff(errors.Fields{
			"instanceId": instanceID}), "error matching instance id")
	}

	instanceID = strings.Replace(instanceID, "instance-", "", 1)
	return instanceID, nil
}

func Init(cfg *config.Config) (storage.Driver, error) {

	fields := ef()
	instanceID, err := getInstanceID(cfg)
	if err != nil {
		return nil, err
	}
	fields["instanceId"] = instanceID

	region, err := getInstanceRegion(cfg)
	if err != nil {
		return nil, err
	}
	fields["region"] = region

	authOpts := getAuthOptions(cfg)

	fields["identityEndpoint"] = cfg.RackspaceAuthUrl
	fields["userId"] = cfg.RackspaceUserId
	fields["userName"] = cfg.RackspaceUserName
	if cfg.RackspacePassword == "" {
		fields["password"] = ""
	} else {
		fields["password"] = "******"
	}
	fields["tenantId"] = cfg.RackspaceTenantId
	fields["tenantName"] = cfg.RackspaceTenantName
	fields["domainId"] = cfg.RackspaceDomainId
	fields["domainName"] = cfg.RackspaceDomainName

	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return nil,
			errors.WithFieldsE(fields, "error getting authenticated client", err)
	}

	region = strings.ToUpper(region)
	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: region,
	})
	if err != nil {
		return nil,
			errors.WithFieldsE(fields, "error getting newComputeV2", err)
	}

	clientBlockStorage, err := openstack.NewBlockStorageV1(provider, gophercloud.EndpointOpts{
		Region: region,
	})
	if err != nil {
		return nil, errors.WithFieldsE(
			fields, "error getting newBlockStorageV1", err)
	}

	driver := &Driver{
		Provider:           provider,
		Client:             client,
		ClientBlockStorage: clientBlockStorage,
		Region:             region,
		InstanceID:         instanceID,
		Config:             cfg,
	}

	return driver, nil
}

func getAuthOptions(cfg *config.Config) gophercloud.AuthOptions {
	return gophercloud.AuthOptions{
		IdentityEndpoint: config.RackspaceAuthUrl,
		UserID:           config.RackspaceUserId,
		Username:         config.RackspaceUserName,
		Password:         config.RackspacePassword,
		TenantID:         config.RackspaceTenantId,
		TenantName:       config.RackspaceTenantName,
		DomainID:         config.RackspaceDomainId,
		DomainName:       config.RackspaceDomainName,
	}
}

func (driver *Driver) getInstance() (*servers.Server, error) {
	server, err := servers.Get(driver.Client, driver.InstanceID).Extract()
	if err != nil {
		return nil,
			errors.WithFieldsE(ef(), "error getting server instance", err)
	}

	return server, nil
}

func (driver *Driver) GetInstance() (*storage.Instance, error) {
	server, err := driver.getInstance()
	if err != nil {
		return nil,
			errors.WithFieldsE(ef(), "error getting driver instance", err)
	}

	instance := &storage.Instance{
		ProviderName: ProviderName,
		InstanceID:   driver.InstanceID,
		Region:       driver.Region,
		Name:         server.Name,
	}

	return instance, nil
}

func (driver *Driver) GetVolumeMapping() ([]*storage.BlockDevice, error) {
	blockDevices, err := driver.getBlockDevices(driver.InstanceID)
	if err != nil {
		return nil,
			errors.WithFieldsE(eff(errors.Fields{
				"instanceId": driver.InstanceID,
			}), "error getting block devices", err)
	}

	var BlockDevices []*storage.BlockDevice
	for _, blockDevice := range blockDevices {
		sdBlockDevice := &storage.BlockDevice{
			ProviderName: ProviderName,
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
		return []volumeattach.VolumeAttachment{},
			errors.WithFieldsE(eff(errors.Fields{
				"instanceId": instanceID}),
				"error extracting volume attachments", err)
	}

	return volumeAttachments, nil

}

func getInstanceRegion(cfg *config.Config) (string, error) {
	cmd := newCmd(
		cfg, "/usr/bin/xenstore-read",
		"vm-data/provider_data/region")

	cmdOut, err := cmd.Output()
	if err != nil {
		return "",
			errors.WithFields(eff(errors.Fields{
				"cmd.Path": cmd.Path,
				"cmd.Args": cmd.Args,
				"cmd.Out":  cmdOut,
			}), "error getting instance region")
	}

	region := strings.Replace(string(cmdOut), "\n", "", -1)
	return region, nil
}

func (driver *Driver) getVolume(volumeID, volumeName string) (volumesRet []volumes.Volume, err error) {
	if volumeID != "" {
		volume, err := volumes.Get(driver.ClientBlockStorage, volumeID).Extract()
		if err != nil {
			return []volumes.Volume{},
				errors.WithFieldsE(eff(errors.Fields{
					"volumeId":   volumeID,
					"volumeName": volumeName}),
					"error getting volumes", err)
		}
		volumesRet = append(volumesRet, *volume)
	} else {
		listOpts := &volumes.ListOpts{
		//Name:       volumeName,
		}

		allPages, err := volumes.List(driver.ClientBlockStorage, listOpts).AllPages()
		if err != nil {
			return []volumes.Volume{},
				errors.WithFieldsE(eff(errors.Fields{
					"volumeId":   volumeID,
					"volumeName": volumeName}),
					"error listing volumes", err)
		}
		volumesRet, err = volumes.ExtractVolumes(allPages)
		if err != nil {
			return []volumes.Volume{},
				errors.WithFieldsE(eff(errors.Fields{
					"volumeId":   volumeID,
					"volumeName": volumeName}),
					"error extracting volumes", err)
		}

		var volumesRetFiltered []volumes.Volume
		if volumeName != "" {
			var found bool
			for _, volume := range volumesRet {
				if volume.Name == volumeName {
					volumesRetFiltered = append(volumesRetFiltered, volume)
					found = true
					break
				}
			}
			if !found {
				return []volumes.Volume{}, nil
			}
			volumesRet = volumesRetFiltered
		}
	}

	return volumesRet, nil
}

func (driver *Driver) GetVolume(volumeID, volumeName string) ([]*storage.Volume, error) {

	volumesRet, err := driver.getVolume(volumeID, volumeName)
	if err != nil {
		return []*storage.Volume{},
			errors.WithFieldsE(eff(errors.Fields{
				"volumeId":   volumeID,
				"volumeName": volumeName}),
				"error getting volume", err)
	}

	var volumesSD []*storage.Volume
	for _, volume := range volumesRet {
		var attachmentsSD []*storage.VolumeAttachment
		for _, attachment := range volume.Attachments {
			attachmentSD := &storage.VolumeAttachment{
				VolumeID:   attachment["volume_id"].(string),
				InstanceID: attachment["server_id"].(string),
				DeviceName: attachment["device"].(string),
				Status:     "",
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		volumeSD := &storage.Volume{
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

func (driver *Driver) GetVolumeAttach(volumeID, instanceID string) ([]*storage.VolumeAttachment, error) {

	fields := eff(map[string]interface{}{
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return []*storage.VolumeAttachment{},
			errors.WithFields(fields, "volumeId is required")
	}
	volume, err := driver.GetVolume(volumeID, "")
	if err != nil {
		return []*storage.VolumeAttachment{},
			errors.WithFieldsE(fields, "error getting volume attach", err)
	}

	if instanceID != "" {
		var attached bool
		for _, volumeAttachment := range volume[0].Attachments {
			if volumeAttachment.InstanceID == instanceID {
				return volume[0].Attachments, nil
			}
		}
		if !attached {
			return []*storage.VolumeAttachment{}, nil
		}
	}
	return volume[0].Attachments, nil
}

func (driver *Driver) getSnapshot(volumeID, snapshotID, snapshotName string) (allSnapshots []snapshots.Snapshot, err error) {

	fields := eff(map[string]interface{}{
		"volumeId":     volumeID,
		"snapshotId":   snapshotID,
		"snapshotName": snapshotName,
	})

	if snapshotID != "" {
		snapshot, err := snapshots.Get(driver.ClientBlockStorage, snapshotID).Extract()
		if err != nil {
			return []snapshots.Snapshot{},
				errors.WithFieldsE(fields, "error getting snapshot", err)
		}

		allSnapshots = append(allSnapshots, *snapshot)
	} else {
		opts := snapshots.ListOpts{
			VolumeID: volumeID,
			Name:     snapshotName,
		}

		allPages, err := snapshots.List(driver.ClientBlockStorage, opts).AllPages()
		if err != nil {
			return []snapshots.Snapshot{},
				errors.WithFieldsE(fields, "error listing snapshot", err)
		}

		allSnapshots, err = snapshots.ExtractSnapshots(allPages)
		if err != nil {
			return []snapshots.Snapshot{},
				errors.WithFieldsE(fields, "error extracting snapshot", err)
		}
	}

	return allSnapshots, nil
}

func (driver *Driver) GetSnapshot(volumeID, snapshotID, snapshotName string) ([]*storage.Snapshot, error) {
	snapshots, err := driver.getSnapshot(volumeID, snapshotID, snapshotName)
	if err != nil {
		return nil,
			errors.WithFieldsE(eff(errors.Fields{
				"volumeId":     volumeID,
				"snapshotId":   snapshotID,
				"snapshotName": snapshotName}),
				"error getting snapshot", err)
	}

	var snapshotsInt []*storage.Snapshot
	for _, snapshot := range snapshots {
		snapshotSD := &storage.Snapshot{
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

func (driver *Driver) CreateSnapshot(runAsync bool, snapshotName, volumeID, description string) ([]*storage.Snapshot, error) {

	fields := eff(map[string]interface{}{
		"runAsync":     runAsync,
		"snapshotName": snapshotName,
		"volumeId":     volumeID,
		"description":  description,
	})

	opts := snapshots.CreateOpts{
		Name:        snapshotName,
		VolumeID:    volumeID,
		Description: description,
		Force:       true,
	}

	resp, err := snapshots.Create(driver.ClientBlockStorage, opts).Extract()
	if err != nil {
		return nil,
			errors.WithFieldsE(fields, "error creating snapshot", err)
	}

	if !runAsync {
		log.Debug("waiting for snapshot creation to complete")
		err = snapshots.WaitForStatus(driver.ClientBlockStorage, resp.ID, "available", 120)
		if err != nil {
			return nil,
				errors.WithFieldsE(fields,
					"error waiting for snapshot creation to complete", err)
		}
	}

	snapshot, err := driver.GetSnapshot("", resp.ID, "")
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"runAsync":     runAsync,
		"snapshotName": snapshotName,
		"volumeId":     volumeID,
		"description":  description}).Debug("created snapshot")

	return snapshot, nil

}

func (driver *Driver) RemoveSnapshot(snapshotID string) error {
	resp := snapshots.Delete(driver.ClientBlockStorage, snapshotID)
	if resp.Err != nil {
		return errors.WithFieldE(
			"snapshotId", snapshotID, "error removing snapshot", resp.Err)
	}

	log.WithField("snapshotId", snapshotID).Debug("removed snapshot")

	return nil
}

func (driver *Driver) CreateVolume(
	runAsync bool,
	volumeName string,
	volumeID string,
	snapshotID string,
	volumeType string,
	IOPS int64,
	size int64,
	availabilityZone string) (*storage.Volume, error) {

	fields := map[string]interface{}{
		"provider":         ProviderName,
		"runAsync":         runAsync,
		"volumeName":       volumeName,
		"volumeId":         volumeID,
		"snapshotId":       snapshotID,
		"volumeType":       volumeType,
		"iops":             IOPS,
		"size":             size,
		"availabilityZone": availabilityZone,
	}

	if volumeID != "" && runAsync {
		return nil, errors.WithFields(fields,
			"cannot create volume from volume & run async")
	}

	if snapshotID != "" {
		snapshot, err := driver.GetSnapshot("", snapshotID, "")
		if err != nil {
			return nil,
				errors.WithFieldsE(fields, "error getting snapshot", err)
		}

		if len(snapshot) == 0 {
			return nil,
				errors.WithFields(fields, "snapshot array is empty")
		}

		volSize := snapshot[0].VolumeSize
		sizeInt, err := strconv.Atoi(volSize)
		if err != nil {
			f := errors.Fields{
				"volumeSize": volSize,
			}
			for k, v := range fields {
				f[k] = v
			}
			return nil,
				errors.WithFieldsE(f, "error casting volume size", err)
		}
		size = int64(sizeInt)
	}

	var volume []*storage.Volume
	var err error
	if volumeID != "" {
		volume, err = driver.GetVolume(volumeID, "")
		if err != nil {
			return nil, errors.WithFields(fields, "error getting volume")
		}

		if len(volume) == 0 {
			return nil,
				errors.WithFields(fields, "volume array is empty")
		}

		volSize := volume[0].Size
		sizeInt, err := strconv.Atoi(volSize)
		if err != nil {
			f := errors.Fields{
				"volumeSize": volSize,
			}
			for k, v := range fields {
				f[k] = v
			}
			return nil,
				errors.WithFieldsE(f, "error casting volume size", err)
		}
		size = int64(sizeInt)

		volumeID := volume[0].VolumeID
		snapshot, err := driver.CreateSnapshot(
			false, fmt.Sprintf("temp-%s", volumeID), volumeID, "")
		if err != nil {
			return nil,
				errors.WithFields(fields, "error creating snapshot")
		}

		snapshotID = snapshot[0].SnapshotID

		if availabilityZone == "" {
			availabilityZone = volume[0].AvailabilityZone
		}

	}

	if size != 0 && size < minSize {
		size = minSize
	}

	options := &volumes.CreateOpts{
		Name:         volumeName,
		Size:         int(size),
		SnapshotID:   snapshotID,
		VolumeType:   volumeType,
		Availability: availabilityZone,
	}
	resp, err := volumes.Create(driver.ClientBlockStorage, options).Extract()
	if err != nil {
		return nil,
			errors.WithFields(fields, "error creating volume")
	}

	if !runAsync {
		log.Debug("waiting for volume creation to complete")
		err = volumes.WaitForStatus(driver.ClientBlockStorage, resp.ID, "available", 120)
		if err != nil {
			return nil,
				errors.WithFields(fields,
					"error waiting for volume creation to complete")
		}

		if volumeID != "" {
			err := driver.RemoveSnapshot(snapshotID)
			if err != nil {
				return nil,
					errors.WithFields(fields,
						"error removing snapshot")
			}
		}
	}

	fields["volumeId"] = resp.ID
	fields["volumeName"] = ""

	volume, err = driver.GetVolume(resp.ID, "")
	if err != nil {
		return nil, errors.WithFields(fields,
			"error removing snapshot")
	}

	log.WithFields(fields).Debug("created volume")
	return volume[0], nil
}

func (driver *Driver) RemoveVolume(volumeID string) error {
	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})
	if volumeID == "" {
		return errors.WithFields(fields, "volumeId is required")
	}
	res := volumes.Delete(driver.ClientBlockStorage, volumeID)
	if res.Err != nil {
		return errors.WithFieldsE(fields, "error removing volume", res.Err)
	}

	log.WithFields(fields).Debug("removed volume")
	return nil
}

func (driver *Driver) GetDeviceNextAvailable() (string, error) {
	letters := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"}
	blockDeviceNames := make(map[string]bool)

	blockDeviceMapping, err := driver.GetVolumeMapping()
	if err != nil {
		return "", err
	}

	for _, blockDevice := range blockDeviceMapping {
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
			return nextDeviceName, nil
		}
	}
	return "", errors.New("No available device")
}

func getLocalDevices() (deviceNames []string, err error) {
	file := "/proc/partitions"
	contentBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return []string{},
			errors.WithFieldsE(
				eff(errors.Fields{"file": file}), "error reading file", err)
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

func (driver *Driver) AttachVolume(runAsync bool, volumeID, instanceID string) ([]*storage.VolumeAttachment, error) {
	fields := eff(map[string]interface{}{
		"runAsync":   runAsync,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	nextDeviceName, err := driver.GetDeviceNextAvailable()
	if err != nil {
		return nil, errors.WithFieldsE(
			fields, "error getting next available device", err)
	}

	options := &volumeattach.CreateOpts{
		Device:   nextDeviceName,
		VolumeID: volumeID,
	}

	_, err = volumeattach.Create(driver.Client, instanceID, options).Extract()
	if err != nil {
		return nil, errors.WithFieldsE(
			fields, "error attaching volume", err)
	}

	if !runAsync {
		log.WithFields(fields).Debug("waiting for volume to attach")
		err = driver.waitVolumeAttach(volumeID)
		if err != nil {
			return nil, errors.WithFieldsE(
				fields, "error waiting for volume to detach", err)
		}
	}

	volumeAttachment, err := driver.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, err
	}

	log.WithFields(fields).Debug("volume attached")
	return volumeAttachment, nil
}

func (driver *Driver) DetachVolume(runAsync bool, volumeID, instanceID string) error {
	fields := eff(map[string]interface{}{
		"runAsync":   runAsync,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return errors.WithFields(fields, "volumeId is required")
	}
	volume, err := driver.GetVolume(volumeID, "")
	if err != nil {
		return errors.WithFieldsE(fields, "error getting volume", err)
	}

	fields["instanceId"] = volume[0].Attachments[0].InstanceID
	resp := volumeattach.Delete(
		driver.Client, volume[0].Attachments[0].InstanceID, volumeID)
	if resp.Err != nil {
		return errors.WithFieldsE(fields, "error deleting volume", err)
	}

	if !runAsync {
		log.WithFields(fields).Debug("waiting for volume to detach")
		err = driver.waitVolumeDetach(volumeID)
		if err != nil {
			return errors.WithFieldsE(
				fields, "error waiting for volume to detach", err)
		}
	}

	log.WithFields(fields).Debug("volume detached")
	return nil
}

func (driver *Driver) waitVolumeAttach(volumeID string) error {

	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})

	if volumeID == "" {
		return errors.WithFields(fields, "volumeId is required")
	}
	for {
		volume, err := driver.GetVolume(volumeID, "")
		if err != nil {
			return errors.WithFieldsE(fields, "error getting volume", err)
		}
		if volume[0].Status == "in-use" {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (driver *Driver) waitVolumeDetach(volumeID string) error {

	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})

	if volumeID == "" {
		return errors.WithFields(fields, "volumeId is required")
	}
	for {
		volume, err := driver.GetVolume(volumeID, "")
		if err != nil {
			return errors.WithFieldsE(fields, "error getting volume", err)
		}
		if len(volume[0].Attachments) == 0 {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (driver *Driver) CopySnapshot(runAsync bool, volumeID, snapshotID, snapshotName, destinationSnapshotName, destinationRegion string) (*storage.Snapshot, error) {
	return nil, errors.New("This driver does not implement CopySnapshot")
}
