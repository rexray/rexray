package openstack

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/config"
	"github.com/emccode/rexray/core/errors"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/snapshots"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/volumes"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

const (
	providerName = "Openstack"
	minSize      = 1 // openstack has no minimum
)

type driver struct {
	provider           *gophercloud.ProviderClient
	client             *gophercloud.ServiceClient
	clientBlockStorage *gophercloud.ServiceClient
	region             string
	availabilityZone   string
	instanceID         string
	r                  *core.RexRay
}

func ef() errors.Fields {
	return errors.Fields{
		"provider": providerName,
	}
}

func eff(fields errors.Fields) map[string]interface{} {
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

func init() {
	core.RegisterDriver(providerName, newDriver)
}

func newDriver() core.Driver {
	return &driver{}
}

func (d *driver) Init(r *core.RexRay) error {
	d.r = r
	fields := ef()
	var err error

	if d.instanceID, err = getInstanceID(d.r.Config); err != nil {
		return err
	}

	fields["instanceId"] = d.instanceID

	if d.r.Config.OpenstackRegionName == "" {
		if d.region, err = getInstanceRegion(d.r.Config); err != nil {
			return err
		}
	} else {
		d.region = d.r.Config.OpenstackRegionName
	}
	fields["region"] = d.region

	if d.r.Config.OpenstackAvailabilityZoneName == "" {
		if d.availabilityZone, err = getInstanceAvailabilityZone(); err != nil {
			return err
		}
	} else {
		d.availabilityZone = d.r.Config.OpenstackAvailabilityZoneName
	}
	fields["availabilityZone"] = d.availabilityZone

	authOpts := getAuthOptions(d.r.Config)

	fields["identityEndpoint"] = d.r.Config.RackspaceAuthURL
	fields["userId"] = d.r.Config.RackspaceUserID
	fields["userName"] = d.r.Config.RackspaceUserName
	if d.r.Config.RackspacePassword == "" {
		fields["password"] = ""
	} else {
		fields["password"] = "******"
	}
	fields["tenantId"] = d.r.Config.RackspaceTenantID
	fields["tenantName"] = d.r.Config.RackspaceTenantName
	fields["domainId"] = d.r.Config.RackspaceDomainID
	fields["domainName"] = d.r.Config.RackspaceDomainName

	if d.provider, err = openstack.AuthenticatedClient(authOpts); err != nil {
		return errors.WithFieldsE(fields,
			"error getting authenticated client", err)
	}

	if d.client, err = openstack.NewComputeV2(d.provider,
		gophercloud.EndpointOpts{Region: d.region}); err != nil {
		errors.WithFieldsE(fields, "error getting newComputeV2", err)
	}

	if d.clientBlockStorage, err = openstack.NewBlockStorageV1(d.provider,
		gophercloud.EndpointOpts{Region: d.region}); err != nil {
		return errors.WithFieldsE(fields,
			"error getting newBlockStorageV1", err)
	}

	log.WithField("provider", providerName).Debug("storage driver initialized")

	return nil
}

func (d *driver) Name() string {
	return providerName
}

func (d *driver) newCmd(name string, args ...string) *exec.Cmd {
	return newCmd(d.r.Config, name, args...)
}

func newCmd(c *config.Config, name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Env = c.EnvVars()
	return cmd
}

func getInstanceID(c *config.Config) (string, error) {
	cmd := newCmd(c, "/usr/sbin/dmidecode")
	cmdOut, err := cmd.Output()

	if err != nil {
		return "",
			errors.WithFields(eff(errors.Fields{
				"cmd.Path": cmd.Path,
				"cmd.Args": cmd.Args,
				"cmd.Out":  cmdOut,
			}), "error getting instance id")
	}

	rp := regexp.MustCompile("UUID:(.*)")
	uuid := strings.Replace(rp.FindString(string(cmdOut)), "UUID: ", "", -1)

	return strings.ToLower(uuid), nil
}

func getAuthOptions(cfg *config.Config) gophercloud.AuthOptions {
	return gophercloud.AuthOptions{
		IdentityEndpoint: cfg.OpenstackAuthURL,
		UserID:           cfg.OpenstackUserID,
		Username:         cfg.OpenstackUserName,
		Password:         cfg.OpenstackPassword,
		TenantID:         cfg.OpenstackTenantID,
		TenantName:       cfg.OpenstackTenantName,
		DomainID:         cfg.OpenstackDomainID,
		DomainName:       cfg.OpenstackDomainName,
	}
}

func (d *driver) getInstance() (*servers.Server, error) {
	server, err := servers.Get(d.client, d.instanceID).Extract()
	if err != nil {
		return nil,
			errors.WithFieldsE(ef(), "error getting server instance", err)
	}

	return server, nil
}

func (d *driver) GetInstance() (*core.Instance, error) {
	server, err := d.getInstance()
	if err != nil {
		return nil,
			errors.WithFieldsE(ef(), "error getting driver instance", err)
	}

	instance := &core.Instance{
		ProviderName: providerName,
		InstanceID:   d.instanceID,
		Region:       d.region,
		Name:         server.Name,
	}

	return instance, nil
}

func (d *driver) GetVolumeMapping() ([]*core.BlockDevice, error) {
	blockDevices, err := d.getBlockDevices(d.instanceID)
	if err != nil {
		return nil,
			errors.WithFieldsE(eff(errors.Fields{
				"instanceId": d.instanceID,
			}), "error getting block devices", err)
	}

	var BlockDevices []*core.BlockDevice
	for _, blockDevice := range blockDevices {
		sdBlockDevice := &core.BlockDevice{
			ProviderName: providerName,
			InstanceID:   d.instanceID,
			VolumeID:     blockDevice.VolumeID,
			DeviceName:   blockDevice.Device,
			Region:       d.region,
			Status:       "",
		}
		BlockDevices = append(BlockDevices, sdBlockDevice)
	}

	return BlockDevices, nil

}

func (d *driver) getBlockDevices(
	instanceID string) ([]volumeattach.VolumeAttachment, error) {

	// volumes := volumeattach.Get(driver.client, driver.instanceId, "")
	allPages, err := volumeattach.List(d.client, d.instanceID).AllPages()

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

func getInstanceAvailabilityZone() (string, error) {
	conn, err := net.DialTimeout("tcp", "169.254.169.254:80", 50*time.Millisecond)
	if err != nil {
		return "", fmt.Errorf("Error: %v\n", err)
	}
	defer conn.Close()

	url := "http://169.254.169.254/2009-04-04/meta-data/placement/availability-zone"
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Error: %v\n", err)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error: %v\n", err)
	}

	return string(data), nil
}

