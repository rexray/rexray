package gce

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

const providerName = "gce"
const defaultVolumeType = "pd-standard"

// The GCE storage driver.
type driver struct {
	currentInstanceID string
	client            *compute.Service
	r                 *core.RexRay
	zone              string
	project           string
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

	var err error

	if d.zone, err = getCurrentZone(); err != nil {
		return goof.WithError("error getting current zone", err)
	}
	if d.project, err = getCurrentProjectID(); err != nil {
		return goof.WithError("error getting current project ID", err)
	}
	serviceAccountJSON, err := ioutil.ReadFile(d.r.Config.GetString("gce.keyfile"))
	if err != nil {
		log.WithField("provider", providerName).Fatalf("Could not read service account credentials file, %s => {%s}", d.r.Config.GetString("gce.keyfile"), err)
		return err
	}

	config, err := google.JWTConfigFromJSON(serviceAccountJSON,
		compute.ComputeScope,
	)
	if err != nil {
		goof.WithFieldE("provider", providerName, "could not create JWT Config From JSON", err)
		return err
	}

	client, err := compute.New(config.Client(context.Background()))
	if err != nil {
		log.WithField("provider", providerName).Fatalf("Could not create compute client => {%s}", err)
		return err
	}

	d.client = client
	instanceID, err := getCurrentInstanceID()
	if err != nil {
		log.WithField("provider", providerName).Fatalf("Could not get current  instance => {%s}", err)
		return err
	}
	d.currentInstanceID = instanceID
	log.WithField("provider", providerName).Info("storage driver initialized")
	return nil
}

func getCurrentInstanceID() (string, error) {
	return getMetadata("http://metadata.google.internal/computeMetadata/v1/instance/id")
}

func getCurrentProjectID() (string, error) {
	return getMetadata("http://metadata.google.internal/computeMetadata/v1/project/project-id")
}

func getCurrentZone() (zone string, err error) {
	if zone, err = getMetadata("http://metadata.google.internal/computeMetadata/v1/instance/zone"); err != nil {
		return "", err
	}
	zone = getIndex(zone)
	return
}

func getMetadata(url string) (string, error) {
	conn, err := net.DialTimeout("tcp", "metadata.google.internal:80", 50*time.Millisecond)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error: %v\n", err)
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	s := buf.String()
	return s, nil
}

func (d *driver) Name() string {
	return providerName
}

func (d *driver) GetVolumeMapping() ([]*core.BlockDevice, error) {
	log.WithField("provider", providerName).Debug("GetVolumeMapping")

	diskMap := make(map[string]*compute.Disk)
	disks, err := d.client.Disks.List(d.project, d.zone).Do()
	if err != nil {
		return []*core.BlockDevice{}, err
	}
	for _, disk := range disks.Items {
		diskMap[disk.SelfLink] = disk
	}

	instance, err := d.getInstance()
	if err != nil {
		return nil, err
	}

	var ret []*core.BlockDevice
	for _, disk := range instance.Disks {
		region := getIndex(diskMap[disk.Source].Zone)
		deviceName := fmt.Sprintf("/dev/disk/by-id/google-%s", disk.DeviceName)
		ret = append(ret, &core.BlockDevice{
			ProviderName: "gce",
			InstanceID:   instance.Name,
			VolumeID:     getIndex(disk.Source),
			DeviceName:   deviceName,
			Region:       region,
			Status:       diskMap[disk.Source].Status,
		})
	}

	return ret, nil
}

func (d *driver) getInstance() (*compute.Instance, error) {
	instance, err := d.getInstances()
	if err != nil {
		return nil, err
	}

	for _, instance := range instance.Items {
		if strconv.FormatUint(instance.Id, 10) == d.currentInstanceID {
			return instance, nil
		}
	}
	return nil, goof.New("instance not found")
}

func (d *driver) getInstances() (*compute.InstanceList, error) {
	query := d.client.Instances.List(d.project, d.zone)
	return query.Do()
}

func (d *driver) GetInstance() (*core.Instance, error) {
	log.WithField("provider", providerName).Debug("GetInstance")

	instance, err := d.getInstance()
	if err != nil {
		return nil, err
	}

	return &core.Instance{
		ProviderName: "gce",
		InstanceID:   instance.Name,
		Region:       instance.Zone,
		Name:         instance.Name,
	}, nil
}

