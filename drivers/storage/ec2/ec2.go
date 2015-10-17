package ec2

import (
	log "github.com/Sirupsen/logrus"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/ec2"
)

const providerName = "ec2"

// The EC2 storage driver.
type driver struct {
	instanceDocument *instanceIdentityDocument
	ec2Instance      *ec2.EC2
	r                *core.RexRay
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

	var err error
	d.instanceDocument, err = getInstanceIdendityDocument()
	if err != nil {
		return errors.WithFields(ef(), "error getting instance id doc")
	}

	auth := aws.Auth{
		AccessKey: d.r.Config.AwsAccessKey,
		SecretKey: d.r.Config.AwsSecretKey,
	}
	region := d.r.Config.AwsRegion
	if region == "" {
		region = d.instanceDocument.Region
	}
	d.ec2Instance = ec2.New(
		auth,
		aws.Regions[region],
	)

	log.WithField("provider", providerName).Debug("storage driver initialized")

	return nil
}

type instanceIdentityDocument struct {
	InstanceID         string      `json:"instanceId"`
	BillingProducts    interface{} `json:"billingProducts"`
	AccountID          string      `json:"accountId"`
	ImageID            string      `json:"imageId"`
	InstanceType       string      `json:"instanceType"`
	KernelID           string      `json:"kernelId"`
	RamdiskID          string      `json:"ramdiskId"`
	PendingTime        string      `json:"pendingTime"`
	Architecture       string      `json:"architecture"`
	Region             string      `json:"region"`
	Version            string      `json:"version"`
	AvailabilityZone   string      `json:"availabilityZone"`
	DevpayproductCodes interface{} `json:"devpayProductCodes"`
	PrivateIP          string      `json:"privateIp"`
}

func (d *driver) Name() string {
	return providerName
}

func (d *driver) GetVolumeMapping() ([]*core.BlockDevice, error) {
	blockDevices, err := d.getBlockDevices(d.instanceDocument.InstanceID)
	if err != nil {
		return nil, err
	}

	var BlockDevices []*core.BlockDevice
	for _, blockDevice := range blockDevices {
		sdBlockDevice := &core.BlockDevice{
			ProviderName: providerName,
			InstanceID:   d.instanceDocument.InstanceID,
			Region:       d.instanceDocument.Region,
			DeviceName:   blockDevice.DeviceName,
			VolumeID:     blockDevice.EBS.VolumeId,
			Status:       blockDevice.EBS.Status,
		}
		BlockDevices = append(BlockDevices, sdBlockDevice)
	}

	// log.Println("Got Block Device Mappings: " + fmt.Sprintf("%+v", BlockDevices))
	return BlockDevices, nil
}

func getInstanceIdendityDocument() (*instanceIdentityDocument, error) {
	conn, err := net.DialTimeout("tcp", "169.254.169.254:80", 50*time.Millisecond)
	if err != nil {
		return &instanceIdentityDocument{}, fmt.Errorf("Error: %v\n", err)
	}
	defer conn.Close()

	url := "http://169.254.169.254/latest/dynamic/instance-identity/document"
	resp, err := http.Get(url)
	if err != nil {
		return &instanceIdentityDocument{}, fmt.Errorf("Error: %v\n", err)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &instanceIdentityDocument{}, fmt.Errorf("Error: %v\n", err)
	}

	var document instanceIdentityDocument
	err = json.Unmarshal(data, &document)
	if err != nil {
		return &instanceIdentityDocument{}, fmt.Errorf("Error: %v\n", err)
	}

	return &document, nil
}

func (d *driver) getBlockDevices(instanceID string) ([]ec2.BlockDevice, error) {

	instance, err := d.getInstance()
	if err != nil {
		return []ec2.BlockDevice{}, err
	}

	return instance.BlockDevices, nil

}

func getInstanceName(server ec2.Instance) string {
	return getTag(server, "Name")
}

func getTag(server ec2.Instance, key string) string {
	for _, tag := range server.Tags {
		if tag.Key == key {
			return tag.Value
		}
	}
	return ""
}

func (d *driver) GetInstance() (*core.Instance, error) {

	server, err := d.getInstance()
	if err != nil {
		return &core.Instance{}, err
	}

	instance := &core.Instance{
		ProviderName: providerName,
		InstanceID:   d.instanceDocument.InstanceID,
		Region:       d.instanceDocument.Region,
		Name:         getInstanceName(server),
	}

	// log.Println("Got Instance: " + fmt.Sprintf("%+v", instance))
	return instance, nil
}

