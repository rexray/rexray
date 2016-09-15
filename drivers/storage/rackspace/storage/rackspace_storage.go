package storage

import (
	"os/exec"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/rackspace"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/snapshots"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/volumes"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/volumeattach"
)

const (
	providerName = "Rackspace"
	minSize      = 75 //rackspace is 75
)

type driver struct {
	provider           *gophercloud.ProviderClient
	client             *gophercloud.ServiceClient
	clientBlockStorage *gophercloud.ServiceClient
	region             string
	instanceID         string
	config             gofig.Config
}

func init() {
	registry.RegisterStorageDriver(rackspace.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{}
}

func (d *driver) Name() string {
	return rackspace.Name
}

func (d *driver) Init(context types.Context, config gofig.Config) error {
	d.config = config

	fields := eff(map[string]interface{}{})
	var err error

	if d.instanceID, err = d.getInstanceID(); err != nil {
		return err
	}

	fields["moduleName"] = context
	fields["instanceId"] = d.instanceID

	if d.region, err = d.getInstanceRegion(); err != nil {
		return err
	}

	fields["region"] = d.region
	d.region = strings.ToUpper(d.region)

	authOpts := d.getAuthOptions()

	fields["identityEndpoint"] = d.authURL()
	fields["userId"] = d.userID()
	fields["userName"] = d.userName()
	if d.password() == "" {
		fields["password"] = ""
	} else {
		fields["password"] = "******"
	}
	fields["tenantId"] = d.tenantID()
	fields["tenantName"] = d.tenantName()
	fields["domainId"] = d.domainID()
	fields["domainName"] = d.domainName()

	if d.provider, err = openstack.AuthenticatedClient(authOpts); err != nil {
		return goof.WithFieldsE(fields,
			"error getting authenticated client", err)
	}

	if d.client, err = openstack.NewComputeV2(d.provider,
		gophercloud.EndpointOpts{Region: d.region}); err != nil {
		goof.WithFieldsE(fields, "error getting newComputeV2", err)
	}

	if d.clientBlockStorage, err = openstack.NewBlockStorageV1(d.provider,
		gophercloud.EndpointOpts{Region: d.region}); err != nil {
		return goof.WithFieldsE(fields,
			"error getting newBlockStorageV1", err)
	}

	log.WithFields(fields).Info("storage driver initialized")
	return nil

}

// 	// Type returns the type of storage the driver provides.
func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Block, nil
}

// 	// NextDeviceInfo returns the information about the driver's next available
// 	// device workflow.
func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return nil, nil
}

// 	// InstanceInspect returns an instance.
func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	iid := context.MustInstanceID(ctx)
	if iid.ID != "" {
		return &types.Instance{InstanceID: iid}, nil
	}
	var rsSubnetID string
	if err := iid.UnmarshalMetadata(&rsSubnetID); err != nil {
		return nil, err
	}
	instanceID := &types.InstanceID{ID: rsSubnetID, Driver: d.Name()}
	return &types.Instance{InstanceID: instanceID}, nil
}

// 	// Volumes returns all volumes or a filtered list of volumes.
func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {
	// always return attachments to align against other drivers for now
	return d.getVolume(ctx, "", "", true)
}

// 	// VolumeInspect inspects a single volume.
func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {
	if volumeID == "" {
		return nil, goof.New("no volumeID specified")
	}

	vols, err := d.getVolume(ctx, volumeID, "", opts.Attachments)
	if err != nil {
		return nil, err
	}

	if vols == nil {
		return nil, nil
	}
	return vols[0], nil
}

// 	// VolumeCreate creates a new volume.
func (d *driver) VolumeCreate(ctx types.Context, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	return d.createVolume(ctx, volumeName, "", "", opts)
}

// 	// VolumeCreateFromSnapshot creates a new volume from an existing snapshot.
func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	return d.createVolume(ctx, volumeName, "", snapshotID, opts)

}

