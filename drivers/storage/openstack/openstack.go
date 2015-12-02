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
	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/snapshots"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/volumes"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v2/extensions/volumeactions"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

const (
	providerName = "Openstack"
	minSize      = 1 // openstack has no minimum
)

type driver struct {
	provider             *gophercloud.ProviderClient
	client               *gophercloud.ServiceClient
	clientBlockStorage   *gophercloud.ServiceClient
	clientBlockStoragev2 *gophercloud.ServiceClient
	region               string
	availabilityZone     string
	instanceID           string
	r                    *core.RexRay
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

func init() {
	core.RegisterDriver(providerName, newDriver)
	gofig.Register(configRegistration())
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

	if d.regionName() == "" {
		if d.region, err = getInstanceRegion(d.r.Config); err != nil {
			return err
		}
	} else {
		d.region = d.regionName()
	}
	fields["region"] = d.region

	if d.availabilityZoneName() == "" {
		if d.availabilityZone, err = getInstanceAvailabilityZone(); err != nil {
			return err
		}
	} else {
		d.availabilityZone = d.availabilityZoneName()
	}
	fields["availabilityZone"] = d.availabilityZone

	authOpts := d.getAuthOptions()

	fields["identityEndpoint"] = d.authURL()
	fields["userId"] = d.userID()
	fields["userName"] = d.userName()
	if d.password() == "" {
		fields["password"] = ""
	} else {
		fields["password"] = "******"
	}
	fields["tenantId"] = d.tenantID()
	fields["tenantName"] = d.tenantName()
	fields["domainId"] = d.domainID()
	fields["domainName"] = d.domainName()

	if d.provider, err = openstack.AuthenticatedClient(authOpts); err != nil {
		return goof.WithFieldsE(fields,
			"error getting authenticated client", err)
	}

	if d.client, err = openstack.NewComputeV2(d.provider,
		gophercloud.EndpointOpts{Region: d.region}); err != nil {
		goof.WithFieldsE(fields, "error getting newComputeV2", err)
	}

	if d.clientBlockStorage, err = openstack.NewBlockStorageV1(d.provider,
		gophercloud.EndpointOpts{Region: d.region}); err != nil {
		return goof.WithFieldsE(fields,
			"error getting newBlockStorageV1", err)
	}

	fmt.Println(fmt.Sprintf("%v", d.clientBlockStorage))

	if d.clientBlockStoragev2, err = openstack.NewBlockStorageV2(d.provider,
		gophercloud.EndpointOpts{Region: d.region}); err != nil {
		return goof.WithFieldsE(fields,
			"error getting newBlockStorageV2", err)
	}
	fmt.Println(fmt.Sprintf("%v", d.clientBlockStoragev2))

	log.WithField("provider", providerName).Info("storage driver initialized")

	return nil
}

func (d *driver) Name() string {
	return providerName
}

func (d *driver) newCmd(name string, args ...string) *exec.Cmd {
	return newCmd(d.r.Config, name, args...)
}

func newCmd(c gofig.Config, name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Env = c.EnvVars()
	return cmd
}

func getInstanceID(c gofig.Config) (string, error) {
	cmd := newCmd(c, "/usr/sbin/dmidecode")
	cmdOut, err := cmd.Output()

	if err != nil {
		return "",
			goof.WithFields(eff(goof.Fields{
				"cmd.Path": cmd.Path,
				"cmd.Args": cmd.Args,
				"cmd.Out":  cmdOut,
			}), "error getting instance id")
	}

	rp := regexp.MustCompile("UUID:(.*)")
	uuid := strings.Replace(rp.FindString(string(cmdOut)), "UUID: ", "", -1)

	return strings.ToLower(uuid), nil
}

func (d *driver) getAuthOptions() gophercloud.AuthOptions {
	return gophercloud.AuthOptions{
		IdentityEndpoint: d.authURL(),
		UserID:           d.userID(),
		Username:         d.userName(),
		Password:         d.password(),
		TenantID:         d.tenantID(),
		TenantName:       d.tenantName(),
		DomainID:         d.domainID(),
		DomainName:       d.domainName(),
	}
}

func (d *driver) getInstance() (*servers.Server, error) {
	server, err := servers.Get(d.client, d.instanceID).Extract()
	if err != nil {
		return nil,
			goof.WithFieldsE(ef(), "error getting server instance", err)
	}

	return server, nil
}

func (d *driver) GetInstance() (*core.Instance, error) {
	server, err := d.getInstance()
	if err != nil {
		return nil,
			goof.WithFieldsE(ef(), "error getting driver instance", err)
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
			goof.WithFieldsE(eff(goof.Fields{
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
			goof.WithFieldsE(eff(goof.Fields{
				"instanceId": instanceID}),
				"error extracting volume attachments", err)
	}

	return volumeAttachments, nil

}

func getInstanceRegion(cfg gofig.Config) (string, error) {
	cmd := newCmd(
		cfg, "/usr/bin/xenstore-read",
		"vm-data/provider_data/region")

	cmdOut, err := cmd.Output()
	if err != nil {
		return "",
			goof.WithFields(eff(goof.Fields{
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
				goof.WithFieldsE(eff(goof.Fields{
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
				goof.WithFieldsE(eff(goof.Fields{
					"volumeId":   volumeID,
					"volumeName": volumeName}),
					"error listing volumes", err)
		}
		volumesRet, err = volumes.ExtractVolumes(allPages)
		if err != nil {
			return []volumes.Volume{},
				goof.WithFieldsE(eff(goof.Fields{
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
			goof.WithFieldsE(eff(goof.Fields{
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
			goof.WithFields(fields, "volumeId is required")
	}
	volume, err := d.GetVolume(volumeID, "")
	if err != nil {
		return []*core.VolumeAttachment{},
			goof.WithFieldsE(fields, "error getting volume attach", err)
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
				goof.WithFieldsE(fields, "error getting snapshot", err)
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
				goof.WithFieldsE(fields, "error listing snapshot", err)
		}

		allSnapshots, err = snapshots.ExtractSnapshots(allPages)
		if err != nil {
			return []snapshots.Snapshot{},
				goof.WithFieldsE(fields, "error extracting snapshot", err)
		}
	}

	return allSnapshots, nil
}

func (d *driver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {

	snapshots, err := d.getSnapshot(volumeID, snapshotID, snapshotName)
	if err != nil {
		return nil,
			goof.WithFieldsE(eff(goof.Fields{
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
			goof.WithFieldsE(fields, "error creating snapshot", err)
	}

	if !runAsync {
		log.Debug("waiting for snapshot creation to complete")
		err = snapshots.WaitForStatus(d.clientBlockStorage, resp.ID, "available", 120)
		if err != nil {
			return nil,
				goof.WithFieldsE(fields,
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
		return goof.WithFieldE(
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
		return nil, errors.ErrRunAsyncFromVolume
	}

	d.createVolumeEnsureAvailabilityZone(&availabilityZone)

	var err error

	if err = d.createVolumeHandleSnapshotID(
		&size, snapshotID, fields); err != nil {
		return nil, err
	}

	var volume []*core.Volume
	if volume, err = d.createVolumeHandleVolumeID(
		&availabilityZone, &snapshotID, &volumeID, &size, fields); err != nil {
		return nil, err
	}

	createVolumeEnsureSize(&size)

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
			goof.WithFields(fields, "error creating volume")
	}

	if !runAsync {
		log.Debug("waiting for volume creation to complete")
		err = volumes.WaitForStatus(d.clientBlockStorage, resp.ID, "available", 120)
		if err != nil {
			return nil,
				goof.WithFields(fields,
					"error waiting for volume creation to complete")
		}

		if volumeID != "" {
			err := d.RemoveSnapshot(snapshotID)
			if err != nil {
				return nil,
					goof.WithFields(fields,
						"error removing snapshot")
			}
		}
	}

	fields["volumeId"] = resp.ID
	fields["volumeName"] = ""

	volume, err = d.GetVolume(resp.ID, "")
	if err != nil {
		return nil, goof.WithFields(fields,
			"error removing snapshot")
	}

	log.WithFields(fields).Debug("created volume")
	return volume[0], nil
}

func (d *driver) createVolumeEnsureAvailabilityZone(availabilityZone *string) {
	if *availabilityZone == "" {
		*availabilityZone = d.availabilityZone
	}
}

func createVolumeEnsureSize(size *int64) {
	if *size != 0 && *size < minSize {
		*size = minSize
	}
}

func (d *driver) createVolumeHandleSnapshotID(
	size *int64, snapshotID string, fields map[string]interface{}) error {
	if snapshotID == "" {
		return nil
	}
	snapshots, err := d.GetSnapshot("", snapshotID, "")
	if err != nil {
		return goof.WithFieldsE(fields, "error getting snapshot", err)
	}

	if len(snapshots) == 0 {
		return goof.WithFields(fields, "snapshot array is empty")
	}

	volSize := snapshots[0].VolumeSize
	sizeInt, err := strconv.Atoi(volSize)
	if err != nil {
		f := goof.Fields{
			"volumeSize": volSize,
		}
		for k, v := range fields {
			f[k] = v
		}
		return goof.WithFieldsE(f, "error casting volume size", err)
	}
	*size = int64(sizeInt)
	return nil
}

func (d *driver) createVolumeHandleVolumeID(
	availabilityZone, snapshotID, volumeID *string,
	size *int64,
	fields map[string]interface{}) ([]*core.Volume, error) {

	if *volumeID == "" {
		return nil, nil
	}

	var err error
	var volume []*core.Volume

	if volume, err = d.GetVolume(*volumeID, ""); err != nil {
		return nil, goof.WithFieldsE(fields, "error getting volumes", err)
	}

	if len(volume) == 0 {
		return nil, goof.WithFieldsE(fields, "", errors.ErrNoVolumesReturned)
	}

	volSize := volume[0].Size
	sizeInt, err := strconv.Atoi(volSize)
	if err != nil {
		f := goof.Fields{
			"volumeSize": volSize,
		}
		for k, v := range fields {
			f[k] = v
		}
		return nil,
			goof.WithFieldsE(f, "error casting volume size", err)
	}
	*size = int64(sizeInt)

	*volumeID = volume[0].VolumeID
	snapshot, err := d.CreateSnapshot(
		false, fmt.Sprintf("temp-%s", *volumeID), *volumeID, "")
	if err != nil {
		return nil,
			goof.WithFields(fields, "error creating snapshot")
	}

	*snapshotID = snapshot[0].SnapshotID

	if *availabilityZone == "" {
		*availabilityZone = volume[0].AvailabilityZone
	}

	return volume, nil
}

func (d *driver) RemoveVolume(volumeID string) error {
	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})
	if volumeID == "" {
		return goof.WithFields(fields, "volumeId is required")
	}
	res := volumes.Delete(d.clientBlockStorage, volumeID)
	if res.Err != nil {
		return goof.WithFieldsE(fields, "error removing volume", res.Err)
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
	return "", goof.New("No available device")
}

func getLocalDevices() (deviceNames []string, err error) {
	file := "/proc/partitions"
	contentBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return []string{},
			goof.WithFieldsE(
				eff(goof.Fields{"file": file}), "error reading file", err)
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
	runAsync bool, volumeID, instanceID string, force bool) ([]*core.VolumeAttachment, error) {

	fields := eff(map[string]interface{}{
		"runAsync":   runAsync,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	nextDeviceName, err := d.GetDeviceNextAvailable()
	if err != nil {
		return nil, goof.WithFieldsE(
			fields, "error getting next available device", err)
	}

	if force {
		if err := d.DetachVolume(false, volumeID, "", true); err != nil {
			return nil, err
		}
	}

	options := &volumeattach.CreateOpts{
		Device:   nextDeviceName,
		VolumeID: volumeID,
	}

	_, err = volumeattach.Create(d.client, instanceID, options).Extract()
	if err != nil {
		return nil, goof.WithFieldsE(
			fields, "error attaching volume", err)
	}

	if !runAsync {
		log.WithFields(fields).Debug("waiting for volume to attach")
		err = d.waitVolumeAttach(volumeID)
		if err != nil {
			return nil, goof.WithFieldsE(
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
	runAsync bool, volumeID, instanceID string, force bool) error {

	fields := eff(map[string]interface{}{
		"runAsync":   runAsync,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return goof.WithFields(fields, "volumeId is required")
	}
	volume, err := d.GetVolume(volumeID, "")
	if err != nil {
		return goof.WithFieldsE(fields, "error getting volume", err)
	}

	if len(volume) == 0 {
		return goof.WithFields(fields, "no volumes returned")
	}

	if len(volume[0].Attachments) == 0 {
		return nil
	}

	fields["instanceId"] = volume[0].Attachments[0].InstanceID
	if force {
		if resp := volumeactions.ForceDetach(d.clientBlockStoragev2, volumeID); resp.Err != nil {
			log.Info(fmt.Sprintf("%+v", resp.Err))
			return goof.WithFieldsE(fields, "error forcing detach volume", resp.Err)
		}
	} else {
		if resp := volumeattach.Delete(
			d.client, volume[0].Attachments[0].InstanceID, volumeID); resp.Err != nil {
			return goof.WithFieldsE(fields, "error detaching volume", resp.Err)
		}
	}

	if !runAsync {
		log.WithFields(fields).Debug("waiting for volume to detach")
		err = d.waitVolumeDetach(volumeID)
		if err != nil {
			return goof.WithFieldsE(
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
		return goof.WithFields(fields, "volumeId is required")
	}
	for {
		volume, err := d.GetVolume(volumeID, "")
		if err != nil {
			return goof.WithFieldsE(fields, "error getting volume", err)
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
		return goof.WithFields(fields, "volumeId is required")
	}
	for {
		volume, err := d.GetVolume(volumeID, "")
		if err != nil {
			return goof.WithFieldsE(fields, "error getting volume", err)
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
	return nil, goof.New("This driver does not implement CopySnapshot")
}

func (d *driver) authURL() string {
	return d.r.Config.GetString("openstack.authURL")
}

func (d *driver) userID() string {
	return d.r.Config.GetString("openstack.userID")
}

func (d *driver) userName() string {
	return d.r.Config.GetString("openstack.userName")
}

func (d *driver) password() string {
	return d.r.Config.GetString("openstack.password")
}

func (d *driver) tenantID() string {
	return d.r.Config.GetString("openstack.tenantID")
}

func (d *driver) tenantName() string {
	return d.r.Config.GetString("openstack.tenantName")
}

func (d *driver) domainID() string {
	return d.r.Config.GetString("openstack.domainID")
}

func (d *driver) domainName() string {
	return d.r.Config.GetString("openstack.domainName")
}

func (d *driver) regionName() string {
	return d.r.Config.GetString("openstack.regionName")
}

func (d *driver) availabilityZoneName() string {
	return d.r.Config.GetString("openstack.availabilityZoneName")
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("Openstack")
	r.Key(gofig.String, "", "", "", "openstack.authURL")
	r.Key(gofig.String, "", "", "", "openstack.userID")
	r.Key(gofig.String, "", "", "", "openstack.userName")
	r.Key(gofig.String, "", "", "", "openstack.password")
	r.Key(gofig.String, "", "", "", "openstack.tenantID")
	r.Key(gofig.String, "", "", "", "openstack.tenantName")
	r.Key(gofig.String, "", "", "", "openstack.domainID")
	r.Key(gofig.String, "", "", "", "openstack.domainName")
	r.Key(gofig.String, "", "", "", "openstack.regionName")
	r.Key(gofig.String, "", "", "", "openstack.availabilityZoneName")
	return r
}
