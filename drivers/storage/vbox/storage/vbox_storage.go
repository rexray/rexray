package storage

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	vboxw "github.com/appropriate/go-virtualboxclient/vboxwebsrv"
	vboxc "github.com/appropriate/go-virtualboxclient/virtualboxclient"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/vbox"
)

// Driver represents a vbox driver implementation of StorageDriver
type driver struct {
	sync.Mutex
	config gofig.Config
	vbox   *vboxc.VirtualBox
}

func init() {
	registry.RegisterStorageDriver(vbox.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{}
}

// Name returns the name of the driver
func (d *driver) Name() string {
	return vbox.Name
}

// Init initializes the driver.
func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config

	fields := map[string]interface{}{
		"provider":        vbox.Name,
		"moduleName":      vbox.Name,
		"endpoint":        d.endpoint(),
		"userName":        d.username(),
		"tls":             d.tls(),
		"volumePath":      d.volumePath(),
		"controllerName":  d.controllerName(),
		"machineNameOrId": d.machineNameID(ctx),
	}

	log.Info("initializing driver: ", fields)
	d.vbox = vboxc.New(d.username(), d.password(),
		d.endpoint(), d.tls(), d.controllerName())

	if err := d.vbox.Logon(); err != nil {
		return goof.WithFieldsE(fields,
			"error logging in", err)
	}

	log.WithFields(fields).Info("storage driver initialized")
	return nil
}

// LocalDevices returns a map of the system's local devices.
func (d *driver) LocalDevices(
	ctx types.Context,
	opts types.Store) (map[string]string, error) {
	return ctx.LocalDevices(), nil
}

// NextDevice returns the next available device (not implemented).
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "", nil
}

// Type returns the type of storage a driver provides
func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Block, nil
}

// NextDeviceInfo returns the information about the driver's next available
// device workflow.
func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return nil, nil
}

func getMacs(ctx types.Context) []string {
	if ctx.InstanceID() == nil {
		return nil
	}
	ctx.WithField("instanceID", ctx.InstanceID().ID).Debug("checking instance ID")
	iidj, _ := ctx.InstanceID().Metadata.MarshalJSON()
	var iid []string
	json.Unmarshal(iidj, &iid)
	return iid
}

// getInstanceID returns the local system's InstanceID.
func (d *driver) getInstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	m, err := d.findMachine(ctx, d.machineNameID(ctx), getMacs(ctx))
	if err != nil {
		goof.WithFieldE("metadata", getMacs(ctx),
			"failed to find local machine", err)
	}

	if m == nil {
		return nil, goof.New("machine not found")
	}

	return &types.InstanceID{ID: m.ID}, nil
}

// InstanceInspect returns an instance.
func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	d.refreshSession(ctx)

	instanceID, err := d.getInstanceID(ctx, opts)
	if err != nil {
		return nil, err
	}
	instanceID.Formatted = true

	return &types.Instance{InstanceID: instanceID}, nil
}

// Volumes returns all volumes or a filtered list of volumes.
func (d *driver) getVolumeMapping(
	ctx types.Context) ([]*types.Volume, error) {

	var err error
	var mapDiskByID map[string]string
	var mas []*vboxw.IMediumAttachment
	var m *vboxc.Machine

	m, err = d.findMachine(ctx, d.machineNameID(ctx), getMacs(ctx))
	if err != nil {
		return nil, err
	}

	if err := m.Refresh(); err != nil {
		return nil, err
	}
	defer m.Release()

	mapDiskByID = ctx.LocalDevices()

	mas, err = m.GetMediumAttachments()
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

// VolumeCreate creates a new volume.
func (d *driver) VolumeCreate(ctx types.Context, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	d.Lock()
	defer d.Unlock()
	d.refreshSession(ctx)

	if opts.Size == nil {
		return nil, goof.New("missing volume size")
	}

	fields := map[string]interface{}{
		"provider":   vbox.Name,
		"volumeName": volumeName,
		"size":       *opts.Size,
	}

	size := *opts.Size * 1024 * 1024 * 1024

	vol, err := d.getVolume(ctx, "", volumeName, false)
	if err != nil {
		return nil, err
	}

	if vol != nil {
		return nil, goof.New("volume already exists")
	}

	med, err := d.createVolume(ctx, volumeName, size)
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error creating new volume", err)
	}

	var iops int64
	if opts.IOPS != nil {
		iops = *opts.IOPS
	}

	newVol := &types.Volume{
		ID:   med.ID,
		Name: med.Name,
		Size: med.LogicalSize / 1024 / 1024 / 1024,
		IOPS: iops,
		Type: string(med.DeviceType),
	}

	return newVol, nil
}

// VolumeCreateFromSnapshot (not implemented).
func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	return nil, types.ErrNotImplemented
}

// VolumeCopy copies an existing volume (not implemented)
func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {
	return nil, types.ErrNotImplemented
}

// VolumeSnapshot snapshots a volume (not implemented)
func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, types.ErrNotImplemented
}

// VolumeRemove removes a volume.
func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {

	d.Lock()
	defer d.Unlock()
	d.refreshSession(ctx)

	fields := map[string]interface{}{
		"provider": vbox.Name,
		"volumeID": volumeID,
	}

	err := d.vbox.RemoveMedium(volumeID)
	if err != nil {
		return goof.WithFieldsE(fields, "error deleting volume", err)
	}

	return nil
}

