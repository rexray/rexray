package storage

import (
	"fmt"
	"net"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	isi "github.com/emccode/goisilon"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/isilon"
)

const (
	bytesPerGb  = int64(1024 * 1024 * 1024)
	idDelimiter = "/"
)

// Driver represents a vbox driver implementation of StorageDriver
type driver struct {
	sync.Mutex
	config gofig.Config
	client *isi.Client
}

func init() {
	registry.RegisterStorageDriver(isilon.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{}
}

// Name returns the name of the driver
func (d *driver) Name() string {
	return isilon.Name
}

// Init initializes the driver.
func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config

	fields := log.Fields{
		"endpoint":   d.endpoint(),
		"userName":   d.userName(),
		"group":      d.group(),
		"insecure":   d.insecure(),
		"volumePath": d.volumePath(),
		"dataSubnet": d.dataSubnet(),
	}

	if d.password() == "" {
		fields["password"] = ""
	} else {
		fields["password"] = "******"
	}

	var err error

	if d.client, err = isi.NewClientWithArgs(
		d.endpoint(),
		d.insecure(),
		d.userName(),
		d.group(),
		d.password(),
		d.volumePath()); err != nil {
		return goof.WithFieldsE(fields,
			"error creating isilon client", err)
	}

	log.WithFields(fields).Info("storage driver initialized")
	return nil
}

func (d *driver) getInstanceID(ctx types.Context) (string, error) {

	iid := context.MustInstanceID(ctx)
	var nets []string
	if err := iid.UnmarshalMetadata(&nets); err != nil {
		return "", err
	}

	_, dataSubnet, err := net.ParseCIDR(d.dataSubnet())
	if err != nil {
		return "", goof.WithFieldE("dataSubnet", d.dataSubnet(),
			"invalid data subnet", err)
	}

	var idList []string
	for _, addr := range nets {
		ip, _, err := net.ParseCIDR(addr)
		if err != nil {
			return "", err
		}
		if dataSubnet.Contains(ip) == true {
			idList = append(idList, ip.String())
		}
	}

	if len(idList) == 0 {
		return "", goof.WithFieldsE(
			log.Fields{
				"dataSubnet": d.dataSubnet(),
			}, "no IPs in the data subnet", err)
	}

	return createInstanceID(idList), nil
}

// Create an instance ID from a list of client IP addresses
func createInstanceID(clients []string) string {
	return strings.Join(clients, idDelimiter)
}

// InstanceInspect returns an instance.
func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	iid := context.MustInstanceID(ctx)
	if iid.ID != "" {
		return &types.Instance{InstanceID: iid}, nil
	}

	id, err := d.getInstanceID(ctx)
	if err != nil {
		return nil, err
	}
	instanceID := &types.InstanceID{ID: id, Driver: d.Name()}

	return &types.Instance{InstanceID: instanceID}, nil
}

// LocalDevices returns a map of the system's local devices.
func (d *driver) LocalDevices(
	ctx types.Context,
	opts types.Store) (*types.LocalDevices, error) {

	if ld, ok := context.LocalDevices(ctx); ok {
		return ld, nil
	}
	return nil, goof.New("missing local devices")
}

// NextDevice returns the next available device (not implemented).
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "", nil
}

// Type returns the type of storage a driver provides
func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.NAS, nil
}

// NextDeviceInfo returns the information about the driver's next available
// device workflow.
func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return nil, nil
}

func (d *driver) getVolumeAttachments(ctx types.Context) (
	[]*types.VolumeAttachment, error) {
	exports, err := d.client.GetVolumeExports()
	if err != nil {
		return nil, err
	}

	// should be using this, but instanceID is coming back either blank or
	// with metadata today
	// iid := ctx.InstanceID()

	ii, err := d.InstanceInspect(ctx, nil)
	if err != nil {
		return nil, err
	}
	iid := ii.InstanceID

	ld, err := d.LocalDevices(ctx, nil)
	if err != nil {
		return nil, err
	}

	var atts []*types.VolumeAttachment
	for _, export := range exports {
		var dev string
		var status string
		for _, c := range export.Clients {
			if c == iid.ID {
				dev = d.nfsMountPath(export.ExportPath)
				if _, ok := ld.DeviceMap[dev]; ok {
					status = "Exported and Mounted"
				} else {
					status = "Exported and Unmounted"
				}
			} else {
				status = "Exported"
			}
			attachmentSD := &types.VolumeAttachment{
				VolumeID:   export.Volume.Name,
				InstanceID: &types.InstanceID{ID: c, Driver: d.Name()},
				DeviceName: dev,
				Status:     status,
			}
			atts = append(atts, attachmentSD)
		}
	}

	return atts, nil
}

