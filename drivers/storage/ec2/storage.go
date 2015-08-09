package ec2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/emccode/rexray/drivers/storage"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/ec2"
)

var (
	providerName string
)

type Driver struct {
	InstanceDocument *instanceIdentityDocument
	EC2Instance      *ec2.EC2
}

var (
	ErrMissingVolumeID         = errors.New("Missing VolumeID")
	ErrMultipleVolumesReturned = errors.New("Multiple Volumes returned")
	ErrNoVolumesReturned       = errors.New("No Volumes returned")
)

func init() {
	providerName = "ec2"
	storagedriver.Register("ec2", Init)
}

func Init() (storagedriver.Driver, error) {
	instanceDocument, err := getInstanceIdendityDocument()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", storagedriver.ErrDriverInstanceDiscovery, err)
	}

	auth := aws.Auth{AccessKey: os.Getenv("AWS_ACCESS_KEY"), SecretKey: os.Getenv("AWS_SECRET_KEY")}
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = instanceDocument.Region
	}
	ec2Instance := ec2.New(
		auth,
		aws.Regions[region],
	)

	// table := InitDD(auth, aws.Regions[instanceDocument.Region])

	driver := &Driver{
		EC2Instance:      ec2Instance,
		InstanceDocument: instanceDocument,
		// DDTable:          table,
	}

	if os.Getenv("REXRAY_DEBUG") == "true" {
		log.Println("Storage Driver Initialized: " + providerName)
	}

	return driver, nil
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

