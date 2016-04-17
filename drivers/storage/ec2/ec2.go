package ec2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

const (
	providerName = "ec2"
	rexrayTag    = "rexraySet"
)

// The EC2 storage driver.
type driver struct {
	instanceDocument *instanceIdentityDocument
	ec2Instance      *ec2.EC2
	ec2Tag           string
	ec2creds         *credentials.Credentials
	r                *core.RexRay
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

	fields := eff(map[string]interface{}{
		"moduleName": d.r.Context,
		"accessKey":  d.accessKey(),
	})

	if d.secretKey() == "" {
		fields["secretKey"] = ""
	} else {
		fields["secretKey"] = "******"
	}

	var err error
	d.instanceDocument, err = getInstanceIdendityDocument()
	if err != nil {
		return goof.WithFields(ef(), "error getting instance id doc")
	}

	region := d.region()
	if region == "" {
		region = d.instanceDocument.Region
	}

	d.ec2Tag = d.rexrayTag()

	mySession := session.New()

	d.ec2creds = credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.StaticProvider{Value: credentials.Value{AccessKeyID: d.accessKey(), SecretAccessKey: d.secretKey()}},
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(mySession),
			},
		})

	config := aws.NewConfig().WithCredentials(d.ec2creds).WithRegion(region)

	d.ec2Instance = ec2.New(mySession, config)

	log.WithFields(fields).Info("storage driver initialized")

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
			DeviceName:   *blockDevice.DeviceName,
			VolumeID:     *((*blockDevice.Ebs).VolumeId),
			Status:       *((*blockDevice.Ebs).Status),
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

func (d *driver) getBlockDevices(instanceID string) ([]*ec2.InstanceBlockDeviceMapping, error) {

	instance, err := d.getInstance()
	if err != nil {
		return nil, err
	}

	return instance.BlockDeviceMappings, nil
}

func getInstanceName(server ec2.Instance) string {
	return getTag(server, "Name")
}

func getTag(server ec2.Instance, key string) string {
	for _, tag := range server.Tags {
		if *tag.Key == key {
			return *tag.Value
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

	diInput := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{&d.instanceDocument.InstanceID},
	}
	resp, err := d.ec2Instance.DescribeInstances(diInput)
	if err != nil {
		return ec2.Instance{}, err
	}

	return *resp.Reservations[0].Instances[0], nil
}

func (d *driver) CreateSnapshot(
	runAsync bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {

	csInput := &ec2.CreateSnapshotInput{
		VolumeId:    &volumeID, // Required
		Description: &description,
	}

	resp, err := d.ec2Instance.CreateSnapshot(csInput)
	if err != nil {
		return nil, err
	}

	if snapshotName != "" || d.ec2Tag != "" {

		var ctInput *ec2.CreateTagsInput
		initCTInput := func() {
			if ctInput != nil {
				return
			}
			ctInput = &ec2.CreateTagsInput{
				Resources: []*string{resp.SnapshotId},
				Tags:      []*ec2.Tag{},
			}
		}

		if snapshotName != "" {
			initCTInput()
			ctInput.Tags = append(
				ctInput.Tags,
				&ec2.Tag{
					Key:   aws.String("Name"),
					Value: &snapshotName,
				})
		}

		if d.ec2Tag != "" {
			initCTInput()
			ctInput.Tags = append(
				ctInput.Tags,
				&ec2.Tag{
					Key:   aws.String(rexrayTag),
					Value: &d.ec2Tag,
				})
		}

		_, err = d.ec2Instance.CreateTags(ctInput)
		if err != nil {
			return nil, err
		}
	}

	if !runAsync {
		log.Println("Waiting for snapshot to complete")
		err = d.waitSnapshotComplete(*resp.SnapshotId)
		if err != nil {
			return nil, err
		}
	}

	snapshot, err := d.GetSnapshot("", *resp.SnapshotId, "")
	if err != nil {
		return nil, err
	}

	log.Println("Created Snapshot: " + snapshot[0].SnapshotID)
	return snapshot, nil
}

func (d *driver) getSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*ec2.Snapshot, error) {

	filters := []*ec2.Filter{}
	if snapshotName != "" {
		filters = append(filters, &ec2.Filter{
			Name: aws.String("tag:Name"), Values: []*string{&snapshotName}})
	}

	if volumeID != "" {
		filters = append(filters, &ec2.Filter{
			Name: aws.String("volume-id"), Values: []*string{&volumeID}})
	}

	if snapshotID != "" {
		//using SnapshotIds in request is returning stale data
		filters = append(filters, &ec2.Filter{
			Name: aws.String("snapshot-id"), Values: []*string{&snapshotID}})
	}

	if d.ec2Tag != "" {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", rexrayTag)),
			Values: []*string{&d.ec2Tag}})
	}

	dsInput := &ec2.DescribeSnapshotsInput{}

	if len(filters) > 0 {
		dsInput.Filters = filters
	}

	resp, err := d.ec2Instance.DescribeSnapshots(dsInput)
	if err != nil {
		return nil, err
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
			VolumeID:    *snapshot.VolumeId,
			SnapshotID:  *snapshot.SnapshotId,
			VolumeSize:  fmt.Sprintf("%d", *snapshot.VolumeSize),
			StartTime:   (*snapshot.StartTime).Format(time.RFC3339),
			Description: *snapshot.Description,
			Status:      *snapshot.State,
		}
		snapshotsInt = append(snapshotsInt, snapshotSD)
	}

	// log.Println("Got Snapshots: " + fmt.Sprintf("%+v", snapshotsInt))
	return snapshotsInt, nil
}