func (d *driver) nfsMountPath(mountPath string) string {
	return fmt.Sprintf("%s:%s", d.nfsHost(), mountPath)
}

func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	// always return attachments to align against other drivers for now
	return d.getVolume(ctx, "", "", true)
}

func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	vols, err := d.getVolume(ctx, volumeID, "", opts.Attachments)
	if err != nil {
		return nil, err
	}

	if vols == nil {
		return nil, nil
	}

	return vols[0], nil
}

// VolumeCreate creates a new volume.
func (d *driver) VolumeCreate(ctx types.Context, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	vol, err := d.VolumeInspect(ctx, volumeName,
		&types.VolumeInspectOpts{Attachments: false})
	if err != nil {
		return nil, err
	}

	if vol != nil {
		return nil, goof.New("volume name already exists")
	}

	_, err = d.client.CreateVolume(volumeName)
	if err != nil {
		return nil, goof.WithFieldE("volumeName", volumeName, "Error creating volume", err)
	}

	// Set or update the quota for volume
	if d.quotas() {
		quota, err := d.client.GetQuota(volumeName)
		if quota == nil {
			// PAPI uses bytes for it's size units, but REX-Ray uses gigs
			err = d.client.SetQuotaSize(volumeName, *opts.Size*bytesPerGb)
			if err != nil {
				// TODO: not sure how to handle this situation.  Delete created volume
				// and return an error?  Ignore and continue?
				return nil, goof.WithFieldE("volumeName", volumeName,
					"Error creating volume", err)
			}
		} else {
			// PAPI uses bytes for it's size units, but REX-Ray uses gigs
			err = d.client.UpdateQuotaSize(volumeName, *opts.Size*bytesPerGb)
			if err != nil {
				// TODO: not sure how to handle this situation.  Delete created volume
				// and return an error?  Ignore and continue?
				return nil, goof.WithFieldE("volumeName", volumeName,
					"Error creating volume", err)
			}
		}
	}

	return d.VolumeInspect(ctx, volumeName,
		&types.VolumeInspectOpts{Attachments: false})
}

// VolumeRemove removes a volume.
func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {

	if d.quotas() {
		err := d.client.ClearQuota(volumeID)
		if err != nil {
			return err
		}
	}

	err := d.client.DeleteVolume(volumeID)
	if err != nil {
		return err
	}
	return nil
}

// VolumeAttach attaches a volume.
func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	d.Lock()
	defer d.Unlock()

	instanceID, err := d.InstanceInspect(ctx, nil)
	if err != nil {
		return nil, "", err
	}

	// ensure the volume exists and is exported
	vol, err := d.VolumeInspect(ctx, volumeID,
		&types.VolumeInspectOpts{Attachments: true})
	if err != nil {
		return nil, "", err
	}
	if vol == nil {
		return nil, "", goof.New("no volumes returned")
	}
	if err := d.client.ExportVolume(volumeID); err != nil {
		return nil, "", goof.WithError("problem exporting volume", err)
	}
	// see if anyone is attached already
	clients, err := d.client.GetExportClients(volumeID)
	if err != nil {
		return nil, "", goof.WithError("problem getting export client", err)
	}

	// clear out any existing clients if necessary.  if force is false and
	// we have existing clients, we need to exit.
	if len(clients) > 0 && !d.sharedMounts() && opts.Force == false {
		for _, c := range clients {
			if c == instanceID.InstanceID.ID {
				return nil, "", goof.New("volume already attached to instance")
			}
		}

		return nil, "", goof.New("volume already attached to another host")
	}

	if d.sharedMounts() {
		clients = append(clients, instanceID.InstanceID.ID)
	} else {
		clients = []string{instanceID.InstanceID.ID}
	}

	log.WithField("clients", clients).Info("setting exports")
	err = d.client.SetExportClients(volumeID, clients)
	if err != nil {
		return nil, "", err
	}

	vol, err = d.VolumeInspect(ctx, volumeID,
		&types.VolumeInspectOpts{Attachments: true})
	if err != nil {
		return nil, "", err
	}

	return vol, "", err
}

