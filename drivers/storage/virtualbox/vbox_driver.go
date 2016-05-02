package virtualbox

import (
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	vbox "github.com/appropriate/go-virtualboxclient/virtualboxclient"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/virtualbox/executor"
)

const (
	// Name is the name of the driver.
	Name = executor.Name
)

// Driver represents a vbox driver implementation of StorageDriver
type Driver struct {
	exec executor.Executor
	sync.Mutex
	config         gofig.Config
	vbox           *vbox.VirtualBox
	machine        *vbox.Machine
	machineNameID  string
	uname          string
	passwd         string
	endpoint       string
	volumePath     string
	useTLS         bool
	controllerName string
}

func init() {
	gofig.Register(executor.LoadConfig())
	registry.RegisterStorageDriver(Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &Driver{}
}

// Name returns the name of the driver
func (d *Driver) Name() string {
	return Name
}

// Init initializes the driver.
func (d *Driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	d.exec.Config = config
	d.uname = d.config.GetString("virtualbox.username")
	d.passwd = d.config.GetString("virtualbox.username")
	d.endpoint = d.config.GetString("virtualbox.endpoint")
	d.volumePath = d.config.GetString("virtualbox.volumePath")
	d.useTLS = d.config.GetBool("virtualbox.tls")
	d.machineNameID = d.config.GetString("virtualbox.localMachineNameOrId")

	fields := map[string]interface{}{
		"provider":        Name,
		"moduleName":      Name,
		"endpoint":        d.endpoint,
		"userName":        d.uname,
		"tls":             d.useTLS,
		"volumePath":      d.volumePath,
		"machineNameOrId": d.machineNameID,
	}

	log.Info("initializing driver: ", fields)
	d.vbox = vbox.New(d.uname, d.passwd,
		d.endpoint, d.useTLS, d.controllerName)

	if err := d.vbox.Logon(); err != nil {
		return goof.WithFieldsE(fields,
			"error logging in", err)
	}

	if m, err := d.findLocalMachine(d.machineNameID); err != nil {
		goof.WithFieldsE(fields,
			"failed to find local machine", err)
	} else {
		d.machine = m
	}

	log.WithFields(fields).Info("storage driver initialized")
	return nil
}

// InstanceID returns the local system's InstanceID.
func (d *Driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {
	if d.machine == nil {
		log.Error("invalid object id for machine")
		return nil, goof.New("invalid machine object id")
	}
	d.findLocalMachine(d.machineNameID)
	return &types.InstanceID{ID: d.machine.ID}, nil
}

// LocalDevices returns a map of the system's local devices.
func (d *Driver) LocalDevices(
	ctx types.Context,
	opts types.Store) (map[string]string, error) {
	return d.exec.LocalDevices(ctx, opts)
}

// NextDevice returns the next available device (not implemented).
func (d *Driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return d.exec.NextDevice(ctx, opts)
}

// Type returns the type of storage a driver provides
func (d *Driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Block, nil
}

// NextDeviceInfo returns the information about the driver's next available
// device workflow.
func (d *Driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return nil, nil
}

// InstanceInspect returns an instance.
func (d *Driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {
	instanceID, _ := d.InstanceID(ctx, opts)
	return &types.Instance{InstanceID: instanceID}, nil
}

// Volumes returns all volumes or a filtered list of volumes.
func (d *Driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {
	d.Lock()
	defer d.Unlock()
	d.refreshSession()

	if err := d.machine.Refresh(); err != nil {
		return nil, err
	}
	defer d.machine.Release()

	mapDiskByID, err := d.LocalDevices(ctx, nil)
	if err != nil {
		return nil, err
	}

	mas, err := d.machine.GetMediumAttachments()
	if err != nil {
		return nil, err
	}

	var blockDevices []*types.Volume
	for _, ma := range mas {
		medium := d.vbox.NewMedium(ma.Medium)
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
		sdBlockDevice := &types.Volume{
			Name:   bdn,
			ID:     mid,
			Status: location,
		}
		blockDevices = append(blockDevices, sdBlockDevice)

	}
	return blockDevices, nil
}

// VolumeInspect inspects a single volume.
func (d *Driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {
	return nil, types.ErrNotImplemented
}

// VolumeCreate creates a new volume.
func (d *Driver) VolumeCreate(
	ctx types.Context,
	name string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	if opts.Size == nil {
		return nil, goof.New("missing volume size")
	}

	fields := map[string]interface{}{
		"provider":   Name,
		"volumeName": name,
		"size":       *opts.Size,
	}

	size := *opts.Size * 1024 * 1024 * 1024

	d.refreshSession()
	volumes, err := d.GetVolume(ctx, "", name)
	if err != nil {
		return nil, err
	}

	if len(volumes) > 0 {
		return nil, goof.WithFields(fields, "volume exists already")
	}

	volume, err := d.createVolume(name, size)
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error creating new volume", err)
	}

	// double check
	volumes, err = d.GetVolume(ctx, volume.ID, "")
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, goof.New("failed to get new volume")
	}

	newVol := &types.Volume{
		ID:   volume.ID,
		Name: volume.Name,
		Size: volume.Size,
		IOPS: *opts.IOPS,
		Type: string(volume.DeviceType),
	}

	return newVol, nil
}

// VolumeCreateFromSnapshot (not implemented).
func (d *Driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	return nil, types.ErrNotImplemented
}

// VolumeCopy copies an existing volume (not implemented)
func (d *Driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {
	return nil, types.ErrNotImplemented
}

// VolumeSnapshot snapshots a volume (not implemented)
func (d *Driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

// VolumeRemove removes a volume.
func (d *Driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {

	d.Lock()
	defer d.Unlock()
	d.refreshSession()

	fields := map[string]interface{}{
		"provider": Name,
		"volumeID": volumeID,
	}

	err := d.vbox.RemoveMedium(volumeID)
	if err != nil {
		return goof.WithFieldsE(fields, "error deleting volume", err)
	}

	return nil
}

// VolumeAttach attaches a volume.
func (d *Driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, error) {

	if volumeID == "" {
		return nil, goof.New("missing volume id")
	}

	volumes, err := d.GetVolume(ctx, volumeID, "")
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, goof.New("no volume found")
	}

	if len(volumes[0].Attachments) > 0 && !opts.Force {
		return nil, goof.New("volume already attached to a host")
	}
	if opts.Force {
		if _, err := d.VolumeDetach(ctx, volumeID, nil); err != nil {
			return nil, err
		}
	}

	err = d.attachVolume(volumeID, "")
	if err != nil {
		return nil, goof.WithFieldsE(
			log.Fields{
				"provider": Name,
				"volumeID": volumeID},
			"error attaching volume",
			err,
		)
	}

	d.rescanScsiHosts()

	return volumes[0], nil
}

// VolumeDetach detaches a volume.
func (d *Driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	if volumeID == "" {
		return nil, goof.New("missing volume id")
	}

	volumes, err := d.GetVolume(ctx, volumeID, "")
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, goof.New("no volume returned")
	}

	if err := d.detachVolume(volumeID, ""); err != nil {
		return nil, goof.WithFieldsE(
			log.Fields{
				"provier":  Name,
				"volumeID": volumeID}, "error detaching volume", err)
	}

	log.Info("detached volume", volumeID)
	return d.VolumeInspect(
		ctx, volumeID, &types.VolumeInspectOpts{Attachments: true})
}

//Snapshots (not implmented)
func (d *Driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

// SnapshotInspect (not implemented)
func (d *Driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

// SnapshotCopy (not implemented)
func (d *Driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

// SnapshotRemove (not implmeented)
func (d *Driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {
	return types.ErrNotImplemented
}

// GetVolume searches and returns a volume matching criteria
func (d *Driver) GetVolume(
	ctx types.Context,
	volumeID, volumeName string) ([]*types.Volume, error) {
	d.Lock()
	d.refreshSession()

	volumes, err := d.vbox.GetMedium(volumeID, volumeName)
	if err != nil {
		return nil, err
	}
	d.Unlock()

	if len(volumes) == 0 {
		return nil, nil
	}

	volumeMapping, err := d.Volumes(ctx, nil)
	if err != nil {
		return nil, err
	}

	mapDN := make(map[string]string)
	for _, vm := range volumeMapping {
		mapDN[vm.ID] = vm.Name
	}

	var volumesSD []*types.Volume

	for _, v := range volumes {
		var attachmentsSD []*types.VolumeAttachment
		for _, mid := range v.MachineIDs {
			dn, _ := mapDN[v.ID]
			attachmentSD := &types.VolumeAttachment{
				VolumeID:   v.ID,
				InstanceID: &types.InstanceID{ID: mid},
				DeviceName: dn,
				Status:     v.Location,
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}

		volumeSD := &types.Volume{
			Name:        v.Name,
			ID:          v.ID,
			Size:        int64(v.LogicalSize / 1024 / 1024 / 1024),
			Status:      v.Location,
			Attachments: attachmentsSD,
		}
		volumesSD = append(volumesSD, volumeSD)
	}

	return volumesSD, nil
}

// GetVolumeAttach returns volume attachments
func (d *Driver) GetVolumeAttach(
	ctx types.Context,
	volumeID, instanceID string) ([]*types.VolumeAttachment, error) {

	if volumeID == "" {
		return nil, goof.New("missing volume id")
	}
	volume, err := d.GetVolume(ctx, volumeID, "")
	if err != nil {
		return nil, err
	}

	// TODO - Logic looks suspicious, attached always false
	if instanceID != "" {
		var attached bool
		for _, volumeAttachment := range volume[0].Attachments {
			if volumeAttachment.InstanceID.ID == instanceID {
				return volume[0].Attachments, nil
			}
		}
		if !attached {
			return []*types.VolumeAttachment{}, nil
		}
	}
	return volume[0].Attachments, nil
}

func (d *Driver) findLocalMachine(nameOrID string) (*vbox.Machine, error) {
	d.Lock()
	defer d.Unlock()

	log.Debug("Finding local machine for ID: ", nameOrID)
	if nameOrID != "" {
		m, err := d.vbox.FindMachine(nameOrID)
		if err != nil {
			return nil, err
		}
		if m == nil {
			return nil, goof.New("could not find machine")
		}

		if id, err := m.GetID(); err == nil {
			m.ID = id
		} else {
			return nil, err
		}

		if name, err := m.GetName(); err == nil {
			m.Name = name
		} else {
			return nil, err
		}

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

	machines, err := d.vbox.GetMachines()
	if err != nil {
		return nil, err
	}

	sp, err := d.vbox.GetSystemProperties()
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

// TODO too costly, need better way to validate session (i.e. some delay)
func (d *Driver) refreshSession() {
	_, err := d.vbox.FindMachine(d.machine.ID)
	if err != nil {
		log.Debug("logging in again")
		d.vbox.Logon()
	}
}

func (d *Driver) createVolume(name string, size int64) (*vbox.Medium, error) {
	d.Lock()
	defer d.Unlock()
	d.refreshSession()

	if name == "" {
		return nil, goof.New("name is empty")
	}
	path := filepath.Join(d.volumePath, name)
	return d.vbox.CreateMedium("vmdk", path, size)
}

func (d *Driver) attachVolume(volumeID, volumeName string) error {
	d.Lock()
	defer d.Unlock()
	d.refreshSession()

	medium, err := d.vbox.GetMedium(volumeID, volumeName)
	if err != nil {
		return err
	}

	if len(medium) == 0 {
		return goof.New("no volume returned")
	}
	if len(medium) > 1 {
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

func (d *Driver) detachVolume(volumeID, volumeName string) error {
	d.Lock()
	defer d.Unlock()
	d.refreshSession()

	media, err := d.vbox.GetMedium(volumeID, volumeName)
	if err != nil {
		return err
	}

	if len(media) == 0 {
		return goof.New("no volume returned")
	}
	if len(media) > 1 {
		return goof.New("too many volumes returned")
	}

	if err := d.machine.Refresh(); err != nil {
		return err
	}
	defer d.machine.Release()

	if err := media[0].DetachMachines(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) rescanScsiHosts() {
	hosts := d.config.GetString("virtualbox.scsiHostPath")
	if dirs, err := ioutil.ReadDir(hosts); err == nil {
		for _, f := range dirs {
			name := hosts + f.Name() + "/scan"
			data := []byte("- - -")
			ioutil.WriteFile(name, data, 0666)
		}
	}
	time.Sleep(1 * time.Second)
}