// 	// VolumeCopy copies an existing volume.
func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {
	volume, err := d.VolumeInspect(ctx, volumeID, &types.VolumeInspectOpts{})
	if err != nil {
		return nil,
			goof.New("error getting reference volume for volume copy")
	}

	volumeCreateOpts := &types.VolumeCreateOpts{
		Type:             &volume.Type,
		AvailabilityZone: &volume.AvailabilityZone,
	}

	return d.createVolume(ctx, volumeName, volumeID, "", volumeCreateOpts)
}

// 	// VolumeSnapshot snapshots a volume.
func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {

	fields := eff(map[string]interface{}{
		"moduleName":   ctx,
		"snapshotName": snapshotName,
		"volumeId":     volumeID,
	})

	createOpts := snapshots.CreateOpts{
		Name:     snapshotName,
		VolumeID: volumeID,
		Force:    true,
	}

	resp, err := snapshots.Create(d.clientBlockStorage, createOpts).Extract()
	if err != nil {
		return nil,
			goof.WithFieldsE(fields, "error creating snapshot", err)
	}

	log.Debug("waiting for snapshot creation to complete")
	d.waitSnapshotStatus(ctx, resp.ID)
	return translateSnapshot(resp), nil
}

// 	// VolumeRemove removes a volume.
func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {
	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})
	if volumeID == "" {
		return goof.WithFields(fields, "volumeId is required")
	}

	attached, err := d.volumeAttached(ctx, volumeID)
	if err != nil {
		return goof.WithFieldsE(fields, "error retrieving attachment status", err)
	}

	if attached {
		_, err := d.VolumeDetach(ctx, volumeID, &types.VolumeDetachOpts{})
		if err != nil {
			return goof.WithFieldsE(fields, "error detaching before volume removal", err)
		}
	}

	res := volumes.Delete(d.clientBlockStorage, volumeID)
	if res.Err != nil {
		return goof.WithFieldsE(fields, "error removing volume", res.Err)
	}

	return nil
}

// 	// VolumeAttach attaches a volume and provides a token clients can use
// 	// to validate that device has appeared locally.
func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {
	iid := context.MustInstanceID(ctx)
	fields := eff(map[string]interface{}{
		"volumeId":   volumeID,
		"instanceId": iid.ID,
	})

	if opts.Force {
		if _, err := d.VolumeDetach(ctx, volumeID,
			&types.VolumeDetachOpts{}); err != nil {
			return nil, "", err
		}
	}

	options := &volumeattach.CreateOpts{
		VolumeID: volumeID,
	}
	if opts.NextDevice != nil {
		options.Device = *opts.NextDevice
	}

	volumeAttach, err := volumeattach.Create(d.client, iid.ID, options).Extract()
	if err != nil {
		return nil, "", goof.WithFieldsE(
			fields, "error attaching volume", err)
	}

	ctx.WithFields(fields).Debug("waiting for volume to attach")
	volume, err := d.waitVolumeAttachStatus(ctx, volumeID, true)
	if err != nil {
		return nil, "", goof.WithFieldsE(
			fields, "error waiting for volume to attach", err)
	}
	return volume, volumeAttach.Device, nil
}

// 	// VolumeDetach detaches a volume.
func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {
	fields := eff(map[string]interface{}{
		"moduleName": ctx,
		"volumeId":   volumeID,
	})

	if volumeID == "" {
		return nil, goof.WithFields(fields, "volumeId is required for VolumeDetach")
	}
	vols, err := d.getVolume(ctx, volumeID, "", true)
	if err != nil {
		return nil, err
	}

	resp := volumeattach.Delete(
		d.client, vols[0].Attachments[0].InstanceID.ID, volumeID)
	if resp.Err != nil {
		return nil, goof.WithFieldsE(fields, "error detaching volume", resp.Err)
	}
	ctx.WithFields(fields).Debug("waiting for volume to detach")
	volume, err := d.waitVolumeAttachStatus(ctx, volumeID, false)
	if err == nil {
		return volume, nil
	}
	log.WithFields(fields).Debug("volume detached")
	return nil, nil
}

