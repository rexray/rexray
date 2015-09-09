package scaleio

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
	"github.com/emccode/rexray/config"
	"github.com/emccode/rexray/errors"
	"github.com/emccode/rexray/storage"
)

const ProviderName = "XtremIO"

type Driver struct {
	Client          *xtio.Client
	Initiator       xtio.Initiator
	VolumesSig      string
	LunMapsSig      string
	VolumesByNAA    map[string]xtio.Volume
	UseDeviceMapper bool
	UseMultipath    bool
	Config          *config.Config
}

var (
	ErrMissingVolumeID         = errors.New("Missing VolumeID")
	ErrMultipleVolumesReturned = errors.New("Multiple Volumes returned")
	ErrNoVolumesReturned       = errors.New("No Volumes returned")
	ErrLocalVolumeMaps         = errors.New("Getting local volume mounts")
)

func ef() errors.Fields {
	return errors.Fields{
		"provider": ProviderName,
	}
}

func eff(fields errors.Fields) map[string]interface{} {
	errFields := map[string]interface{}{
		"provider": ProviderName,
	}
	if fields != nil {
		for k, v := range fields {
			errFields[k] = v
		}
	}
	return errFields
}

func init() {
	storage.Register(ProviderName, Init)
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

func Init(cfg *config.Config) (storage.Driver, error) {

	fields := eff(map[string]interface{}{
		"endpoint":         cfg.XtremIoEndpoint,
		"userName":         cfg.XtremIoUserName,
		"deviceMapper":     cfg.XtremIoDeviceMapper,
		"multipath":        cfg.XtremIoMultipath,
		"remoteManagement": cfg.XtremIoRemoteManagement,
		"insecure":         cfg.XtremIoInsecure,
	})

	if cfg.XtremIoPassword == "" {
		fields["password"] = ""
	} else {
		fields["password"] = "******"
	}

	if !isXtremIOAttached() && !cfg.XtremIoRemoteManagement {
		return nil, errors.WithFields(fields, "device not detected")
	}

	client, cErr := xtio.NewClientWithArgs(
		cfg.XtremIoEndpoint,
		cfg.XtremIoInsecure,
		cfg.XtremIoUserName,
		cfg.XtremIoPassword)

	if cErr != nil {
		return nil,
			errors.WithFieldsE(fields, "error creating xtremio client", cErr)
	}

	var iqn string
	var ini xtio.Initiator
	if !cfg.XtremIoRemoteManagement {
		var iqnErr error
		iqn, iqnErr = getIQN()
		if iqnErr != nil {
			return nil, iqnErr
		}

		var iniErr error
		ini, iniErr = client.GetInitiator("", iqn)
		if iniErr != nil {
			return nil, iniErr
		}
	}

	useDeviceMapper, _ := strconv.ParseBool(os.Getenv("REXRAY_XTREMIO_DM"))
	useMultipath, _ := strconv.ParseBool(os.Getenv("REXRAY_XTREMIO_MULTIPATH"))

	driver := &Driver{
		Client:          client,
		Initiator:       ini,
		UseDeviceMapper: useDeviceMapper,
		UseMultipath:    useMultipath,
		Config:          cfg,
		VolumesByNAA:    map[string]xtio.Volume{},
	}

	log.WithField("provider", ProviderName).Debug(
		"storage driver initialized")

	return driver, nil
}

func (driver *Driver) getVolumesSig() (string, error) {
	volumes, err := driver.Client.GetVolumes()
	if err != nil {
		return "", err
	}

	var volumesNameHref sort.StringSlice
	for _, volume := range volumes {
		volumesNameHref = append(volumesNameHref, fmt.Sprintf("%s-%s", volume.Name, volume.Href))
	}
	volumesNameHref.Sort()
	return strings.Join(volumesNameHref, ";"), err
}

func (driver *Driver) getLunMapsSig() (string, error) {
	lunMap, err := driver.Client.GetLunMaps()
	if err != nil {
		return "", err
	}

	var lunMapsNameHref sort.StringSlice
	for _, lunMap := range lunMap {
		lunMapsNameHref = append(lunMapsNameHref, fmt.Sprintf("%s-%s", lunMap.Name, lunMap.Href))
	}
	lunMapsNameHref.Sort()
	return strings.Join(lunMapsNameHref, ";"), err
}

func (driver *Driver) isVolumesSigEqual() (bool, string, error) {
	volumesSig, err := driver.getVolumesSig()
	if err != nil {
		return false, "", err
	}

	if volumesSig == driver.VolumesSig {
		return true, volumesSig, nil
	} else {
		return false, volumesSig, nil
	}
}

func (driver *Driver) isLunMapsSigEqual() (bool, string, error) {
	lunMapsSig, err := driver.getLunMapsSig()
	if err != nil {
		return false, "", err
	}

	if lunMapsSig == driver.LunMapsSig {
		return true, lunMapsSig, nil
	} else {
		return false, lunMapsSig, nil
	}
}

func (driver *Driver) updateVolumesSig() error {
	oldVolumesSig := driver.VolumesSig
	checkSig, volumesSig, err := driver.isVolumesSigEqual()
	if err != nil {
		return err
	}

	oldLunMapsSig := driver.LunMapsSig
	checkMapsSig, lunMapsSig, err := driver.isLunMapsSigEqual()
	if err != nil {
		return err
	}

	if checkSig && checkMapsSig {
		return nil
	}

	if oldVolumesSig != "" || oldLunMapsSig != "" {
		log.Println("volumeSig or volumeMapsSig updated")
	}

	volumes, err := driver.Client.GetVolumes()
	if err != nil {
		return err
	}

	for _, volume := range volumes {
		index := getIndex(volume.Href)
		volumeDetail, err := driver.Client.GetVolume(index, "")
		if err != nil {
			return err
		}
		driver.VolumesByNAA[volumeDetail.NaaName] = volumeDetail
	}

	driver.VolumesSig = volumesSig
	driver.LunMapsSig = lunMapsSig
	return nil
}

func (driver *Driver) getInitiator() (xtio.Initiator, error) {
	return driver.Initiator, nil
}

func (driver *Driver) getLocalDeviceByID() (map[string]string, error) {
	mapDiskByID := make(map[string]string)
	diskIDPath := "/dev/disk/by-id"
	files, err := ioutil.ReadDir(diskIDPath)
	if err != nil {
		return nil, err
	}

	var match1 *regexp.Regexp
	var match2 string

	if driver.UseDeviceMapper || driver.UseMultipath {
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

func (driver *Driver) GetInstance() (*storage.Instance, error) {

	initiator, err := driver.getInitiator()
	if err != nil {
		return &storage.Instance{}, err
	}

	instance := &storage.Instance{
		ProviderName: ProviderName,
		InstanceID:   strconv.Itoa(initiator.Index),
		Region:       "",
		Name:         initiator.Name,
	}

	// log.Println("Got Instance: " + fmt.Sprintf("%+v", instance))
	return instance, nil
}

func (driver *Driver) GetVolumeMapping() ([]*storage.BlockDevice, error) {

	mapDiskByID, err := driver.getLocalDeviceByID()
	if err != nil {
		return nil, err
	}

	err = driver.updateVolumesSig()
	if err != nil {
		return nil, err
	}

	var BlockDevices []*storage.BlockDevice
	for naa, blockDeviceName := range mapDiskByID {
		if volume, ok := driver.VolumesByNAA[naa]; ok {
			for _, level1 := range volume.LunMappingList {
				for _, level2 := range level1 {
					if reflect.TypeOf(level2).String() == "[]interface {}" && level2.([]interface{})[1].(string) == driver.Initiator.Name {
						//maybe bug in API where initiator group name changes don't update to volume
						sdBlockDevice := &storage.BlockDevice{
							ProviderName: ProviderName,
							InstanceID:   driver.Initiator.Name,
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

func (driver *Driver) getVolume(volumeID, volumeName string) ([]xtio.Volume, error) {
	var volumes []xtio.Volume
	if volumeID != "" || volumeName != "" {
		volume, err := driver.Client.GetVolume(volumeID, volumeName)
		if err != nil {
			return nil, err
		}
		volumes = append(volumes, volume)
	} else {
		allVolumes, err := driver.Client.GetVolumes()
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

func (driver *Driver) GetVolume(volumeID, volumeName string) ([]*storage.Volume, error) {

	volumes, err := driver.getVolume(volumeID, volumeName)
	if err != nil && err.Error() == "obj_not_found" {
		return []*storage.Volume{}, nil
	} else if err != nil {
		return nil, err
	}

	localVolumeMappings, err := driver.GetVolumeMapping()
	if err != nil {
		return nil, err
	}

	blockDeviceMap := make(map[string]*storage.BlockDevice)
	for _, volume := range localVolumeMappings {
		blockDeviceMap[volume.VolumeID] = volume
	}

	var volumesSD []*storage.Volume
	for _, volume := range volumes {
		var attachmentsSD []*storage.VolumeAttachment
		if _, exists := blockDeviceMap[strconv.Itoa(volume.Index)]; exists {
			attachmentSD := &storage.VolumeAttachment{
				VolumeID:   strconv.Itoa(volume.Index),
				InstanceID: strconv.Itoa(driver.Initiator.Index),
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
		volumeSD := &storage.Volume{
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

func (driver *Driver) CreateVolume(
	notUsed bool,
	volumeName, volumeID, snapshotID, NUvolumeType string,
	NUIOPS, size int64, NUavailabilityZone string) (*storage.Volume, error) {

	var volumes []*storage.Volume
	if volumeID == "" && snapshotID == "" {
		req := &xtio.NewVolumeOptions{
			VolName: volumeName,
			VolSize: int(size) * 1024 * 1024,
		}
		res, err := driver.Client.NewVolume(req)
		if err != nil {
			return nil, err
		}

		index := getIndex(res.Links[0].Href)
		volumes, err = driver.GetVolume(index, "")
		if err != nil {
			return nil, err
		}
	} else {
		if snapshotID != "" {
			snapshots, err := driver.GetSnapshot("", snapshotID, "")
			if err != nil {
				return nil, err
			}
			volumeID = snapshots[0].VolumeID
		}
		snapshot, err := driver.CreateSnapshot(false, volumeName, volumeID, "")
		if err != nil {
			return nil, err
		}

		snapshotID := snapshot[0].SnapshotID
		snapshots, err := driver.getSnapshot(snapshotID, "")
		if err != nil {
			return nil, err
		}

		volumes, err = driver.GetVolume(strconv.Itoa(int(snapshots[0].VolID[2].(float64))), "")
		if err != nil {
			return nil, err
		}
	}

	return volumes[0], nil
}

func (driver *Driver) RemoveVolume(volumeID string) error {
	err := driver.Client.DeleteVolume(volumeID, "")
	if err != nil {
		return err
	}

	log.Println("Deleted Volume: " + volumeID)
	return nil
}

func (driver *Driver) getSnapshot(snapshotID, snapshotName string) ([]xtio.Snapshot, error) {
	var snapshots []xtio.Snapshot
	if snapshotID != "" || snapshotName != "" {
		snapshot, err := driver.Client.GetSnapshot(snapshotID, snapshotName)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	} else {
		allSnapshots, err := driver.Client.GetSnapshots()
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
func (driver *Driver) GetSnapshot(volumeID, snapshotID, snapshotName string) ([]*storage.Snapshot, error) {
	var snapshotsInt []*storage.Snapshot
	if volumeID != "" {
		volumes, err := driver.getVolume(volumeID, "")
		if err != nil {
			return nil, err
		}

		for _, volume := range volumes {
			for _, destSnap := range volume.DestSnapList {
				snapshot, err := driver.getSnapshot(strconv.Itoa(int(destSnap.([]interface{})[2].(float64))), "")
				if err != nil {
					return nil, err
				}

				volSize, _ := strconv.Atoi(volume.VolSize)
				snapshotSD := &storage.Snapshot{
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
		snapshots, err := driver.getSnapshot(snapshotID, snapshotName)
		if err != nil {
			return nil, err
		}

		for _, snapshot := range snapshots {
			snapshot, err := driver.Client.GetSnapshot(
				strconv.Itoa(snapshot.Index), "")
			if err != nil {
				return nil, err
			}

			volume, err := driver.getVolume(strconv.Itoa(int(snapshot.AncestorVolID[2].(float64))), "")
			if err != nil {
				return nil, err
			}

			volSize, _ := strconv.Atoi(volume[0].VolSize)
			snapshotSD := &storage.Snapshot{
				Name:        snapshot.Name,
				VolumeID:    strconv.Itoa(int(snapshot.AncestorVolID[2].(float64))),
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

func (driver *Driver) CreateSnapshot(notUsed bool, snapshotName, volumeID, description string) ([]*storage.Snapshot, error) {
	volume, err := driver.Client.GetVolume(volumeID, "")
	if err != nil {
		return nil, err
	}

	//getfolder of volume
	req := &xtio.NewSnapshotOptions{
		SnapList: xtio.NewSnapListItems(volume.Name, snapshotName),
		FolderID: "/",
	}

	postSnapshotsResp, err := driver.Client.NewSnapshot(req)
	if err != nil {
		return nil, err
	}

	index := getIndex(postSnapshotsResp.Links[0].Href)
	snapshot, err := driver.GetSnapshot("", index, "")
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (driver *Driver) RemoveSnapshot(snapshotID string) error {
	err := driver.Client.DeleteSnapshot(snapshotID, "")
	if err != nil {
		return err
	}

	return nil
}

func (driver *Driver) GetVolumeAttach(volumeID, instanceID string) ([]*storage.VolumeAttachment, error) {
	if volumeID == "" {
		return []*storage.VolumeAttachment{}, ErrMissingVolumeID
	}
	volume, err := driver.GetVolume(volumeID, "")
	if err != nil {
		return []*storage.VolumeAttachment{}, err
	}

	if instanceID != "" {
		var attached bool
		for _, volumeAttachment := range volume[0].Attachments {
			if volumeAttachment.InstanceID == instanceID {
				return volume[0].Attachments, nil
			}
		}
		if !attached {
			return []*storage.VolumeAttachment{}, nil
		}
	}
	return volume[0].Attachments, nil
}

func (driver *Driver) waitAttach(volumeID string) (*storage.BlockDevice, error) {

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(10 * time.Second)
		timeout <- true
	}()

	successCh := make(chan *storage.BlockDevice, 1)
	errorCh := make(chan error, 1)
	go func(volumeID string) {
		log.Println("XtremIO: waiting for volume attach")
		for {
			if driver.UseMultipath {
				_, err := exec.Command("/sbin/multipath").Output()
				if err != nil {
					errorCh <- fmt.Errorf("Error refreshing multipath: %s", err)
				}
			}

			blockDevices, err := driver.GetVolumeMapping()
			if err != nil {
				errorCh <- fmt.Errorf("XtremIO: problem getting local block devices: %s", err)
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
		log.Println(fmt.Sprintf("XtremIO: got attachedVolume %s at %s", blockDevice.VolumeID, blockDevice.DeviceName))
		return blockDevice, nil
	case err := <-errorCh:
		return nil, err
	case <-timeout:
		return nil, fmt.Errorf("XtremIO: timed out waiting for mount")
	}

}

func (driver *Driver) AttachVolume(runAsync bool, volumeID, instanceID string) ([]*storage.VolumeAttachment, error) {
	if volumeID == "" {
		return nil, ErrMissingVolumeID
	}

	// doing a lookup here for intiator name as IG name, so limited to IG name as initiator name to work for now
	//1) need to consider when instanceid is blank,need to lookup ig from instanceid

	// if instanceID == "" {
	initiatorGroup, err := driver.Client.GetInitiatorGroup("", driver.Initiator.Name)
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

	_, err = driver.Client.NewLunMap(req)
	if err != nil {
		return nil, err
	}

	if !runAsync {
		_, err := driver.waitAttach(volumeID)
		if err != nil {
			return nil, err
		}
	} else {
		_, _ = driver.GetVolumeMapping()
	}

	volumeAttachment, err := driver.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, err
	}

	return volumeAttachment, nil

}

func (driver *Driver) getLunMaps(initiatorName, volumeID string) (xtio.Refs, error) {
	if initiatorName == "" {
		return nil, errors.New("Missing initiatorName")
	}

	initiatorGroup, err := driver.Client.GetInitiatorGroup("", initiatorName)
	if err != nil {
		return nil, err
	}

	lunMaps, err := driver.Client.GetLunMaps()
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

func (driver *Driver) DetachVolume(notUsed bool, volumeID string, blank string) error {
	if volumeID == "" {
		return ErrMissingVolumeID
	}

	volumes, err := driver.getVolume(volumeID, "")
	if err != nil {
		return err
	}

	if len(volumes) == 0 {
		return ErrNoVolumesReturned
	}

	if driver.UseMultipath {

		_, err := exec.Command("/sbin/multipath", "-f", fmt.Sprintf("3%s", volumes[0].NaaName)).Output()
		if err != nil {
			return fmt.Errorf("Error removing multipath: %s", err)
		}

	}

	lunMaps, err := driver.getLunMaps(
		driver.Initiator.Name, strconv.Itoa(volumes[0].Index))
	if err != nil {
		return err
	}

	if len(lunMaps) == 0 {
		return nil
	}

	index := getIndex(lunMaps[0].Href)
	if err = driver.Client.DeleteLunMap(index, ""); err != nil {
		return err
	}

	log.Println("Detached volume", volumeID)
	return nil
}

func (driver *Driver) CopySnapshot(runAsync bool, volumeID, snapshotID, snapshotName, destinationSnapshotName, destinationRegion string) (*storage.Snapshot, error) {
	return nil, errors.New("This driver does not implement CopySnapshot")
}

func (driver *Driver) GetDeviceNextAvailable() (string, error) {
	return "", errors.New("This driver does not implment, since it cannot determine device name when attaching")
}
