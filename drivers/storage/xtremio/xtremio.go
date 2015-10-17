package xtremio

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	xtio "github.com/emccode/goxtremio"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

const providerName = "XtremIO"

// The XtremIO storage driver.
type xtremIODriver struct {
	client       *xtio.Client
	initiator    xtio.Initiator
	volumesSig   string
	lunMapsSig   string
	volumesByNaa map[string]xtio.Volume
	r            *core.RexRay
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
	return &xtremIODriver{}
}

func (d *xtremIODriver) Init(r *core.RexRay) error {

	d.r = r
	d.volumesByNaa = map[string]xtio.Volume{}

	fields := eff(map[string]interface{}{
		"endpoint":         r.Config.XtremIOEndpoint,
		"userName":         r.Config.XtremIOUserName,
		"deviceMapper":     r.Config.XtremIODeviceMapper,
		"multipath":        r.Config.XtremIOMultipath,
		"remoteManagement": r.Config.XtremIORemoteManagement,
		"insecure":         r.Config.XtremIOInsecure,
	})

	if r.Config.XtremIoPassword == "" {
		fields["password"] = ""
	} else {
		fields["password"] = "******"
	}

	if !isXtremIOAttached() && !d.r.Config.XtremIORemoteManagement {
		return errors.WithFields(fields, "device not detected")
	}

	var err error

	if d.client, err = xtio.NewClientWithArgs(
		r.Config.XtremIOEndpoint,
		r.Config.XtremIOInsecure,
		r.Config.XtremIOUserName,
		r.Config.XtremIoPassword); err != nil {
		return errors.WithFieldsE(fields,
			"error creating xtremio client", err)
	}

	if !d.r.Config.XtremIORemoteManagement {
		var iqn string
		if iqn, err = getIQN(); err != nil {
			return err
		}
		if d.initiator, err = d.client.GetInitiator("", iqn); err != nil {
			return err
		}
	}

	log.WithField("provider", providerName).Debug("storage driver initialized")

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

var isXtremIOAttached = func() bool {
	filepath.Walk("/sys/class/scsi_device/", walkDevices)
	for _, vendor := range scsiDeviceVendors {
		if vendor == "XtremIO" {
			return true
		}
	}
	return false
}

func getIQN() (string, error) {
	data, err := ioutil.ReadFile("/etc/iscsi/initiatorname.iscsi")
	if err != nil {
		return "", err
	}

	result := string(data)
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		split := strings.Split(line, "=")
		if split[0] == "InitiatorName" {
			return split[1], nil
		}
	}
	return "", errors.New("IQN not found")
}

func (d *xtremIODriver) Name() string {
	return providerName
}

func (d *xtremIODriver) getVolumesSig() (string, error) {
	volumes, err := d.client.GetVolumes()
	if err != nil {
		return "", err
	}

	var volumesNameHref sort.StringSlice
	for _, volume := range volumes {
		volumesNameHref = append(volumesNameHref,
			fmt.Sprintf("%s-%s", volume.Name, volume.Href))
	}
	volumesNameHref.Sort()
	return strings.Join(volumesNameHref, ";"), err
}

func (d *xtremIODriver) getLunMapsSig() (string, error) {
	lunMap, err := d.client.GetLunMaps()
	if err != nil {
		return "", err
	}

	var lunMapsNameHref sort.StringSlice
	for _, lunMap := range lunMap {
		lunMapsNameHref = append(lunMapsNameHref,
			fmt.Sprintf("%s-%s", lunMap.Name, lunMap.Href))
	}
	lunMapsNameHref.Sort()
	return strings.Join(lunMapsNameHref, ";"), err
}

func (d *xtremIODriver) isVolumesSigEqual() (bool, string, error) {
	volumesSig, err := d.getVolumesSig()
	if err != nil {
		return false, "", err
	}
	if volumesSig == d.volumesSig {
		return true, volumesSig, nil
	}
	return false, volumesSig, nil
}

func (d *xtremIODriver) isLunMapsSigEqual() (bool, string, error) {
	lunMapsSig, err := d.getLunMapsSig()
	if err != nil {
		return false, "", err
	}
	if lunMapsSig == d.lunMapsSig {
		return true, lunMapsSig, nil
	}
	return false, lunMapsSig, nil
}

