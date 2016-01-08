package isilon

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"strconv"
	"strings"

	isi "github.com/emccode/goisilon"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

const providerName = "Isilon"
const bytesPerGb = int64(1024 * 1024 * 1024)
const idDelimiter = "/"

// The Isilon storage driver.
type driver struct {
	client *isi.Client
	r      *core.RexRay
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
		"endpoint":   d.endpoint(),
		"userName":   d.userName(),
		"group":      d.group(),
		"insecure":   d.insecure(),
		"volumePath": d.volumePath(),
		"dataSubnet": d.dataSubnet(),
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
		d.group(),
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
	vendorFilePath := fmt.Sprintf("%s/device/vendor", path)
	// fmt.Printf("vendorFilePath: %+v\n", string(vendorFilePath))
	data, _ := ioutil.ReadFile(vendorFilePath)
	scsiDeviceVendors = append(scsiDeviceVendors, strings.TrimSpace(string(data)))
	return nil
}

var isIsilonAttached = func() bool {
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
	return providerName
}

// Create an instance ID from a list of client IP addresses
func createInstanceId(clients []string) string {
	return strings.Join(clients, idDelimiter)
}

// Parse an instance ID into a list of client IP addresses
func parseInstanceId(id string) []string {
	return strings.Split(id, idDelimiter)
}

func (d *driver) GetInstance() (*core.Instance, error) {

	// parse the data subnet
	_, dataSubnet, err := net.ParseCIDR(d.dataSubnet())
	if err != nil {
		return nil, goof.WithFieldsE(
			eff(goof.Fields{"dataSubnet": d.dataSubnet()}), "Invalid data subnet", err)
	}

	// find all local IP addresses on the data subnet
	ipList, err := net.InterfaceAddrs()
	if err != nil {
		return nil, goof.WithError("No Network Interface Addresses", err)
	}
	var idList []string
	for _, addr := range ipList {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			return nil, err
		}
		if dataSubnet.Contains(ip) == true {
			idList = append(idList, ip.String())
		}
	}

	if len(idList) == 0 {
		return nil, goof.WithFieldsE(
			eff(goof.Fields{"dataSubnet": d.dataSubnet()}), "No IPs in the data subnet", err)
	}

	instance := &core.Instance{
		ProviderName: providerName,
		InstanceID:   createInstanceId(idList),
		Region:       "",
		Name:         "",
	}

	return instance, nil
}

func (d *driver) nfsMountPath(mountPath string) string {
	return fmt.Sprintf("%s:%s", d.nfsHost(), mountPath)
}

