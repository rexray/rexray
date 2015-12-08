package isilon

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	//	"reflect"

	//	"sort"
	//	"strconv"
	"strings"

	isi "github.com/emccode/goisilon"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

const providerName = "Isilon"

// The Isilon storage driver.
type driver struct {
	client *isi.Client
	r      *core.RexRay
}

func ef() goof.Fields {
	log.Println("Start ef()")
	return goof.Fields{
		"provider": providerName,
	}
}

func eff(fields goof.Fields) map[string]interface{} {
	log.Println("Start eff()")
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
	log.Println("Start init()")

	core.RegisterDriver(providerName, newDriver)
	gofig.Register(configRegistration())
}

func newDriver() core.Driver {
	log.Println("Start newDriver()")
	return &driver{}
}

func (d *driver) Init(r *core.RexRay) error {
	log.Println("Start Init()")

	d.r = r

	fields := eff(map[string]interface{}{
		"endpoint":   d.endpoint(),
		"userName":   d.userName(),
		"insecure":   d.insecure(),
		"volumePath": d.volumePath(),
	})

	if d.password() == "" {
		fields["password"] = ""
	} else {
		fields["password"] = "******"
	}

	if !isIsilonAttached() {
		return goof.WithFields(fields, "device not detected")
	}

	var err error

	if d.client, err = isi.NewClientWithArgs(
		d.endpoint(),
		d.insecure(),
		d.userName(),
		d.password(),
		d.volumePath()); err != nil {
		return goof.WithFieldsE(fields,
			"error creating isilon client", err)
	}

	log.WithField("provider", providerName).Info("storage driver initialized")

	return nil
}

var scsiDeviceVendors []string

func walkDevices(path string, f os.FileInfo, err error) error {
	log.Println("Start walkDevices()")
	vendorFilePath := fmt.Sprintf("%s/device/vendor", path)
	// fmt.Printf("vendorFilePath: %+v\n", string(vendorFilePath))
	data, _ := ioutil.ReadFile(vendorFilePath)
	scsiDeviceVendors = append(scsiDeviceVendors, strings.TrimSpace(string(data)))
	return nil
}

var isIsilonAttached = func() bool {
	log.Println("Start isIsilonAttached()")
	return true
	filepath.Walk("/sys/class/scsi_device/", walkDevices)
	for _, vendor := range scsiDeviceVendors {
		if vendor == "Isilon" {
			return true
		}
	}
	return false
}

func (d *driver) Name() string {
	log.Println("Start Name()")
	return providerName
}

func (d *driver) GetInstance() (*core.Instance, error) {
	log.Println("Start GetInstance()")

	return &core.Instance{}, nil

	//	instance := &core.Instance{
	//		ProviderName: providerName,
	//		InstanceID:   strconv.Itoa(initiator.Index),
	//		Region:       "",
	//		Name:         initiator.Name,
	//	}

	// log.Println("Got Instance: " + fmt.Sprintf("%+v", instance))
	//	return instance, nil
}

func (d *driver) nfsMountPath(mountPath string) string {
	return fmt.Sprintf("%s:%s", d.nfsHost(), mountPath)
}

func (d *driver) GetVolumeMapping() ([]*core.BlockDevice, error) {
	log.Println("Start GetVolumeMapping()")

	exports, err := d.client.GetVolumeExports()
	if err != nil {
		return nil, err
	}

	var BlockDevices []*core.BlockDevice
	for _, export := range exports {
		device := &core.BlockDevice{
			ProviderName: providerName,
			InstanceID:   "",
			Region:       "",
			DeviceName:   d.nfsMountPath(export.ExportPath),
			VolumeID:     export.Volume.Name,
			NetworkName:  export.ExportPath,
			Status:       "",
		}
		BlockDevices = append(BlockDevices, device)
	}

	return BlockDevices, nil
}

func (d *driver) getVolume(volumeID, volumeName string) ([]isi.Volume, error) {
	log.Println("Start getVolume()")
	var volumes []isi.Volume
	if volumeID != "" || volumeName != "" {
		volume, err := d.client.GetVolume(volumeID, volumeName)
		if err != nil && !strings.Contains(err.Error(), "Unable to open object") {
			return nil, err
		}
		if volume != nil {
			volumes = append(volumes, volume)
		}
	} else {
		var err error
		volumes, err = d.client.GetVolumes()
		if err != nil {
			return nil, err
		}

	}
	return volumes, nil
}

