package scaleio

import (
	log "github.com/Sirupsen/logrus"
	"strconv"
	"strings"
	"time"

	"github.com/emccode/goscaleio"
	types "github.com/emccode/goscaleio/types/v1"
	"github.com/emccode/rexray/config"
	"github.com/emccode/rexray/errors"
	"github.com/emccode/rexray/storage"
)

const ProviderName = "ScaleIO"

type Driver struct {
	Client           *goscaleio.Client
	System           *goscaleio.System
	ProtectionDomain *goscaleio.ProtectionDomain
	StoragePool      *goscaleio.StoragePool
	Sdc              *goscaleio.Sdc
	Config           *config.Config
}

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

func Init(cfg *config.Config) (storage.Driver, error) {

	fields := eff(map[string]interface{}{
		"endpoint": cfg.ScaleIoEndpoint,
		"insecure": cfg.ScaleIoInsecure,
		"useCerts": cfg.ScaleIoUseCerts,
	})

	client, err := goscaleio.NewClientWithArgs(
		cfg.ScaleIoEndpoint,
		cfg.ScaleIoInsecure,
		cfg.ScaleIoUseCerts)

	if err != nil {
		return nil, errors.WithFieldsE(fields,
			"error constructing new client", err)
	}

	_, err = client.Authenticate(&goscaleio.ConfigConnect{
		cfg.ScaleIoEndpoint, cfg.ScaleIoUserName, cfg.ScaleIoPassword})
	if err != nil {
		fields["userName"] = cfg.ScaleIoUserName
		if cfg.ScaleIoPassword != "" {
			fields["password"] = "******"
		}
		return nil, errors.WithFieldsE(fields,
			"error authenticating", err)
	}

	system, err := client.FindSystem(
		cfg.ScaleIoSystemId, cfg.ScaleIoSystemName, "")
	if err != nil {
		fields["systemId"] = cfg.ScaleIoSystemId
		fields["systemName"] = cfg.ScaleIoSystemName
		return nil, errors.WithFieldsE(fields,
			"error finding system", err)
	}

	pd, err := system.FindProtectionDomain(
		cfg.ScaleIoProtectionDomainId, cfg.ScaleIoProtectionDomainName, "")
	if err != nil {
		fields["domainId"] = cfg.ScaleIoProtectionDomainId
		fields["domainName"] = cfg.ScaleIoProtectionDomainName
		return nil, errors.WithFieldsE(fields,
			"error finding protection domain", err)
	}

	protectionDomain := goscaleio.NewProtectionDomain(client)
	protectionDomain.ProtectionDomain = pd

	sp, err := protectionDomain.FindStoragePool(
		cfg.ScaleIoStoragePoolId, cfg.ScaleIoStoragePoolName, "")
	if err != nil {
		fields["storagePoolId"] = cfg.ScaleIoStoragePoolId
		fields["storagePoolName"] = cfg.ScaleIoStoragePoolName
		return nil, errors.WithFieldsE(fields,
			"error finding storage pool", err)
	}

	storagePool := goscaleio.NewStoragePool(client)
	storagePool.StoragePool = sp

	sdcguid, err := goscaleio.GetSdcLocalGUID()
	if err != nil {
		return nil, errors.WithFieldsE(fields,
			"error getting sdc local guid", err)
	}

	sdc, err := system.FindSdc("SdcGuid", strings.ToUpper(sdcguid))
	if err != nil {
		fields["sdcGuid"] = sdcguid
		return nil, errors.WithFieldsE(fields,
			"error finding sdc", err)
	}

	driver := &Driver{
		Client:           client,
		System:           system,
		ProtectionDomain: protectionDomain,
		StoragePool:      storagePool,
		Sdc:              sdc,
	}

	log.WithField("providerName", ProviderName).Debug(
		"storage driver initialized")

	return driver, nil
}

func (driver *Driver) getInstance() (*goscaleio.Sdc, error) {
	return driver.Sdc, nil
}

func (driver *Driver) GetInstance() (*storage.Instance, error) {

	server, err := driver.getInstance()
	if err != nil {
		return &storage.Instance{}, err
	}

	instance := &storage.Instance{
		ProviderName: ProviderName,
		InstanceID:   server.Sdc.ID,
		Region:       "",
		Name:         server.Sdc.Name,
	}

	log.WithFields(log.Fields{
		"provider": ProviderName,
		"instance": instance,
	}).Debug("got instance")
	return instance, nil
}

