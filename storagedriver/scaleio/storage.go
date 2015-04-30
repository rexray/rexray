package scaleio

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/emccode/goscaleio"
	types "github.com/emccode/goscaleio/types/v1"
	"github.com/emccode/rexray/storagedriver"
)

var (
	providerName string
)

type Driver struct {
	Client           *goscaleio.Client
	System           *goscaleio.System
	ProtectionDomain *goscaleio.ProtectionDomain
	StoragePool      *goscaleio.StoragePool
	Sdc              *goscaleio.Sdc
}

var (
	ErrMissingVolumeID         = errors.New("Missing VolumeID")
	ErrMultipleVolumesReturned = errors.New("Multiple Volumes returned")
	ErrNoVolumesReturned       = errors.New("No Volumes returned")
	ErrLocalVolumeMaps         = errors.New("Getting local volume mounts")
)

func init() {
	providerName = "scaleio"
	storagedriver.Register("scaleio", Init)
}

func Init() (storagedriver.Driver, error) {

	var (
		username           = os.Getenv("GOSCALEIO_USERNAME")
		password           = os.Getenv("GOSCALEIO_PASSWORD")
		endpoint           = os.Getenv("GOSCALEIO_ENDPOINT")
		systemID           = os.Getenv("GOSCALEIO_SYSTEMID")
		protectionDomainID = os.Getenv("GOSCALEIO_PROTECTIONDOMAINID")
		storagePoolID      = os.Getenv("GOSCALEIO_STORAGEPOOLID")
	)

	client, err := goscaleio.NewClient()
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	_, err = client.Authenticate(&goscaleio.ConfigConnect{endpoint, username, password})
	if err != nil {
		log.Fatalf("error authenticating: %v", err)
	}

	system, err := client.FindSystem(systemID, "")
	if err != nil {
		log.Fatalf("err: problem getting system: %v", err)
	}

	pd, err := system.FindProtectionDomain(protectionDomainID, "", "")
	if err != nil {
		log.Fatalf("err: problem getting protection domain: %v", err)
	}

	protectionDomain := goscaleio.NewProtectionDomain(client)
	protectionDomain.ProtectionDomain = pd

	sp, err := protectionDomain.FindStoragePool(storagePoolID, "", "")
	if err != nil {
		log.Fatalf("err: problem getting protection domain: %v", err)
	}

	storagePool := goscaleio.NewStoragePool(client)
	storagePool.StoragePool = sp

	sdcguid, err := goscaleio.GetSdcLocalGUID()
	if err != nil {
		log.Fatalf("Error getting local sdc guid: %s", err)
	}

	sdc, err := system.FindSdc("SdcGuid", strings.ToUpper(sdcguid))
	if err != nil {
		log.Fatalf("Error finding Sdc %s: %s", sdcguid, err)
	}

	driver := &Driver{
		Client:           client,
		System:           system,
		ProtectionDomain: protectionDomain,
		StoragePool:      storagePool,
		Sdc:              sdc,
	}

	if os.Getenv("REXRAY_DEBUG") == "true" {
		log.Println("Driver Initialized: " + providerName)
	}

	return driver, nil
}

func (driver *Driver) getInstance() (*goscaleio.Sdc, error) {
	return driver.Sdc, nil
}

func (driver *Driver) GetInstance() (interface{}, error) {

	server, err := driver.getInstance()
	if err != nil {
		return storagedriver.Instance{}, err
	}

	instance := &storagedriver.Instance{
		ProviderName: providerName,
		InstanceID:   server.Sdc.ID,
		Region:       "",
		Name:         server.Sdc.Name,
	}

	// log.Println("Got Instance: " + fmt.Sprintf("%+v", instance))
	return instance, nil
}

func (driver *Driver) getBlockDevices() ([]*goscaleio.SdcMappedVolume, error) {
	volumeMaps, err := goscaleio.GetLocalVolumeMap()
	if err != nil {
		return []*goscaleio.SdcMappedVolume{}, ErrLocalVolumeMaps
	}
	return volumeMaps, nil
}

