package scaleio

import (
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/goscaleio"
	types "github.com/emccode/goscaleio/types/v1"

	"github.com/emccode/rexray/core"
)

const providerName = "ScaleIO"
const cc = 31

// The ScaleIO storage driver.
type driver struct {
	client           *goscaleio.Client
	system           *goscaleio.System
	protectionDomain *goscaleio.ProtectionDomain
	storagePool      *goscaleio.StoragePool
	sdc              *goscaleio.Sdc
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

	fields := eff(map[string]interface{}{
		"moduleName": d.r.Context,
		"endpoint":   d.endpoint(),
		"insecure":   d.insecure(),
		"useCerts":   d.useCerts(),
	})

	var err error

	if d.client, err = goscaleio.NewClientWithArgs(
		d.endpoint(),
		d.insecure(),
		d.useCerts()); err != nil {
		return goof.WithFieldsE(fields, "error constructing new client", err)
	}

	if _, err := d.client.Authenticate(
		&goscaleio.ConfigConnect{
			d.endpoint(),
			d.version(),
			d.userName(),
			d.password()}); err != nil {
		fields["userName"] = d.userName()
		if d.password() != "" {
			fields["password"] = "******"
		}
		if d.version() != "" {
			fields["version"] = d.version()
		}
		return goof.WithFieldsE(fields, "error authenticating", err)
	}

	if d.system, err = d.client.FindSystem(
		d.systemID(),
		d.systemName(), ""); err != nil {
		fields["systemId"] = d.systemID()
		fields["systemName"] = d.systemName()
		return goof.WithFieldsE(fields, "error finding system", err)
	}

	var pd *types.ProtectionDomain
	if pd, err = d.system.FindProtectionDomain(
		d.protectionDomainID(),
		d.protectionDomainName(), ""); err != nil {
		fields["domainId"] = d.protectionDomainID()
		fields["domainName"] = d.protectionDomainName()
		return goof.WithFieldsE(fields,
			"error finding protection domain", err)
	}
	d.protectionDomain = goscaleio.NewProtectionDomain(d.client)
	d.protectionDomain.ProtectionDomain = pd

	var sp *types.StoragePool
	if sp, err = d.protectionDomain.FindStoragePool(
		d.storagePoolID(),
		d.storagePoolName(), ""); err != nil {
		fields["storagePoolId"] = d.storagePoolID()
		fields["storagePoolName"] = d.storagePoolName()
		return goof.WithFieldsE(fields, "error finding storage pool", err)
	}
	d.storagePool = goscaleio.NewStoragePool(d.client)
	d.storagePool.StoragePool = sp

	var sdcGUID string
	if sdcGUID, err = goscaleio.GetSdcLocalGUID(); err != nil {
		return goof.WithFieldsE(fields, "error getting sdc local guid", err)
	}

	if d.sdc, err = d.system.FindSdc(
		"SdcGuid",
		strings.ToUpper(sdcGUID)); err != nil {
		fields["sdcGuid"] = sdcGUID
		return goof.WithFieldsE(fields, "error finding sdc", err)
	}

	log.WithFields(fields).Info("storage driver initialized")

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
		"moduleName": d.r.Context,
		"provider":   providerName,
		"instance":   instance,
	}).Debug("got instance")
	return instance, nil
}

func (d *driver) getBlockDevices() ([]*goscaleio.SdcMappedVolume, error) {
	volumeMaps, err := goscaleio.GetLocalVolumeMap()
	if err != nil {
		return []*goscaleio.SdcMappedVolume{},
			goof.WithFieldsE(eff(map[string]interface{}{
				"moduleName": d.r.Context,
			}), "error getting local volume map", err)
	}
	return volumeMaps, nil
}

func (d *driver) GetVolumeMapping() ([]*core.BlockDevice, error) {
	blockDevices, err := d.getBlockDevices()
	if err != nil {
		return nil,
			goof.WithFieldsE(eff(map[string]interface{}{
				"moduleName": d.r.Context,
			}), "error getting block devices", err)
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
		"moduleName":   d.r.Context,
		"provider":     providerName,
		"blockDevices": BlockDevices,
	}).Debug("got block device mappings")
	return BlockDevices, nil
}

func shrink(n string) string {
	if len(n) > cc {
		return n[:cc]
	}
	return n
}

func (d *driver) getVolume(
	volumeID, volumeName string, getSnapshots bool) ([]*types.Volume, error) {

	volumeName = shrink(volumeName)

	volumes, err := d.client.GetVolume("", volumeID, "", volumeName, getSnapshots)
	if err != nil {
		return nil, err
	}
	return volumes, nil
}