func (driver *Driver) getBlockDevices() ([]*goscaleio.SdcMappedVolume, error) {
	volumeMaps, err := goscaleio.GetLocalVolumeMap()
	if err != nil {
		return []*goscaleio.SdcMappedVolume{},
			errors.WithFieldsE(ef(), "error getting local volume map", err)
	}
	return volumeMaps, nil
}

func (driver *Driver) GetVolumeMapping() ([]*storage.BlockDevice, error) {
	blockDevices, err := driver.getBlockDevices()
	if err != nil {
		return nil,
			errors.WithFieldsE(ef(), "error getting block devices", err)
	}

	var BlockDevices []*storage.BlockDevice
	for _, blockDevice := range blockDevices {
		sdBlockDevice := &storage.BlockDevice{
			ProviderName: ProviderName,
			InstanceID:   driver.Sdc.Sdc.ID,
			Region:       blockDevice.MdmID,
			DeviceName:   blockDevice.SdcDevice,
			VolumeID:     blockDevice.VolumeID,
			Status:       "",
		}
		BlockDevices = append(BlockDevices, sdBlockDevice)
	}

	log.WithFields(log.Fields{
		"provider":     ProviderName,
		"blockDevices": BlockDevices,
	}).Debug("got block device mappings")
	return BlockDevices, nil
}

func (driver *Driver) getVolume(volumeID, volumeName string) ([]*types.Volume, error) {
	volumes, err := driver.StoragePool.GetVolume("", volumeID, "", volumeName)
	if err != nil {
		return nil, err
	}
	return volumes, nil
}

