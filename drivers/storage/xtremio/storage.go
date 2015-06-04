package scaleio

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	goxtremio "github.com/emccode/goxtremio"
	xmsv3 "github.com/emccode/goxtremio/api/v3"
	"github.com/emccode/rexray/drivers/storage"
)

var (
	providerName string
)

type Driver struct {
	Initiator       *xmsv3.Initiator
	VolumesSig      string
	LunMapsSig      string
	VolumesByNAA    map[string]*xmsv3.Volume
	UseDeviceMapper bool
	UseMultipath    bool
}

var (
	ErrMissingVolumeID         = errors.New("Missing VolumeID")
	ErrMultipleVolumesReturned = errors.New("Multiple Volumes returned")
	ErrNoVolumesReturned       = errors.New("No Volumes returned")
	ErrLocalVolumeMaps         = errors.New("Getting local volume mounts")
)

func init() {
	providerName = "XtremIO"
	storagedriver.Register("XtremIO", Init)
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

func Init() (storagedriver.Driver, error) {

	if !isXtremIOAttached() {
		return nil, fmt.Errorf("%s: %s", storagedriver.ErrDriverInstanceDiscovery, "Device not detected")
	}

	iqn, err := getIQN()
	if err != nil {
		return nil, err
	}

	initiator, err := goxtremio.GetInitiator("", iqn)
	if err != nil {
		return nil, err
	}

	useDeviceMapper, _ := strconv.ParseBool(os.Getenv("GOREXRAY_XTREMIO_DM"))
	useMultipath, _ := strconv.ParseBool(os.Getenv("GOREXRAY_XTREMIO_MULTIPATH"))

	driver := &Driver{
		Initiator:       initiator,
		UseDeviceMapper: useDeviceMapper,
		UseMultipath:    useMultipath,
	}
	driver.VolumesByNAA = make(map[string]*xmsv3.Volume)

	if os.Getenv("REXRAY_DEBUG") == "true" {
		log.Println("Storage Driver Initialized: " + providerName)
	}

	return driver, nil
}

func (driver *Driver) getVolumesSig() (string, error) {
	volumes, err := goxtremio.GetVolumes()
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
	lunMap, err := goxtremio.GetLunMaps()
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

	volumes, err := goxtremio.GetVolumes()
	if err != nil {
		return err
	}

	for _, volume := range volumes {
		index := getIndex(volume.Href)
		volumeDetail, err := goxtremio.GetVolume(index, "")
		if err != nil {
			return err
		}
		driver.VolumesByNAA[volumeDetail.NaaName] = volumeDetail
	}

	driver.VolumesSig = volumesSig
	driver.LunMapsSig = lunMapsSig
	return nil
}

func (driver *Driver) getInitiator() (*xmsv3.Initiator, error) {
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

func (driver *Driver) GetInstance() (interface{}, error) {

	initiator, err := driver.getInitiator()
	if err != nil {
		return storagedriver.Instance{}, err
	}

	instance := &storagedriver.Instance{
		ProviderName: providerName,
		InstanceID:   strconv.Itoa(initiator.Index),
		Region:       "",
		Name:         initiator.Name,
	}

	// log.Println("Got Instance: " + fmt.Sprintf("%+v", instance))
	return instance, nil
}

func (driver *Driver) GetVolumeMapping() (interface{}, error) {

	mapDiskByID, err := driver.getLocalDeviceByID()
	if err != nil {
		return nil, err
	}

	err = driver.updateVolumesSig()
	if err != nil {
		return nil, err
	}

	var BlockDevices []*storagedriver.BlockDevice
	for naa, blockDeviceName := range mapDiskByID {
		if volume, ok := driver.VolumesByNAA[naa]; ok {
			for _, level1 := range volume.LunMappingList {
				for _, level2 := range level1 {
					if reflect.TypeOf(level2).String() == "[]interface {}" && level2.([]interface{})[1].(string) == driver.Initiator.Name {
						//maybe bug in API where initiator group name changes don't update to volume
						sdBlockDevice := &storagedriver.BlockDevice{
							ProviderName: providerName,
							InstanceID:   driver.Initiator.Name,
							Region:       volume.SysID[0].(string),
							DeviceName:   blockDeviceName,
							VolumeID:     strconv.Itoa(volume.Index),
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

func (driver *Driver) getVolume(volumeID, volumeName string) ([]*xmsv3.Volume, error) {
	var volumes []*xmsv3.Volume
	if volumeID != "" || volumeName != "" {
		volume, err := goxtremio.GetVolume(volumeID, volumeName)
		if err != nil {
			return nil, err
		}
		volumes = append(volumes, volume)
	} else {
		allVolumes, err := goxtremio.GetVolumes()
		if err != nil {
			return nil, err
		}
		for _, volume := range allVolumes {
			hrefFields := strings.Split(volume.Href, "/")
			index, _ := strconv.Atoi(hrefFields[len(hrefFields)-1])
			volumes = append(volumes, &xmsv3.Volume{
				Name:  volume.Name,
				Index: index,
			})
		}
	}
	return volumes, nil
}

func (driver *Driver) GetVolume(volumeID, volumeName string) (interface{}, error) {

	volumes, err := driver.getVolume(volumeID, volumeName)
	if err != nil && err.Error() == "obj_not_found" {
		return []*storagedriver.Volume{}, nil
	} else if err != nil {
		return nil, err
	}

	localVolumeMappings, err := driver.GetVolumeMapping()
	if err != nil {
		return nil, err
	}

	blockDeviceMap := make(map[string]*storagedriver.BlockDevice)
	for _, volume := range localVolumeMappings.([]*storagedriver.BlockDevice) {
		blockDeviceMap[volume.VolumeID] = volume
	}

	var volumesSD []*storagedriver.Volume
	for _, volume := range volumes {
		var attachmentsSD []*storagedriver.VolumeAttachment
		if _, exists := blockDeviceMap[strconv.Itoa(volume.Index)]; exists {
			attachmentSD := &storagedriver.VolumeAttachment{
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
		volumeSD := &storagedriver.Volume{
			Name:             volume.Name,
			VolumeID:         strconv.Itoa(volume.Index),
			Size:             strconv.Itoa(volSize / 1024 / 1024),
			AvailabilityZone: az,
			Attachments:      attachmentsSD,
		}
		volumesSD = append(volumesSD, volumeSD)
	}

	return volumesSD, nil
}

func (driver *Driver) CreateVolume(notUsed bool, volumeName string, volumeID string, snapshotID string, NUvolumeType string, NUIOPS int64, size int64, NUavailabilityZone string) (interface{}, error) {

	var volumes interface{}
	if volumeID == "" && snapshotID == "" {
		req := &xmsv3.PostVolumesReq{
			VolName: volumeName,
			VolSize: int(size) * 1024 * 1024,
		}
		postVolumesResp, err := goxtremio.NewVolume(req)
		if err != nil {
			return nil, err
		}

		index := getIndex(postVolumesResp.Links[0].Href)
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
			volumeID = snapshots.([]*storagedriver.Snapshot)[0].VolumeID
		}
		snapshot, err := driver.CreateSnapshot(false, volumeName, volumeID, "")
		if err != nil {
			return nil, err
		}

		snapshotID := snapshot.([]*storagedriver.Snapshot)[0].SnapshotID
		snapshots, err := driver.getSnapshot(snapshotID, "")
		if err != nil {
			return nil, err
		}

		volumes, err = driver.GetVolume(strconv.Itoa(int(snapshots[0].VolID[2].(float64))), "")
		if err != nil {
			return nil, err
		}
	}

	return volumes.([]*storagedriver.Volume)[0], nil
}

func (driver *Driver) RemoveVolume(volumeID string) error {
	err := goxtremio.DeleteVolume(volumeID, "")
	if err != nil {
		return err
	}

	log.Println("Deleted Volume: " + volumeID)
	return nil
}

func (driver *Driver) getSnapshot(snapshotID, snapshotName string) ([]*xmsv3.Snapshot, error) {
	var snapshots []*xmsv3.Snapshot
	if snapshotID != "" || snapshotName != "" {
		snapshot, err := goxtremio.GetSnapshot(snapshotID, snapshotName)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	} else {
		allSnapshots, err := goxtremio.GetSnapshots()
		if err != nil {
			return nil, err
		}
		for _, snapshot := range allSnapshots {
			hrefFields := strings.Split(snapshot.Href, "/")
			index, _ := strconv.Atoi(hrefFields[len(hrefFields)-1])
			snapshots = append(snapshots, &xmsv3.Snapshot{
				Name:  snapshot.Name,
				Index: index,
			})
		}
	}
	return snapshots, nil
}

//GetSnapshot returns snapshots from a volume or a specific snapshot
func (driver *Driver) GetSnapshot(volumeID, snapshotID, snapshotName string) (interface{}, error) {
	var snapshotsInt []*storagedriver.Snapshot
	if volumeID != "" {
		volumes, err := driver.getVolume(volumeID, "")
		if err != nil {
			return []*storagedriver.Snapshot{}, err
		}

		for _, volume := range volumes {
			for _, destSnap := range volume.DestSnapList {
				snapshot, err := driver.getSnapshot(strconv.Itoa(int(destSnap.([]interface{})[2].(float64))), "")
				if err != nil {
					return []*storagedriver.Snapshot{}, err
				}

				volSize, _ := strconv.Atoi(volume.VolSize)
				snapshotSD := &storagedriver.Snapshot{
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
			return []*storagedriver.Snapshot{}, err
		}

		for _, snapshot := range snapshots {
			snapshot, err := goxtremio.GetSnapshot(strconv.Itoa(snapshot.Index), "")
			if err != nil {
				return nil, err
			}

			volume, err := driver.getVolume(strconv.Itoa(int(snapshot.AncestorVolID[2].(float64))), "")
			if err != nil {
				return nil, err
			}

			volSize, _ := strconv.Atoi(volume[0].VolSize)
			snapshotSD := &storagedriver.Snapshot{
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

func (driver *Driver) CreateSnapshot(notUsed bool, snapshotName, volumeID, description string) (interface{}, error) {
	volume, err := goxtremio.GetVolume(volumeID, "")
	if err != nil {
		return nil, err
	}

	var snapList []*xmsv3.SnapListItem
	snapList = append(snapList, &xmsv3.SnapListItem{
		AncestorVolID: volume.Name,
		SnapVolName:   snapshotName,
	})

	//getfolder of volume
	req := &xmsv3.PostSnapshotsReq{
		SnapList: snapList,
		FolderID: "/",
	}

	postSnapshotsResp, err := goxtremio.NewSnapshot(req)
	if err != nil {
		return nil, err
	}

	index := getIndex(postSnapshotsResp.Links[0].Href)
	snapshot, err := driver.GetSnapshot("", index, "")
	if err != nil {
		return nil, err
	}

	return snapshot.([]*storagedriver.Snapshot), nil
}

func (driver *Driver) RemoveSnapshot(snapshotID string) error {
	err := goxtremio.DeleteSnapshot(snapshotID, "")
	if err != nil {
		return err
	}

	return nil
}

func (driver *Driver) GetVolumeAttach(volumeID, instanceID string) (interface{}, error) {
	if volumeID == "" {
		return []*storagedriver.VolumeAttachment{}, ErrMissingVolumeID
	}
	volume, err := driver.GetVolume(volumeID, "")
	if err != nil {
		return []*storagedriver.VolumeAttachment{}, err
	}

	if instanceID != "" {
		var attached bool
		for _, volumeAttachment := range volume.([]*storagedriver.Volume)[0].Attachments {
			if volumeAttachment.InstanceID == instanceID {
				return volume.([]*storagedriver.Volume)[0].Attachments, nil
			}
		}
		if !attached {
			return []*storagedriver.VolumeAttachment{}, nil
		}
	}
	return volume.([]*storagedriver.Volume)[0].Attachments, nil
}

func (driver *Driver) waitAttach(volumeID string) (*storagedriver.BlockDevice, error) {

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(10 * time.Second)
		timeout <- true
	}()

	successCh := make(chan *storagedriver.BlockDevice, 1)
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

			for _, blockDevice := range blockDevices.([]*storagedriver.BlockDevice) {
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

func (driver *Driver) AttachVolume(runAsync bool, volumeID, instanceID string) (interface{}, error) {
	if volumeID == "" {
		return storagedriver.VolumeAttachment{}, ErrMissingVolumeID
	}

	// doing a lookup here for intiator name as IG name, so limited to IG name as initiator name to work for now

	initiatorGroup, err := goxtremio.GetInitiatorGroup("", driver.Initiator.Name)
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
	req := &xmsv3.PostLunMapsReq{
		VolID: volumeIDI,
		IgID:  initiatorGroupIDI,
	}

	_, err = goxtremio.NewLunMap(req)
	if err != nil {
		return nil, err
	}

	if !runAsync {
		_, err := driver.waitAttach(volumeID)
		if err != nil {
			return nil, err
		}
	}

	volumeAttachment, err := driver.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return storagedriver.VolumeAttachment{}, err
	}

	return volumeAttachment, nil

}

func getLunMaps(initiatorName, volumeID string) ([]*xmsv3.Ref, error) {
	if initiatorName == "" {
		return nil, errors.New("Missing initiatorName")
	}

	initiatorGroup, err := goxtremio.GetInitiatorGroup("", initiatorName)
	if err != nil {
		return nil, err
	}

	lunMaps, err := goxtremio.GetLunMaps()
	if err != nil {
		return nil, err
	}

	var refs []*xmsv3.Ref
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

	lunMaps, err := getLunMaps(driver.Initiator.Name, strconv.Itoa(volumes[0].Index))
	if err != nil {
		return err
	}
	index := getIndex(lunMaps[0].Href)
	if err = goxtremio.DeleteLunMap(index, ""); err != nil {
		return err
	}

	log.Println("Detached volume", volumeID)
	return nil
}

func (driver *Driver) CopySnapshot(runAsync bool, volumeID, snapshotID, snapshotName, destinationSnapshotName, destinationRegion string) (interface{}, error) {
	return nil, errors.New("This driver does not implement CopySnapshot")
}

func (driver *Driver) GetDeviceNextAvailable() (string, error) {
	return "", errors.New("This driver does not implment, since it cannot determine device name when attaching")
}