func (d *driver) getStoragePoolIDs() (map[string]*types.StoragePool, error) {
	storagePools, err := d.client.GetStoragePool("")
	if err != nil {
		return nil, err
	}

	mapPoolID := make(map[string]*types.StoragePool)

	for _, pool := range storagePools {
		mapPoolID[pool.ID] = pool
	}
	return mapPoolID, nil
}

func (d *driver) getProtectionDomainIDs() (map[string]*types.ProtectionDomain, error) {
	protectionDomains, err := d.system.GetProtectionDomain("")
	if err != nil {
		return nil, err
	}

	mapProtectionDomainID := make(map[string]*types.ProtectionDomain)

	for _, protectionDomain := range protectionDomains {
		mapProtectionDomainID[protectionDomain.ID] = protectionDomain
	}
	return mapProtectionDomainID, nil
}

func (d *driver) GetVolume(
	volumeID, volumeName string) ([]*core.Volume, error) {

	sdcMappedVolumes, err := goscaleio.GetLocalVolumeMap()
	if err != nil {
		return []*core.Volume{}, err
	}

	mapStoragePoolName, err := d.getStoragePoolIDs()
	if err != nil {
		return nil, err
	}

	mapProtectionDomainName, err := d.getProtectionDomainIDs()
	if err != nil {
		return nil, err
	}

	getStoragePoolName := func(ID string) string {
		if pool, ok := mapStoragePoolName[ID]; ok {
			return pool.Name
		}
		return ""
	}

	getProtectionDomainName := func(poolID string) string {
		var ok bool
		var pool *types.StoragePool

		if pool, ok = mapStoragePoolName[poolID]; !ok {
			return ""
		}

		if protectionDomain, ok := mapProtectionDomainName[pool.ProtectionDomainID]; ok {
			return protectionDomain.Name
		}
		return ""
	}

	sdcDeviceMap := make(map[string]*goscaleio.SdcMappedVolume)
	for _, sdcMappedVolume := range sdcMappedVolumes {
		sdcDeviceMap[sdcMappedVolume.VolumeID] = sdcMappedVolume
	}

	volumes, err := d.getVolume(volumeID, volumeName, false)
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
			AvailabilityZone: getProtectionDomainName(volume.StoragePoolID),
			Status:           "",
			VolumeType:       getStoragePoolName(volume.StoragePoolID),
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
		"moduleName": d.r.Context,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return []*core.VolumeAttachment{},
			goof.WithFields(fields, "volumeId is required")
	}
	volume, err := d.GetVolume(volumeID, "")
	if err != nil {
		return []*core.VolumeAttachment{},
			goof.WithFieldsE(fields, "error getting volume", err)
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

	volumes, err := d.getVolume(volumeID, snapshotName, true)
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
		"moduleName": d.r.Context,
		"provider":   providerName,
		"snapshots":  snapshotsInt,
	}).Debug("got snapshots")
	return snapshotsInt, nil
}

func (d *driver) CreateSnapshot(
	notUsed bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {

	fields := eff(map[string]interface{}{
		"moduleName":   d.r.Context,
		"volumeID":     volumeID,
		"snapshotName": snapshotName,
		"description":  description,
	})

	if snapshotName == "" {
		return nil, goof.New("no snapshot name specified")
	}

	volumes, err := d.GetVolume("", snapshotName)
	if err != nil {
		return nil, err
	}

	if len(volumes) > 0 {
		return nil, goof.WithFieldsE(fields, "volume name already exists", err)
	}

	snapshotDef := &types.SnapshotDef{
		VolumeID:     volumeID,
		SnapshotName: snapshotName,
	}

	var snapshotDefs []*types.SnapshotDef
	snapshotDefs = append(snapshotDefs, snapshotDef)
	snapshotVolumesParam := &types.SnapshotVolumesParam{
		SnapshotDefs: snapshotDefs,
	}

	var snapshot []*core.Snapshot
	var snapshotVolumes *types.SnapshotVolumesResp

	if snapshotVolumes, err =
		d.system.CreateSnapshotConsistencyGroup(
			snapshotVolumesParam); err != nil {
		return nil, goof.WithFieldsE(fields, "failed to create snapshot", err)
	}

	if snapshot, err = d.GetSnapshot(
		"", snapshotVolumes.VolumeIDList[0], ""); err != nil {
		return nil, goof.WithFieldsE(fields, "error getting new snapshot", err)
	}

	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"provider":   providerName,
		"snapshot":   snapshot}).Debug("created snapshot")
	return snapshot, nil

}