// VolumeDetach detaches a volume.
func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	d.Lock()
	defer d.Unlock()

	vol, err := d.VolumeInspect(ctx, volumeID,
		&types.VolumeInspectOpts{
			Attachments: false,
		})
	if err != nil {
		return nil, err
	}

	if vol == nil {
		return nil, goof.New("no volumes returned")
	}

	instanceID, err := d.InstanceInspect(ctx, nil)
	if err != nil {
		return nil, err
	}

	clients, err := d.client.GetExportClients(volumeID)
	if err != nil {
		return nil, goof.WithError("problem getting export client", err)
	}

	var newClients []string
	for _, c := range clients {
		if c != instanceID.InstanceID.ID {
			newClients = append(newClients, c)
		}
	}

	if len(newClients) > 0 {
		log.WithField("clients", clients).Info("setting exports")
		err = d.client.SetExportClients(volumeID, newClients)
		if err != nil {
			return nil, err
		}
	} else {
		if err := d.client.UnexportVolume(volumeID); err != nil {
			return nil, goof.WithError("problem unexporting volume", err)
		}
	}

	return d.VolumeInspect(ctx, volumeID, &types.VolumeInspectOpts{
		Attachments: true,
	})
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

func (d *driver) getVolume(ctx types.Context, volumeID, volumeName string,
	attachments bool) ([]*types.Volume, error) {
	var volumes []isi.Volume
	if volumeID != "" || volumeName != "" {
		volume, err := d.client.GetVolume(volumeID, volumeName)
		if err != nil && !strings.Contains(err.Error(), "Unable to open object") {
			return nil, err
		}
		if volume != nil {
			volumes = append(volumes, volume)
		}
	} else {
		var err error
		volumes, err = d.client.GetVolumes()
		if err != nil {
			return nil, err
		}
	}

	if len(volumes) == 0 {
		return nil, nil
	}

	var atts []*types.VolumeAttachment
	if attachments {
		var err error
		atts, err = d.getVolumeAttachments(ctx)
		if err != nil {
			return nil, err
		}
	}

	attMap := make(map[string][]*types.VolumeAttachment)
	for _, att := range atts {
		if attMap[att.VolumeID] == nil {
			attMap[att.VolumeID] = make([]*types.VolumeAttachment, 0)
		}
		attMap[att.VolumeID] = append(attMap[att.VolumeID], att)
	}

	var volumesSD []*types.Volume
	for _, volume := range volumes {
		volSize, err := d.getSize(volume.Name, volume.Name)
		if err != nil {
			return nil, err
		}

		vatts, _ := attMap[volume.Name]
		volumeSD := &types.Volume{
			Name:        volume.Name,
			ID:          volume.Name,
			Size:        volSize,
			Attachments: vatts,
		}
		volumesSD = append(volumesSD, volumeSD)
	}

	return volumesSD, nil
}

func (d *driver) getSize(volumeID, volumeName string) (int64, error) {
	if d.quotas() == false {
		return 0, nil
	}

	if volumeID != "" {
		volumeName = volumeID
	}
	if volumeName == "" {
		return 0, goof.New("volume name or ID not set")
	}

	quota, err := d.client.GetQuota(volumeName)
	if err != nil {
		return 0, nil
	}
	// PAPI returns the size in bytes, REX-Ray uses gigs
	if quota.Thresholds.Hard != 0 {
		return quota.Thresholds.Hard / bytesPerGb, nil
	}

	return 0, nil

}

func (d *driver) endpoint() string {
	return d.config.GetString("isilon.endpoint")
}

func (d *driver) insecure() bool {
	return d.config.GetBool("isilon.insecure")
}

func (d *driver) userName() string {
	return d.config.GetString("isilon.userName")
}

func (d *driver) group() string {
	return d.config.GetString("isilon.group")
}

func (d *driver) password() string {
	return d.config.GetString("isilon.password")
}

func (d *driver) volumePath() string {
	return d.config.GetString("isilon.volumePath")
}

func (d *driver) nfsHost() string {
	return d.config.GetString("isilon.nfsHost")
}

func (d *driver) dataSubnet() string {
	return d.config.GetString("isilon.dataSubnet")
}

func (d *driver) quotas() bool {
	return d.config.GetBool("isilon.quotas")
}

func (d *driver) sharedMounts() bool {
	return d.config.GetBool("isilon.sharedMounts")
}
