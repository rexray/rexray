package vmax

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	govmax "github.com/emccode/govmax/api/v1"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

const providerName = "VMAX"
const jobTimeout = 30

// The XtremIO storage driver.
type driver struct {
	client *govmax.SMIS
	// initiator        xtio.Initiator
	// volumesSig       string
	// lunMapsSig       string
	// initiatorsSig    string
	// volumesByNaa     map[string]xtio.Volume
	// initiatorsByName map[string]xtio.Initiator
	arrayID    string
	volPrefix  string
	instanceID string
	vmh        *govmax.VMHost
	r          *core.RexRay
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
		"userName": d.userName(),
		"smisHost": d.smisHost(),
		"smisPort": d.smisPort(),
		"insecure": d.insecure(),
		"sid":      d.sid(),
	})

	if d.password() == "" {
		fields["password"] = ""
	} else {
		fields["password"] = "******"
	}

	var err error
	if d.client, err = govmax.New(
		d.smisHost(),
		d.smisPort(),
		d.insecure(),
		d.userName(),
		d.password()); err != nil {
		return goof.WithFieldsE(fields,
			"error creating govmax client", err)
	}

	d.arrayID = d.sid()
	d.volPrefix = d.volumePrefix()

	vmFields := eff(map[string]interface{}{
		"userName": d.vmhUserName(),
		"smisHost": d.vmhHost(),
		"insecure": d.vmhInsecure(),
	})

	if d.vmh, err = govmax.NewVMHost(
		d.vmhInsecure(),
		d.vmhHost(),
		d.vmhUserName(),
		d.vmhPassword(),
	); err != nil {
		return goof.WithFieldsE(vmFields,
			"error retrieving VM host info", err)
	}

	d.instanceID = d.vmh.Vm.Reference().Value

	log.WithField("provider", providerName).Info("storage driver initialized")

	return nil
}

func (d *driver) Name() string {
	return providerName
}

func (d *driver) GetInstance() (*core.Instance, error) {
	instance := &core.Instance{
		ProviderName: providerName,
		InstanceID:   d.vmh.Vm.Reference().Value,
		Region:       d.arrayID,
		Name:         d.vmh.Vm.Reference().Value,
	}
	return instance, nil
}

func (d *driver) GetVolumeMapping() ([]*core.BlockDevice, error) {

	volumes, err := d.GetVolume("", "")
	if err != nil {
		return nil, err
	}

	var blockDevices []*core.BlockDevice
	for _, volume := range volumes {
		if len(volume.Attachments) == 0 {
			continue
		}

		sdBlockDevice := &core.BlockDevice{
			ProviderName: providerName,
			InstanceID:   d.instanceID,
			Region:       d.arrayID,
			DeviceName:   volume.Attachments[0].DeviceName,
			VolumeID:     volume.VolumeID,
			NetworkName:  volume.NetworkName,
			Status:       volume.Status,
		}

		blockDevices = append(blockDevices, sdBlockDevice)
	}

	return blockDevices, nil
}

func (d *driver) isValidVolume(volName string) bool {
	if d.volPrefix == "" {
		return true
	}
	return strings.HasPrefix(volName, d.volPrefix)
}

func (d *driver) prefixVolumeName(volName string) string {
	return fmt.Sprintf("%s%s", d.volPrefix, volName)
}

func (d *driver) unprefixVolumeName(volName string) string {
	return strings.TrimPrefix(volName, d.volPrefix)
}