func (d *driver) RemoveSnapshot(snapshotID string) error {
	dsInput := &ec2.DeleteSnapshotInput{
		SnapshotId: &snapshotID,
	}
	_, err := d.ec2Instance.DeleteSnapshot(dsInput)
	if err != nil {
		return err
	}

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
			log.WithFields(log.Fields{
				"moduleName":     d.r.Context,
				"driverName":     d.Name(),
				"nextDeviceName": nextDeviceName}).Info("got next device name")
			return nextDeviceName, nil
		}
	}
	return "", goof.New("No available device")
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

	volumes, err := d.GetVolume("", volumeName)
	if err != nil {
		return nil, err
	}

	if len(volumes) > 0 {
		return nil, goof.WithFields(goof.Fields{
			"moduleName": d.r.Context,
			"driverName": d.Name(),
			"volumeName": volumeName}, "volume name already exists")
	}

	resp, err := d.createVolume(
		runAsync, volumeName, volumeID, snapshotID, volumeType,
		IOPS, size, availabilityZone)

	if err != nil {
		return nil, err
	}

	volumes, err = d.GetVolume(*resp.VolumeId, "")
	if err != nil {
		return nil, err
	}

	// log.Println(fmt.Sprintf("Created volume: %+v", volumes[0]))
	return volumes[0], nil
}

func (d *driver) createVolume(
	runAsync bool, volumeName, volumeID, snapshotID, volumeType string,
	IOPS, size int64,
	availabilityZone string) (*ec2.Volume, error) {

	if volumeID != "" && runAsync {
		return &ec2.Volume{}, errors.ErrRunAsyncFromVolume
	}

	var err error

	var server ec2.Instance
	if server, err = d.getInstance(); err != nil {
		return &ec2.Volume{}, err
	}

	if volumeID != "" {
		if snapshotID, err = d.createVolumeCreateSnapshot(
			volumeID, snapshotID); err != nil {
			return &ec2.Volume{}, err
		}
	}

	d.createVolumeEnsureAvailabilityZone(&availabilityZone, &server)

	options := &ec2.CreateVolumeInput{
		Size:             &size,
		SnapshotId:       &snapshotID,
		AvailabilityZone: &availabilityZone,
		VolumeType:       &volumeType,
	}

	if IOPS > 0 {
		options.Iops = &IOPS
	}

	var resp *ec2.Volume
	if resp, err = d.createVolumeCreateVolume(options); err != nil {
		return &ec2.Volume{}, err
	}

	if err = d.createVolumeCreateTags(volumeName, resp); err != nil {
		return &ec2.Volume{}, err
	}

	if err = d.createVolumeWait(
		runAsync, snapshotID, volumeID, resp); err != nil {
		return &ec2.Volume{}, err
	}

	return resp, nil
}