func (d *driver) CreateSnapshot(
	runAsync bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {

	log.WithField("provider", providerName).Debug("CreateSnapshot")

	volumes, err := d.GetVolume(volumeID, "")

	if len(volumes) == 0 {
		return nil, goof.New("no volume returned by ID")
	}

	if err := d.createSnapshot(runAsync, snapshotName, volumes[0]); err != nil {
		return nil, err
	}

	snapshot, err := d.GetSnapshot("", snapshotName, "")
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

func (d *driver) createSnapshot(runAsync bool, snapshotName string, volume *core.Volume) error {
	sourceDisk := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/disks/%s",
		d.project, d.zone, volume.Name)
	snapshot := &compute.Snapshot{
		SourceDisk: sourceDisk,
		Name:       snapshotName,
	}
	operation, err := d.client.Disks.CreateSnapshot(d.project, d.zone, getIndex(sourceDisk), snapshot).Do()
	if err != nil {
		return goof.WithError("error creating snapshot", err)
	}

	if !runAsync {
		err := d.waitUntilOperationIsFinished(operation)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {

	log.WithField("provider", providerName).Debug("GetSnapshot")

	var diskList *compute.DiskList
	if volumeID != "" {
		var err error
		diskList, err = d.getVolume(volumeID, "")
		if err != nil {
			return nil, err
		}
		if len(diskList.Items) > 0 {
			volumeID = strconv.FormatUint(diskList.Items[0].Id, 10)
		}
	}

	snapshotList, err := d.getSnapshot(volumeID, snapshotID, snapshotName)
	if err != nil {
		return nil, goof.WithError("problem getting snapshots", err)
	}

	var snapshots []*core.Snapshot
	for _, snapshot := range snapshotList.Items {
		snapshotSD := &core.Snapshot{
			Name:       snapshot.Name,
			VolumeID:   getIndex(snapshot.SourceDisk),
			SnapshotID: snapshot.Name,
			VolumeSize: strconv.FormatInt(snapshot.DiskSizeGb, 10),
			StartTime:  snapshot.CreationTimestamp,
			Status:     snapshot.Status,
		}
		snapshots = append(snapshots, snapshotSD)
	}

	return snapshots, nil
}

func (d *driver) getSnapshot(volumeID, snapshotID, snapshotName string) (*compute.SnapshotList, error) {
	query := d.client.Snapshots.List(d.project)
	if snapshotID != "" {
		query.Filter(fmt.Sprintf("name eq '%s'", snapshotID))
	} else if snapshotName != "" {
		query.Filter(fmt.Sprintf("name eq '%s'", snapshotName))
	}

	if volumeID != "" {
		query.Filter(fmt.Sprintf("sourceDiskId eq '%s'", volumeID))
	}

	return query.Do()
}

func (d *driver) RemoveSnapshot(snapshotID string) error {
	log.WithField("provider", providerName).Debug("RemoveSnapshot :%s", snapshotID)
	if _, err := d.client.Snapshots.Delete(d.project, snapshotID).Do(); err != nil {
		return goof.WithError("problem removing snapshot", err)
	}
	return nil
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

func (d *driver) GetDeviceNextAvailable() (string, error) {
	return "", nil
}
func (d *driver) waitUntilOperationIsFinished(operation *compute.Operation) error {
	opName := operation.Name
OpLoop:
	for {
		time.Sleep(100 * time.Millisecond)
		op, err := d.client.ZoneOperations.Get(d.project, d.zone, opName).Do()
		if err != nil {
			return err
		}

		switch op.Status {
		case "PENDING", "RUNNING":
			continue
		case "DONE":
			if op.Error != nil {
				bytea, _ := op.Error.MarshalJSON()
				return goof.New(string(bytea))
			}
			break OpLoop
		default:
			log.WithField("provider", providerName).Fatalf("Unknown status %q: %+v", op.Status, op)
			return nil
		}
	}
	return nil
}

func (d *driver) CreateVolume(
	runAsync bool, volumeName, volumeID, snapshotID, volumeType string,
	IOPS, size int64, availabilityZone string) (*core.Volume, error) {
	log.WithFields(log.Fields{
		"provider":         providerName,
		"volumeName":       volumeName,
		"volumeID":         volumeID,
		"snapshotID":       snapshotID,
		"volumeType":       volumeType,
		"size":             size,
		"availabilityZone": availabilityZone}).Debug("CreateVolume")

	if availabilityZone == "" {
		availabilityZone = d.zone
	}

	if volumeType == "" {
		volumeType = defaultVolumeType
	}
	diskType := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/diskTypes/%s",
		d.project, d.zone, volumeType)

	var snapshots []*core.Snapshot
	if volumeID != "" {
		volume, err := d.GetVolume(volumeName, "")
		if err != nil {
			return nil, err
		}

		if len(volume) > 0 {
			return nil, goof.New("volume already exists by name")
		}

		tmpSnapshotName := fmt.Sprintf("temp-%s", volumeID)
		snapshots, err = d.CreateSnapshot(false, tmpSnapshotName, volumeID, "")
		if err != nil {
			return nil, err
		}
		if len(snapshots) == 0 {
			return nil, goof.New("no snapshot returned")
		}
		snapshotID = snapshots[0].SnapshotID
	}

	var snapshotURL string
	if snapshotID != "" {
		snapshotURL = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/snapshots/%s",
			d.project, snapshotID)
	}

	disk := &compute.Disk{
		Name:           volumeName,
		Zone:           availabilityZone,
		Type:           diskType,
		SizeGb:         size,
		SourceSnapshot: snapshotURL,
	}
	createdVolume, err := d.client.Disks.Insert(d.project, d.zone, disk).Do()
	if err != nil {
		return nil, err
	}
	if !runAsync && volumeID != "" {
		err := d.waitUntilOperationIsFinished(createdVolume)
		if err != nil {
			return nil, err
		}
	}

	if volumeID != "" {
		if err := d.RemoveSnapshot(snapshots[0].SnapshotID); err != nil {
			return nil, err
		}
	}

	volume, err := d.GetVolume(volumeName, "")
	if err != nil {
		return nil, err
	}
	return volume[0], nil
}

func (d *driver) getVolumesAttachedToInstance(instances []*compute.Instance, volumeIDMapByName map[string]string, volumeID string) []*core.VolumeAttachment {
	var attachments []*core.VolumeAttachment
	for _, instance := range instances {
		for _, disk := range instance.Disks {
			shortVolume := getIndex(disk.Source)

			if volumeID != "" && shortVolume != volumeID {
				continue
			}
			instanceID, _ := volumeIDMapByName[instance.Name]
			attachments = append(attachments, d.convertGCEAttachedDisk(instanceID, disk))
		}
	}

	return attachments
}

func (d *driver) getVolume(volumeName, volumeID string) (*compute.DiskList, error) {
	query := d.client.Disks.List(d.project, d.zone)
	if volumeID != "" {
		query.Filter(fmt.Sprintf("name eq '%s'", volumeID))
	} else if volumeName != "" {
		query.Filter(fmt.Sprintf("name eq '%s'", volumeName))
	}

	return query.Do()
}

func (d *driver) GetVolume(
	volumeID, volumeName string) ([]*core.Volume, error) {

	log.WithField("provider", providerName).Debugf("GetVolume :%s %s", volumeID, volumeName)
	instanceList, err := d.getInstances()
	if err != nil {
		return nil, err
	}

	mapInstanceBySource := make(map[string]*compute.Instance)
	for _, instance := range instanceList.Items {
		mapInstanceBySource[instance.SelfLink] = instance
	}

	diskList, err := d.getVolume(volumeName, volumeID)
	if err != nil {
		return nil, err
	}

	var volumesSD []*core.Volume
	for _, disk := range diskList.Items {

		var diskAttachments []*core.VolumeAttachment
		for _, user := range disk.Users {
			if instance, ok := mapInstanceBySource[user]; ok {
				for _, idisk := range instance.Disks {
					if idisk.Source == disk.SelfLink {
						diskAttachments = append(diskAttachments, &core.VolumeAttachment{
							InstanceID: getIndex(instance.SelfLink),
							DeviceName: idisk.DeviceName,
							Status:     idisk.Mode,
							VolumeID:   disk.Name,
						})
					}
				}
			}
		}

		volumeSD := &core.Volume{
			Name:             disk.Name,
			VolumeID:         disk.Name,
			AvailabilityZone: getIndex(disk.Zone),
			Status:           disk.Status,
			VolumeType:       getIndex(disk.Type),
			IOPS:             0,
			Size:             strconv.FormatInt(disk.SizeGb, 10),
			Attachments:      diskAttachments,
		}
		volumesSD = append(volumesSD, volumeSD)

	}
	return volumesSD, nil
}

func getIndex(href string) string {
	hrefFields := strings.Split(href, "/")
	return hrefFields[len(hrefFields)-1]
}

func (d *driver) convertGCEAttachedDisk(instanceID string, disk *compute.AttachedDisk) *core.VolumeAttachment {
	deviceName := fmt.Sprintf("/dev/disk/by-id/google-%s", disk.DeviceName)
	return &core.VolumeAttachment{
		InstanceID: instanceID,
		DeviceName: deviceName,
		Status:     disk.Mode,
		VolumeID:   getIndex(disk.Source),
	}
}

func (d *driver) GetVolumeAttach(
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
	log.WithField("provider", providerName).Debugf("GetVolumeAttach :%s %s", volumeID, instanceID)
	query := d.client.Instances.List(d.project, d.zone)
	if instanceID != "" {
		query.Filter(fmt.Sprintf("name eq '%s'", instanceID))
	}
	instances, err := query.Do()
	if err != nil {
		return nil, err
	}

	volumes, err := d.GetVolume("", "")
	if err != nil {
		return nil, err
	}

	volumeIDMapByName := make(map[string]string)
	for _, volume := range volumes {
		volumeIDMapByName[volume.Name] = volume.VolumeID
	}

	return d.getVolumesAttachedToInstance(instances.Items, volumeIDMapByName, volumeID), nil
}

func (d *driver) RemoveVolume(volumeID string) error {
	log.WithField("provider", providerName).Debugf("RemoveVolume :%s", volumeID)
	if _, err := d.client.Disks.Delete(d.project, d.zone, volumeID).Do(); err != nil {
		return goof.WithError("problem removing volume", err)
	}
	return nil
}

func (d *driver) AttachVolume(
	runAsync bool,
	volumeID, instanceID string, force bool) ([]*core.VolumeAttachment, error) {

	log.WithField("provider", providerName).Debugf("AttachVolume %s %s", volumeID, instanceID)

	if volumeID == "" {
		return nil, errors.ErrMissingVolumeID
	}

	if instanceID == "" {
		return nil, goof.New("missing instance ID")
	}

	volumes, err := d.GetVolume(volumeID, "")
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, errors.ErrNoVolumesReturned
	}

	if len(volumes[0].Attachments) > 0 && !force {
		return nil, goof.New("Volume already attached to another host")
	} else if len(volumes[0].Attachments) > 0 && force {
		if err := d.DetachVolume(false, volumeID, "", true); err != nil {
			return nil, err
		}
	}

	if err := d.attachDisk(false, instanceID, volumes[0]); err != nil {
		return nil, err
	}

	return d.GetVolumeAttach(volumeID, instanceID)

}

func (d *driver) attachDisk(runAsync bool, instanceID string, volume *core.Volume) error {
	disk := &compute.AttachedDisk{
		AutoDelete: false,
		Boot:       false,
		Source: fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/disks/%s",
			d.project, d.zone, volume.Name),
	}
	operation, err := d.client.Instances.AttachDisk(d.project, d.zone, instanceID, disk).Do()
	if err != nil {
		return err
	}
	if !runAsync {
		err := d.waitUntilOperationIsFinished(operation)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) DetachVolume(
	runAsync bool,
	volumeID, instanceID string, force bool) error {

	fields := eff(map[string]interface{}{
		"runAsync":   runAsync,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return goof.WithFields(fields, "volumeId is required")
	}
	volumes, err := d.GetVolume(volumeID, "")
	if err != nil {
		return goof.WithFieldsE(fields, "error getting volume", err)
	}

	if len(volumes) == 0 {
		return goof.WithFields(fields, "no volumes returned")
	}

	if len(volumes[0].Attachments) == 0 {
		return nil
	}

	for _, attachment := range volumes[0].Attachments {
		operation, err := d.client.Instances.DetachDisk(d.project, d.zone, attachment.InstanceID, attachment.DeviceName).Do()
		if err != nil {
			return err
		}
		if !runAsync {
			err := d.waitUntilOperationIsFinished(operation)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *driver) CopySnapshot(runAsync bool,
	volumeID, snapshotID, snapshotName, destinationSnapshotName,
	destinationRegion string) (*core.Snapshot, error) {
	log.WithField("provider", providerName).Debug("CopySnapshot")
	return nil, nil
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("Google GCE")
	r.Key(gofig.String, "", "", "", "gce.keyfile")
	return r
}
