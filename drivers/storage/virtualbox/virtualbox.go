package virtualbox

import (
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	vbox "github.com/appropriate/go-virtualboxclient/virtualboxclient"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

const providerName = "virtualbox"

type driver struct {
	virtualbox *vbox.VirtualBox
	machine    *vbox.Machine
	r          *core.RexRay
	m          sync.Mutex
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
		"endpoint":             d.endpoint(),
		"userName":             d.userName(),
		"tls":                  d.tls(),
		"volumePath":           d.volumePath(),
		"localMachineNameOrId": d.localMachineNameOrId(),
	})

	if d.volumePath() == "" {
		return goof.New("missing volumePath")
	}

	if d.endpoint() == "" {
		return goof.New("missing endpoint")
	}

	if d.password() == "" {
		fields["password"] = ""
	} else {
		fields["password"] = "******"
	}

	d.virtualbox = vbox.New(d.userName(), d.password(),
		d.endpoint(), d.tls(), d.controllerName())

	if err := d.login(); err != nil {
		return goof.WithFieldsE(fields,
			"error logging in", err)
	}

	if m, err := d.findLocalMachine(d.localMachineNameOrId()); err != nil {
		goof.WithFieldsE(fields,
			"failed to find local machine", err)
	} else {
		d.machine = m
	}

	log.WithField("provider", providerName).Info("storage driver initialized")

	return nil
}

func (d *driver) checkSession() error {
	_, err := d.virtualbox.FindMachine(d.machine.ID)
	if err != nil {
		log.Debug("logging in again")
		d.login()
	}
	return nil
}

func (d *driver) login() error {
	if err := d.virtualbox.Logon(); err != nil {
		return err
	}
	return nil
}

func (d *driver) Name() string {
	return providerName
}

func (d *driver) GetInstance() (*core.Instance, error) {

	instance := &core.Instance{
		ProviderName: providerName,
		InstanceID:   d.machine.ID,
		Region:       "",
		Name:         d.machine.Name,
	}

	return instance, nil
}

func (d *driver) getShortDeviceID(f string) string {
	sid := strings.Split(f, "VBOX_HARDDISK_VB")
	if len(sid) < 1 {
		return ""
	}

	aid := strings.Split(sid[1], "-")
	if len(aid) < 1 {
		return ""
	}
	return aid[0]
}

func (d *driver) getLocalDeviceByID() (map[string]string, error) {
	mapDiskByID := make(map[string]string)
	diskIDPath := "/dev/disk/by-id"
	files, err := ioutil.ReadDir(diskIDPath)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if strings.Contains(f.Name(), "VBOX_HARDDISK_VB") {
			sid := d.getShortDeviceID(f.Name())
			if sid == "" {
				continue
			}
			devPath, _ := filepath.EvalSymlinks(fmt.Sprintf("%s/%s", diskIDPath, f.Name()))
			mapDiskByID[sid] = devPath
		}
	}
	return mapDiskByID, nil
}

func (d *driver) GetVolumeMapping() ([]*core.BlockDevice, error) {
	d.m.Lock()
	defer d.m.Unlock()
	d.checkSession()

	if err := d.machine.Refresh(); err != nil {
		return nil, err
	}
	defer d.machine.Release()

	mapDiskByID, err := d.getLocalDeviceByID()
	if err != nil {
		return nil, err
	}

	mas, err := d.machine.GetMediumAttachments()
	if err != nil {
		return nil, err
	}

	var blockDevices []*core.BlockDevice
	for _, ma := range mas {
		medium := d.virtualbox.NewMedium(ma.Medium)
		defer medium.Release()

		mid, err := medium.GetID()
		if err != nil {
			return nil, err
		}
		smid := strings.Split(mid, "-")
		if len(smid) == 0 {
			continue
		}

		location, err := medium.GetLocation()
		if err != nil {
			return nil, err
		}

		var bdn string
		var ok bool
		if bdn, ok = mapDiskByID[smid[0]]; !ok {
			continue
		}
		sdBlockDevice := &core.BlockDevice{
			ProviderName: providerName,
			InstanceID:   d.machine.ID,
			DeviceName:   bdn,
			VolumeID:     mid,
			Status:       location,
		}
		blockDevices = append(blockDevices, sdBlockDevice)

	}
	return blockDevices, nil

}