func (d *driver) GetVolume(volumeID, volumeName string) ([]*core.Volume, error) {
	log.Println("Start GetVolume()")

	volumes, err := d.getVolume(volumeID, volumeName)
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, nil
	}

	localVolumeMappings, err := d.GetVolumeMapping()
	if err != nil {
		return nil, err
	}

	blockDeviceMap := make(map[string]*core.BlockDevice)
	for _, volume := range localVolumeMappings {
		blockDeviceMap[volume.VolumeID] = volume
	}

	var volumesSD []*core.Volume
	for _, volume := range volumes {
		var attachmentsSD []*core.VolumeAttachment
		if _, exists := blockDeviceMap[volume.Name]; exists {
			attachmentSD := &core.VolumeAttachment{
				VolumeID: volume.Name,
				//				InstanceID: strconv.Itoa(d.initiator.Index),
				DeviceName: blockDeviceMap[volume.Name].DeviceName,
				Status:     "",
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		volumeSD := &core.Volume{
			Name:             volume.Name,
			VolumeID:         volume.Name,
			Size:             "", //strconv.Itoa(volSize / 1024 / 1024),
			AvailabilityZone: "",
			NetworkName:      d.client.Path(volume.Name), //volume.NaaName,
			Attachments:      attachmentsSD,
		}
		volumesSD = append(volumesSD, volumeSD)
	}

	return volumesSD, nil
}

func (d *driver) CreateVolume(
	notUsed bool,
	volumeName, volumeID, snapshotID, NUvolumeType string,
	NUIOPS, size int64, NUavailabilityZone string) (*core.Volume, error) {
	log.Println("Start CreateVolume() (", volumeName, ") (", volumeID, ")")

	newIsiVolume, _ := d.client.CreateVolume(volumeName)
	volumes, _ := d.GetVolume(newIsiVolume.Name, newIsiVolume.Name)

	return volumes[0], nil
}

func (d *driver) RemoveVolume(volumeID string) error {
	log.Println("Start RemoveVolume()")
	err := d.client.DeleteVolume(volumeID)
	if err != nil {
		return err
	}

	log.Println("Deleted Volume: " + volumeID)
	return nil
}

//GetSnapshot returns snapshots from a volume or a specific snapshot
func (d *driver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {
	return nil, nil
}

func getIndex(href string) string {
	log.Println("Start getIndex()")
	hrefFields := strings.Split(href, "/")
	return hrefFields[len(hrefFields)-1]
}

func (d *driver) CreateSnapshot(
	notUsed bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {
	return nil, nil
}

func (d *driver) RemoveSnapshot(snapshotID string) error {
	return nil
}

func (d *driver) GetVolumeAttach(volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
	log.Println("Start GetVolumeAttach()")
	if volumeID == "" {
		return []*core.VolumeAttachment{}, errors.ErrMissingVolumeID
	}
	volume, err := d.GetVolume(volumeID, "")
	if err != nil {
		return []*core.VolumeAttachment{}, err
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

func (d *driver) AttachVolume(
	notused bool,
	volumeID, instanceID string, force bool) ([]*core.VolumeAttachment, error) {
	log.Println("Start AttachVolume()")

	if volumeID == "" {
		return nil, errors.ErrMissingVolumeID
	}

	volumes, err := d.GetVolume(volumeID, "")
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, errors.ErrNoVolumesReturned
	}

	if err := d.client.ExportVolume(volumeID); err != nil {
		return nil, goof.WithError("problem exporting volume", err)
	}

	volumeAttachment, err := d.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, err
	}

	return volumeAttachment, nil

}

func (d *driver) DetachVolume(notUsed bool, volumeID string, blank string, force bool) error {
	log.Println("Start DetachVolume()")
	if volumeID == "" {
		return errors.ErrMissingVolumeID
	}

	volumes, err := d.GetVolume(volumeID, "")
	if err != nil {
		return err
	}

	if len(volumes) == 0 {
		return errors.ErrNoVolumesReturned
	}

	if err := d.client.UnexportVolume(volumeID); err != nil {
		return goof.WithError("problem unexporting volume", err)
	}

	log.Println("Detached volume", volumeID)
	return nil
}

func (d *driver) CopySnapshot(
	runAsync bool,
	volumeID, snapshotID, snapshotName,
	destinationSnapshotName, destinationRegion string) (*core.Snapshot, error) {
	log.Println("Start CopySnapshot()")
	return nil, errors.ErrNotImplemented
}

func (d *driver) GetDeviceNextAvailable() (string, error) {
	log.Println("Start GetDeviceNextAvailable()")
	return "", errors.ErrNotImplemented
}

func (d *driver) endpoint() string {
	return d.r.Config.GetString("isilon.endpoint")
}

func (d *driver) insecure() bool {
	return d.r.Config.GetBool("isilon.insecure")
}

func (d *driver) userName() string {
	return d.r.Config.GetString("isilon.userName")
}

func (d *driver) password() string {
	return d.r.Config.GetString("isilon.password")
}

func (d *driver) volumePath() string {
	return d.r.Config.GetString("isilon.volumePath")
}

func (d *driver) nfsHost() string {
	return d.r.Config.GetString("isilon.nfsHost")
}

func configRegistration() *gofig.Registration {
	log.Println("Start configRegistration()")
	r := gofig.NewRegistration("Isilon")
	r.Key(gofig.String, "", "", "", "isilon.endpoint")
	r.Key(gofig.Bool, "", false, "", "isilon.insecure")
	r.Key(gofig.String, "", "", "", "isilon.userName")
	r.Key(gofig.String, "", "", "", "isilon.password")
	r.Key(gofig.String, "", "", "", "isilon.volumePath")
	r.Key(gofig.String, "", "", "", "isilon.nfsHost")
	return r
}
