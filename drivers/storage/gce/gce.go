package gce

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/config"
	"github.com/emccode/rexray/core/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"
)

const providerName = "gce"

// The GCE storage driver.
type driver struct {
	currentInstanceId string
	client            *compute.Service
	r                 *core.RexRay
	zone              string
	project           string
}

func init() {
	core.RegisterDriver(providerName, newDriver)
	config.Register(configRegistration())
}

func newDriver() core.Driver {
	return &driver{}
}

func (d *driver) Init(r *core.RexRay) error {
	d.r = r

	var err error

	d.zone = d.r.Config.GetString("gce.zone")
	d.project = d.r.Config.GetString("gce.project")
	serviceAccountJSON, err := ioutil.ReadFile(d.r.Config.GetString("gce.keyfile"))
	if err != nil {
		log.WithField("provider", providerName).Fatalf("Could not read service account credentials file, %s => {%s}", d.r.Config.GetString("gce.keyfile"), err)
		return err
	}

	config, err := google.JWTConfigFromJSON(serviceAccountJSON,
		compute.ComputeScope,
	)
	client, err := compute.New(config.Client(context.Background()))

	if err != nil {
		log.WithField("provider", providerName).Fatalf("Could not create compute client => {%s}", err)
		return err
	}
	d.client = client
	instanceId, err := getCurrentInstanceId()
	if err != nil {
		log.WithField("provider", providerName).Fatalf("Could not get current  instance => {%s}", err)
		return err
	}
	d.currentInstanceId = instanceId
	log.WithField("provider", providerName).Info("storage driver initialized")
	return nil
}

func getCurrentInstanceId() (string, error) {
	conn, err := net.DialTimeout("tcp", "metadata.google.internal:80", 50*time.Millisecond)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	url := "http://metadata.google.internal/computeMetadata/v1/instance/id"
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

	instances, err := d.client.Instances.List(d.project, d.zone).Do()
	if err != nil {
		return []*core.BlockDevice{}, err
	}
	var ret []*core.BlockDevice
	for _, instance := range instances.Items {
		for _, disk := range instance.Disks {
			ret = append(ret, &core.BlockDevice{
				ProviderName: "gce",
				InstanceID:   instance.Name,
				VolumeID:     diskMap[disk.Source].Name,
				DeviceName:   disk.DeviceName,
				Region:       diskMap[disk.Source].Zone,
				Status:       diskMap[disk.Source].Status,
				NetworkName:  disk.Source,
			})

		}
	}
	return ret, nil
}

func (d *driver) GetInstance() (*core.Instance, error) {
	log.WithField("provider", providerName).Debug("GetInstance")
	query := d.client.Instances.List(d.project, d.zone)
	instances, err := query.Do()
	if err != nil {
		return nil, err
	}
	for _, instance := range instances.Items {
		if strconv.FormatUint(instance.Id, 10) == d.currentInstanceId {
			return &core.Instance{
				ProviderName: "gce",
				InstanceID:   instance.Name,
				Region:       instance.Zone,
				Name:         instance.Name,
			}, nil
		}

	}
	return nil, nil
}

func (d *driver) CreateSnapshot(
	runAsync bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {

	log.WithField("provider", providerName).Debug("CreateSnapshot")
	return nil, nil

}

func (d *driver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {

	log.WithField("provider", providerName).Debug("GetSnapshot")
	return nil, nil
}

func (d *driver) RemoveSnapshot(snapshotID string) error {
	log.WithField("provider", providerName).Debug("RemoveSnapshot")
	return nil
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
				return err
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
		"availabilityZone": availabilityZone}).Debug("CreateVolume")

	if availabilityZone == "" {
		availabilityZone = d.zone
	}
	disk := &compute.Disk{
		Name:   volumeName,
		Zone:   availabilityZone,
		Type:   "https://www.googleapis.com/compute/v1/projects/gce-dev-1060/zones/europe-west1-b/diskTypes/pd-standard",
		SizeGb: size,
	}
	createdVolume, err := d.client.Disks.Insert(d.project, d.zone, disk).Do()
	if err != nil {
		return nil, err
	}
	if !runAsync {
		err := d.waitUntilOperationIsFinished(createdVolume)
		if err != nil {
			return nil, err
		}

	}
	volume, err := d.GetVolume(volumeName, "")
	if err != nil {
		return nil, err
	}
	return volume[0], nil
}