func (d *xtremIODriver) updateVolumesSig() error {
	oldVolumesSig := d.volumesSig
	checkSig, volumesSig, err := d.isVolumesSigEqual()
	if err != nil {
		return err
	}

	oldLunMapsSig := d.lunMapsSig
	checkMapsSig, lunMapsSig, err := d.isLunMapsSigEqual()
	if err != nil {
		return err
	}

	if checkSig && checkMapsSig {
		return nil
	}

	if oldVolumesSig != "" || oldLunMapsSig != "" {
		log.Println("volumeSig or volumeMapsSig updated")
	}

	volumes, err := d.client.GetVolumes()
	if err != nil {
		return err
	}

	for _, volume := range volumes {
		index := getIndex(volume.Href)
		volumeDetail, err := d.client.GetVolume(index, "")
		if err != nil {
			return err
		}
		d.volumesByNaa[volumeDetail.NaaName] = volumeDetail
	}

	d.volumesSig = volumesSig
	d.lunMapsSig = lunMapsSig
	return nil
}

func (d *xtremIODriver) getInitiator() (xtio.Initiator, error) {
	return d.initiator, nil
}

func (d *xtremIODriver) getLocalDeviceByID() (map[string]string, error) {
	mapDiskByID := make(map[string]string)
	diskIDPath := "/dev/disk/by-id"
	files, err := ioutil.ReadDir(diskIDPath)
	if err != nil {
		return nil, err
	}

	var match1 *regexp.Regexp
	var match2 string

	if d.r.Config.XtremIODeviceMapper || d.r.Config.XtremIOMultipath {
		match1, _ = regexp.Compile(`^dm-name-\w*$`)
		match2 = `^dm-name-\d+`
	} else {
		match1, _ = regexp.Compile(`^wwn-0x\w*$`)
		match2 = `^wwn-0x`
	}

	for _, f := range files {
		if match1.MatchString(f.Name()) {
			naaName := strings.Replace(f.Name(), match2, "", 1)
			naaName = naaName[len(naaName)-16:]
			devPath, _ := filepath.EvalSymlinks(fmt.Sprintf("%s/%s", diskIDPath, f.Name()))
			mapDiskByID[naaName] = devPath
		}
	}
	return mapDiskByID, nil
}

func (d *xtremIODriver) GetInstance() (*core.Instance, error) {

	initiator, err := d.getInitiator()
	if err != nil {
		return &core.Instance{}, err
	}

	instance := &core.Instance{
		ProviderName: providerName,
		InstanceID:   strconv.Itoa(initiator.Index),
		Region:       "",
		Name:         initiator.Name,
	}

	// log.Println("Got Instance: " + fmt.Sprintf("%+v", instance))
	return instance, nil
}

func (d *xtremIODriver) GetVolumeMapping() ([]*core.BlockDevice, error) {

	mapDiskByID, err := d.getLocalDeviceByID()
	if err != nil {
		return nil, err
	}

	err = d.updateVolumesSig()
	if err != nil {
		return nil, err
	}

	var BlockDevices []*core.BlockDevice
	for naa, blockDeviceName := range mapDiskByID {
		if volume, ok := d.volumesByNaa[naa]; ok {
			for _, level1 := range volume.LunMappingList {
				for _, level2 := range level1 {
					if reflect.TypeOf(level2).String() ==
						"[]interface {}" &&
						level2.([]interface{})[1].(string) == d.initiator.Name {
						// maybe bug in API where initiator group name changes
						// don't update to volume
						sdBlockDevice := &core.BlockDevice{
							ProviderName: providerName,
							InstanceID:   d.initiator.Name,
							Region:       volume.SysID[0].(string),
							DeviceName:   blockDeviceName,
							VolumeID:     strconv.Itoa(volume.Index),
							NetworkName:  naa,
							Status:       "",
						}
						BlockDevices = append(BlockDevices, sdBlockDevice)
					}
				}
			}
		}
	}

	return BlockDevices, nil
}