func (d *driver) GetVolumeMapping() ([]*core.BlockDevice, error) {
	exports, err := d.client.GetVolumeExports()
	if err != nil {
		return nil, err
	}

	var BlockDevices []*core.BlockDevice
	for _, export := range exports {

		device := &core.BlockDevice{
			ProviderName: providerName,
			InstanceID:   createInstanceId(export.Clients),
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

func (d *driver) getSize(volumeID, volumeName string) (int64, error) {
	if d.quotas() == false {
		return 0, nil
	}

	if volumeID != "" {
		volumeName = volumeID
	}
	if volumeName != "" {
		quota, err := d.client.GetQuota(volumeName)
		if err != nil {
			return 0, nil
		}
		// PAPI returns the size in bytes, REX-Ray uses gigs
		return quota.Thresholds.Hard / bytesPerGb, nil
	}

	return 0, errors.ErrMissingVolumeID

}

func (d *driver) GetVolume(volumeID, volumeName string) ([]*core.Volume, error) {
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
				VolumeID:   volume.Name,
				InstanceID: blockDeviceMap[volume.Name].InstanceID,
				DeviceName: blockDeviceMap[volume.Name].DeviceName,
				Status:     "",
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		volSize, _ := d.getSize(volume.Name, volume.Name)

		volumeSD := &core.Volume{
			Name:             volume.Name,
			VolumeID:         volume.Name,
			Size:             strconv.FormatInt(volSize, 10),
			AvailabilityZone: "",
			NetworkName:      d.client.Path(volume.Name),
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

	var err error

	fields := eff(map[string]interface{}{
		"volumeName": volumeName,
		"volumeId":   volumeID,
		"snapshotId": snapshotID,
		"size":       size,
	})

	log.WithFields(fields).Debug("Creating volume")

	if volumeName == "" {
		return nil, goof.New("Cannot create a volume.  No volume name provided")
	}
	if volumeID != "" {
		// a volume id was passed in.  copy that volume into the new one
		// instead of creating an empty volume.
		_, err = d.client.CopyVolume(volumeID, volumeName)
		if err != nil {
			return nil, goof.WithFieldsE(
				eff(goof.Fields{"volumeName": volumeName}), "Error copying volume", err)
		}
		if size < 1 {
			// get the size from the existing volume
			size, err = d.getSize(volumeID, "")
			if err != nil {
				return nil, goof.WithFieldsE(
					eff(goof.Fields{"volumeName": volumeName}), "Error reading size of existing volume", err)
			}
		}
	} else if snapshotID != "" {
		// a snapshot id was passed in.  copy the volume out of the given
		// snapshot into the new one instead of creating an empty one.
		_, err = d.client.CopySnapshot(-1, snapshotID, volumeName)
		if err != nil {
			return nil, goof.WithFieldsE(
				eff(goof.Fields{"volumeName": volumeName}), "Error creating volume from snapshot", err)
		}
		if size < 1 {
			// get the size from the volume in the snapshot
			snapshots, err := d.GetSnapshot("", snapshotID, snapshotID)
			if err != nil {
				return nil, goof.WithFieldsE(
					eff(goof.Fields{"volumeName": volumeName}), fmt.Sprintf("Error creating volume from snapshot.  Can't query snapshot with id: (%s) (%s)", snapshotID, err), err)
			}
			if len(snapshots) <= 0 {
				return nil, goof.WithFieldsE(
					eff(goof.Fields{"volumeName": volumeName}), fmt.Sprintf("Error creating volume from snapshot.  No snapshot with id: (%s)", snapshotID), err)
			}
			size, err = d.getSize("", snapshots[0].VolumeID)
			if err != nil {
				return nil, goof.WithFieldsE(
					eff(goof.Fields{"volumeName": volumeName}), "Error finding size for volume from snapshot", err)
			}
		}
	} else {
		_, err = d.client.CreateVolume(volumeName)
		if err != nil {
			return nil, goof.WithFieldsE(
				eff(goof.Fields{"volumeName": volumeName}), "Error creating volume", err)
		}
	}

	// Set or update the quota for volume
	if d.quotas() {
		quota, err := d.client.GetQuota(volumeName)
		if quota == nil {
			// PAPI uses bytes for it's size units, but REX-Ray uses gigs
			err = d.client.SetQuotaSize(volumeName, size*bytesPerGb)
			if err != nil {
				// TODO: not sure how to handle this situation.  Delete created volume
				// and return an error?  Ignore and continue?
				return nil, goof.WithFieldsE(
					eff(goof.Fields{"volumeName": volumeName}), "Error setting quota for new volume", err)
			}
		} else {
			// PAPI uses bytes for it's size units, but REX-Ray uses gigs
			err = d.client.UpdateQuotaSize(volumeName, size*bytesPerGb)
			if err != nil {
				// TODO: not sure how to handle this situation.  Delete created volume
				// and return an error?  Ignore and continue?
				return nil, goof.WithFieldsE(
					eff(goof.Fields{"volumeName": volumeName}), "Error updating quota for existing volume", err)
			}
		}
	}

	volumes, _ := d.GetVolume("", volumeName)
	if volumes == nil || err != nil {
		return nil, goof.WithFieldsE(
			eff(goof.Fields{"volumeName": volumeName}), "Error getting volume", err)
	}

	return volumes[0], nil
}

func (d *driver) RemoveVolume(volumeID string) error {
	if d.quotas() {
		err := d.client.ClearQuota(volumeID)
		if err != nil {
			return err
		}
	}

	err := d.client.DeleteVolume(volumeID)
	if err != nil {
		return err
	}

	return nil
}

//GetSnapshot returns snapshots from a volume or a specific snapshot
func (d *driver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {

	var snapshotsSD []*core.Snapshot

	if snapshotID != "" || snapshotName != "" {
		idInt, err := strconv.ParseInt(snapshotID, 10, 64)
		if err != nil {
			idInt = -1
		}
		snapshot, err := d.client.GetSnapshot(idInt, snapshotName)
		if err != nil {
			return nil, err
		}
		if snapshot == nil {
			return nil, nil
		}

		volumeName := d.client.NameFromPath(snapshot.Path)
		if volumeID != "" && volumeID != volumeName {
			return nil, goof.New(fmt.Sprintf("Snapshot volume name does not match volumeID: Snapshot volume: (%s) volumeID: (%s)", volumeName, volumeID))
		}
		size, err := d.getSize("", volumeName)
		if err != nil {
			return nil, err
		}

		snapshotSD := &core.Snapshot{
			Name:        snapshot.Name,
			VolumeID:    volumeName,
			SnapshotID:  strconv.FormatInt(snapshot.Id, 10),
			VolumeSize:  strconv.FormatInt(size, 10),
			StartTime:   strconv.FormatInt(snapshot.Created, 10),
			Description: "",
			Status:      snapshot.State,
		}
		snapshotsSD = append(snapshotsSD, snapshotSD)
	} else if volumeID != "" {

		snapshots, err := d.client.GetSnapshotsByPath(volumeID)
		if err != nil {
			return nil, err
		}
		if snapshots == nil {
			return nil, nil
		}

		for _, snapshot := range snapshots {
			volumeName := d.client.NameFromPath(snapshot.Path)
			size, err := d.getSize("", volumeName)
			if err != nil {
				return nil, err
			}

			snapshotSD := &core.Snapshot{
				Name:        snapshot.Name,
				VolumeID:    volumeName,
				SnapshotID:  strconv.FormatInt(snapshot.Id, 10),
				VolumeSize:  strconv.FormatInt(size, 10),
				StartTime:   strconv.FormatInt(snapshot.Created, 10),
				Description: "",
				Status:      snapshot.State,
			}
			snapshotsSD = append(snapshotsSD, snapshotSD)
		}
	}

	return snapshotsSD, nil
}

func getIndex(href string) string {
	hrefFields := strings.Split(href, "/")
	return hrefFields[len(hrefFields)-1]
}

func (d *driver) CreateSnapshot(
	notUsed bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {

	if volumeID == "" {
		return []*core.Snapshot{}, errors.ErrMissingVolumeID
	}
	volume, err := d.GetVolume(volumeID, "")
	if err != nil {
		return []*core.Snapshot{}, err
	}

	snapshot, err := d.client.CreateSnapshot(volume[0].Name, snapshotName)
	if err != nil {
		return []*core.Snapshot{}, err
	}
	if snapshot == nil {
		return nil, nil
	}

	var snapshotsSD []*core.Snapshot
	size, err := d.getSize(volumeID, "")
	if err != nil {
		return nil, err
	}

	snapshotSD := &core.Snapshot{
		Name:        snapshot.Name,
		VolumeID:    volumeID,
		SnapshotID:  strconv.FormatInt(snapshot.Id, 10),
		VolumeSize:  strconv.FormatInt(size, 64),
		StartTime:   strconv.FormatInt(snapshot.Created, 10),
		Description: "",
		Status:      snapshot.State,
	}
	snapshotsSD = append(snapshotsSD, snapshotSD)

	return snapshotsSD, nil
}

func (d *driver) RemoveSnapshot(snapshotID string) error {
	idInt, err := strconv.ParseInt(snapshotID, 10, 64)
	if err != nil {
		return err
	}
	return d.client.RemoveSnapshot(idInt, "")
}

func (d *driver) GetVolumeAttach(volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
	if volumeID == "" {
		return []*core.VolumeAttachment{}, errors.ErrMissingVolumeID
	}
	volume, err := d.GetVolume(volumeID, "")
	if err != nil {
		return []*core.VolumeAttachment{}, err
	}

	if instanceID != "" {
		for _, volumeAttachment := range volume[0].Attachments {
			if volumeAttachment.InstanceID == instanceID {
				return volume[0].Attachments, nil
			}
		}
		// not attached
		return []*core.VolumeAttachment{}, nil
	}
	return volume[0].Attachments, nil
}

func (d *driver) AttachVolume(
	notused bool,
	volumeID, instanceID string, force bool) ([]*core.VolumeAttachment, error) {

	// sanity check the input
	if volumeID == "" {
		return nil, errors.ErrMissingVolumeID
	}
	if instanceID == "" {
		return nil, goof.New("Missing Instance ID")
	}
	// ensure the volume exists and is exported
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
	// see if anyone is attached already
	clients, err := d.client.GetExportClients(volumeID)
	if err != nil {
		return nil, goof.WithError("problem getting export client", err)
	}

	// clear out any existing clients if necessary.  if force is false and
	// we have existing clients, we need to exit.
	if len(clients) > 0 {
		if force == false {
			return nil, goof.New("Volume already attached to another host")
		}

		// remove all clients
		err = d.client.ClearExportClients(volumeID)
		if err != nil {
			return nil, err
		}
	}

	err = d.client.SetExportClients(volumeID, parseInstanceId(instanceID))
	if err != nil {
		return nil, err
	}

	volumeAttachment, err := d.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, err
	}

	return volumeAttachment, nil

}

func (d *driver) DetachVolume(notUsed bool, volumeID string, blank string, notused bool) error {
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

	return nil
}

func (d *driver) CopySnapshot(
	runAsync bool,
	volumeID, snapshotID, snapshotName,
	destinationSnapshotName, destinationRegion string) (*core.Snapshot, error) {
	return nil, goof.New("This driver does not implement CopySnapshot")
}

func (d *driver) GetDeviceNextAvailable() (string, error) {
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

func (d *driver) group() string {
	return d.r.Config.GetString("isilon.group")
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

func (d *driver) dataSubnet() string {
	return d.r.Config.GetString("isilon.dataSubnet")
}

func (d *driver) quotas() bool {
	return d.r.Config.GetBool("isilon.quotas")
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("Isilon")
	r.Key(gofig.String, "", "", "", "isilon.endpoint")
	r.Key(gofig.Bool, "", false, "", "isilon.insecure")
	r.Key(gofig.String, "", "", "", "isilon.userName")
	r.Key(gofig.String, "", "", "", "isilon.group")
	r.Key(gofig.String, "", "", "", "isilon.password")
	r.Key(gofig.String, "", "", "", "isilon.volumePath")
	r.Key(gofig.String, "", "", "", "isilon.nfsHost")
	r.Key(gofig.String, "", "", "", "isilon.dataSubnet")
	r.Key(gofig.Bool, "", false, "", "isilon.quotas")
	return r
}