func (driver *Driver) GetVolumeMapping() ([]*storagedriver.BlockDevice, error) {
	blockDevices, err := driver.getBlockDevices(driver.InstanceDocument.InstanceID)
	if err != nil {
		return nil, err
	}

	var BlockDevices []*storagedriver.BlockDevice
	for _, blockDevice := range blockDevices {
		sdBlockDevice := &storagedriver.BlockDevice{
			ProviderName: providerName,
			InstanceID:   driver.InstanceDocument.InstanceID,
			Region:       driver.InstanceDocument.Region,
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

func (driver *Driver) getBlockDevices(instanceID string) ([]ec2.BlockDevice, error) {

	instance, err := driver.getInstance()
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

func (driver *Driver) GetInstance() (*storagedriver.Instance, error) {

	server, err := driver.getInstance()
	if err != nil {
		return &storagedriver.Instance{}, err
	}

	instance := &storagedriver.Instance{
		ProviderName: providerName,
		InstanceID:   driver.InstanceDocument.InstanceID,
		Region:       driver.InstanceDocument.Region,
		Name:         getInstanceName(server),
	}

	// log.Println("Got Instance: " + fmt.Sprintf("%+v", instance))
	return instance, nil
}

func (driver *Driver) getInstance() (ec2.Instance, error) {

	resp, err := driver.EC2Instance.DescribeInstances([]string{driver.InstanceDocument.InstanceID}, &ec2.Filter{})
	if err != nil {
		return ec2.Instance{}, err
	}

	return resp.Reservations[0].Instances[0], nil
}

func (driver *Driver) CreateSnapshot(runAsync bool, snapshotName, volumeID, description string) ([]*storagedriver.Snapshot, error) {
	resp, err := driver.EC2Instance.CreateSnapshot(volumeID, description)
	if err != nil {
		return nil, err
	}

	if snapshotName != "" {
		_, err := driver.EC2Instance.CreateTags([]string{resp.Id}, []ec2.Tag{{"Name", snapshotName}})
		if err != nil {
			return nil, err
		}
	}

	if !runAsync {
		log.Println("Waiting for snapshot to complete")
		err = driver.waitSnapshotComplete(resp.Snapshot.Id)
		if err != nil {
			return nil, err
		}
	}

	snapshot, err := driver.GetSnapshot("", resp.Snapshot.Id, "")
	if err != nil {
		return nil, err
	}

	log.Println("Created Snapshot: " + snapshot[0].SnapshotID)
	return snapshot, nil

}

func (driver *Driver) getSnapshot(volumeID, snapshotID, snapshotName string) ([]ec2.Snapshot, error) {
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

	resp, err := driver.EC2Instance.Snapshots(snapshotList, filter)
	if err != nil {
		return []ec2.Snapshot{}, err
	}

	return resp.Snapshots, nil
}

func (driver *Driver) GetSnapshot(volumeID, snapshotID, snapshotName string) ([]*storagedriver.Snapshot, error) {
	snapshots, err := driver.getSnapshot(volumeID, snapshotID, snapshotName)
	if err != nil {
		return nil, err
	}

	var snapshotsInt []*storagedriver.Snapshot
	for _, snapshot := range snapshots {
		name := getName(snapshot.Tags)
		snapshotSD := &storagedriver.Snapshot{
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

func (driver *Driver) RemoveSnapshot(snapshotID string) error {
	_, err := driver.EC2Instance.DeleteSnapshots([]string{snapshotID})
	if err != nil {
		return err
	}

	log.Println("Removed Snapshot: " + snapshotID)
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

func (driver *Driver) CreateVolume(runAsync bool, volumeName string, volumeID string, snapshotID string, volumeType string, IOPS int64, size int64, availabilityZone string) (*storagedriver.Volume, error) {
	resp, err := driver.createVolume(runAsync, volumeName, volumeID, snapshotID, volumeType, IOPS, size, availabilityZone)
	if err != nil {
		return nil, err
	}

	volumes, err := driver.GetVolume(resp.VolumeId, "")
	if err != nil {
		return nil, err
	}

	// log.Println(fmt.Sprintf("Created volume: %+v", volumes[0]))
	return volumes[0], nil

}

func (driver *Driver) createVolume(runAsync bool, volumeName string, volumeID string, snapshotID string, volumeType string, IOPS int64, size int64, availabilityZone string) (*ec2.CreateVolumeResp, error) {

	if volumeID != "" && runAsync {
		return &ec2.CreateVolumeResp{}, errors.New("Cannot create volume from volume and run asynchronously")
	}

	server, err := driver.getInstance()
	if err != nil {
		return &ec2.CreateVolumeResp{}, err
	}

	var snapshot *storagedriver.Snapshot
	if volumeID != "" {
		snapshotInt, err := driver.CreateSnapshot(true, fmt.Sprintf("temp-%v", volumeID), volumeID, "created for createVolume")
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
		resp, err = driver.EC2Instance.CreateVolume(options)
		if err != nil {
			if err.Error() == "Snapshot is in invalid state - pending (IncorrectState)" {
				time.Sleep(1 * time.Second)
				continue
			}
			return &ec2.CreateVolumeResp{}, err
		}
		break
	}

	if volumeName != "" {
		_, err := driver.EC2Instance.CreateTags([]string{resp.VolumeId}, []ec2.Tag{{"Name", volumeName}})
		if err != nil {
			return &ec2.CreateVolumeResp{}, err
		}
	}

	if !runAsync {
		log.Println("Waiting for volume creation to complete")
		err = driver.waitVolumeComplete(resp.VolumeId)
		if err != nil {
			return &ec2.CreateVolumeResp{}, err
		}

		if volumeID != "" {
			err := driver.RemoveSnapshot(snapshotID)
			if err != nil {
				return &ec2.CreateVolumeResp{}, err
			}
		}
	}

	return resp, nil
}

func (driver *Driver) getVolume(volumeID, volumeName string) ([]ec2.Volume, error) {
	filter := ec2.NewFilter()
	if volumeName != "" {
		filter.Add("tag:Name", fmt.Sprintf("%s", volumeName))
	}

	volumeList := []string{}
	if volumeID != "" {
		volumeList = append(volumeList, volumeID)
	}

	resp, err := driver.EC2Instance.Volumes(volumeList, filter)
	if err != nil {
		return []ec2.Volume{}, err
	}

	return resp.Volumes, nil
}

func getName(tags []ec2.Tag) string {
	for _, tag := range tags {
		if tag.Key == "Name" {
			return tag.Value
			break
		}
	}
	return ""
}

func (driver *Driver) GetVolume(volumeID, volumeName string) ([]*storagedriver.Volume, error) {

	volumes, err := driver.getVolume(volumeID, volumeName)
	if err != nil {
		return []*storagedriver.Volume{}, err
	}

	var volumesSD []*storagedriver.Volume
	for _, volume := range volumes {
		var attachmentsSD []*storagedriver.VolumeAttachment
		for _, attachment := range volume.Attachments {
			attachmentSD := &storagedriver.VolumeAttachment{
				VolumeID:   attachment.VolumeId,
				InstanceID: attachment.InstanceId,
				DeviceName: attachment.Device,
				Status:     attachment.Status,
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		name := getName(volume.Tags)

		volumeSD := &storagedriver.Volume{
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

func (driver *Driver) GetVolumeAttach(volumeID, instanceID string) ([]*storagedriver.VolumeAttachment, error) {
	if volumeID == "" {
		return []*storagedriver.VolumeAttachment{}, ErrMissingVolumeID
	}

	volumes, err := driver.GetVolume(volumeID, "")
	if err != nil {
		return []*storagedriver.VolumeAttachment{}, err
	}

	if instanceID != "" {
		var attached bool
		for _, volumeAttachment := range volumes[0].Attachments {
			if volumeAttachment.InstanceID == instanceID {
				return volumes[0].Attachments, nil
			}
		}
		if !attached {
			return []*storagedriver.VolumeAttachment{}, nil
		}
	}
	return volumes[0].Attachments, nil
}

func (driver *Driver) waitSnapshotComplete(snapshotID string) error {
	for {

		snapshots, err := driver.getSnapshot("", snapshotID, "")
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

func (driver *Driver) waitVolumeComplete(volumeID string) error {
	if volumeID == "" {
		return ErrMissingVolumeID
	}

	for {
		volumes, err := driver.getVolume(volumeID, "")
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

func (driver *Driver) waitVolumeAttach(volumeID, instanceID string) error {
	if volumeID == "" {
		return ErrMissingVolumeID
	}

	for {
		volume, err := driver.GetVolumeAttach(volumeID, instanceID)
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

func (driver *Driver) waitVolumeDetach(volumeID string) error {
	if volumeID == "" {
		return ErrMissingVolumeID
	}

	for {
		volume, err := driver.GetVolumeAttach(volumeID, "")
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

func (driver *Driver) RemoveVolume(volumeID string) error {
	if volumeID == "" {
		return ErrMissingVolumeID
	}

	_, err := driver.EC2Instance.DeleteVolume(volumeID)
	if err != nil {
		return err
	}

	log.Println("Deleted Volume: " + volumeID)
	return nil
}

func (driver *Driver) AttachVolume(runAsync bool, volumeID, instanceID string) ([]*storagedriver.VolumeAttachment, error) {
	if volumeID == "" {
		return nil, ErrMissingVolumeID
	}

	nextDeviceName, err := driver.GetDeviceNextAvailable()
	if err != nil {
		return nil, err
	}

	_, err = driver.EC2Instance.AttachVolume(volumeID, instanceID, nextDeviceName)
	if err != nil {
		return nil, err
	}

	if !runAsync {
		log.Println("Waiting for volume attachment to complete")
		err = driver.waitVolumeAttach(volumeID, instanceID)
		if err != nil {
			return nil, err
		}
	}

	volumeAttachment, err := driver.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("Attached volume %s to instance %s", volumeID, instanceID))
	return volumeAttachment, nil
}

func (driver *Driver) DetachVolume(runAsync bool, volumeID string, blank string) error {
	if volumeID == "" {
		return ErrMissingVolumeID
	}

	_, err := driver.EC2Instance.DetachVolume(volumeID)
	if err != nil {
		return err
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

func (driver *Driver) CopySnapshot(runAsync bool, volumeID, snapshotID, snapshotName, destinationSnapshotName, destinationRegion string) (*storagedriver.Snapshot, error) {
	if volumeID == "" && snapshotID == "" && snapshotName == "" {
		return nil, errors.New("Missing volumeID, snapshotID, or snapshotName")
	}

	snapshots, err := driver.getSnapshot(volumeID, snapshotID, snapshotName)
	if err != nil {
		return nil, err
	}

	if len(snapshots) > 1 {
		return nil, ErrMultipleVolumesReturned
	} else if len(snapshots) == 0 {
		return nil, ErrNoVolumesReturned
	}

	snapshotID = snapshots[0].Id

	options := &ec2.CopySnapshot{
		SourceRegion:      driver.EC2Instance.Region.Name,
		DestinationRegion: destinationRegion,
		SourceSnapshotId:  snapshotID,
		Description:       fmt.Sprintf("[Copied %s from %s]", snapshotID, driver.EC2Instance.Region.Name),
	}
	resp := &ec2.CopySnapshotResp{}

	auth := aws.Auth{AccessKey: os.Getenv("AWS_ACCESS_KEY"), SecretKey: os.Getenv("AWS_SECRET_KEY")}
	destEC2Instance := ec2.New(
		auth,
		aws.Regions[destinationRegion],
	)

	origEC2Instance := driver.EC2Instance
	driver.EC2Instance = destEC2Instance
	defer func() { driver.EC2Instance = origEC2Instance }()

	resp, err = driver.EC2Instance.CopySnapshot(options)
	if err != nil {
		return nil, err
	}

	if destinationSnapshotName != "" {
		_, err := driver.EC2Instance.CreateTags([]string{resp.SnapshotId}, []ec2.Tag{{"Name", destinationSnapshotName}})
		if err != nil {
			return nil, err
		}
	}

	if !runAsync {
		log.Println("Waiting for snapshot copy to complete")
		err = driver.waitSnapshotComplete(resp.SnapshotId)
		if err != nil {
			return nil, err
		}
	}

	snapshot, err := driver.GetSnapshot("", resp.SnapshotId, "")
	if err != nil {
		return nil, err
	}

	return snapshot[0], nil
}