func (d *driver) GetVolume(volumeID, volumeName string) ([]*core.Volume, error) {

	localDeviceMap, err := d.getLocalWWNDeviceByID()
	if err != nil {
		return nil, goof.WithError("error getting local devices", err)
	}

	var volumesResp *govmax.GetVolumesResp
	if volumeID != "" {
		volumesResp, err = d.client.GetVolumeByID(d.sid(), volumeID)
	} else if volumeName != "" {
		volumesResp, err = d.client.GetVolumeByName(d.sid(), d.prefixVolumeName(volumeName))
	} else {
		volumesResp, err = d.client.GetVolumes(d.sid())
	}
	if err != nil {
		return nil, goof.WithError("problem getting volumes", err)
	}

	var volumesSD []*core.Volume
	for _, entry := range volumesResp.Entries {
		if d.isValidVolume(entry.Content.I_ElementName) {
			deviceName, _ := localDeviceMap[entry.Content.I_EMCWWN]

			volumeSD := &core.Volume{
				Name:             d.unprefixVolumeName(entry.Content.I_ElementName),
				VolumeID:         entry.Content.I_DeviceID,
				NetworkName:      entry.Content.I_EMCWWN,
				Status:           strings.Join(entry.Content.I_StatusDescriptions, ","),
				VolumeType:       entry.Content.I_Caption,
				AvailabilityZone: d.arrayID,
				Size:             strconv.Itoa((entry.Content.I_BlockSize * entry.Content.I_NumberOfBlocks) / 1024 / 1024 / 1024),
			}
			if deviceName != "" {
				volumeSD.Attachments = append(volumeSD.Attachments, &core.VolumeAttachment{
					VolumeID:   entry.Content.I_DeviceID,
					InstanceID: d.instanceID,
					DeviceName: deviceName,
					Status:     strings.Join(entry.Content.I_StatusDescriptions, ","),
				})
			}
			volumesSD = append(volumesSD, volumeSD)
		}
	}

	return volumesSD, nil
}

func getID(id string) string {
	fields := strings.Split(id, ":")
	return fields[len(fields)-1]
}

func (d *driver) waitJob(instanceID string) (*govmax.GetJobStatusResp, error) {

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(jobTimeout * time.Second)
		timeout <- true
	}()

	successCh := make(chan *govmax.GetJobStatusResp, 1)
	errorCh := make(chan struct {
		err           error
		jobStatusResp *govmax.GetJobStatusResp
	}, 1)
	go func(instanceID string) {
		log.Println("waiting for job to complete")
		for {
			jobStatusResp, jobStatus, err := d.client.GetJobStatus(instanceID)
			if err != nil {
				errorCh <- struct {
					err           error
					jobStatusResp *govmax.GetJobStatusResp
				}{
					goof.WithError(
						"error getting job status", err),
					nil,
				}
			}

			switch {
			case jobStatus == "TERMINATED" || jobStatus == "KILLED" ||
				jobStatus == "EXCEPTION":
				errorCh <- struct {
					err           error
					jobStatusResp *govmax.GetJobStatusResp
				}{
					goof.Newf(
						"problem with job: %s", jobStatus),
					jobStatusResp,
				}
				return
			case jobStatus == "COMPLETED":
				successCh <- jobStatusResp
				return
			}

			time.Sleep(100 * time.Millisecond)
		}
	}(instanceID)

	select {
	case jobStatusResp := <-successCh:
		return jobStatusResp, nil
	case jobStatusRespErr := <-errorCh:
		return jobStatusRespErr.jobStatusResp, jobStatusRespErr.err
	case <-timeout:
		return nil, goof.New("timed out waiting for job")
	}

}

func (d *driver) volumeExists(volumeName string) (bool, error) {
	volumes, err := d.GetVolume("", volumeName)
	if err != nil {
		return false, err
	}

	if len(volumes) > 0 {
		return true, goof.New("volume name already exists")
	}
	return false, nil
}

func (d *driver) CreateVolume(
	runAsync bool,
	volumeName, volumeID, snapshotID, NUvolumeType string,
	NUIOPS, size int64, NUavailabilityZone string) (*core.Volume, error) {

	exists, err := d.volumeExists(volumeName)
	if err != nil && !exists {
		return nil, err
	} else if exists {
		return nil, err
	}

	PostVolRequest := &govmax.PostVolumesReq{
		PostVolumesRequestContent: &govmax.PostVolumesReqContent{
			AtType:             "http://schemas.emc.com/ecom/edaa/root/emc/Symm_StorageConfigurationService",
			ElementName:        d.prefixVolumeName(volumeName),
			ElementType:        "2",
			EMCNumberOfDevices: "1",
			Size:               strconv.Itoa(int(size * 1024 * 1024 * 1024)),
		},
	}
	queuedJob, _, err := d.client.PostVolumes(PostVolRequest, d.arrayID)
	if err != nil {
		return nil, goof.WithError("error creating volume", err)
	}

	if len(queuedJob.Entries) == 0 {
		return nil, goof.New("no jobs returned")
	}

	if !runAsync {
		jobStatusResp, err := d.waitJob(queuedJob.Entries[0].Content.I_Parameters.I_Job.E0_InstanceID)
		if err != nil {
			return nil, err
		}

		if len(jobStatusResp.Entries) == 0 {
			return nil, goof.New("no volume returned")
		}

		fields := strings.Split(jobStatusResp.Entries[0].Content.I_Description, "Output: DeviceIDs=")
		if len(fields) < 2 {
			return nil, goof.New("new volumeID not found")
		}

		volume, err := d.GetVolume(fields[1], "")
		if err != nil {
			return nil, err
		}

		if len(volume) == 0 {
			return nil, goof.New("no new volume returned")
		}
		return volume[0], nil
	}

	return nil, nil
}