func (d *driver) getVolume(
	volumeID, volumeName string) (volumesRet []volumes.Volume, err error) {

	if volumeID != "" {
		volume, err := volumes.Get(d.clientBlockStorage, volumeID).Extract()
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

		allPages, err := volumes.List(d.clientBlockStorage, listOpts).AllPages()
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

func (d *driver) GetVolume(
	volumeID, volumeName string) ([]*core.Volume, error) {

	volumesRet, err := d.getVolume(volumeID, volumeName)
	if err != nil {
		return []*core.Volume{},
			errors.WithFieldsE(eff(errors.Fields{
				"volumeId":   volumeID,
				"volumeName": volumeName}),
				"error getting volume", err)
	}

	var volumesSD []*core.Volume
	for _, volume := range volumesRet {
		var attachmentsSD []*core.VolumeAttachment
		for _, attachment := range volume.Attachments {
			attachmentSD := &core.VolumeAttachment{
				VolumeID:   attachment["volume_id"].(string),
				InstanceID: attachment["server_id"].(string),
				DeviceName: attachment["device"].(string),
				Status:     "",
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		volumeSD := &core.Volume{
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

func (d *driver) GetVolumeAttach(
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {

	fields := eff(map[string]interface{}{
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return []*core.VolumeAttachment{},
			errors.WithFields(fields, "volumeId is required")
	}
	volume, err := d.GetVolume(volumeID, "")
	if err != nil {
		return []*core.VolumeAttachment{},
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
			return []*core.VolumeAttachment{}, nil
		}
	}
	return volume[0].Attachments, nil
}

func (d *driver) getSnapshot(
	volumeID,
	snapshotID,
	snapshotName string) (allSnapshots []snapshots.Snapshot, err error) {

	fields := eff(map[string]interface{}{
		"volumeId":     volumeID,
		"snapshotId":   snapshotID,
		"snapshotName": snapshotName,
	})

	if snapshotID != "" {
		snapshot, err := snapshots.Get(d.clientBlockStorage, snapshotID).Extract()
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

		allPages, err := snapshots.List(d.clientBlockStorage, opts).AllPages()
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

func (d *driver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {

	snapshots, err := d.getSnapshot(volumeID, snapshotID, snapshotName)
	if err != nil {
		return nil,
			errors.WithFieldsE(eff(errors.Fields{
				"volumeId":     volumeID,
				"snapshotId":   snapshotID,
				"snapshotName": snapshotName}),
				"error getting snapshot", err)
	}

	var snapshotsInt []*core.Snapshot
	for _, snapshot := range snapshots {
		snapshotSD := &core.Snapshot{
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

func (d *driver) CreateSnapshot(
	runAsync bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {

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

	resp, err := snapshots.Create(d.clientBlockStorage, opts).Extract()
	if err != nil {
		return nil,
			errors.WithFieldsE(fields, "error creating snapshot", err)
	}

	if !runAsync {
		log.Debug("waiting for snapshot creation to complete")
		err = snapshots.WaitForStatus(d.clientBlockStorage, resp.ID, "available", 120)
		if err != nil {
			return nil,
				errors.WithFieldsE(fields,
					"error waiting for snapshot creation to complete", err)
		}
	}

	snapshot, err := d.GetSnapshot("", resp.ID, "")
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

func (d *driver) RemoveSnapshot(snapshotID string) error {
	resp := snapshots.Delete(d.clientBlockStorage, snapshotID)
	if resp.Err != nil {
		return errors.WithFieldE(
			"snapshotId", snapshotID, "error removing snapshot", resp.Err)
	}

	log.WithField("snapshotId", snapshotID).Debug("removed snapshot")

	return nil
}

func (d *driver) CreateVolume(
	runAsync bool,
	volumeName string,
	volumeID string,
	snapshotID string,
	volumeType string,
	IOPS int64,
	size int64,
	availabilityZone string) (*core.Volume, error) {

	fields := map[string]interface{}{
		"provider":         providerName,
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

	if availabilityZone == "" {
		availabilityZone = d.availabilityZone
	}

	if snapshotID != "" {
		snapshot, err := d.GetSnapshot("", snapshotID, "")
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

	var volume []*core.Volume
	var err error
	if volumeID != "" {
		volume, err = d.GetVolume(volumeID, "")
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
		snapshot, err := d.CreateSnapshot(
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
	resp, err := volumes.Create(d.clientBlockStorage, options).Extract()
	if err != nil {
		return nil,
			errors.WithFields(fields, "error creating volume")
	}

	if !runAsync {
		log.Debug("waiting for volume creation to complete")
		err = volumes.WaitForStatus(d.clientBlockStorage, resp.ID, "available", 120)
		if err != nil {
			return nil,
				errors.WithFields(fields,
					"error waiting for volume creation to complete")
		}

		if volumeID != "" {
			err := d.RemoveSnapshot(snapshotID)
			if err != nil {
				return nil,
					errors.WithFields(fields,
						"error removing snapshot")
			}
		}
	}

	fields["volumeId"] = resp.ID
	fields["volumeName"] = ""

	volume, err = d.GetVolume(resp.ID, "")
	if err != nil {
		return nil, errors.WithFields(fields,
			"error removing snapshot")
	}

	log.WithFields(fields).Debug("created volume")
	return volume[0], nil
}

func (d *driver) RemoveVolume(volumeID string) error {
	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})
	if volumeID == "" {
		return errors.WithFields(fields, "volumeId is required")
	}
	res := volumes.Delete(d.clientBlockStorage, volumeID)
	if res.Err != nil {
		return errors.WithFieldsE(fields, "error removing volume", res.Err)
	}

	log.WithFields(fields).Debug("removed volume")
	return nil
}

func (d *driver) GetDeviceNextAvailable() (string, error) {
	letters := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"}
	blockDeviceNames := make(map[string]bool)

	blockDeviceMapping, err := d.GetVolumeMapping()
	if err != nil {
		return "", err
	}

	for _, blockDevice := range blockDeviceMapping {
		re, _ := regexp.Compile(`^/dev/vd([a-z])`)
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
		re, _ := regexp.Compile(`^vd([a-z])`)
		res := re.FindStringSubmatch(localDevice)
		if len(res) > 0 {
			blockDeviceNames[res[1]] = true
		}
	}

	for _, letter := range letters {
		if !blockDeviceNames[letter] {
			nextDeviceName := "/dev/vd" + letter
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

func (d *driver) AttachVolume(
	runAsync bool, volumeID, instanceID string) ([]*core.VolumeAttachment, error) {

	fields := eff(map[string]interface{}{
		"runAsync":   runAsync,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	nextDeviceName, err := d.GetDeviceNextAvailable()
	if err != nil {
		return nil, errors.WithFieldsE(
			fields, "error getting next available device", err)
	}

	options := &volumeattach.CreateOpts{
		Device:   nextDeviceName,
		VolumeID: volumeID,
	}

	_, err = volumeattach.Create(d.client, instanceID, options).Extract()
	if err != nil {
		return nil, errors.WithFieldsE(
			fields, "error attaching volume", err)
	}

	if !runAsync {
		log.WithFields(fields).Debug("waiting for volume to attach")
		err = d.waitVolumeAttach(volumeID)
		if err != nil {
			return nil, errors.WithFieldsE(
				fields, "error waiting for volume to detach", err)
		}
	}

	volumeAttachment, err := d.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, err
	}

	log.WithFields(fields).Debug("volume attached")
	return volumeAttachment, nil
}

func (d *driver) DetachVolume(
	runAsync bool, volumeID, instanceID string) error {

	fields := eff(map[string]interface{}{
		"runAsync":   runAsync,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return errors.WithFields(fields, "volumeId is required")
	}
	volume, err := d.GetVolume(volumeID, "")
	if err != nil {
		return errors.WithFieldsE(fields, "error getting volume", err)
	}

	fields["instanceId"] = volume[0].Attachments[0].InstanceID
	resp := volumeattach.Delete(
		d.client, volume[0].Attachments[0].InstanceID, volumeID)
	if resp.Err != nil {
		return errors.WithFieldsE(fields, "error deleting volume", err)
	}

	if !runAsync {
		log.WithFields(fields).Debug("waiting for volume to detach")
		err = d.waitVolumeDetach(volumeID)
		if err != nil {
			return errors.WithFieldsE(
				fields, "error waiting for volume to detach", err)
		}
	}

	log.WithFields(fields).Debug("volume detached")
	return nil
}

func (d *driver) waitVolumeAttach(volumeID string) error {

	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})

	if volumeID == "" {
		return errors.WithFields(fields, "volumeId is required")
	}
	for {
		volume, err := d.GetVolume(volumeID, "")
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

func (d *driver) waitVolumeDetach(volumeID string) error {

	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})

	if volumeID == "" {
		return errors.WithFields(fields, "volumeId is required")
	}
	for {
		volume, err := d.GetVolume(volumeID, "")
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

func (d *driver) CopySnapshot(
	runAsync bool, volumeID, snapshotID, snapshotName, destinationSnapshotName,
	destinationRegion string) (*core.Snapshot, error) {
	return nil, errors.New("This driver does not implement CopySnapshot")
}