func (d *driver) createVolumeCreateSnapshot(
	volumeID string, snapshotID string) (string, error) {

	var err error
	var snapshots []*core.Snapshot

	if snapshots, err = d.CreateSnapshot(
		true, fmt.Sprintf("temp-%v", volumeID),
		volumeID, "created for createVolume"); err != nil {
		return "", err
	}

	err = d.waitSnapshotComplete(snapshots[0].SnapshotID)

	if err != nil {
		return "", err
	}

	return snapshots[0].SnapshotID, nil
}

func (d *driver) createVolumeEnsureAvailabilityZone(
	availabilityZone *string, server *ec2.Instance) {
	if *availabilityZone == "" {
		*availabilityZone = *server.Placement.AvailabilityZone
	}
}

func (d *driver) createVolumeCreateVolume(
	options *ec2.CreateVolumeInput) (resp *ec2.Volume, err error) {
	for {
		resp, err = d.ec2Instance.CreateVolume(options)
		if err != nil {
			if awsErrMessage(err) ==
				"Snapshot is in invalid state - pending" {
				// Really, snapshot should be created already
				time.Sleep(1 * time.Second)
				continue
			}
			return nil, err
		}
		break
	}
	return
}

func (d *driver) createVolumeCreateTags(
	volumeName string, resp *ec2.Volume) (err error) {
	if volumeName == "" && d.ec2Tag == "" {
		return
	}

	var ctInput *ec2.CreateTagsInput
	initCTInput := func() {
		if ctInput != nil {
			return
		}
		ctInput = &ec2.CreateTagsInput{
			Resources: []*string{resp.VolumeId},
			Tags:      []*ec2.Tag{},
		}
	}

	if volumeName != "" {
		initCTInput()
		ctInput.Tags = append(
			ctInput.Tags,
			&ec2.Tag{
				Key:   aws.String("Name"),
				Value: &volumeName,
			})
	}

	if d.ec2Tag != "" {
		initCTInput()
		ctInput.Tags = append(
			ctInput.Tags,
			&ec2.Tag{
				Key:   aws.String(rexrayTag),
				Value: &d.ec2Tag,
			})
	}

	_, err = d.ec2Instance.CreateTags(ctInput)
	if err != nil {
		return err
	}
	return nil
}

func (d *driver) createVolumeWait(
	runAsync bool, snapshotID, volumeID string,
	resp *ec2.Volume) (err error) {
	if runAsync {
		return
	}
	log.WithFields(log.Fields{
		"moduleName":    d.r.Context,
		"driverName":    d.Name(),
		"runAsync":      runAsync,
		"snapshotID":    snapshotID,
		"volumeID":      volumeID,
		"resp.VolumeId": resp.VolumeId}).Info("waiting for volume creation to complete")
	if err = d.waitVolumeComplete(*resp.VolumeId); err != nil {
		return
	}

	if volumeID != "" {
		if err = d.RemoveSnapshot(snapshotID); err != nil {
			return
		}
	}

	return
}

