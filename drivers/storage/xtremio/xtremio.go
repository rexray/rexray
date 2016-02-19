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
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	xtio "github.com/emccode/goxtremio"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

const providerName = "XtremIO"

// The XtremIO storage driver.
type driver struct {
	client           *xtio.Client
	initiator        xtio.Initiator
	volumesSig       string
	lunMapsSig       string
	initiatorsSig    string
	volumesByNaa     map[string]xtio.Volume
	initiatorsByName map[string]xtio.Initiator
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
	d.volumesByNaa = map[string]xtio.Volume{}

	fields := eff(map[string]interface{}{
		"moduleName":       d.r.Context,
		"endpoint":         d.endpoint(),
		"userName":         d.userName(),
		"deviceMapper":     d.deviceMapper(),
		"multipath":        d.multipath(),
		"remoteManagement": d.remoteManagement(),
		"insecure":         d.insecure(),
	})

	if d.password() == "" {
		fields["password"] = ""
	} else {
		fields["password"] = "******"
	}

	if !isXtremIOAttached() && !d.remoteManagement() {
		return goof.WithFields(fields, "device not detected")
	}

	var err error

	if d.client, err = xtio.NewClientWithArgs(
		d.endpoint(),
		d.insecure(),
		d.userName(),
		d.password()); err != nil {
		return goof.WithFieldsE(fields,
			"error creating xtremio client", err)
	}

	if !d.remoteManagement() {
		var iqn string
		if iqn, err = getIQN(); err != nil {
			return goof.WithFieldsE(fields,
				"error getting IQN", err)
		}
		if d.initiator, err = d.client.GetInitiator("", iqn); err != nil {
			return goof.WithFieldsE(fields,
				"error getting initiator", err)
		}
	}

	log.WithFields(fields).Info("storage driver initialized")

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
		return "", goof.WithError("problem reading /etc/iscsi/initiatorname.iscsi", err)
	}

	result := string(data)
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		split := strings.Split(line, "=")
		if split[0] == "InitiatorName" {
			return split[1], nil
		}
	}
	return "", goof.New("IQN not found")
}

func (d *driver) Name() string {
	return providerName
}