//  // Not a part of storage interface
// Not implemented in Anywhere???
func (d *driver) VolumeDetachAll(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {
	return nil
}

// 	// Snapshots returns all volumes or a filtered list of snapshots.
func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {
	allPages, err := snapshots.List(d.clientBlockStorage, nil).AllPages()
	if err != nil {
		return []*types.Snapshot{},
			goof.WithError("error listing volume snapshots", err)
	}
	allSnapshots, err := snapshots.ExtractSnapshots(allPages)
	if err != nil {
		return []*types.Snapshot{},
			goof.WithError("error listing volume snapshots", err)
	}

	var libstorageSnapshots []*types.Snapshot
	for _, snapshot := range allSnapshots {
		libstorageSnapshots = append(libstorageSnapshots, translateSnapshot(&snapshot))
	}

	return libstorageSnapshots, nil
}

// 	// SnapshotInspect inspects a single snapshot.
func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	fields := eff(map[string]interface{}{
		"snapshotId": snapshotID,
	})

	snapshot, err := snapshots.Get(d.clientBlockStorage, snapshotID).Extract()
	if err != nil {
		return nil,
			goof.WithFieldsE(fields, "error getting snapshot", err)
	}

	return translateSnapshot(snapshot), nil
}

// 	// SnapshotCopy copies an existing snapshot.
func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {

	// TODO
	return nil, types.ErrNotImplemented
}

// 	// SnapshotRemove removes a snapshot.
func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {
	resp := snapshots.Delete(d.clientBlockStorage, snapshotID)
	if resp.Err != nil {
		return goof.WithFieldsE(goof.Fields{
			"snapshotId": snapshotID}, "error removing snapshot", resp.Err)
	}

	return nil
}

///////////////////////////////////////////////////////////////////////
///// HELPER FUNCTIONS FOR RACKSPACE DRIVER FROM THIS POINT ON ////////
///////////////////////////////////////////////////////////////////////

func (d *driver) getAuthOptions() gophercloud.AuthOptions {
	return gophercloud.AuthOptions{
		IdentityEndpoint: d.authURL(),
		UserID:           d.userID(),
		Username:         d.userName(),
		Password:         d.password(),
		TenantID:         d.tenantID(),
		TenantName:       d.tenantName(),
		DomainID:         d.domainID(),
		DomainName:       d.domainName(),
	}
}

func (d *driver) getInstanceID() (string, error) {
	cmd := exec.Command("xenstore-read", "name")
	cmd.Env = d.config.EnvVars()
	cmdOut, err := cmd.Output()

	if err != nil {
		return "",
			goof.WithError("problem getting instance ID", err)
	}

	instanceID := strings.Replace(string(cmdOut), "\n", "", -1)

	validInstanceID := regexp.MustCompile(`^instance-`)
	valid := validInstanceID.MatchString(instanceID)
	if !valid {
		return "",
			goof.WithError("InstanceID not valid", err)
	}
	instanceID = strings.Replace(instanceID, "instance-", "", 1)
	return instanceID, nil
}

func (d *driver) getInstanceRegion() (string, error) {
	cmd := exec.Command("xenstore-read",
		"vm-data/provider_data/region")
	cmd.Env = d.config.EnvVars()
	cmdOut, err := cmd.Output()

	if err != nil {
		return "",
			goof.WithError("problem getting instance region", err)
		//return "",
		// goof.WithFields(eff(goof.Fields{
		// 	"moduleName": d.r.Context,
		// 	"cmd.Path":   cmd.Path,
		// 	"cmd.Args":   cmd.Args,
		// 	"cmd.Out":    cmdOut,
		// }), "error getting instance region")
	}

	region := strings.Replace(string(cmdOut), "\n", "", -1)
	return region, nil
}