func (d *driver) getVolumesAttachedToInstance(instances []*compute.Instance) []*core.VolumeAttachment {
	var attachments []*core.VolumeAttachment
	for _, instance := range instances {
		for _, disk := range instance.Disks {
			attachments = append(attachments, d.convertGCEAttachedDisk(instance.Name, disk))
		}
	}

	return attachments
}
func (d *driver) GetVolume(
	volumeID, volumeName string) ([]*core.Volume, error) {
	log.WithField("provider", providerName).Debugf("GetVolume :%s %s", volumeID, volumeName)

	instances, err := d.client.Instances.List(d.project, d.zone).Do()
	if err != nil {
		return []*core.Volume{}, err
	}
	attachments := d.getVolumesAttachedToInstance(instances.Items)

	query := d.client.Disks.List(d.project, d.zone)
	if volumeID != "" {
		query.Filter(fmt.Sprintf("name eq %s", volumeID))
	}
	if volumeName != "" {
		query.Filter(fmt.Sprintf("name eq %s", volumeName))
	}
	disks, err := query.Do()
	if err != nil {
		return []*core.Volume{}, err
	}
	var volumesSD []*core.Volume
	for _, disk := range disks.Items {
		var diskAttachments []*core.VolumeAttachment
		for _, attachment := range attachments {
			if attachment.VolumeID == disk.SelfLink {
				diskAttachments = append(diskAttachments, &core.VolumeAttachment{
					InstanceID: attachment.InstanceID,
					DeviceName: attachment.DeviceName,
					Status:     attachment.Status,
					VolumeID:   strconv.FormatUint(disk.Id, 10),
				})
			}
		}
		volumeSD := &core.Volume{
			Name:             disk.Name,
			VolumeID:         disk.Name,
			AvailabilityZone: disk.Zone,
			Status:           disk.Status,
			VolumeType:       disk.Kind,
			NetworkName:      disk.SelfLink,
			IOPS:             0,
			Size:             strconv.FormatInt(disk.SizeGb, 10),
			Attachments:      diskAttachments,
		}
		volumesSD = append(volumesSD, volumeSD)

	}
	return volumesSD, nil
}

func (d *driver) convertGCEAttachedDisk(instanceName string, disk *compute.AttachedDisk) *core.VolumeAttachment {
	return &core.VolumeAttachment{
		InstanceID: instanceName,
		DeviceName: disk.DeviceName,
		Status:     disk.Mode,
		VolumeID:   disk.Source,
	}
}

func (d *driver) GetVolumeAttach(
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
	log.WithField("provider", providerName).Debugf("GetVolumeAttach :%s %s", volumeID, instanceID)
	query := d.client.Instances.List(d.project, d.zone)
	if instanceID != "" {
		query.Filter(fmt.Sprintf("name eq %s", instanceID))
	}
	instances, err := query.Do()
	if err != nil {
		return []*core.VolumeAttachment{}, err
	}
	return d.getVolumesAttachedToInstance(instances.Items), nil
}

func (d *driver) RemoveVolume(volumeID string) error {
	log.WithField("provider", providerName).Debugf("RemoveVolume :%s", volumeID)
	_, err := d.client.Disks.Delete(d.project, d.zone, volumeID).Do()
	return err

}

func (d *driver) AttachVolume(
	runAsync bool,
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
	if instanceID == "" {
		instanceID = d.currentInstanceId
	}
	instance, err := d.GetInstance()
	if err != nil {
		return nil, err
	}
	instanceID = instance.Name
	log.WithField("provider", providerName).Debugf("AttachVolume %s %s", volumeID, instance.Name)
	query := d.client.Disks.List(d.project, d.zone)
	query.Filter(fmt.Sprintf("name eq %s", volumeID))
	disks, err := query.Do()
	if err != nil {
		return nil, err
	}
	if len(disks.Items) != 1 {
		return nil, errors.New("No available device")
	}

	disk := &compute.AttachedDisk{
		AutoDelete: false,
		Boot:       false,
		Source:     disks.Items[0].SelfLink,
	}
	_, err = d.client.Instances.AttachDisk(d.project, d.zone, instanceID, disk).Do()
	if err != nil {
		return nil, err
	}

	return d.GetVolumeAttach("", instanceID)

}

func (d *driver) DetachVolume(
	runAsync bool,
	volumeID, blank string) error {
	instance, err := d.GetInstance()
	if err != nil {
		return err
	}
	instanceID := instance.Name
	log.WithField("provider", providerName).Debugf("DetachVolume %s %s", volumeID, instance.Name)
	query := d.client.Disks.List(d.project, d.zone)
	query.Filter(fmt.Sprintf("name eq %s", volumeID))
	disks, err := query.Do()
	if err != nil {
		return err
	}
	if len(disks.Items) != 1 {
		return errors.New("No available device")
	}
	attachements, err := d.GetVolumeAttach(volumeID, instanceID)
	for _, attachement := range attachements {
		if attachement.VolumeID == volumeID && attachement.InstanceID == instanceID {
			_, err = d.client.Instances.DetachDisk(d.project, d.zone, instanceID, attachement.DeviceName).Do()
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

func configRegistration() *config.Registration {
	r := config.NewRegistration("Google GCE")
	r.Key(config.String, "", "", "", "gce.zone")
	r.Key(config.String, "", "", "", "gce.project")
	r.Key(config.String, "", "", "", "gce.keyfile")
	return r
}