func (d *driver) GetVolume(volumeID, volumeName string) ([]*core.Volume, error) {
	d.m.Lock()
	d.checkSession()

	volumes, err := d.virtualbox.GetMedium(volumeID, volumeName)
	if err != nil {
		return nil, err
	}
	d.m.Unlock()

	if len(volumes) == 0 {
		return nil, nil
	}

	volumeMapping, err := d.GetVolumeMapping()
	if err != nil {
		return nil, err
	}

	mapDN := make(map[string]string)
	for _, vm := range volumeMapping {
		mapDN[vm.VolumeID] = vm.DeviceName
	}

	var volumesSD []*core.Volume

	for _, v := range volumes {
		var attachmentsSD []*core.VolumeAttachment
		for _, mid := range v.MachineIDs {
			dn, _ := mapDN[v.ID]
			attachmentSD := &core.VolumeAttachment{
				VolumeID:   v.ID,
				InstanceID: mid,
				DeviceName: dn,
				Status:     v.Location,
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		volumeSD := &core.Volume{
			Name:     v.Name,
			VolumeID: v.ID,
			Size:     strconv.Itoa(int(v.LogicalSize / 1024 / 1024 / 1024)),
			// VolumeType: v.MediumFormat,
			// NetworkName:      volume.NaaName,
			Status:      v.Location,
			Attachments: attachmentsSD,
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
		"volumeID":   volumeID,
		"volumeName": volumeName,
		"snapshotID": snapshotID,
		"size":       size,
	})

	size = size * 1024 * 1024 * 1024

	volumes, err := d.GetVolume("", volumeName)
	if err != nil {
		return nil, err
	}

	if len(volumes) > 0 {
		return nil, goof.WithField(volumeName, "volumeName", "volume exists already")
	}

	volume, err := d.createVolume(volumeName, size)
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error creating new volume", err)
	}

	volumes, err = d.GetVolume(volume.ID, "")
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, goof.New("failed to get new volume")
	}

	return volumes[0], nil
}

func (d *driver) RemoveVolume(volumeID string) error {
	d.m.Lock()
	defer d.m.Unlock()
	d.checkSession()

	fields := eff(map[string]interface{}{
		"volumeID": volumeID,
	})

	err := d.virtualbox.RemoveMedium(volumeID)
	if err != nil {
		return goof.WithFieldsE(fields, "error deleting volume", err)
	}

	return nil
}

//GetSnapshot returns snapshots from a volume or a specific snapshot
func (d *driver) GetSnapshot(
	volumeID, snapshotID, snapshotName string) ([]*core.Snapshot, error) {
	return nil, errors.ErrNotImplemented
}

func (d *driver) CreateSnapshot(
	notUsed bool,
	snapshotName, volumeID, description string) ([]*core.Snapshot, error) {
	return nil, errors.ErrNotImplemented
}