func (d *driver) RemoveVolume(volumeID string) error {
	fields := eff(map[string]interface{}{
		"volumeID": volumeID,
	})

	deleteVolumeRequest := &govmax.DeleteVolReq{
		DeleteVolRequestContent: &govmax.DeleteVolReqContent{
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_StorageConfigurationService",
			DeleteVolRequestContentElement: &govmax.DeleteVolReqContentElement{
				AtType:                  "http://schemas.emc.com/ecom/edaa/root/emc/Symm_StorageVolume",
				DeviceID:                volumeID,
				CreationClassName:       "Symm_StorageVolume",
				SystemName:              "SYMMETRIX-+-" + d.arrayID,
				SystemCreationClassName: "Symm_StorageSystem",
			},
		},
	}
	queuedJob, err := d.client.PostDeleteVol(deleteVolumeRequest, d.arrayID)
	if err != nil {
		return goof.WithFieldsE(fields, "error deleteing volume", err)
	}

	if len(queuedJob.Entries) == 0 {
		return goof.New("no jobs returned")
	}

	_, err = d.waitJob(queuedJob.Entries[0].Content.I_Parameters.I_Job.E0_InstanceID)
	if err != nil {
		return err
	}

	log.Println("Deleted Volume: " + volumeID)
	return nil
}

func (d *driver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {
	return nil, goof.New("not implemented in driver")
}

func (d *driver) CreateSnapshot(
	notUsed bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {
	return nil, goof.New("not implemented in driver")
}

func (d *driver) RemoveSnapshot(snapshotID string) error {
	return goof.New("not implemented in driver")
}

func (d *driver) GetVolumeAttach(volumeID, instanceID string) ([]*core.VolumeAttachment, error) {
	if volumeID == "" {
		return nil, errors.ErrMissingVolumeID
	}
	volume, err := d.GetVolume(volumeID, "")
	if err != nil {
		return nil, err
	}

	if instanceID != "" {
		var attached bool
		for _, volumeAttachment := range volume[0].Attachments {
			if volumeAttachment.InstanceID == instanceID {
				return volume[0].Attachments, nil
			}
		}
		if !attached {
			return nil, nil
		}
	}
	return volume[0].Attachments, nil
}

func (d *driver) rescanScsiHosts() {
	hosts := "/sys/class/scsi_host/"
	if dirs, err := ioutil.ReadDir(hosts); err == nil {
		for _, f := range dirs {
			name := hosts + f.Name() + "/scan"
			data := []byte("- - -")
			ioutil.WriteFile(name, data, 0666)
		}
	}
	time.Sleep(1 * time.Second)
}

func (d *driver) multipath() bool {
	return false
}

func (d *driver) deviceMapper() bool {
	return false
}

func (d *driver) getLocalWWNDeviceByID() (map[string]string, error) {
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
			//32 for WWN
			naaName = naaName[len(naaName)-32:]
			devPath, _ := filepath.EvalSymlinks(fmt.Sprintf("%s/%s", diskIDPath, f.Name()))
			mapDiskByID[naaName] = devPath
		}
	}
	return mapDiskByID, nil
}