func (d *driver) createVolume(
	notUsed bool,
	volumeName, volumeID, snapshotID, volumeType string,
	IOPS, size int64, availabilityZone string) (*types.VolumeResp, error) {

	volumeName = shrink(volumeName)

	fields := eff(map[string]interface{}{
		"moduleName":       d.r.Context,
		"volumeID":         volumeID,
		"volumeName":       volumeName,
		"snapshotID":       snapshotID,
		"volumeType":       volumeType,
		"IOPS":             IOPS,
		"size":             size,
		"availabilityZone": availabilityZone,
	})

	snapshot := &core.Snapshot{}
	if volumeID != "" {
		snapshotInt, err := d.CreateSnapshot(
			true, volumeName, volumeID, "created for createVolume")
		if err != nil {
			return nil, goof.WithFieldsE(fields, "error creating volume from snapshot", err)
		}
		snapshot = snapshotInt[0]
		return &types.VolumeResp{ID: snapshot.SnapshotID}, nil
	}

	volumeParam := &types.VolumeParam{
		Name:           volumeName,
		VolumeSizeInKb: strconv.Itoa(int(size) * 1024 * 1024),
		VolumeType:     d.thinOrThick(),
	}

	if volumeType == "" {
		volumeType = d.storagePool.StoragePool.Name
		fields["volumeType"] = volumeType
	}

	volumeResp, err := d.client.CreateVolume(volumeParam, volumeType)
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error creating volume", err)
	}

	return volumeResp, nil
}

func (d *driver) CreateVolume(
	notUsed bool,
	volumeName, volumeID, snapshotID, volumeType string,
	IOPS, size int64, availabilityZone string) (*core.Volume, error) {

	if volumeName == "" {
		return nil, goof.WithFields(eff(map[string]interface{}{
			"moduleName": d.r.Context}),
			"no volume name specified")
	}

	volumes, err := d.GetVolume("", volumeName)
	if err != nil {
		return nil, err
	}

	if len(volumes) > 0 {
		return nil, goof.WithFields(eff(map[string]interface{}{
			"moduleName": d.r.Context,
			"volumeName": volumeName}),
			"volume name already exists")
	}

	resp, err := d.createVolume(
		notUsed, volumeName, volumeID, snapshotID,
		volumeType, IOPS, size, availabilityZone)

	if err != nil {
		return nil, err
	}

	volumes, err = d.GetVolume(resp.ID, "")
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"provider":   providerName,
		"volume":     volumes[0],
	}).Debug("created volume")
	return volumes[0], nil

}

func (d *driver) RemoveVolume(volumeID string) error {

	fields := eff(map[string]interface{}{
		"moduleName": d.r.Context,
		"volumeId":   volumeID,
	})

	if volumeID == "" {
		return goof.WithFields(fields, "volumeId is required")
	}

	var err error
	var volumes []*types.Volume

	if volumes, err = d.getVolume(volumeID, "", false); err != nil {
		return goof.WithFieldsE(fields, "error getting volume", err)
	}

	targetVolume := goscaleio.NewVolume(d.client)
	targetVolume.Volume = volumes[0]

	if err = targetVolume.RemoveVolume("ONLY_ME"); err != nil {
		return goof.WithFieldsE(fields, "error removing volume", err)
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
	volumeID, instanceID string, force bool) ([]*core.VolumeAttachment, error) {

	fields := eff(map[string]interface{}{
		"moduleName": d.r.Context,
		"runAsync":   runAsync,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	})

	if volumeID == "" {
		return nil, goof.WithFields(fields, "volumeId is required")
	}

	if force {
		if err := d.DetachVolume(false, volumeID, "", true); err != nil {
			return nil, err
		}
	}

	mapVolumeSdcParam := &types.MapVolumeSdcParam{
		SdcID: d.sdc.Sdc.ID,
		AllowMultipleMappings: "false",
		AllSdcs:               "",
	}

	volumes, err := d.getVolume(volumeID, "", false)
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error getting volume", err)
	}

	if len(volumes) == 0 {
		return nil, goof.WithFields(fields, "no volumes returned")
	}

	targetVolume := goscaleio.NewVolume(d.client)
	targetVolume.Volume = volumes[0]

	err = targetVolume.MapVolumeSdc(mapVolumeSdcParam)
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error mapping volume sdc", err)
	}

	_, err = d.waitMount(volumes[0].ID)
	if err != nil {
		fields["volumeId"] = volumes[0].ID
		return nil, goof.WithFieldsE(
			fields, "error waiting on volume to mount", err)
	}

	volumeAttachment, err := d.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, goof.WithFieldsE(
			fields, "error getting volume attachments", err)
	}

	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"provider":   providerName,
		"volumeId":   volumeID,
		"instanceId": instanceID,
	}).Debug("attached volume to instance")
	return volumeAttachment, nil
}