func (d *driver) getVolume(ctx types.Context, volumeID, volumeName string,
	attachments bool) ([]*types.Volume, error) {
	var volumesRet []volumes.Volume
	fields := eff(goof.Fields{
		"moduleName": ctx,
		"volumeId":   volumeID,
		"volumeName": volumeName})

	if volumeID != "" {
		volume, err := volumes.Get(d.clientBlockStorage, volumeID).Extract()
		if err != nil {
			return nil,
				goof.WithFieldsE(fields, "error getting volumes", err)
		}
		volumesRet = append(volumesRet, *volume)
	} else {
		listOpts := &volumes.ListOpts{
		//Name:       volumeName,
		}

		allPages, err := volumes.List(d.clientBlockStorage, listOpts).AllPages()
		if err != nil {
			return nil,
				goof.WithFieldsE(fields, "error listing volumes", err)
		}
		volumesRet, err = volumes.ExtractVolumes(allPages)
		if err != nil {
			return nil,
				goof.WithFieldsE(fields, "error extracting volumes", err)
		}

		var volumesRetFiltered []volumes.Volume
		if volumeName != "" {
			for _, volumer := range volumesRet { //volumer avoids any namespace confict
				if volumer.Name == volumeName {
					volumesRetFiltered = append(volumesRetFiltered, volumer)
					break
				}
			}
			volumesRet = volumesRetFiltered
		}
	}
	//now cast from []volumes.Volume to []types.Volume
	var volumesSD []*types.Volume
	for _, volume := range volumesRet {
		volumesSD = append(volumesSD, translateVolume(&volume, attachments))
	}
	return volumesSD, nil
}

func createVolumeEnsureSize(size *int64) {
	if *size != 0 && *size < minSize {
		*size = minSize
	}
}

func (d *driver) createVolume(
	ctx types.Context,
	volumeName string,
	volumeSourceID string,
	snapshotID string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	var (
		volumeType       string
		IOPS             int64
		size             int64
		availabilityZone string
	)
	if opts.Type != nil {
		volumeType = *(opts.Type)
	}
	if opts.IOPS != nil {
		IOPS = *(opts.IOPS)
	}
	if opts.Size != nil {
		size = *(opts.Size)
	}
	if opts.AvailabilityZone != nil {
		availabilityZone = *(opts.AvailabilityZone)
	}

	//check some fields...
	createVolumeEnsureSize(&size)
	vsize := int(size)

	fields := map[string]interface{}{
		"availabilityZone": availabilityZone,
		"iops":             IOPS,
		"provider":         d.Name(),
		"size":             size,
		"snapshotId":       snapshotID,
		"volumeName":       volumeName,
		"volumeSourceID":   volumeSourceID,
		"volumeType":       volumeType,
	}

	options := &volumes.CreateOpts{
		Name:       volumeName,
		Size:       vsize,
		SnapshotID: snapshotID,
		VolumeType: volumeType,
		//AvailabilityZone: availabilityZone, //Not in old Rackspace
		//SourceReplica:    volumeSourceID,
	}
	resp, err := volumes.Create(d.clientBlockStorage, options).Extract()
	if err != nil {
		return nil,
			goof.WithFields(fields, "error creating volume")
	}
	fields["volumeId"] = resp.ID
	//for openstack must test before rackspace integration
	err = volumes.WaitForStatus(d.clientBlockStorage, resp.ID, "available", 120)
	if err != nil {
		return nil,
			goof.WithFieldsE(fields,
				"error waiting for volume creation to complete", err)
	}
	log.WithFields(fields).Debug("created volume")
	return translateVolume(resp, true), nil
}

//Reformats from volumes.Volume to types.Volume credit to github.com/MatMaul
func translateVolume(volume *volumes.Volume, includeAttachments bool) *types.Volume {
	var attachments []*types.VolumeAttachment
	if includeAttachments {
		for _, attachment := range volume.Attachments {
			libstorageAttachment := &types.VolumeAttachment{
				VolumeID: attachment["volume_id"].(string),
				InstanceID: &types.InstanceID{
					ID:     attachment["server_id"].(string),
					Driver: rackspace.Name},
				DeviceName: attachment["device"].(string),
				Status:     "",
			}
			attachments = append(attachments, libstorageAttachment)
		}
	} else {
		for _, attachment := range volume.Attachments {
			libstorageAttachment := &types.VolumeAttachment{
				VolumeID:   attachment["volume_id"].(string),
				InstanceID: &types.InstanceID{ID: attachment["server_id"].(string), Driver: rackspace.Name},
				DeviceName: "",
				Status:     "",
			}
			attachments = append(attachments, libstorageAttachment)
			break
		}
	}

	return &types.Volume{
		Name:             volume.Name,
		ID:               volume.ID,
		AvailabilityZone: volume.AvailabilityZone,
		Status:           volume.Status,
		Type:             volume.VolumeType,
		IOPS:             0,
		Size:             int64(volume.Size),
		Attachments:      attachments,
	}
}