func (d *driver) getVolume(
	volumeID, volumeName string) ([]*ec2.Volume, error) {

	filters := []*ec2.Filter{}
	if volumeName != "" {
		filters = append(filters, &ec2.Filter{
			Name: aws.String("tag:Name"), Values: []*string{&volumeName}})
	}

	if volumeID != "" {
		filters = append(filters, &ec2.Filter{
			Name: aws.String("volume-id"), Values: []*string{&volumeID}})
	}

	if d.ec2Tag != "" {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", rexrayTag)),
			Values: []*string{&d.ec2Tag}})
	}

	dvInput := &ec2.DescribeVolumesInput{}

	if len(filters) > 0 {
		dvInput.Filters = filters
	}

	if volumeID != "" {
		dvInput.VolumeIds = []*string{&volumeID}
	}

	resp, err := d.ec2Instance.DescribeVolumes(dvInput)
	if err != nil {
		return []*ec2.Volume{}, err
	}

	return resp.Volumes, nil
}

func getName(tags []*ec2.Tag) string {
	for _, tag := range tags {
		if *tag.Key == "Name" {
			return *tag.Value
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
				VolumeID:   *attachment.VolumeId,
				InstanceID: *attachment.InstanceId,
				DeviceName: *attachment.Device,
				Status:     *attachment.State,
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		name := getName(volume.Tags)

		volumeSD := &core.Volume{
			Name:             name,
			VolumeID:         *volume.VolumeId,
			AvailabilityZone: *volume.AvailabilityZone,
			Status:           *volume.State,
			VolumeType:       *volume.VolumeType,
			Size:             fmt.Sprintf("%d", *volume.Size),
			Attachments:      attachmentsSD,
		}
		// Some volume types have no IOPS, so we get nil in volume.Iops
		if volume.Iops != nil {
			volumeSD.IOPS = *volume.Iops
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

		if len(snapshots) == 0 {
			return errors.ErrDriverSnapshotDiscovery
		}
		snapshot := snapshots[0]
		if *snapshot.State == ec2.SnapshotStateCompleted {
			break
		}
		if *snapshot.State == ec2.SnapshotStateError {
			return fmt.Errorf("%s", *snapshot.StateMessage)
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

		if *volumes[0].State == ec2.VolumeStateAvailable {
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

		if len(volume) == 0 {
			break
		}

		if volume[0].Status == ec2.VolumeAttachmentStateAttached {
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

	dvInput := &ec2.DeleteVolumeInput{
		VolumeId: &volumeID,
	}
	_, err := d.ec2Instance.DeleteVolume(dvInput)
	if err != nil {
		return err
	}

	return nil
}

func (d *driver) AttachVolume(
	runAsync bool,
	volumeID, instanceID string, force bool) ([]*core.VolumeAttachment, error) {

	if volumeID == "" {
		return nil, errors.ErrMissingVolumeID
	}

	nextDeviceName, err := d.GetDeviceNextAvailable()
	if err != nil {
		return nil, err
	}

	if force {
		if err := d.DetachVolume(false, volumeID, "", true); err != nil {
			return nil, err
		}
	}

	if instanceID == "" {
		instanceID = d.instanceDocument.InstanceID
	}

	avInput := &ec2.AttachVolumeInput{
		Device:     &nextDeviceName,
		InstanceId: &instanceID,
		VolumeId:   &volumeID,
	}
	_, err = d.ec2Instance.AttachVolume(avInput)

	if err != nil {
		return nil, err
	}

	if !runAsync {
		log.WithFields(log.Fields{
			"moduleName": d.r.Context,
			"driverName": d.Name(),
			"runAsync":   runAsync,
			"volumeID":   volumeID,
			"instanceID": instanceID,
			"force":      force}).Info("waiting for volume attachment to complete")

		err = d.waitVolumeAttach(volumeID, instanceID)
		if err != nil {
			return nil, err
		}
	}

	volumeAttachment, err := d.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, err
	}

	return volumeAttachment, nil
}

func (d *driver) DetachVolume(
	runAsync bool,
	volumeID, blank string, force bool) error {

	if volumeID == "" {
		return errors.ErrMissingVolumeID
	}

	volumes, err := d.getVolume(volumeID, "")
	if err != nil {
		return err
	}

	if *volumes[0].State == ec2.VolumeStateAvailable {
		return nil
	}

	dvInput := &ec2.DetachVolumeInput{
		VolumeId: &volumeID,
		Force:    &force,
	}

	_, err = d.ec2Instance.DetachVolume(dvInput)
	if err != nil {
		return err
	}

	if !runAsync {
		log.WithFields(log.Fields{
			"moduleName": d.r.Context,
			"driverName": d.Name(),
			"runAsync":   runAsync,
			"volumeID":   volumeID,
			"force":      force}).Info("waiting for volume detachment to complete")

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
		return nil, goof.New("Missing volumeID, snapshotID, or snapshotName")
	}

	if destinationRegion == "" {
		destinationRegion = d.instanceDocument.Region
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

	snapshotID = *(snapshots[0]).SnapshotId

	options := &ec2.CopySnapshotInput{
		SourceRegion:      &d.instanceDocument.Region,
		DestinationRegion: &destinationRegion,
		SourceSnapshotId:  &snapshotID,
		Description: aws.String(fmt.Sprintf("[Copied %s from %s]",
			snapshotID, d.instanceDocument.Region)),
	}
	resp := &ec2.CopySnapshotOutput{}

	destec2Instance := ec2.New(
		session.New(),
		aws.NewConfig().WithCredentials(d.ec2creds).WithRegion(destinationRegion),
	)

	origec2Instance := d.ec2Instance
	d.ec2Instance = destec2Instance
	defer func() { d.ec2Instance = origec2Instance }()

	resp, err = d.ec2Instance.CopySnapshot(options)
	if err != nil {
		return nil, err
	}

	if destinationSnapshotName != "" || d.ec2Tag != "" {

		var ctInput *ec2.CreateTagsInput
		initCTInput := func() {
			if ctInput != nil {
				return
			}
			ctInput = &ec2.CreateTagsInput{
				Resources: []*string{resp.SnapshotId},
				Tags:      []*ec2.Tag{},
			}
		}

		if destinationSnapshotName != "" {
			initCTInput()
			ctInput.Tags = append(
				ctInput.Tags,
				&ec2.Tag{
					Key:   aws.String("Name"),
					Value: &destinationSnapshotName,
				})
		}

		if d.ec2Tag != "" {
			initCTInput()
			ctInput.Tags = append(
				ctInput.Tags,
				&ec2.Tag{
					Key:   aws.String(rexrayTag),
					Value: &d.ec2Tag,
				})
		}

		_, err = d.ec2Instance.CreateTags(ctInput)

		if err != nil {
			return nil, err
		}
	}

	if !runAsync {
		log.WithFields(log.Fields{
			"moduleName":      d.r.Context,
			"driverName":      d.Name(),
			"runAsync":        runAsync,
			"resp.SnapshotId": resp.SnapshotId}).Info("waiting for snapshot to complete")

		err = d.waitSnapshotComplete(*resp.SnapshotId)
		if err != nil {
			return nil, err
		}
	}

	snapshot, err := d.GetSnapshot("", *resp.SnapshotId, "")
	if err != nil {
		return nil, err
	}

	return snapshot[0], nil
}

func (d *driver) accessKey() string {
	return d.r.Config.GetString("aws.accessKey")
}

func (d *driver) secretKey() string {
	return d.r.Config.GetString("aws.secretKey")
}

func (d *driver) region() string {
	return d.r.Config.GetString("aws.region")
}

func (d *driver) rexrayTag() string {
	return d.r.Config.GetString("aws.rexrayTag")
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("Amazon EC2")
	r.Key(gofig.String, "", "", "", "aws.accessKey")
	r.Key(gofig.String, "", "", "", "aws.secretKey")
	r.Key(gofig.String, "", "", "", "aws.region")
	r.Key(gofig.String, "", "", "", "aws.rexrayTag")
	return r
}

func awsErrCode(err error) string {
	if awsErr, ok := err.(awserr.Error); ok {
		return awsErr.Code()
	}
	return ""
}

func awsErrMessage(err error) string {
	if awsErr, ok := err.(awserr.Error); ok {
		return awsErr.Message()
	}
	return ""
}