func (d *driver) DetachVolume(
	runAsync bool, volumeID string, blank string, force bool) error {

	fields := eff(map[string]interface{}{
		"moduleName": d.r.Context,
		"runAsync":   runAsync,
		"volumeId":   volumeID,
		"blank":      blank,
	})

	if volumeID == "" {
		return goof.WithFields(fields, "volumeId is required")
	}

	volumes, err := d.getVolume(volumeID, "", false)
	if err != nil {
		return goof.WithFieldsE(fields, "error getting volume", err)
	}

	if len(volumes) == 0 {
		return goof.WithFields(fields, "no volumes returned")
	}

	targetVolume := goscaleio.NewVolume(d.client)
	targetVolume.Volume = volumes[0]

	unmapVolumeSdcParam := &types.UnmapVolumeSdcParam{
		SdcID:                "",
		IgnoreScsiInitiators: "true",
		AllSdcs:              "",
	}

	if force {
		unmapVolumeSdcParam.AllSdcs = "true"
	} else {
		unmapVolumeSdcParam.SdcID = d.sdc.Sdc.ID
	}

	_ = targetVolume.UnmapVolumeSdc(unmapVolumeSdcParam)

	log.WithFields(log.Fields{
		"moduleName": d.r.Context,
		"provider":   providerName,
		"volumeId":   volumeID}).Debug("detached volume")
	return nil
}

func (d *driver) CopySnapshot(
	runAsync bool,
	volumeID, snapshotID,
	snapshotName, destinationSnapshotName,
	destinationRegion string) (*core.Snapshot, error) {
	return nil, goof.New("This driver does not implement CopySnapshot")
}

func (d *driver) waitMount(volumeID string) (*goscaleio.SdcMappedVolume, error) {

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
				errorCh <- goof.WithFieldE(
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
			"moduleName": d.r.Context,
			"provider":   providerName,
			"volumeId":   sdcMappedVolume.VolumeID,
			"volume":     sdcMappedVolume.SdcDevice,
		}).Debug("got sdcMappedVolume")
		return sdcMappedVolume, nil
	case err := <-errorCh:
		return &goscaleio.SdcMappedVolume{}, err
	case <-timeout:
		return &goscaleio.SdcMappedVolume{}, goof.WithFields(
			ef(), "timed out waiting for mount")
	}

}

func (d *driver) endpoint() string {
	return d.r.Config.GetString("scaleio.endpoint")
}

func (d *driver) insecure() bool {
	return d.r.Config.GetBool("scaleio.insecure")
}

func (d *driver) useCerts() bool {
	return d.r.Config.GetBool("scaleio.useCerts")
}

func (d *driver) userID() string {
	return d.r.Config.GetString("scaleio.userID")
}

func (d *driver) userName() string {
	return d.r.Config.GetString("scaleio.userName")
}

func (d *driver) password() string {
	return d.r.Config.GetString("scaleio.password")
}

func (d *driver) systemID() string {
	return d.r.Config.GetString("scaleio.systemID")
}

func (d *driver) systemName() string {
	return d.r.Config.GetString("scaleio.systemName")
}

func (d *driver) protectionDomainID() string {
	return d.r.Config.GetString("scaleio.protectionDomainID")
}

func (d *driver) protectionDomainName() string {
	return d.r.Config.GetString("scaleio.protectionDomainName")
}

func (d *driver) storagePoolID() string {
	return d.r.Config.GetString("scaleio.storagePoolID")
}

func (d *driver) storagePoolName() string {
	return d.r.Config.GetString("scaleio.storagePoolName")
}

func (d *driver) thinOrThick() string {
	thinOrThick := d.r.Config.GetString("scaleio.thinOrThick")
	if thinOrThick == "" {
		return "ThinProvisioned"
	}
	return thinOrThick
}

func (d *driver) version() string {
	return d.r.Config.GetString("scaleio.version")
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("ScaleIO")
	r.Key(gofig.String, "", "", "", "scaleio.endpoint")
	r.Key(gofig.Bool, "", false, "", "scaleio.insecure")
	r.Key(gofig.Bool, "", false, "", "scaleio.useCerts")
	r.Key(gofig.String, "", "", "", "scaleio.userID")
	r.Key(gofig.String, "", "", "", "scaleio.userName")
	r.Key(gofig.String, "", "", "", "scaleio.password")
	r.Key(gofig.String, "", "", "", "scaleio.systemID")
	r.Key(gofig.String, "", "", "", "scaleio.systemName")
	r.Key(gofig.String, "", "", "", "scaleio.protectionDomainID")
	r.Key(gofig.String, "", "", "", "scaleio.protectionDomainName")
	r.Key(gofig.String, "", "", "", "scaleio.storagePoolID")
	r.Key(gofig.String, "", "", "", "scaleio.storagePoolName")
	r.Key(gofig.String, "", "", "", "scaleio.thinOrThick")
	r.Key(gofig.String, "", "", "", "scaleio.version")
	return r
}