func (d *driver) attachVolumeToSG(runAsync bool, volumeID string) error {
	PostVol2SGRequest := &govmax.PostVolumesToSGReq{
		PostVolumesToSGRequestContent: &govmax.PostVolumesToSGReqContent{
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_ControllerconfigurationService",
			PostVolumesToSGRequestContentMG: &govmax.PostVolumesToSGReqContentMG{
				AtType:     "http://schemas.emc.com/ecom/edaa/root/emc/SE_DeviceMaskingGroup",
				InstanceID: "SYMMETRIX-+-" + d.arrayID + "-+-" + d.storageGroup(),
			},
			PostVolumesToSGRequestContentMember: []*govmax.PostVolumesToSGReqContentMember{
				&govmax.PostVolumesToSGReqContentMember{
					AtType:            "http://schemas.emc.com/ecom/edaa/root/emc/Symm_StorageVolume",
					CreationClassName: "Symm_StorageVolume",
					//Change DeviceID to existing Volume ID
					DeviceID:                volumeID,
					SystemCreationClassName: "Symm_StorageSystem",
					SystemName:              "SYMMETRIX-+-" + d.arrayID,
				},
			},
		},
	}

	queuedJob, err := d.client.PostVolumesToSG(PostVol2SGRequest, d.arrayID)
	if err != nil {
		return err
	}

	if !runAsync {
		if len(queuedJob.Entries) == 0 {
			return goof.New("no jobs returned")
		}

		jobResp, err := d.waitJob(
			queuedJob.Entries[0].Content.I_Parameters.I_Job.E0_InstanceID)
		if err != nil {
			if len(jobResp.Entries) > 0 {
				if !strings.Contains(jobResp.Entries[0].Content.I_ErrorDescription,
					"already exists in specified group") {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}

func (d *driver) AttachVolume(
	runAsync bool,
	volumeID, instanceID string, force bool) ([]*core.VolumeAttachment, error) {

	volumes, err := d.GetVolume(volumeID, "")
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, goof.New("volume not found")
	}

	if err := d.attachVolumeToSG(runAsync, volumeID); err != nil {
		return nil, goof.WithError("error adding volume to storage group", err)
	}

	hostSystem, err := d.vmh.Vm.HostSystem(d.vmh.Ctx)
	if err != nil {
		return nil, err
	}

	if err := d.vmh.RescanAllHba(hostSystem); err != nil {
		return nil, goof.WithError("error rescanning all hbas", err)
	}

	if err := d.vmh.AttachRDM(d.vmh.Vm, volumes[0].NetworkName); err != nil {
		return nil, goof.WithError("error attaching volume as RDM", err)
	}

	hosts, err := d.vmh.FindHosts(d.vmh.Vm)
	if err != nil {
		return nil, goof.WithError("error getting clustered hosts", err)
	}

	for _, host := range hosts {
		if hostSystem.ConfigManager().Reference() == host.ConfigManager().Reference() {
			continue
		}

		go d.vmh.RescanAllHba(host)
	}

	d.rescanScsiHosts()

	localDeviceMap, err := d.getLocalWWNDeviceByID()
	if err != nil {
		return nil, goof.WithError("error getting local devices", err)
	}

	var deviceName string
	var ok bool
	if deviceName, ok = localDeviceMap[volumes[0].NetworkName]; !ok {
		return nil, goof.New("local device not found for volume")
	}

	log.WithFields(log.Fields{"deviceName": deviceName}).Println("discovered device")

	volumeAttachment, err := d.GetVolumeAttach(volumeID, instanceID)
	if err != nil {
		return nil, err
	}

	return volumeAttachment, nil

}

func (d *driver) deleteScsiDevice(deviceID string) error {
	mapDiskByID, err := d.getLocalWWNDeviceByID()
	if err != nil {
		return goof.WithError("error getting devices", err)
	}

	var devicePath string
	var ok bool
	if devicePath, ok = mapDiskByID[deviceID]; !ok {
		return nil
	}

	deviceName := strings.Replace(devicePath, "/dev/", "", -1)

	device := "/sys/class/block/" + deviceName + "/device/delete"
	data := []byte("1")
	ioutil.WriteFile(device, data, 0666)
	return nil

}

func (d *driver) detachVolumeFromSG(runAsync bool, volumeID string) error {
	RemVol2SGRequest := &govmax.PostVolumesToSGReq{
		PostVolumesToSGRequestContent: &govmax.PostVolumesToSGReqContent{
			AtType: "http://schemas.emc.com/ecom/edaa/root/emc/Symm_ControllerconfigurationService",
			PostVolumesToSGRequestContentMG: &govmax.PostVolumesToSGReqContentMG{
				AtType:     "http://schemas.emc.com/ecom/edaa/root/emc/SE_DeviceMaskingGroup",
				InstanceID: "SYMMETRIX-+-" + d.arrayID + "-+-" + d.storageGroup(),
			},
			PostVolumesToSGRequestContentMember: []*govmax.PostVolumesToSGReqContentMember{
				&govmax.PostVolumesToSGReqContentMember{
					AtType:                  "http://schemas.emc.com/ecom/edaa/root/emc/Symm_StorageVolume",
					CreationClassName:       "Symm_StorageVolume",
					DeviceID:                volumeID,
					SystemCreationClassName: "Symm_StorageSystem",
					SystemName:              "SYMMETRIX-+-" + d.arrayID,
				},
			},
		},
	}
	queuedJob, err := d.client.RemoveVolumeFromSG(RemVol2SGRequest, d.arrayID)
	if err != nil {
		return err
	}

	if !runAsync {
		if len(queuedJob.Entries) == 0 {
			return goof.New("no jobs returned")
		}

		jobResp, err := d.waitJob(
			queuedJob.Entries[0].Content.I_Parameters.I_Job.E0_InstanceID)
		if err != nil {
			if len(jobResp.Entries) > 0 {
				if !strings.Contains(jobResp.Entries[0].Content.I_ErrorDescription,
					"The specified device was not found") {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}

func (d *driver) DetachVolume(runAsync bool, volumeID string, blank string, notused bool) error {

	volumes, err := d.GetVolume(volumeID, "")
	if err != nil {
		return err
	}

	if len(volumes) == 0 {
		return goof.New("volume not found")
	}

	if err := d.deleteScsiDevice(volumes[0].NetworkName); err != nil {
		return goof.WithError("error deleting scsi device from host", err)
	}

	if err := d.vmh.DetachRDM(d.vmh.Vm, volumes[0].NetworkName); err != nil {
		return goof.WithError("error removing RDM from vm", err)
	}

	if err := d.detachVolumeFromSG(runAsync, volumeID); err != nil {
		return goof.WithError("error detaching volume from storage group", err)
	}

	log.Println("Detached volume", volumeID)
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

func (d *driver) smisHost() string {
	return d.r.Config.GetString("vmax.smishost")
}

func (d *driver) smisPort() string {
	return d.r.Config.GetString("vmax.smisport")
}

func (d *driver) insecure() bool {
	return d.r.Config.GetBool("vmax.insecure")
}

func (d *driver) userName() string {
	return d.r.Config.GetString("vmax.userName")
}

func (d *driver) password() string {
	return d.r.Config.GetString("vmax.password")
}

func (d *driver) sid() string {
	return d.r.Config.GetString("vmax.sid")
}

func (d *driver) volumePrefix() string {
	return d.r.Config.GetString("vmax.volumePrefix")
}

func (d *driver) vmhInsecure() bool {
	return d.r.Config.GetBool("vmax.vmh.insecure")
}

func (d *driver) vmhHost() string {
	return d.r.Config.GetString("vmax.vmh.host")
}

func (d *driver) vmhUserName() string {
	return d.r.Config.GetString("vmax.vmh.userName")
}

func (d *driver) vmhPassword() string {
	return d.r.Config.GetString("vmax.vmh.password")
}

func (d *driver) storageGroup() string {
	return d.r.Config.GetString("vmax.storageGroup")
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("GOVMAX")
	r.Key(gofig.String, "", "", "", "vmax.smishost")
	r.Key(gofig.String, "", "", "", "vmax.smisport")
	r.Key(gofig.Bool, "", false, "", "vmax.insecure")
	r.Key(gofig.String, "", "", "", "vmax.userName")
	r.Key(gofig.String, "", "", "", "vmax.password")
	r.Key(gofig.String, "", "", "", "vmax.sid")
	r.Key(gofig.String, "", "", "", "vmax.volumePrefix")
	r.Key(gofig.String, "", "", "", "vmax.storageGroup")
	r.Key(gofig.Bool, "", false, "", "vmax.vmh.insecure")
	r.Key(gofig.String, "", "", "", "vmax.vmh.userName")
	r.Key(gofig.String, "", "", "", "vmax.vmh.password")
	r.Key(gofig.String, "", "", "", "vmax.vmh.host")

	return r
}