func (d *xtremIODriver) getVolume(volumeID, volumeName string) ([]xtio.Volume, error) {
	var volumes []xtio.Volume
	if volumeID != "" || volumeName != "" {
		volume, err := d.client.GetVolume(volumeID, volumeName)
		if err != nil {
			return nil, err
		}
		volumes = append(volumes, volume)
	} else {
		allVolumes, err := d.client.GetVolumes()
		if err != nil {
			return nil, err
		}
		for _, volume := range allVolumes {
			hrefFields := strings.Split(volume.Href, "/")
			index, _ := strconv.Atoi(hrefFields[len(hrefFields)-1])
			volumes = append(volumes,
				xtio.VolumeCtorNameIndex(volume.Name, index))
		}
	}
	return volumes, nil
}

func (d *xtremIODriver) GetVolume(volumeID, volumeName string) ([]*core.Volume, error) {

	volumes, err := d.getVolume(volumeID, volumeName)
	if err != nil && err.Error() == "obj_not_found" {
		return []*core.Volume{}, nil
	} else if err != nil {
		return nil, err
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
		if _, exists := blockDeviceMap[strconv.Itoa(volume.Index)]; exists {
			attachmentSD := &core.VolumeAttachment{
				VolumeID:   strconv.Itoa(volume.Index),
				InstanceID: strconv.Itoa(d.initiator.Index),
				DeviceName: blockDeviceMap[strconv.Itoa(volume.Index)].DeviceName,
				Status:     "",
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		var az string
		if reflect.TypeOf(volume.SysID).String() == "[]interface {}" {
			if len(volume.SysID) > 0 {
				az = volume.SysID[0].(string)
			}
		}

		volSize, _ := strconv.Atoi(volume.VolSize)
		volumeSD := &core.Volume{
			Name:             volume.Name,
			VolumeID:         strconv.Itoa(volume.Index),
			Size:             strconv.Itoa(volSize / 1024 / 1024),
			AvailabilityZone: az,
			NetworkName:      volume.NaaName,
			Attachments:      attachmentsSD,
		}
		volumesSD = append(volumesSD, volumeSD)
	}

	return volumesSD, nil
}

func (d *xtremIODriver) CreateVolume(
	notUsed bool,
	volumeName, volumeID, snapshotID, NUvolumeType string,
	NUIOPS, size int64, NUavailabilityZone string) (*core.Volume, error) {

	var volumes []*core.Volume
	if volumeID == "" && snapshotID == "" {
		req := &xtio.NewVolumeOptions{
			VolName: volumeName,
			VolSize: int(size) * 1024 * 1024,
		}
		res, err := d.client.NewVolume(req)
		if err != nil {
			return nil, err
		}

		index := getIndex(res.Links[0].Href)
		volumes, err = d.GetVolume(index, "")
		if err != nil {
			return nil, err
		}
	} else {
		if snapshotID != "" {
			snapshots, err := d.GetSnapshot("", snapshotID, "")
			if err != nil {
				return nil, err
			}
			volumeID = snapshots[0].VolumeID
		}
		snapshot, err := d.CreateSnapshot(false, volumeName, volumeID, "")
		if err != nil {
			return nil, err
		}

		snapshotID := snapshot[0].SnapshotID
		snapshots, err := d.getSnapshot(snapshotID, "")
		if err != nil {
			return nil, err
		}

		volumes, err = d.GetVolume(
			strconv.Itoa(int(snapshots[0].VolID[2].(float64))), "")
		if err != nil {
			return nil, err
		}
	}

	return volumes[0], nil
}

func (d *xtremIODriver) RemoveVolume(volumeID string) error {
	err := d.client.DeleteVolume(volumeID, "")
	if err != nil {
		return err
	}

	log.Println("Deleted Volume: " + volumeID)
	return nil
}

func (d *xtremIODriver) getSnapshot(snapshotID, snapshotName string) ([]xtio.Snapshot, error) {
	var snapshots []xtio.Snapshot
	if snapshotID != "" || snapshotName != "" {
		snapshot, err := d.client.GetSnapshot(snapshotID, snapshotName)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	} else {
		allSnapshots, err := d.client.GetSnapshots()
		if err != nil {
			return nil, err
		}
		for _, snapshot := range allSnapshots {
			hrefFields := strings.Split(snapshot.Href, "/")
			index, _ := strconv.Atoi(hrefFields[len(hrefFields)-1])
			snapshots = append(snapshots,
				xtio.SnapshotCtorNameIndex(snapshot.Name, index))
		}
	}
	return snapshots, nil
}

//GetSnapshot returns snapshots from a volume or a specific snapshot
func (d *xtremIODriver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {

	var snapshotsInt []*core.Snapshot
	if volumeID != "" {
		volumes, err := d.getVolume(volumeID, "")
		if err != nil {
			return nil, err
		}

		for _, volume := range volumes {
			for _, destSnap := range volume.DestSnapList {
				snapshot, err := d.getSnapshot(strconv.Itoa(
					int(destSnap.([]interface{})[2].(float64))), "")
				if err != nil {
					return nil, err
				}

				volSize, _ := strconv.Atoi(volume.VolSize)
				snapshotSD := &core.Snapshot{
					Name:        snapshot[0].Name,
					VolumeID:    strconv.Itoa(volume.Index),
					SnapshotID:  strconv.Itoa(snapshot[0].Index),
					VolumeSize:  strconv.Itoa(volSize / 1024 / 1024),
					StartTime:   snapshot[0].CreationTime,
					Description: "",
					Status:      "",
				}
				snapshotsInt = append(snapshotsInt, snapshotSD)
			}
		}
	} else {
		snapshots, err := d.getSnapshot(snapshotID, snapshotName)
		if err != nil {
			return nil, err
		}

		for _, snapshot := range snapshots {
			snapshot, err := d.client.GetSnapshot(
				strconv.Itoa(snapshot.Index), "")
			if err != nil {
				return nil, err
			}

			volume, err := d.getVolume(
				strconv.Itoa(int(snapshot.AncestorVolID[2].(float64))), "")
			if err != nil {
				return nil, err
			}

			volSize, _ := strconv.Atoi(volume[0].VolSize)
			snapshotSD := &core.Snapshot{
				Name: snapshot.Name,
				VolumeID: strconv.Itoa(
					int(snapshot.AncestorVolID[2].(float64))),
				SnapshotID:  strconv.Itoa(snapshot.Index),
				VolumeSize:  strconv.Itoa(volSize / 1024 / 1024),
				StartTime:   snapshot.CreationTime,
				Description: "",
				Status:      "",
			}
			snapshotsInt = append(snapshotsInt, snapshotSD)
		}

	}

	return snapshotsInt, nil
}

func getIndex(href string) string {
	hrefFields := strings.Split(href, "/")
	return hrefFields[len(hrefFields)-1]
}

func (d *xtremIODriver) CreateSnapshot(
	notUsed bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {

	volume, err := d.client.GetVolume(volumeID, "")
	if err != nil {
		return nil, err
	}

	//getfolder of volume
	req := &xtio.NewSnapshotOptions{
		SnapList: xtio.NewSnapListItems(volume.Name, snapshotName),
		FolderID: "/",
	}

	postSnapshotsResp, err := d.client.NewSnapshot(req)
	if err != nil {
		return nil, err
	}

	index := getIndex(postSnapshotsResp.Links[0].Href)
	snapshot, err := d.GetSnapshot("", index, "")
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (d *xtremIODriver) RemoveSnapshot(snapshotID string) error {
	err := d.client.DeleteSnapshot(snapshotID, "")
	if err != nil {
		return err
	}

	return nil
}

func (d *xtremIODriver) GetVolumeAttach(volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
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

func (d *xtremIODriver) waitAttach(volumeID string) (*core.BlockDevice, error) {

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(10 * time.Second)
		timeout <- true
	}()

	successCh := make(chan *core.BlockDevice, 1)
	errorCh := make(chan error, 1)
	go func(volumeID string) {
		log.Println("XtremIO: waiting for volume attach")
		for {
			if d.r.Config.XtremIOMultipath {
				_, err := exec.Command("/sbin/multipath").Output()
				if err != nil {
					errorCh <- fmt.Errorf(
						"Error refreshing multipath: %s", err)
				}
			}

			blockDevices, err := d.GetVolumeMapping()
			if err != nil {
				errorCh <- fmt.Errorf(
					"XtremIO: problem getting local block devices: %s", err)
				return
			}

			for _, blockDevice := range blockDevices {
				if blockDevice.VolumeID == volumeID {
					successCh <- blockDevice
					return
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}(volumeID)

	select {
	case blockDevice := <-successCh:
		log.Println(fmt.Sprintf("XtremIO: got attachedVolume %s at %s",
			blockDevice.VolumeID, blockDevice.DeviceName))
		return blockDevice, nil
	case err := <-errorCh:
		return nil, err
	case <-timeout:
		return nil, fmt.Errorf("XtremIO: timed out waiting for mount")
	}

}

func (d *xtremIODriver) AttachVolume(
	runAsync bool,
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {

	if volumeID == "" {
		return nil, errors.ErrMissingVolumeID
	}

	// doing a lookup here for intiator name as IG name, so limited to IG name
	// as initiator name to work for now 1) need to consider when instanceid is
	// blank,need to lookup ig from instanceid

	// if instanceID == "" {
	initiatorGroup, err := d.client.GetInitiatorGroup("", d.initiator.Name)
	if err != nil {
		return nil, err
	}

	initiatorGroupID := strconv.Itoa(initiatorGroup.Index)

	initiatorGroupIDI, err := strconv.Atoi(initiatorGroupID)
	if err != nil {
		return nil, err
	}
	volumeIDI, err := strconv.Atoi(volumeID)
	if err != nil {
		return nil, err
	}
	req := &xtio.NewLunMapOptions{
		VolID: volumeIDI,
		IgID:  initiatorGroupIDI,
	}

	_, err = d.client.NewLunMap(req)
	if err != nil {
		return nil, err
	}

	if !runAsync {
		_, err := d.waitAttach(volumeID)
		if err != nil {
			return nil, err
		}
	} else {
		_, _ = d.GetVolumeMapping()
	}

	volumeAttachment, err := d.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, err
	}

	return volumeAttachment, nil

}

func (d *xtremIODriver) getLunMaps(initiatorName, volumeID string) (xtio.Refs, error) {
	if initiatorName == "" {
		return nil, errors.New("Missing initiatorName")
	}

	initiatorGroup, err := d.client.GetInitiatorGroup("", initiatorName)
	if err != nil {
		return nil, err
	}

	lunMaps, err := d.client.GetLunMaps()
	if err != nil {
		return nil, err
	}

	var refs xtio.Refs
	for _, ref := range lunMaps {

		idents := strings.Split(ref.Name, "_")
		if len(idents) < 3 {
			continue
		} else if strconv.Itoa(initiatorGroup.Index) == idents[1] && volumeID == idents[0] {
			refs = append(refs, ref)
		}
	}

	return refs, nil
}

func (d *xtremIODriver) DetachVolume(notUsed bool, volumeID string, blank string) error {
	if volumeID == "" {
		return errors.ErrMissingVolumeID
	}

	volumes, err := d.getVolume(volumeID, "")
	if err != nil {
		return err
	}

	if len(volumes) == 0 {
		return errors.ErrNoVolumesReturned
	}

	if d.r.Config.XtremIOMultipath {
		_, err := exec.Command("/sbin/multipath", "-f",
			fmt.Sprintf("3%s", volumes[0].NaaName)).Output()
		if err != nil {
			return fmt.Errorf("Error removing multipath: %s", err)
		}
	}

	lunMaps, err := d.getLunMaps(
		d.initiator.Name, strconv.Itoa(volumes[0].Index))
	if err != nil {
		return err
	}

	if len(lunMaps) == 0 {
		return nil
	}

	index := getIndex(lunMaps[0].Href)
	if err = d.client.DeleteLunMap(index, ""); err != nil {
		return err
	}

	log.Println("Detached volume", volumeID)
	return nil
}

func (d *xtremIODriver) CopySnapshot(
	runAsync bool,
	volumeID, snapshotID, snapshotName,
	destinationSnapshotName, destinationRegion string) (*core.Snapshot, error) {
	return nil, errors.ErrNotImplemented
}

func (d *xtremIODriver) GetDeviceNextAvailable() (string, error) {
	return "", errors.ErrNotImplemented
}