func (d *driver) getVolumesSig() (string, error) {
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

func (d *driver) getLunMapsSig() (string, error) {
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

func (d *driver) getInitiatorsSig() (string, error) {
	inits, err := d.client.GetInitiators()
	if err != nil {
		return "", err
	}

	var initsNameHref sort.StringSlice
	for _, init := range inits {
		initsNameHref = append(initsNameHref,
			fmt.Sprintf("%s-%s", init.Name, init.Href))
	}
	initsNameHref.Sort()
	return strings.Join(initsNameHref, ";"), err
}

func (d *driver) isVolumesSigEqual() (bool, string, error) {
	volumesSig, err := d.getVolumesSig()
	if err != nil {
		return false, "", err
	}
	if volumesSig == d.volumesSig {
		return true, volumesSig, nil
	}
	return false, volumesSig, nil
}

func (d *driver) isLunMapsSigEqual() (bool, string, error) {
	lunMapsSig, err := d.getLunMapsSig()
	if err != nil {
		return false, "", err
	}
	if lunMapsSig == d.lunMapsSig {
		return true, lunMapsSig, nil
	}
	return false, lunMapsSig, nil
}

func (d *driver) isInitiatorsSigEqual() (bool, string, error) {
	initiatorsSig, err := d.getInitiatorsSig()
	if err != nil {
		return false, "", err
	}
	if initiatorsSig == d.initiatorsSig {
		return true, initiatorsSig, nil
	}
	return false, initiatorsSig, nil
}

func (d *driver) updateInitiatorMap() error {
	initiators, err := d.client.GetInitiators()
	if err != nil {
		return err
	}

	d.initiatorsByName = make(map[string]xtio.Initiator)

	for _, initiator := range initiators {
		index := getIndex(initiator.Href)
		initiatorDetail, err := d.client.GetInitiator(index, "")
		if err != nil {
			return err
		}
		d.initiatorsByName[initiatorDetail.Name] = initiatorDetail
	}

	return nil
}

func (d *driver) updateInitiatorsSig() error {
	checkSig, initiatorsSig, err := d.isInitiatorsSigEqual()
	if err != nil {
		return err
	}

	if checkSig {
		return nil
	}

	if err := d.updateInitiatorMap(); err != nil {
		return err
	}

	d.initiatorsSig = initiatorsSig

	return nil
}

func (d *driver) updateVolumesSig() error {
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

func (d *driver) getInitiator() (xtio.Initiator, error) {
	return d.initiator, nil
}

func (d *driver) getLocalDeviceByID() (map[string]string, error) {
	mapDiskByID := make(map[string]string)
	diskIDPath := "/dev/disk/by-id"
	files, err := ioutil.ReadDir(diskIDPath)
	if err != nil {
		return nil, err
	}

	var match1 *regexp.Regexp
	var match2 string

	if d.deviceMapper() || d.multipath() {
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

func (d *driver) GetInstance() (*core.Instance, error) {

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

func (d *driver) GetVolumeMapping() ([]*core.BlockDevice, error) {

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

func (d *driver) getVolume(volumeID, volumeName string) ([]xtio.Volume, error) {

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

func (d *driver) GetVolume(volumeID, volumeName string) ([]*core.Volume, error) {
	fields := eff(map[string]interface{}{
		"moduleName": d.r.Context,
		"volumeID":   volumeID,
		"volumeName": volumeName,
	})

	volumes, err := d.getVolume(volumeID, volumeName)
	if err != nil && err.Error() == "obj_not_found" {
		return []*core.Volume{}, nil
	} else if err != nil {
		return nil, goof.WithFieldsE(fields, "error getting volumes", err)
	}

	mapDiskByID, err := d.getLocalDeviceByID()
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error getting local device by ID", err)
	}

	if err := d.updateInitiatorsSig(); err != nil {
		return nil, goof.WithFieldsE(fields, "error updating initiators signatures", err)
	}

	if err := d.updateVolumesSig(); err != nil {
		return nil, goof.WithFieldsE(fields, "error updating volumes signature", err)
	}

	var volumesSD []*core.Volume
	for _, volume := range volumes {
		blockDeviceName, _ := mapDiskByID[volume.NaaName]
		var attachmentsSD []*core.VolumeAttachment
		for _, level1 := range volume.LunMappingList {
			for _, level2 := range level1 {
				var initiatorIndex int
				if reflect.TypeOf(level2).String() == "[]interface {}" {
					initiatorName := level2.([]interface{})[1].(string)
					if initiator, ok := d.initiatorsByName[initiatorName]; ok {
						initiatorIndex = initiator.Index
					}
				}
				if initiatorIndex != 0 {
					attachmentSD := &core.VolumeAttachment{
						VolumeID:   strconv.Itoa(volume.Index),
						InstanceID: strconv.Itoa(initiatorIndex),
						DeviceName: blockDeviceName,
						Status:     "",
					}
					attachmentsSD = append(attachmentsSD, attachmentSD)
				}
			}
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

func (d *driver) CreateVolume(
	notUsed bool,
	volumeName, volumeID, snapshotID, NUvolumeType string,
	NUIOPS, size int64, NUavailabilityZone string) (*core.Volume, error) {

	fields := eff(map[string]interface{}{
		"moduleName": d.r.Context,
		"volumeID":   volumeID,
		"volumeName": volumeName,
		"snapshotID": snapshotID,
		"size":       size,
	})

	var volumes []*core.Volume
	if volumeID == "" && snapshotID == "" {
		req := &xtio.NewVolumeOptions{
			VolName: volumeName,
			VolSize: int(size) * 1024 * 1024,
		}
		res, err := d.client.NewVolume(req)
		if err != nil {
			return nil, goof.WithFieldsE(fields, "error creating new volume", err)
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

func (d *driver) RemoveVolume(volumeID string) error {
	fields := eff(map[string]interface{}{
		"moduleName": d.r.Context,
		"volumeID":   volumeID,
	})

	err := d.client.DeleteVolume(volumeID, "")
	if err != nil {
		return goof.WithFieldsE(fields, "error deleting volume", err)
	}

	return nil
}

func (d *driver) getSnapshot(snapshotID, snapshotName string) ([]xtio.Snapshot, error) {
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
func (d *driver) GetSnapshot(
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

func (d *driver) CreateSnapshot(
	notUsed bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {

	volume, err := d.client.GetVolume(volumeID, "")
	if err != nil {
		return nil, goof.WithFieldE("volumeID", volumeID, "error getting volume", err)
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

func (d *driver) RemoveSnapshot(snapshotID string) error {
	err := d.client.DeleteSnapshot(snapshotID, "")
	if err != nil {
		return goof.WithFieldE("snapshotID", snapshotID, "error deleting snapshot", err)
	}

	return nil
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

func (d *driver) waitAttach(volumeID string) (*core.BlockDevice, error) {

	volumes, err := d.GetVolume(volumeID, "")
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, goof.New("no volumes returned")
	}

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
			if d.multipath() {
				_, _ = exec.Command("/sbin/multipath", "-f",
					fmt.Sprintf("3%s", volumes[0].NetworkName)).Output()
				_, _ = exec.Command("/sbin/multipath").Output()
			}

			blockDevices, err := d.GetVolumeMapping()
			if err != nil {
				errorCh <- goof.Newf(
					"problem getting local block devices: %s", err)
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
		return nil, goof.New("timed out waiting for mount")
	}

}

func (d *driver) AttachVolume(
	runAsync bool,
	volumeID, instanceID string, force bool) ([]*core.VolumeAttachment, error) {

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

	if len(volumes[0].Attachments) > 0 && !force {
		return nil, goof.New("Volume already attached to another host")
	} else if len(volumes[0].Attachments) > 0 && force {
		if err := d.DetachVolume(false, volumeID, "", true); err != nil {
			return nil, err
		}
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

func (d *driver) getLunMaps(initiatorName, volumeID string) (xtio.Refs, error) {
	if initiatorName == "" {
		return nil, goof.New("Missing initiatorName")
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

	if d.multipath() {
		_, _ = exec.Command("/sbin/multipath", "-f",
			fmt.Sprintf("3%s", volumes[0].NetworkName)).Output()
	}

	if err := d.updateInitiatorsSig(); err != nil {
		return err
	}

	mapInitiatorNamesByID := make(map[string]string)
	for _, initiator := range d.initiatorsByName {
		mapInitiatorNamesByID[strconv.Itoa(initiator.Index)] = initiator.Name
	}

	for _, attachment := range volumes[0].Attachments {
		var initiatorName string
		var ok bool
		if initiatorName, ok = mapInitiatorNamesByID[attachment.InstanceID]; !ok {
			continue
		}
		lunMaps, err := d.getLunMaps(
			initiatorName, attachment.VolumeID)
		if err != nil {
			return err
		}

		if len(lunMaps) == 0 {
			continue
		}

		index := getIndex(lunMaps[0].Href)
		if err = d.client.DeleteLunMap(index, ""); err != nil {
			return goof.WithFieldE("index", index, "error deleting lun map", err)
		}
	}

	return nil
}

func (d *driver) CopySnapshot(
	runAsync bool,
	volumeID, snapshotID, snapshotName,
	destinationSnapshotName, destinationRegion string) (*core.Snapshot, error) {
	return nil, errors.ErrNotImplemented
}

func (d *driver) GetDeviceNextAvailable() (string, error) {
	return "", errors.ErrNotImplemented
}

func (d *driver) endpoint() string {
	return d.r.Config.GetString("xtremio.endpoint")
}

func (d *driver) insecure() bool {
	return d.r.Config.GetBool("xtremio.insecure")
}

func (d *driver) userName() string {
	return d.r.Config.GetString("xtremio.userName")
}

func (d *driver) password() string {
	return d.r.Config.GetString("xtremio.password")
}

func (d *driver) deviceMapper() bool {
	return d.r.Config.GetBool("xtremio.deviceMapper")
}

func (d *driver) multipath() bool {
	return d.r.Config.GetBool("xtremio.multipath")
}

func (d *driver) remoteManagement() bool {
	return d.r.Config.GetBool("xtremio.remoteManagement")
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("XtremIO")
	r.Key(gofig.String, "", "", "", "xtremio.endpoint")
	r.Key(gofig.Bool, "", false, "", "xtremio.insecure")
	r.Key(gofig.String, "", "", "", "xtremio.userName")
	r.Key(gofig.String, "", "", "", "xtremio.password")
	r.Key(gofig.Bool, "", false, "", "xtremio.deviceMapper")
	r.Key(gofig.Bool, "", false, "", "xtremio.multipath")
	r.Key(gofig.Bool, "", false, "", "xtremio.remoteManagement")
	return r
}
