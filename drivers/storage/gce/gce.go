package gce

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/config"
	"github.com/emccode/rexray/core/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

const providerName = "gce"

// The GCE storage driver.
type driver struct {
	client  *compute.Service
	r       *core.RexRay
	zone    string
	project string
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
	}

	config, err := google.JWTConfigFromJSON(serviceAccountJSON,
		compute.ComputeScope,
	)
	client, err := compute.New(config.Client(context.Background()))

	if err != nil {
		log.WithField("provider", providerName).Fatalf("Could not create compute client => {%s}", err)
	}
	d.client = client
	log.WithField("provider", providerName).Info("storage driver initialized")

	return nil
}

func (d *driver) Name() string {
	return providerName
}

func (d *driver) GetVolumeMapping() ([]*core.BlockDevice, error) {
	log.WithField("provider", providerName).Debug("GetVolumeMapping")
	return nil, nil
}

func (d *driver) GetInstance() (*core.Instance, error) {
	log.WithField("provider", providerName).Debug("GetInstance")
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
	log.WithField("provider", providerName).Debug("CreateVolume")
	return nil, nil

}

func (d *driver) createVolumeCreateSnapshot(
	volumeID string, snapshotID string) (string, error) {
	log.WithField("provider", providerName).Debug("CreateVolumeCreateSnapshot")
	return "", nil

}

func (d *driver) GetVolume(
	volumeID, volumeName string) ([]*core.Volume, error) {
	log.WithField("provider", providerName).Debugf("GetVolume :%s %s", volumeID, volumeName)

	query := d.client.Disks.List(d.project, d.zone)
	if volumeID != "" {
		query.Filter(fmt.Sprintf("id eq %s", volumeID))
	}
	if volumeName != "" {
		query.Filter(fmt.Sprintf("name eq %s", volumeName))
	}
	var attachments []*core.VolumeAttachment
	instances, err := d.client.Instances.List(d.project, d.zone).Do()
	for _, instance := range instances.Items {
		for _, disk := range instance.Disks {
			attachment := &core.VolumeAttachment{
				InstanceID: strconv.FormatUint(instance.Id, 10),
				DeviceName: disk.DeviceName,
				Status:     disk.Mode,
				VolumeID:   disk.Source,
			}
			attachments = append(attachments, attachment)

		}
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
			VolumeID:         strconv.FormatUint(disk.Id, 10),
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

func (d *driver) GetVolumeAttach(
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
	log.WithField("provider", providerName).Debug("GetVolumeAttach")
	return nil, nil
}

func (d *driver) waitSnapshotComplete(snapshotID string) error {
	return nil
}

func (d *driver) waitVolumeComplete(volumeID string) error {
	return nil
}

func (d *driver) waitVolumeAttach(volumeID, instanceID string) error {
	return nil
}

func (d *driver) waitVolumeDetach(volumeID string) error {
	return nil
}

func (d *driver) RemoveVolume(volumeID string) error {
	return nil
}

func (d *driver) AttachVolume(
	runAsync bool,
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
	log.WithField("provider", providerName).Debug("AttachVolume")
	return nil, nil

}

func (d *driver) DetachVolume(
	runAsync bool,
	volumeID, blank string) error {
	log.WithField("provider", providerName).Debug("DetachVolume")
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
