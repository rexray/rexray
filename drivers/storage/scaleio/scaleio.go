package scaleio

import (
	log "github.com/Sirupsen/logrus"
	"strconv"
	"strings"
	"time"

	"github.com/emccode/goscaleio"
	types "github.com/emccode/goscaleio/types/v1"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

const providerName = "ScaleIO"

// The ScaleIO storage driver.
type driver struct {
	client           *goscaleio.Client
	system           *goscaleio.System
	protectionDomain *goscaleio.ProtectionDomain
	storagePool      *goscaleio.StoragePool
	sdc              *goscaleio.Sdc
	r                *core.RexRay
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
	return &driver{}
}

func (d *driver) Init(r *core.RexRay) error {
	d.r = r

	fields := eff(map[string]interface{}{
		"endpoint": d.r.Config.ScaleIOEndpoint,
		"insecure": d.r.Config.ScaleIOInsecure,
		"useCerts": d.r.Config.ScaleIOUseCerts,
	})

	var err error

	if d.client, err = goscaleio.NewClientWithArgs(
		d.r.Config.ScaleIOEndpoint,
		d.r.Config.ScaleIOInsecure,
		d.r.Config.ScaleIOUseCerts); err != nil {
		return errors.WithFieldsE(fields, "error constructing new client", err)
	}

	if _, err := d.client.Authenticate(
		&goscaleio.ConfigConnect{
			d.r.Config.ScaleIOEndpoint,
			d.r.Config.ScaleIOUserName,
			d.r.Config.ScaleIoPassword}); err != nil {
		fields["userName"] = d.r.Config.ScaleIOUserName
		if d.r.Config.ScaleIoPassword != "" {
			fields["password"] = "******"
		}
		return errors.WithFieldsE(fields, "error authenticating", err)
	}

	if d.system, err = d.client.FindSystem(
		d.r.Config.ScaleIOSystemID,
		d.r.Config.ScaleIOSystemName, ""); err != nil {
		fields["systemId"] = d.r.Config.ScaleIOSystemID
		fields["systemName"] = d.r.Config.ScaleIOSystemName
		return errors.WithFieldsE(fields, "error finding system", err)
	}

	var pd *types.ProtectionDomain
	if pd, err = d.system.FindProtectionDomain(
		d.r.Config.ScaleIOProtectionDomainID,
		d.r.Config.ScaleIOProtectionDomainName, ""); err != nil {
		fields["domainId"] = d.r.Config.ScaleIOProtectionDomainID
		fields["domainName"] = d.r.Config.ScaleIOProtectionDomainName
		return errors.WithFieldsE(fields,
			"error finding protection domain", err)
	}
	d.protectionDomain = goscaleio.NewProtectionDomain(d.client)
	d.protectionDomain.ProtectionDomain = pd

	var sp *types.StoragePool
	if sp, err = d.protectionDomain.FindStoragePool(
		d.r.Config.ScaleIOStoragePoolID,
		d.r.Config.ScaleIOStoragePoolName, ""); err != nil {
		fields["storagePoolId"] = d.r.Config.ScaleIOStoragePoolID
		fields["storagePoolName"] = d.r.Config.ScaleIOStoragePoolName
		return errors.WithFieldsE(fields, "error finding storage pool", err)
	}
	d.storagePool = goscaleio.NewStoragePool(d.client)
	d.storagePool.StoragePool = sp

	var sdcGUID string
	if sdcGUID, err = goscaleio.GetSdcLocalGUID(); err != nil {
		return errors.WithFieldsE(fields, "error getting sdc local guid", err)
	}

	if d.sdc, err = d.system.FindSdc(
		"SdcGuid",
		strings.ToUpper(sdcGUID)); err != nil {
		fields["sdcGuid"] = sdcGUID
		return errors.WithFieldsE(fields, "error finding sdc", err)
	}

	log.WithField("provider", providerName).Debug("storage driver initialized")

	return nil
}

func (d *driver) Name() string {
	return providerName
}

func (d *driver) getInstance() (*goscaleio.Sdc, error) {
	return d.sdc, nil
}

func (d *driver) GetInstance() (*core.Instance, error) {

	server, err := d.getInstance()
	if err != nil {
		return &core.Instance{}, err
	}

	instance := &core.Instance{
		ProviderName: providerName,
		InstanceID:   server.Sdc.ID,
		Region:       "",
		Name:         server.Sdc.Name,
	}

	log.WithFields(log.Fields{
		"provider": providerName,
		"instance": instance,
	}).Debug("got instance")
	return instance, nil
}

func (d *driver) getBlockDevices() ([]*goscaleio.SdcMappedVolume, error) {
	volumeMaps, err := goscaleio.GetLocalVolumeMap()
	if err != nil {
		return []*goscaleio.SdcMappedVolume{},
			errors.WithFieldsE(ef(), "error getting local volume map", err)
	}
	return volumeMaps, nil
}

func (d *driver) GetVolumeMapping() ([]*core.BlockDevice, error) {
	blockDevices, err := d.getBlockDevices()
	if err != nil {
		return nil,
			errors.WithFieldsE(ef(), "error getting block devices", err)
	}

	var BlockDevices []*core.BlockDevice
	for _, blockDevice := range blockDevices {
		sdBlockDevice := &core.BlockDevice{
			ProviderName: providerName,
			InstanceID:   d.sdc.Sdc.ID,
			Region:       blockDevice.MdmID,
			DeviceName:   blockDevice.SdcDevice,
			VolumeID:     blockDevice.VolumeID,
			Status:       "",
		}
		BlockDevices = append(BlockDevices, sdBlockDevice)
	}

	log.WithFields(log.Fields{
		"provider":     providerName,
		"blockDevices": BlockDevices,
	}).Debug("got block device mappings")
	return BlockDevices, nil
}

func (d *driver) getVolume(
	volumeID, volumeName string) ([]*types.Volume, error) {
	volumes, err := d.storagePool.GetVolume("", volumeID, "", volumeName)
	if err != nil {
		return nil, err
	}
	return volumes, nil
}

func (d *driver) GetVolume(
	volumeID, volumeName string) ([]*core.Volume, error) {

	sdcMappedVolumes, err := goscaleio.GetLocalVolumeMap()
	if err != nil {
		return []*core.Volume{}, err
	}

	sdcDeviceMap := make(map[string]*goscaleio.SdcMappedVolume)
	for _, sdcMappedVolume := range sdcMappedVolumes {
		sdcDeviceMap[sdcMappedVolume.VolumeID] = sdcMappedVolume
	}

	volumes, err := d.getVolume(volumeID, volumeName)
	if err != nil {
		return []*core.Volume{}, err
	}

	var volumesSD []*core.Volume
	for _, volume := range volumes {
		var attachmentsSD []*core.VolumeAttachment
		for _, attachment := range volume.MappedSdcInfo {
			var deviceName string
			if attachment.SdcID == d.sdc.Sdc.ID {
				if _, exists := sdcDeviceMap[volume.ID]; exists {
					deviceName = sdcDeviceMap[volume.ID].SdcDevice
				}
			}
			attachmentSD := &core.VolumeAttachment{
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
		volumeSD := &core.Volume{
			Name:             volume.Name,
			VolumeID:         volume.ID,
			AvailabilityZone: d.protectionDomain.ProtectionDomain.ID,
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

func (d *driver) GetVolumeAttach(
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {

	fields := eff(map[string]interface{}{
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return []*core.VolumeAttachment{},
			errors.WithFields(fields, "volumeId is required")
	}
	volume, err := d.GetVolume(volumeID, "")
	if err != nil {
		return []*core.VolumeAttachment{},
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
			return []*core.VolumeAttachment{}, nil
		}
	}
	return volume[0].Attachments, nil
}

func (d *driver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {

	if snapshotID != "" {
		volumeID = snapshotID
	}

	volumes, err := d.getVolume(volumeID, snapshotName)
	if err != nil {
		return []*core.Snapshot{}, err
	}

	var snapshotsInt []*core.Snapshot
	for _, volume := range volumes {
		if volume.AncestorVolumeID != "" {
			snapshotSD := &core.Snapshot{
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
		"provider":  providerName,
		"snapshots": snapshotsInt,
	}).Debug("got snapshots")
	return snapshotsInt, nil
}

func (d *driver) CreateSnapshot(
	notUsed bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {

	snapshotDef := &types.SnapshotDef{
		VolumeID:     volumeID,
		SnapshotName: snapshotName,
	}

	var snapshotDefs []*types.SnapshotDef
	snapshotDefs = append(snapshotDefs, snapshotDef)
	snapshotVolumesParam := &types.SnapshotVolumesParam{
		SnapshotDefs: snapshotDefs,
	}

	var err error
	var snapshot []*core.Snapshot
	var snapshotVolumes *types.SnapshotVolumesResp

	if snapshotVolumes, err =
		d.system.CreateSnapshotConsistencyGroup(
			snapshotVolumesParam); err != nil {
		return nil, err
	}

	if snapshot, err = d.GetSnapshot(
		"", snapshotVolumes.VolumeIDList[0], ""); err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"provider": providerName,
		"snapshot": snapshot}).Debug("created snapshot")
	return snapshot, nil

}

func (d *driver) createVolume(
	notUsed bool,
	volumeName, volumeID, snapshotID, volumeType string,
	IOPS, size int64, availabilityZone string) (*types.VolumeResp, error) {

	snapshot := &core.Snapshot{}
	if volumeID != "" {
		snapshotInt, err := d.CreateSnapshot(
			true, volumeName, volumeID, "created for createVolume")
		if err != nil {
			return &types.VolumeResp{}, err
		}
		snapshot = snapshotInt[0]
		return &types.VolumeResp{ID: snapshot.SnapshotID}, nil
	}

	volumeParam := &types.VolumeParam{
		Name:           volumeName,
		VolumeSizeInKb: strconv.Itoa(int(size) * 1024 * 1024),
		VolumeType:     volumeType,
	}

	volumeResp, err := d.storagePool.CreateVolume(volumeParam)
	if err != nil {
		return &types.VolumeResp{}, err
	}

	return volumeResp, nil
}

func (d *driver) CreateVolume(
	notUsed bool,
	volumeName, volumeID, snapshotID, volumeType string,
	IOPS, size int64, availabilityZone string) (*core.Volume, error) {

	resp, err := d.createVolume(
		notUsed, volumeName, volumeID, snapshotID,
		volumeType, IOPS, size, availabilityZone)

	if err != nil {
		return nil, err
	}

	volumes, err := d.GetVolume(resp.ID, "")
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"provider": providerName,
		"volume":   volumes[0],
	}).Debug("created volume")
	return volumes[0], nil

}

func (d *driver) RemoveVolume(volumeID string) error {

	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})

	if volumeID == "" {
		return errors.WithFields(fields, "volumeId is required")
	}

	var err error
	var volumes []*types.Volume

	if volumes, err = d.getVolume(volumeID, ""); err != nil {
		return errors.WithFieldsE(fields, "error getting volume", err)
	}

	targetVolume := goscaleio.NewVolume(d.client)
	targetVolume.Volume = volumes[0]

	if err = targetVolume.RemoveVolume("ONLY_ME"); err != nil {
		return errors.WithFieldsE(fields, "error removing volume", err)
	}

	log.WithFields(fields).Debug("removed volume")
	return nil
}

func (d *driver) RemoveSnapshot(snapshotID string) error {
	err := d.RemoveVolume(snapshotID)
	if err != nil {
		return err
	}

	return nil
}

func (d *driver) GetDeviceNextAvailable() (string, error) {
	return "", nil
}

func (d *driver) AttachVolume(
	runAsync bool,
	volumeID, instanceID string) ([]*core.VolumeAttachment, error) {

	fields := eff(map[string]interface{}{
		"runAsync":   runAsync,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return nil, errors.WithFields(fields, "volumeId is required")
	}

	mapVolumeSdcParam := &types.MapVolumeSdcParam{
		SdcID: d.sdc.Sdc.ID,
		AllowMultipleMappings: "false",
		AllSdcs:               "",
	}

	volumes, err := d.getVolume(volumeID, "")
	if err != nil {
		return nil, errors.WithFieldsE(fields, "error getting volume", err)
	}

	if len(volumes) == 0 {
		return nil, errors.WithFields(fields, "no volumes returned")
	}

	targetVolume := goscaleio.NewVolume(d.client)
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

	volumeAttachment, err := d.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, errors.WithFieldsE(
			fields, "error getting volume attachments", err)
	}

	log.WithFields(log.Fields{
		"provider":   providerName,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	}).Debug("attached volume to instance")
	return volumeAttachment, nil
}

func (d *driver) DetachVolume(
	runAsync bool, volumeID string, blank string) error {

	fields := eff(map[string]interface{}{
		"runAsync": runAsync,
		"volumeId": volumeID,
		"blank":    blank,
	})

	if volumeID == "" {
		return errors.WithFields(fields, "volumeId is required")
	}

	volumes, err := d.getVolume(volumeID, "")
	if err != nil {
		return errors.WithFieldsE(fields, "error getting volume", err)
	}

	if len(volumes) == 0 {
		return errors.WithFields(fields, "no volumes returned")
	}

	targetVolume := goscaleio.NewVolume(d.client)
	targetVolume.Volume = volumes[0]

	unmapVolumeSdcParam := &types.UnmapVolumeSdcParam{
		SdcID:                d.sdc.Sdc.ID,
		IgnoreScsiInitiators: "true",
		AllSdcs:              "",
	}

	// need to detect if unmounted first
	err = targetVolume.UnmapVolumeSdc(unmapVolumeSdcParam)
	if err != nil {
		return errors.WithFieldsE(fields, "error unmapping volume sdc", err)
	}

	log.WithFields(log.Fields{
		"provider": providerName,
		"volumeId": volumeID}).Debug("detached volume")
	return nil
}

func (d *driver) CopySnapshot(
	runAsync bool,
	volumeID, snapshotID,
	snapshotName, destinationSnapshotName,
	destinationRegion string) (*core.Snapshot, error) {
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
		log.WithField("provider", providerName).Debug("waiting for volume mount")
		for {
			sdcMappedVolumes, err := goscaleio.GetLocalVolumeMap()
			if err != nil {
				errorCh <- errors.WithFieldE(
					"provider", providerName,
					"problem getting local volume mappings", err)
				return
			}

			sdcMappedVolume := &goscaleio.SdcMappedVolume{}
			var foundVolume bool
			for _, sdcMappedVolume = range sdcMappedVolumes {
				if sdcMappedVolume.VolumeID ==
					volumeID && sdcMappedVolume.SdcDevice != "" {
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
			"provider": providerName,
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