func (driver *Driver) GetVolume(volumeID, volumeName string) ([]*storage.Volume, error) {

	sdcMappedVolumes, err := goscaleio.GetLocalVolumeMap()
	if err != nil {
		return []*storage.Volume{}, err
	}

	sdcDeviceMap := make(map[string]*goscaleio.SdcMappedVolume)
	for _, sdcMappedVolume := range sdcMappedVolumes {
		sdcDeviceMap[sdcMappedVolume.VolumeID] = sdcMappedVolume
	}

	volumes, err := driver.getVolume(volumeID, volumeName)
	if err != nil {
		return []*storage.Volume{}, err
	}

	var volumesSD []*storage.Volume
	for _, volume := range volumes {
		var attachmentsSD []*storage.VolumeAttachment
		for _, attachment := range volume.MappedSdcInfo {
			var deviceName string
			if attachment.SdcID == driver.Sdc.Sdc.ID {
				if _, exists := sdcDeviceMap[volume.ID]; exists {
					deviceName = sdcDeviceMap[volume.ID].SdcDevice
				}
			}
			attachmentSD := &storage.VolumeAttachment{
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
		volumeSD := &storage.Volume{
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

func (driver *Driver) GetVolumeAttach(volumeID, instanceID string) ([]*storage.VolumeAttachment, error) {

	fields := eff(map[string]interface{}{
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return []*storage.VolumeAttachment{},
			errors.WithFields(fields, "volumeId is required")
	}
	volume, err := driver.GetVolume(volumeID, "")
	if err != nil {
		return []*storage.VolumeAttachment{},
			errors.WithFieldsE(fields, "error getting volume", err)
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

func (driver *Driver) GetSnapshot(volumeID, snapshotID, snapshotName string) ([]*storage.Snapshot, error) {
	if snapshotID != "" {
		volumeID = snapshotID
	}

	volumes, err := driver.getVolume(volumeID, snapshotName)
	if err != nil {
		return []*storage.Snapshot{}, err
	}

	var snapshotsInt []*storage.Snapshot
	for _, volume := range volumes {
		if volume.AncestorVolumeID != "" {
			snapshotSD := &storage.Snapshot{
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

	log.WithFields(log.Fields{
		"provider":  ProviderName,
		"snapshots": snapshotsInt,
	}).Debug("got snapshots")
	return snapshotsInt, nil
}

func (driver *Driver) CreateSnapshot(notUsed bool, snapshotName, volumeID, description string) ([]*storage.Snapshot, error) {

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
		return nil, err
	}

	snapshot, err := driver.GetSnapshot("", snapshotVolumesResp.VolumeIDList[0], "")
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"provider": ProviderName,
		"snapshot": snapshot,
	}).Debug("created snapshot")
	return snapshot, nil

}

func (driver *Driver) createVolume(notUsed bool, volumeName string, volumeID string, snapshotID string, volumeType string, IOPS int64, size int64, availabilityZone string) (*types.VolumeResp, error) {

	snapshot := &storage.Snapshot{}
	if volumeID != "" {
		snapshotInt, err := driver.CreateSnapshot(true, volumeName, volumeID, "created for createVolume")
		if err != nil {
			return &types.VolumeResp{}, err
		}
		snapshot = snapshotInt[0]
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

func (driver *Driver) CreateVolume(notUsed bool, volumeName string, volumeID string, snapshotID string, volumeType string, IOPS int64, size int64, availabilityZone string) (*storage.Volume, error) {
	resp, err := driver.createVolume(notUsed, volumeName, volumeID, snapshotID, volumeType, IOPS, size, availabilityZone)
	if err != nil {
		return nil, err
	}

	volumes, err := driver.GetVolume(resp.ID, "")
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"provider": ProviderName,
		"volume":   volumes[0],
	}).Debug("created volume")
	return volumes[0], nil

}

func (driver *Driver) RemoveVolume(volumeID string) error {

	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})

	if volumeID == "" {
		return errors.WithFields(fields, "volumeId is required")
	}

	volumes, err := driver.getVolume(volumeID, "")
	if err != nil {
		return errors.WithFieldsE(fields, "error getting volume", err)
	}

	targetVolume := goscaleio.NewVolume(driver.Client)
	targetVolume.Volume = volumes[0]

	err = targetVolume.RemoveVolume("ONLY_ME")
	if err != nil {
		return errors.WithFieldsE(fields, "error removing volume", err)
	}

	log.WithFields(fields).Debug("removed volume")
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

func (driver *Driver) AttachVolume(runAsync bool, volumeID, instanceID string) ([]*storage.VolumeAttachment, error) {

	fields := eff(map[string]interface{}{
		"runAsync":   runAsync,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return nil, errors.WithFields(fields, "volumeId is required")
	}

	mapVolumeSdcParam := &types.MapVolumeSdcParam{
		SdcID: driver.Sdc.Sdc.ID,
		AllowMultipleMappings: "false",
		AllSdcs:               "",
	}

	volumes, err := driver.getVolume(volumeID, "")
	if err != nil {
		return nil, errors.WithFieldsE(fields, "error getting volume", err)
	}

	if len(volumes) == 0 {
		return nil, errors.WithFields(fields, "no volumes returned")
	}

	targetVolume := goscaleio.NewVolume(driver.Client)
	targetVolume.Volume = volumes[0]

	err = targetVolume.MapVolumeSdc(mapVolumeSdcParam)
	if err != nil {
		return nil, errors.WithFieldsE(fields, "error mapping volume sdc", err)
	}

	_, err = waitMount(volumes[0].ID)
	if err != nil {
		fields["volumeId"] = volumes[0].ID
		return nil, errors.WithFieldsE(
			fields, "error waiting on volume to mount", err)
	}

	volumeAttachment, err := driver.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, errors.WithFieldsE(
			fields, "error getting volume attachments", err)
	}

	log.WithFields(log.Fields{
		"provider":   ProviderName,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	}).Debug("attached volume to instance")
	return volumeAttachment, nil
}

func (driver *Driver) DetachVolume(runAsync bool, volumeID string, blank string) error {

	fields := eff(map[string]interface{}{
		"runAsync": runAsync,
		"volumeId": volumeID,
		"blank":    blank,
	})

	if volumeID == "" {
		return errors.WithFields(fields, "volumeId is required")
	}

	volumes, err := driver.getVolume(volumeID, "")
	if err != nil {
		return errors.WithFieldsE(fields, "error getting volume", err)
	}

	if len(volumes) == 0 {
		return errors.WithFields(fields, "no volumes returned")
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
		return errors.WithFieldsE(fields, "error unmapping volume sdc", err)
	}

	log.WithFields(log.Fields{
		"provider": ProviderName,
		"volumeId": volumeID}).Debug("detached volume")
	return nil
}

func (driver *Driver) CopySnapshot(runAsync bool, volumeID, snapshotID, snapshotName, destinationSnapshotName, destinationRegion string) (*storage.Snapshot, error) {
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
		log.WithField("provider", ProviderName).Debug("waiting for volume mount")
		for {
			sdcMappedVolumes, err := goscaleio.GetLocalVolumeMap()
			if err != nil {
				errorCh <- errors.WithFieldE(
					"provider", ProviderName,
					"problem getting local volume mappings", err)
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
		log.WithFields(log.Fields{
			"provider": ProviderName,
			"volumeId": sdcMappedVolume.VolumeID,
			"volume":   sdcMappedVolume.SdcDevice,
		}).Debug("got sdcMappedVolume")
		return sdcMappedVolume, nil
	case err := <-errorCh:
		return &goscaleio.SdcMappedVolume{}, err
	case <-timeout:
		return &goscaleio.SdcMappedVolume{}, errors.WithFields(
			ef(), "timed out waiting for mount")
	}

}