// VolumeAttach attaches a volume.
func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	d.refreshSession(ctx)

	if volumeID == "" {
		return nil, "", goof.New("missing volume id")
	}

	volumes, err := d.getVolume(ctx, volumeID, "", true)
	if err != nil {
		return nil, "", err
	}

	if len(volumes) == 0 {
		return nil, "", goof.New("no volume found")
	}

	if len(volumes[0].Attachments) > 0 && !opts.Force {
		return nil, "", goof.New("volume already attached to a host")
	}
	if opts.Force {
		if _, err := d.VolumeDetach(ctx, volumeID, nil); err != nil {
			return nil, "", err
		}
	}

	err = d.attachVolume(ctx, volumeID, "")
	if err != nil {
		return nil, "", goof.WithFieldsE(
			log.Fields{
				"provider": vbox.Name,
				"volumeID": volumeID},
			"error attaching volume",
			err,
		)
	}

	volumes, err = d.getVolume(ctx, volumeID, "", true)
	if err != nil {
		return nil, "", err
	}

	if len(volumes) == 0 {
		return nil, "", err
	}

	svid := strings.Split(volumes[0].ID, "-")

	return volumes[0], svid[0], nil
}

// VolumeDetach detaches a volume.
func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	d.refreshSession(ctx)

	if volumeID == "" {
		return nil, goof.New("missing volume id")
	}

	volumes, err := d.getVolume(ctx, volumeID, "", false)
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, goof.New("no volume returned")
	}

	if err := d.detachVolume(ctx, volumeID, ""); err != nil {
		return nil, goof.WithFieldsE(
			log.Fields{
				"provier":  vbox.Name,
				"volumeID": volumeID}, "error detaching volume", err)
	}

	log.Info("detached volume", volumeID)
	return d.VolumeInspect(
		ctx, volumeID, &types.VolumeInspectOpts{Attachments: true})
}

func (d *driver) VolumeDetachAll(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {
	return nil
}

func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	return nil, nil
}

func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {
	return nil
}

func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	d.refreshSession(ctx)

	vols, err := d.getVolume(ctx, "", "", opts.Attachments)
	if err != nil {
		return nil, err
	}
	return vols, nil
}

func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	d.refreshSession(ctx)

	vols, err := d.getVolume(ctx, volumeID, "", opts.Attachments)
	if err != nil {
		return nil, err
	}
	if len(vols) == 0 {
		return nil, goof.New("no volumes returned")
	}
	return vols[0], nil
}

// getVolume searches and returns a volume matching criteria
func (d *driver) getVolume(
	ctx types.Context,
	volumeID string, volumeName string,
	attachments bool) ([]*types.Volume, error) {

	d.refreshSession(ctx)

	volumes, err := d.vbox.GetMedium(volumeID, volumeName)
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return nil, nil
	}

	var mapDN map[string]string
	if attachments {
		volumeMapping, err := d.getVolumeMapping(ctx)
		if err != nil {
			return nil, err
		}

		mapDN = make(map[string]string)
		for _, vm := range volumeMapping {
			mapDN[vm.ID] = vm.Name
		}
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

func (d *driver) findMachine(
	ctx types.Context, nameOrID string, macs []string) (*vboxc.Machine, error) {
	log.WithField("nameOrID", nameOrID).Debug("finding local machine")

	if nameOrID == "" &&
		(ctx.InstanceID() != nil &&
			ctx.InstanceID().ID != "") {
		log.WithField("instanceID", ctx.InstanceID().ID).Debug(
			"proceeding with cached instanceID")
		nameOrID = ctx.InstanceID().ID
	}

	log.WithFields(log.Fields{
		"nameOrID": nameOrID,
		"macs":     macs,
	}).Debug("processing with nameOrID")

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

	macMap := make(map[string]bool)
	for _, mac := range macs {
		macUp := mac
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
func (d *driver) refreshSession(ctx types.Context) {
	d.vbox.Logon()
}

func (d *driver) createVolume(ctx types.Context, name string, size int64) (*vboxc.Medium, error) {
	if name == "" {
		return nil, goof.New("name is empty")
	}
	path := filepath.Join(d.volumePath(), name)
	return d.vbox.CreateMedium("vmdk", path, size)
}

func (d *driver) attachVolume(
	ctx types.Context, volumeID, volumeName string) error {

	m, err := d.findMachine(ctx, d.machineNameID(ctx), getMacs(ctx))
	if err != nil {
		return err
	}

	if err := m.Refresh(); err != nil {
		return err
	}
	defer m.Release()

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

	if err := m.AttachDevice(medium[0]); err != nil {
		return err
	}

	return nil
}

func (d *driver) detachVolume(
	ctx types.Context, volumeID, volumeName string) error {

	m, err := d.findMachine(ctx, d.machineNameID(ctx), getMacs(ctx))
	if err != nil {
		return err
	}

	if err := m.Refresh(); err != nil {
		return err
	}
	defer m.Release()

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

	if err := media[0].DetachMachines(); err != nil {
		return err
	}

	return nil
}

func (d *driver) username() string {
	return d.config.GetString("virtualbox.username")
}

func (d *driver) password() string {
	return d.config.GetString("virtualbox.password")
}

func (d *driver) endpoint() string {
	return d.config.GetString("virtualbox.endpoint")
}

func (d *driver) volumePath() string {
	return d.config.GetString("virtualbox.volumePath")
}

func (d *driver) controllerName() string {
	return d.config.GetString("virtualbox.controllerName")
}

func (d *driver) tls() bool {
	return d.config.GetBool("virtualbox.tls")
}

func (d *driver) machineNameID(ctx types.Context) string {
	if ctx.InstanceID() != nil && ctx.InstanceID().ID != "" {
		return ctx.InstanceID().ID
	}
	return d.config.GetString("virtualbox.localMachineNameOrId")
}