func (d *driver) getInstance() (ec2.Instance, error) {

	resp, err := d.ec2Instance.DescribeInstances(
		[]string{
			d.instanceDocument.InstanceID},
		&ec2.Filter{})
	if err != nil {
		return ec2.Instance{}, err
	}

	return resp.Reservations[0].Instances[0], nil
}

func (d *driver) CreateSnapshot(
	runAsync bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {

	resp, err := d.ec2Instance.CreateSnapshot(volumeID, description)
	if err != nil {
		return nil, err
	}

	if snapshotName != "" {
		_, err := d.ec2Instance.CreateTags(
			[]string{resp.Id}, []ec2.Tag{{"Name", snapshotName}})
		if err != nil {
			return nil, err
		}
	}

	if !runAsync {
		log.Println("Waiting for snapshot to complete")
		err = d.waitSnapshotComplete(resp.Snapshot.Id)
		if err != nil {
			return nil, err
		}
	}

	snapshot, err := d.GetSnapshot("", resp.Snapshot.Id, "")
	if err != nil {
		return nil, err
	}

	log.Println("Created Snapshot: " + snapshot[0].SnapshotID)
	return snapshot, nil

}

func (d *driver) getSnapshot(
	volumeID, snapshotID, snapshotName string) ([]ec2.Snapshot, error) {
	filter := ec2.NewFilter()
	if snapshotName != "" {
		filter.Add("tag:Name", fmt.Sprintf("%s", snapshotName))
	}

	if volumeID != "" {
		filter.Add("volume-id", volumeID)
	}

	snapshotList := []string{}
	if snapshotID != "" {
		//using snapshotList is returning stale data
		//snapshotList = append(snapshotList, snapshotID)
		filter.Add("snapshot-id", snapshotID)
	}

	resp, err := d.ec2Instance.Snapshots(snapshotList, filter)
	if err != nil {
		return []ec2.Snapshot{}, err
	}

	return resp.Snapshots, nil
}

func (d *driver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {

	snapshots, err := d.getSnapshot(volumeID, snapshotID, snapshotName)
	if err != nil {
		return nil, err
	}

	var snapshotsInt []*core.Snapshot
	for _, snapshot := range snapshots {
		name := getName(snapshot.Tags)
		snapshotSD := &core.Snapshot{
			Name:        name,
			VolumeID:    snapshot.VolumeId,
			SnapshotID:  snapshot.Id,
			VolumeSize:  snapshot.VolumeSize,
			StartTime:   snapshot.StartTime,
			Description: snapshot.Description,
			Status:      snapshot.Status,
		}
		snapshotsInt = append(snapshotsInt, snapshotSD)
	}

	// log.Println("Got Snapshots: " + fmt.Sprintf("%+v", snapshotsInt))
	return snapshotsInt, nil
}

func (d *driver) RemoveSnapshot(snapshotID string) error {
	_, err := d.ec2Instance.DeleteSnapshots([]string{snapshotID})
	if err != nil {
		return err
	}

	log.Println("Removed Snapshot: " + snapshotID)
	return nil
}

func (d *driver) GetDeviceNextAvailable() (string, error) {
	letters := []string{
		"a", "b", "c", "d", "e", "f", "g", "h",
		"i", "j", "k", "l", "m", "n", "o", "p"}

	blockDeviceNames := make(map[string]bool)

	blockDeviceMapping, err := d.GetVolumeMapping()
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
		return []string{}, err
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

func (d *driver) CreateVolume(
	runAsync bool, volumeName, volumeID, snapshotID, volumeType string,
	IOPS, size int64, availabilityZone string) (*core.Volume, error) {

	resp, err := d.createVolume(
		runAsync, volumeName, volumeID, snapshotID, volumeType,
		IOPS, size, availabilityZone)

	if err != nil {
		return nil, err
	}

	volumes, err := d.GetVolume(resp.VolumeId, "")
	if err != nil {
		return nil, err
	}

	// log.Println(fmt.Sprintf("Created volume: %+v", volumes[0]))
	return volumes[0], nil

}

func (d *driver) createVolume(
	runAsync bool, volumeName, volumeID, snapshotID, volumeType string,
	IOPS, size int64, availabilityZone string) (*ec2.CreateVolumeResp, error) {

	if volumeID != "" && runAsync {
		return &ec2.CreateVolumeResp{},
			errors.New("Cannot create volume from volume and run asynchronously")
	}

	server, err := d.getInstance()
	if err != nil {
		return &ec2.CreateVolumeResp{}, err
	}

	var snapshot *core.Snapshot
	if volumeID != "" {
		snapshotInt, err := d.CreateSnapshot(
			true, fmt.Sprintf("temp-%v", volumeID),
			volumeID, "created for createVolume")

		if err != nil {
			return &ec2.CreateVolumeResp{}, err
		}
		snapshot = snapshotInt[0]
		snapshotID = snapshot.SnapshotID
	}

	if availabilityZone == "" {
		availabilityZone = server.AvailabilityZone
	}

	options := &ec2.CreateVolume{
		Size:       size,
		SnapshotId: snapshotID,
		AvailZone:  availabilityZone,
		VolumeType: volumeType,
		IOPS:       IOPS,
	}
	resp := &ec2.CreateVolumeResp{}
	for {
		resp, err = d.ec2Instance.CreateVolume(options)
		if err != nil {
			if err.Error() ==
				"Snapshot is in invalid state - pending (IncorrectState)" {
				time.Sleep(1 * time.Second)
				continue
			}
			return &ec2.CreateVolumeResp{}, err
		}
		break
	}

	if volumeName != "" {
		_, err := d.ec2Instance.CreateTags(
			[]string{resp.VolumeId}, []ec2.Tag{{"Name", volumeName}})
		if err != nil {
			return &ec2.CreateVolumeResp{}, err
		}
	}

	if !runAsync {
		log.Println("Waiting for volume creation to complete")
		err = d.waitVolumeComplete(resp.VolumeId)
		if err != nil {
			return &ec2.CreateVolumeResp{}, err
		}

		if volumeID != "" {
			err := d.RemoveSnapshot(snapshotID)
			if err != nil {
				return &ec2.CreateVolumeResp{}, err
			}
		}
	}

	return resp, nil
}

func (d *driver) getVolume(
	volumeID, volumeName string) ([]ec2.Volume, error) {

	filter := ec2.NewFilter()
	if volumeName != "" {
		filter.Add("tag:Name", fmt.Sprintf("%s", volumeName))
	}

	volumeList := []string{}
	if volumeID != "" {
		volumeList = append(volumeList, volumeID)
	}

	resp, err := d.ec2Instance.Volumes(volumeList, filter)
	if err != nil {
		return []ec2.Volume{}, err
	}

	return resp.Volumes, nil
}

func getName(tags []ec2.Tag) string {
	for _, tag := range tags {
		if tag.Key == "Name" {
			return tag.Value
		}
	}
	return ""
}

func (d *driver) GetVolume(
	volumeID, volumeName string) ([]*core.Volume, error) {

	volumes, err := d.getVolume(volumeID, volumeName)
	if err != nil {
		return []*core.Volume{}, err
	}

	var volumesSD []*core.Volume
	for _, volume := range volumes {
		var attachmentsSD []*core.VolumeAttachment
		for _, attachment := range volume.Attachments {
			attachmentSD := &core.VolumeAttachment{
				VolumeID:   attachment.VolumeId,
				InstanceID: attachment.InstanceId,
				DeviceName: attachment.Device,
				Status:     attachment.Status,
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		name := getName(volume.Tags)

		volumeSD := &core.Volume{
			Name:             name,
			VolumeID:         volume.VolumeId,
			AvailabilityZone: volume.AvailZone,
			Status:           volume.Status,
			VolumeType:       volume.VolumeType,
			IOPS:             volume.IOPS,
			Size:             volume.Size,
			Attachments:      attachmentsSD,
		}
		volumesSD = append(volumesSD, volumeSD)
	}

	return volumesSD, nil
}

func (d *driver) GetVolumeAttach(
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {

	if volumeID == "" {
		return []*core.VolumeAttachment{}, errors.ErrMissingVolumeID
	}

	volumes, err := d.GetVolume(volumeID, "")
	if err != nil {
		return []*core.VolumeAttachment{}, err
	}

	if instanceID != "" {
		var attached bool
		for _, volumeAttachment := range volumes[0].Attachments {
			if volumeAttachment.InstanceID == instanceID {
				return volumes[0].Attachments, nil
			}
		}
		if !attached {
			return []*core.VolumeAttachment{}, nil
		}
	}
	return volumes[0].Attachments, nil
}

func (d *driver) waitSnapshotComplete(snapshotID string) error {
	for {

		snapshots, err := d.getSnapshot("", snapshotID, "")
		if err != nil {
			return err
		}

		snapshot := snapshots[0]
		if snapshot.Status == "completed" {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (d *driver) waitVolumeComplete(volumeID string) error {
	if volumeID == "" {
		return errors.ErrMissingVolumeID
	}

	for {
		volumes, err := d.getVolume(volumeID, "")
		if err != nil {
			return err
		}

		if volumes[0].Status == "available" {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (d *driver) waitVolumeAttach(volumeID, instanceID string) error {
	if volumeID == "" {
		return errors.ErrMissingVolumeID
	}

	for {
		volume, err := d.GetVolumeAttach(volumeID, instanceID)
		if err != nil {
			return err
		}
		if volume[0].Status == "attached" {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (d *driver) waitVolumeDetach(volumeID string) error {
	if volumeID == "" {
		return errors.ErrMissingVolumeID
	}

	for {
		volume, err := d.GetVolumeAttach(volumeID, "")
		if err != nil {
			return err
		}

		if len(volume) == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (d *driver) RemoveVolume(volumeID string) error {
	if volumeID == "" {
		return errors.ErrMissingVolumeID
	}

	_, err := d.ec2Instance.DeleteVolume(volumeID)
	if err != nil {
		return err
	}

	log.Println("Deleted Volume: " + volumeID)
	return nil
}

func (d *driver) AttachVolume(
	runAsync bool,
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {

	if volumeID == "" {
		return nil, errors.ErrMissingVolumeID
	}

	nextDeviceName, err := d.GetDeviceNextAvailable()
	if err != nil {
		return nil, err
	}

	_, err = d.ec2Instance.AttachVolume(
		volumeID, instanceID, nextDeviceName)

	if err != nil {
		return nil, err
	}

	if !runAsync {
		log.Println("Waiting for volume attachment to complete")
		err = d.waitVolumeAttach(volumeID, instanceID)
		if err != nil {
			return nil, err
		}
	}

	volumeAttachment, err := d.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf(
		"Attached volume %s to instance %s", volumeID, instanceID))
	return volumeAttachment, nil
}

func (d *driver) DetachVolume(
	runAsync bool,
	volumeID, blank string) error {

	if volumeID == "" {
		return errors.ErrMissingVolumeID
	}

	_, err := d.ec2Instance.DetachVolume(volumeID)
	if err != nil {
		return err
	}

	if !runAsync {
		log.Println("Waiting for volume detachment to complete")
		err = d.waitVolumeDetach(volumeID)
		if err != nil {
			return err
		}
	}

	log.Println("Detached volume", volumeID)
	return nil
}

func (d *driver) CopySnapshot(runAsync bool,
	volumeID, snapshotID, snapshotName, destinationSnapshotName,
	destinationRegion string) (*core.Snapshot, error) {

	if volumeID == "" && snapshotID == "" && snapshotName == "" {
		return nil, errors.New("Missing volumeID, snapshotID, or snapshotName")
	}

	snapshots, err := d.getSnapshot(volumeID, snapshotID, snapshotName)
	if err != nil {
		return nil, err
	}

	if len(snapshots) > 1 {
		return nil, errors.ErrMultipleVolumesReturned
	} else if len(snapshots) == 0 {
		return nil, errors.ErrNoVolumesReturned
	}

	snapshotID = snapshots[0].Id

	options := &ec2.CopySnapshot{
		SourceRegion:      d.ec2Instance.Region.Name,
		DestinationRegion: destinationRegion,
		SourceSnapshotId:  snapshotID,
		Description: fmt.Sprintf("[Copied %s from %s]",
			snapshotID, d.ec2Instance.Region.Name),
	}
	resp := &ec2.CopySnapshotResp{}

	auth := aws.Auth{
		AccessKey: d.r.Config.AwsAccessKey,
		SecretKey: d.r.Config.AwsSecretKey}
	destec2Instance := ec2.New(
		auth,
		aws.Regions[destinationRegion],
	)

	origec2Instance := d.ec2Instance
	d.ec2Instance = destec2Instance
	defer func() { d.ec2Instance = origec2Instance }()

	resp, err = d.ec2Instance.CopySnapshot(options)
	if err != nil {
		return nil, err
	}

	if destinationSnapshotName != "" {
		_, err := d.ec2Instance.CreateTags(
			[]string{resp.SnapshotId},
			[]ec2.Tag{{"Name", destinationSnapshotName}})

		if err != nil {
			return nil, err
		}
	}

	if !runAsync {
		log.Println("Waiting for snapshot copy to complete")
		err = d.waitSnapshotComplete(resp.SnapshotId)
		if err != nil {
			return nil, err
		}
	}

	snapshot, err := d.GetSnapshot("", resp.SnapshotId, "")
	if err != nil {
		return nil, err
	}

	return snapshot[0], nil
}