func (d *driver) RemoveSnapshot(snapshotID string) error {
	return errors.ErrNotImplemented
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
		return nil, goof.New("volume already attached to a host")
	} else if len(volumes[0].Attachments) > 0 && force {
		if err := d.DetachVolume(false, volumeID, "", true); err != nil {
			return nil, err
		}
	}

	if err := d.attachVolume(volumeID, ""); err != nil {
		return nil, goof.WithFieldE("volumeID", volumeID, "error attaching volume", err)
	}

	d.rescanScsiHosts()

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

	if err = d.detachVolume(volumeID, ""); err != nil {
		return goof.WithFieldE("volumeID", volumeID, "error detaching volume", err)
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

func (d *driver) storageLocation(name string) string {
	return filepath.Join(d.volumePath(), name)
}

func (d *driver) mountPoint(name string) string {
	return filepath.Join(d.volumePath(), name)
}

func (d *driver) createVolume(name string, size int64) (*vbox.Medium, error) {
	d.m.Lock()
	defer d.m.Unlock()
	d.checkSession()

	if name == "" {
		return nil, goof.New("name is empty")
	}

	return d.virtualbox.CreateMedium("vmdk", d.storageLocation(name), size)
}

func (d *driver) removeVolume(volumeID string) error {
	d.m.Lock()
	defer d.m.Unlock()
	d.checkSession()

	return d.virtualbox.RemoveMedium(volumeID)
}

func (d *driver) getVolume(volumeID, volumeName string) ([]*vbox.Medium, error) {
	d.m.Lock()
	defer d.m.Unlock()
	d.checkSession()

	return d.virtualbox.GetMedium(volumeID, volumeName)
}

func (d *driver) attachVolume(volumeID, volumeName string) error {
	d.m.Lock()
	defer d.m.Unlock()
	d.checkSession()

	medium, err := d.virtualbox.GetMedium(volumeID, volumeName)
	if err != nil {
		return err
	}

	if len(medium) == 0 {
		return goof.New("no volume returned")
	} else if len(medium) > 1 {
		return goof.New("too many volumes returned")
	}

	if err := d.machine.Refresh(); err != nil {
		return err
	}
	defer d.machine.Release()

	if err := d.machine.AttachDevice(medium[0]); err != nil {
		return err
	}

	return nil
}

func (d *driver) detachVolume(volumeID, volumeName string) error {
	d.m.Lock()
	defer d.m.Unlock()
	d.checkSession()

	medium, err := d.virtualbox.GetMedium(volumeID, volumeName)
	if err != nil {
		return err
	}

	if len(medium) == 0 {
		return goof.New("no volume returned")
	} else if len(medium) > 1 {
		return goof.New("too many volumes returned")
	}

	if err := d.machine.Refresh(); err != nil {
		return err
	}
	defer d.machine.Release()

	if err := medium[0].DetachMachines(); err != nil {
		return err
	}

	return nil
}

func (d *driver) findLocalMachine(nameOrID string) (*vbox.Machine, error) {
	d.m.Lock()
	defer d.m.Unlock()

	if nameOrID != "" {
		m, err := d.virtualbox.FindMachine(nameOrID)
		if err != nil {
			return nil, err
		} else if m == nil {
			return nil, goof.New("could not find machine")
		}
		id, err := m.GetID()
		if err != nil {
			return nil, err
		}
		m.ID = id

		name, err := m.GetName()
		if err != nil {
			return nil, err
		}
		m.Name = name
		return m, nil
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	macMap := make(map[string]bool)
	for _, intf := range interfaces {
		macUp := strings.ToUpper(strings.Replace(intf.HardwareAddr.String(), ":", "", -1))
		macMap[macUp] = true
	}

	machines, err := d.virtualbox.GetMachines()
	if err != nil {
		return nil, err
	}

	sp, err := d.virtualbox.GetSystemProperties()
	if err != nil {
		return nil, err
	}
	defer sp.Release()

	for _, m := range machines {
		defer m.Release()
		chipset, err := m.GetChipsetType()
		if err != nil {
			return nil, err
		}

		mna, err := sp.GetMaxNetworkAdapters(chipset)
		if err != nil {
			return nil, err
		}

		for i := uint32(0); i < mna; i++ {
			na, err := m.GetNetworkAdapter(i)
			if err != nil {
				return nil, err
			}

			mac, err := na.GetMACAddress()
			if err != nil {
				return nil, err
			}

			if _, ok := macMap[mac]; ok {
				id, err := m.GetID()
				if err != nil {
					return nil, err
				}
				m.ID = id

				name, err := m.GetName()
				if err != nil {
					return nil, err
				}
				m.Name = name

				return m, nil
			}
		}
	}

	return nil, goof.New("Unable to find machine")
}

func (d *driver) endpoint() string {
	return d.r.Config.GetString("virtualbox.endpoint")
}

func (d *driver) tls() bool {
	return d.r.Config.GetBool("virtualbox.tls")
}

func (d *driver) volumePath() string {
	return d.r.Config.GetString("virtualbox.volumePath")
}

func (d *driver) userName() string {
	return d.r.Config.GetString("virtualbox.userName")
}

func (d *driver) password() string {
	return d.r.Config.GetString("virtualbox.password")
}

func (d *driver) localMachineNameOrId() string {
	return d.r.Config.GetString("virtualbox.localMachineNameOrId")
}

func (d *driver) controllerName() string {
	cn := d.r.Config.GetString("virtualbox.controllerName")
	if cn == "" {
		return "SATA"
	}
	return cn
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("virtualbox")
	r.Key(gofig.String, "", "", "", "virtualbox.endpoint")
	r.Key(gofig.String, "", "", "", "virtualbox.volumePath")
	r.Key(gofig.String, "", "", "", "virtualbox.localMachineNameOrId")
	r.Key(gofig.String, "", "", "", "virtualbox.username")
	r.Key(gofig.String, "", "", "", "virtualbox.password")
	r.Key(gofig.Bool, "", false, "", "virtualbox.tls")
	r.Key(gofig.String, "", "", "", "virtualbox.controllerName")
	return r
}