//Reformats from snapshots.Snapshot to types.Snapshot credit to github.com/MatMaul
func translateSnapshot(snapshot *snapshots.Snapshot) *types.Snapshot {
	createAtEpoch := int64(0)
	createdAt, err := time.Parse(time.RFC3339Nano, snapshot.CreatedAt)
	if err == nil {
		createAtEpoch = createdAt.Unix()
	}
	return &types.Snapshot{
		Name:        snapshot.Name,
		VolumeID:    snapshot.VolumeID,
		ID:          snapshot.ID,
		VolumeSize:  int64(snapshot.Size),
		StartTime:   createAtEpoch,
		Description: snapshot.Description,
		Status:      snapshot.Status,
	}
}
func (d *driver) volumeAttached(ctx types.Context,
	volumeID string) (bool, error) {
	fields := eff(map[string]interface{}{
		"moduleName": ctx,
		"volumeId":   volumeID,
	})
	if volumeID == "" {
		return true, goof.WithFields(fields, "volumeId is required")
	}
	volume, err := d.VolumeInspect(ctx, volumeID, &types.VolumeInspectOpts{Attachments: true})
	if err != nil {
		return true, goof.WithFieldsE(fields, "error getting volume when waiting", err)
	}
	if len(volume.Attachments) > 0 {
		return true, nil
	}
	if len(volume.Attachments) == 0 {
		return false, nil
	}
	return true, goof.WithFields(fields, "check volume attachement status failed is required")
}

func (d *driver) waitSnapshotStatus(
	ctx types.Context, snapshotID string) error {
	if snapshotID == "" {
		return goof.New("Missing snapshot ID")
	}
	for {
		snapshot, err := d.SnapshotInspect(ctx, snapshotID, nil)
		if err != nil {
			return goof.WithError(
				"Error getting snapshot", err)
		}

		if snapshot.Status != "creating" {
			break
		}

		time.Sleep(30 * time.Second)
	}

	return nil
}

func (d *driver) waitVolumeAttachStatus(
	ctx types.Context,
	volumeID string,
	attachmentNeeded bool) (*types.Volume, error) {
	fields := eff(map[string]interface{}{
		"moduleName": ctx,
		"volumeId":   volumeID,
	})

	if volumeID == "" {
		return nil, goof.WithFields(fields, "volumeId is required")
	}

	for {
		volume, err := d.VolumeInspect(ctx, volumeID, &types.VolumeInspectOpts{Attachments: true})
		if err != nil {
			return nil, goof.WithFieldsE(fields, "error getting volume when waiting", err)
		}

		if attachmentNeeded {
			if len(volume.Attachments) > 0 {
				return volume, nil
			}
		} else {
			if len(volume.Attachments) == 0 {
				return volume, nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return nil, nil
}

//error reporting

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

// Plain Accessors

func (d *driver) authURL() string {
	return d.config.GetString("rackspace.authURL")
}

func (d *driver) userID() string {
	return d.config.GetString("rackspace.userID")
}

func (d *driver) userName() string {
	return d.config.GetString("rackspace.userName")
}

func (d *driver) password() string {
	return d.config.GetString("rackspace.password")
}

func (d *driver) tenantID() string {
	return d.config.GetString("rackspace.tenantID")
}

func (d *driver) tenantName() string {
	return d.config.GetString("rackspace.tenantName")
}

func (d *driver) domainID() string {
	return d.config.GetString("rackspace.domainID")
}

func (d *driver) domainName() string {
	return d.config.GetString("rackspace.domainName")
}