func (driver *Driver) GetBlockDeviceMapping() (interface{}, error) {
	blockDevices, err := driver.getBlockDevices()
	if err != nil {
		return nil, err
	}

	var BlockDevices []*storagedriver.BlockDevice
	for _, blockDevice := range blockDevices {
		sdBlockDevice := &storagedriver.BlockDevice{
			ProviderName: providerName,
			InstanceID:   driver.Sdc.Sdc.ID,
			Region:       blockDevice.MdmID,
			DeviceName:   blockDevice.SdcDevice,
			VolumeID:     blockDevice.VolumeID,
			Status:       "",
		}
		BlockDevices = append(BlockDevices, sdBlockDevice)
	}

	// log.Println("Got Block Device Mappings: " + fmt.Sprintf("%+v", BlockDevices))
	return BlockDevices, nil
}

func (driver *Driver) getVolume(volumeID, volumeName string) ([]*types.Volume, error) {
	volumes, err := driver.StoragePool.GetVolume("", volumeID, "", volumeName)
	if err != nil {
		log.Fatalf("error getting volumes: %v", err)
	}
	return volumes, nil
}

func (driver *Driver) GetVolume(volumeID, volumeName string) (interface{}, error) {

	sdcMappedVolumes, err := goscaleio.GetLocalVolumeMap()
	if err != nil {
		return []*storagedriver.Volume{}, err
	}

	sdcDeviceMap := make(map[string]*goscaleio.SdcMappedVolume)
	for _, sdcMappedVolume := range sdcMappedVolumes {
		sdcDeviceMap[sdcMappedVolume.VolumeID] = sdcMappedVolume
	}

	volumes, err := driver.getVolume(volumeID, volumeName)
	if err != nil {
		return []*storagedriver.Volume{}, err
	}

	var volumesSD []*storagedriver.Volume
	for _, volume := range volumes {
		var attachmentsSD []*storagedriver.VolumeAttachment
		for _, attachment := range volume.MappedSdcInfo {
			var deviceName string
			if attachment.SdcID == driver.Sdc.Sdc.ID {
				if _, exists := sdcDeviceMap[volume.ID]; exists {
					deviceName = sdcDeviceMap[volume.ID].SdcDevice
				}
			}
			attachmentSD := &storagedriver.VolumeAttachment{
				VolumeID:   volume.ID,
				InstanceID: attachment.SdcID,
				DeviceName: deviceName,
				Status:     "",
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		var IOPS int64
		if len(volume.MappedSdcInfo) > 0 {
			IOPS = int64(volume.MappedSdcInfo[0].LimitIops)
		}
		volumeSD := &storagedriver.Volume{
			Name:             volume.Name,
			VolumeID:         volume.ID,
			AvailabilityZone: driver.ProtectionDomain.ProtectionDomain.ID,
			Status:           "",
			VolumeType:       volume.StoragePoolID,
			IOPS:             IOPS,
			Size:             strconv.Itoa(volume.SizeInKb / 1024 / 1024),
			Attachments:      attachmentsSD,
		}
		volumesSD = append(volumesSD, volumeSD)
	}

	return volumesSD, nil
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

func (driver *Driver) GetSnapshot(volumeID, snapshotID, snapshotName string) (interface{}, error) {
	if snapshotID != "" {
		volumeID = snapshotID
	}

	volumes, err := driver.getVolume(volumeID, snapshotName)
	if err != nil {
		return []*storagedriver.Snapshot{}, err
	}

	var snapshotsInt []*storagedriver.Snapshot
	for _, volume := range volumes {
		if volume.AncestorVolumeID != "" {
			snapshotSD := &storagedriver.Snapshot{
				Name:        volume.Name,
				VolumeID:    volume.AncestorVolumeID,
				SnapshotID:  volume.ID,
				VolumeSize:  strconv.Itoa(volume.SizeInKb / 1024 / 1024),
				StartTime:   strconv.Itoa(volume.CreationTime),
				Description: "",
				Status:      "",
			}
			snapshotsInt = append(snapshotsInt, snapshotSD)
		}
	}

	// log.Println("Got Snapshots: " + fmt.Sprintf("%+v", snapshotsInt))
	return snapshotsInt, nil
}

func (driver *Driver) CreateSnapshot(notUsed bool, snapshotName, volumeID, description string) (interface{}, error) {

	snapshotDef := &types.SnapshotDef{
		VolumeID:     volumeID,
		SnapshotName: snapshotName,
	}

	var snapshotDefs []*types.SnapshotDef
	snapshotDefs = append(snapshotDefs, snapshotDef)
	snapshotVolumesParam := &types.SnapshotVolumesParam{
		SnapshotDefs: snapshotDefs,
	}

	snapshotVolumesResp, err := driver.System.CreateSnapshotConsistencyGroup(snapshotVolumesParam)
	if err != nil {
		return storagedriver.Snapshot{}, err
	}

	snapshot, err := driver.GetSnapshot("", snapshotVolumesResp.VolumeIDList[0], "")
	if err != nil {
		return storagedriver.Snapshot{}, err
	}

	// log.Println(fmt.Sprintf("Created Snapshot: %v", snapshot.([]*storagedriver.Snapshot)))
	return snapshot.([]*storagedriver.Snapshot), nil

}

func (driver *Driver) createVolume(notUsed bool, volumeName string, volumeID string, snapshotID string, volumeType string, IOPS int64, size int64, availabilityZone string) (*types.VolumeResp, error) {

	snapshot := &storagedriver.Snapshot{}
	if volumeID != "" {
		snapshotInt, err := driver.CreateSnapshot(true, volumeName, volumeID, "created for createVolume")
		if err != nil {
			return &types.VolumeResp{}, err
		}
		snapshot = snapshotInt.([]*storagedriver.Snapshot)[0]
		return &types.VolumeResp{ID: snapshot.SnapshotID}, nil
	}

	// if availabilityZone == "" {
	// 	availabilityZone = server.AvailabilityZone
	// }

	volumeParam := &types.VolumeParam{
		Name:           volumeName,
		VolumeSizeInKb: strconv.Itoa(int(size) * 1024 * 1024),
		VolumeType:     volumeType,
		// UseRmCache:     volumeusermcache,
	}

	volumeResp, err := driver.StoragePool.CreateVolume(volumeParam)
	if err != nil {
		return &types.VolumeResp{}, err
	}

	return volumeResp, nil
}

func (driver *Driver) CreateVolume(notUsed bool, volumeName string, volumeID string, snapshotID string, volumeType string, IOPS int64, size int64, availabilityZone string) (interface{}, error) {
	resp, err := driver.createVolume(notUsed, volumeName, volumeID, snapshotID, volumeType, IOPS, size, availabilityZone)
	if err != nil {
		return storagedriver.Volume{}, err
	}

	volumes, err := driver.GetVolume(resp.ID, "")
	if err != nil {
		return storagedriver.Volume{}, err
	}

	// log.Println(fmt.Sprintf("Created volume: %+v", volumes.([]*storagedriver.Volume)[0]))
	return volumes.([]*storagedriver.Volume)[0], nil

}

func (driver *Driver) RemoveVolume(volumeID string) error {
	if volumeID == "" {
		return ErrMissingVolumeID
	}

	volumes, err := driver.getVolume(volumeID, "")
	if err != nil {
		return err
	}

	targetVolume := goscaleio.NewVolume(driver.Client)
	targetVolume.Volume = volumes[0]

	err = targetVolume.RemoveVolume("ONLY_ME")
	if err != nil {
		return err
	}

	log.Println("Deleted Volume: " + volumeID)
	return nil
}

func (driver *Driver) RemoveSnapshot(snapshotID string) error {
	err := driver.RemoveVolume(snapshotID)
	if err != nil {
		return err
	}

	return nil
}

func (driver *Driver) GetDeviceNextAvailable() (string, error) {
	return "", nil
}

func (driver *Driver) AttachVolume(runAsync bool, volumeID, instanceID string) (interface{}, error) {
	if volumeID == "" {
		return storagedriver.VolumeAttachment{}, ErrMissingVolumeID
	}

	mapVolumeSdcParam := &types.MapVolumeSdcParam{
		SdcID: driver.Sdc.Sdc.ID,
		AllowMultipleMappings: "false",
		AllSdcs:               "",
	}

	volumes, err := driver.getVolume(volumeID, "")
	if err != nil {
		return storagedriver.VolumeAttachment{}, err
	}

	if len(volumes) == 0 {
		return storagedriver.VolumeAttachment{}, ErrNoVolumesReturned
	}

	targetVolume := goscaleio.NewVolume(driver.Client)
	targetVolume.Volume = volumes[0]

	err = targetVolume.MapVolumeSdc(mapVolumeSdcParam)
	if err != nil {
		return storagedriver.VolumeAttachment{}, err
	}

	_, err = waitMount(volumes[0].ID)
	if err != nil {
		return storagedriver.VolumeAttachment{}, err
	}

	volumeAttachment, err := driver.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return storagedriver.VolumeAttachment{}, err
	}

	// log.Println(fmt.Sprintf("Attached volume %s to instance %s", volumeID, instanceID))
	return volumeAttachment, nil
}

func (driver *Driver) DetachVolume(runAsync bool, volumeID string, blank string) error {
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

	targetVolume := goscaleio.NewVolume(driver.Client)
	targetVolume.Volume = volumes[0]

	unmapVolumeSdcParam := &types.UnmapVolumeSdcParam{
		SdcID:                driver.Sdc.Sdc.ID,
		IgnoreScsiInitiators: "true",
		AllSdcs:              "",
	}

	// need to detect if unmounted first
	err = targetVolume.UnmapVolumeSdc(unmapVolumeSdcParam)
	if err != nil {
		return err
	}

	log.Println("Detached volume", volumeID)
	return nil
}

func (driver *Driver) CopySnapshot(runAsync bool, volumeID, snapshotID, snapshotName, destinationSnapshotName, destinationRegion string) (interface{}, error) {
	return nil, errors.New("This driver does not implement CopySnapshot")
}

func waitMount(volumeID string) (*goscaleio.SdcMappedVolume, error) {

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(10 * time.Second)
		timeout <- true
	}()

	successCh := make(chan *goscaleio.SdcMappedVolume, 1)
	errorCh := make(chan error, 1)
	go func(volumeID string) {
		log.Println("ScaleIO: waiting for volume mount")
		for {
			sdcMappedVolumes, err := goscaleio.GetLocalVolumeMap()
			if err != nil {
				errorCh <- fmt.Errorf("ScaleIO: problem getting local volume mappings: %s", err)
				return
			}

			sdcMappedVolume := &goscaleio.SdcMappedVolume{}
			var foundVolume bool
			for _, sdcMappedVolume = range sdcMappedVolumes {
				if sdcMappedVolume.VolumeID == volumeID && sdcMappedVolume.SdcDevice != "" {
					foundVolume = true
					break
				}
			}

			if foundVolume {
				successCh <- sdcMappedVolume
				return
			}
			time.Sleep(100 * time.Millisecond)
		}

	}(volumeID)

	select {
	case sdcMappedVolume := <-successCh:
		log.Println(fmt.Sprintf("ScaleIO: got sdcMappedVolume %s at %s", sdcMappedVolume.VolumeID, sdcMappedVolume.SdcDevice))
		return sdcMappedVolume, nil
	case err := <-errorCh:
		return &goscaleio.SdcMappedVolume{}, err
	case <-timeout:
		return &goscaleio.SdcMappedVolume{}, fmt.Errorf("ScaleIO: timed out waiting for mount")
	}

}
