package storage

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/cinder"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/volumeactions"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v1/snapshots"
	volumesv1 "github.com/gophercloud/gophercloud/openstack/blockstorage/v1/volumes"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/extensions/trusts"
)

const (
	minSizeGiB = 1
	hiddenText = "******"
)

var (
	errVolAlreadyAttached = goof.New("volume already attached to a host")
	errVolAlreadyDetached = goof.New("volume already detached")
)

type driver struct {
	provider             *gophercloud.ProviderClient
	clientCompute        *gophercloud.ServiceClient
	clientBlockStorage   *gophercloud.ServiceClient
	clientBlockStoragev2 *gophercloud.ServiceClient
	availabilityZone     string
	config               gofig.Config
}

func ef() goof.Fields {
	return goof.Fields{
		"provider": cinder.Name,
	}
}

func eff(fields goof.Fields) map[string]interface{} {
	errFields := map[string]interface{}{
		"provider": cinder.Name,
	}
	if fields != nil {
		for k, v := range fields {
			errFields[k] = v
		}
	}
	return errFields
}

func init() {
	registry.RegisterStorageDriver(cinder.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{}
}

func (d *driver) Name() string {
	return cinder.Name
}

func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Block, nil
}

func (d *driver) Init(context types.Context, config gofig.Config) error {
	d.config = config
	fields := eff(map[string]interface{}{})
	var err error

	endpointOpts := gophercloud.EndpointOpts{}

	endpointOpts.Region = d.regionName()
	fields["region"] = endpointOpts.Region

	d.availabilityZone = d.availabilityZoneName()
	fields["availabilityZone"] = d.availabilityZone

	authOpts := d.getAuthOptions()

	fields["identityEndpoint"] = d.authURL()
	fields["userId"] = d.userID()
	fields["userName"] = d.userName()
	if d.password() == "" {
		fields["password"] = ""
	} else {
		fields["password"] = hiddenText
	}
	if d.tokenID() == "" {
		fields["tokenId"] = ""
	} else {
		fields["tokenId"] = hiddenText
	}
	fields["tenantId"] = d.tenantID()
	fields["tenantName"] = d.tenantName()
	fields["domainId"] = d.domainID()
	fields["domainName"] = d.domainName()

	trustID := d.trustID()
	if trustID == "" {
		fields["trustId"] = ""
	} else {
		fields["trustId"] = hiddenText
	}

	fields["caCert"] = d.caCert()
	fields["insecure"] = d.insecure()

	d.provider, err = openstack.NewClient(authOpts.IdentityEndpoint)
	if err != nil {
		return goof.WithFieldsE(fields, "error creating Keystone client", err)
	}

	d.provider.HTTPClient, err = openstackHTTPClient(d.caCert(), d.insecure())
	if err != nil {
		return goof.WithFieldsE(fields, "error overriding Gophercloud HTTP client", err)
	}

	if trustID != "" {
		authOptionsExt := trusts.AuthOptsExt{
			TrustID:            trustID,
			AuthOptionsBuilder: &authOpts,
		}
		err = openstack.AuthenticateV3(d.provider, authOptionsExt, endpointOpts)
	} else {
		err = openstack.Authenticate(d.provider, authOpts)
	}
	if err != nil {
		return goof.WithFieldsE(fields, "error authenticating", err)
	}

	if d.clientCompute, err = openstack.NewComputeV2(d.provider, endpointOpts); err != nil {
		return goof.WithFieldsE(fields, "error getting newComputeV2", err)
	}

	if d.clientBlockStorage, err = openstack.NewBlockStorageV1(d.provider, endpointOpts); err != nil {
		return goof.WithFieldsE(fields, "error getting newBlockStorageV1", err)
	}

	d.clientBlockStoragev2, err = openstack.NewBlockStorageV2(d.provider, endpointOpts)
	if err != nil {
		// fallback to volume v1
		context.WithFields(fields).Info("BlockStorage API V2 not available, fallback to V1")
		d.clientBlockStoragev2 = nil
	}

	context.WithFields(fields).Info("storage driver initialized")

	return nil
}

func openstackHTTPClient(caCert string, insecure bool) (http.Client, error) {
	if caCert == "" {
		return http.Client{}, nil
	}

	caCertPool := x509.NewCertPool()
	caCertContent, err := ioutil.ReadFile(caCert)
	if err != nil {
		return http.Client{}, errors.New("Can't read certificate file")
	}
	caCertPool.AppendCertsFromPEM(caCertContent)

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: insecure,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	return http.Client{Transport: transport}, nil
}

// InstanceInspect returns an instance.
func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	return &types.Instance{
		InstanceID: context.MustInstanceID(ctx),
	}, nil
}

func (d *driver) getAuthOptions() gophercloud.AuthOptions {
	return gophercloud.AuthOptions{
		IdentityEndpoint: d.authURL(),
		UserID:           d.userID(),
		Username:         d.userName(),
		Password:         d.password(),
		TokenID:          d.tokenID(),
		TenantID:         d.tenantID(),
		TenantName:       d.tenantName(),
		DomainID:         d.domainID(),
		DomainName:       d.domainName(),
		AllowReauth:      true,
	}
}

func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	if d.clientBlockStoragev2 != nil {
		allPages, err := volumes.List(d.clientBlockStoragev2, nil).AllPages()
		if err != nil {
			return nil,
				goof.WithError("error listing volumes", err)
		}
		volumesOS, err := volumes.ExtractVolumes(allPages)
		if err != nil {
			return nil,
				goof.WithError("error listing volumes", err)
		}

		var volumesRet []*types.Volume
		for _, volumeOS := range volumesOS {
			volumesRet = append(volumesRet, translateVolume(
				&volumeOS, opts.Attachments))
		}

		return volumesRet, nil
	}

	allPages, err := volumesv1.List(d.clientBlockStorage, nil).AllPages()
	if err != nil {
		return nil,
			goof.WithError("error listing volumes", err)
	}
	volumesOS, err := volumesv1.ExtractVolumes(allPages)
	if err != nil {
		return nil,
			goof.WithError("error listing volumes", err)
	}

	var volumesRet []*types.Volume
	for _, volumeOS := range volumesOS {
		volumesRet = append(volumesRet, translateVolumeV1(
			&volumeOS, opts.Attachments))
	}

	return volumesRet, nil
}

func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	fields := eff(goof.Fields{
		"volumeId": volumeID,
	})

	if volumeID == "" {
		return nil, goof.New("no volumeID specified")
	}

	if d.clientBlockStoragev2 != nil {
		volume, err := volumes.Get(d.clientBlockStoragev2, volumeID).Extract()

		if err != nil {
			return nil,
				goof.WithFieldsE(fields, "error getting volume", err)
		}

		return translateVolume(volume, opts.Attachments), nil
	}

	volume, err := volumesv1.Get(d.clientBlockStorage, volumeID).Extract()

	if err != nil {
		return nil,
			goof.WithFieldsE(fields, "error getting volume", err)
	}

	return translateVolumeV1(volume, opts.Attachments), nil
}

func translateVolumeV1(
	volume *volumesv1.Volume,
	includeAttachments types.VolumeAttachmentsTypes) *types.Volume {

	var attachments []*types.VolumeAttachment
	if includeAttachments.Requested() {
		for _, attachment := range volume.Attachments {
			libstorageAttachment := &types.VolumeAttachment{
				VolumeID:   attachment["volume_id"].(string),
				InstanceID: &types.InstanceID{ID: attachment["server_id"].(string), Driver: cinder.Name},
				DeviceName: attachment["device"].(string),
				Status:     "",
			}
			attachments = append(attachments, libstorageAttachment)
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

func translateVolume(
	volume *volumes.Volume,
	includeAttachments types.VolumeAttachmentsTypes) *types.Volume {

	var attachments []*types.VolumeAttachment
	if includeAttachments.Requested() {
		for _, attachment := range volume.Attachments {
			libstorageAttachment := &types.VolumeAttachment{
				VolumeID:   attachment.VolumeID,
				InstanceID: &types.InstanceID{ID: attachment.ServerID, Driver: cinder.Name},
				DeviceName: attachment.Device,
				Status:     "",
			}
			attachments = append(attachments, libstorageAttachment)
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

func translateSnapshot(snapshot *snapshots.Snapshot) *types.Snapshot {
	return &types.Snapshot{
		Name:        snapshot.Name,
		VolumeID:    snapshot.VolumeID,
		ID:          snapshot.ID,
		VolumeSize:  int64(snapshot.Size),
		StartTime:   time.Time(snapshot.CreatedAt).Unix(),
		Description: snapshot.Description,
		Status:      snapshot.Status,
	}
}

func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {

	fields := eff(map[string]interface{}{
		"snapshotName": snapshotName,
		"volumeId":     volumeID,
	})

	createOpts := snapshots.CreateOpts{
		Name:     snapshotName,
		VolumeID: volumeID,
		Force:    true,
	}

	snapshot, err := snapshots.Create(d.clientBlockStorage, createOpts).Extract()
	if err != nil {
		return nil,
			goof.WithFieldsE(fields, "error creating snapshot", err)
	}

	ctx.WithFields(fields).Info("waiting for snapshot creation to complete")

	err = snapshots.WaitForStatus(d.clientBlockStorage, snapshot.ID, "available", int(d.snapshotTimeout().Seconds()))
	if err != nil {
		return nil,
			goof.WithFieldsE(fields,
				"error waiting for snapshot creation to complete", err)
	}

	return translateSnapshot(snapshot), nil

}

func waitFor404(c *gophercloud.ServiceClient, url string, secs int) error {
	return gophercloud.WaitFor(secs, func() (bool, error) {
		ret, err := c.Get(url, nil, nil)

		if err != nil {
			if ret != nil && ret.StatusCode == 404 {
				return true, nil
			}
			return false, err
		}

		return false, nil
	})
}

func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {
	resp := snapshots.Delete(d.clientBlockStorage, snapshotID)
	if resp.Err != nil {
		return goof.WithFieldE("snapshotId", snapshotID, "error removing snapshot", resp.Err)
	}

	err := waitFor404(d.clientBlockStorage, d.clientBlockStorage.ServiceURL("snapshots", snapshotID), int(d.deleteTimeout().Seconds()))
	if err != nil {
		return goof.WithFieldE("snapshotId", snapshotID, "error waiting for snapshot removal", err)
	}
	return nil
}

func (d *driver) VolumeCreate(ctx types.Context, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	return d.createVolume(ctx, volumeName, "", "", opts)
}

func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	return d.createVolume(ctx, volumeName, "", snapshotID, opts)
}

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

func (d *driver) createVolume(
	ctx types.Context,
	volumeName string,
	volumeSourceID string,
	snapshotID string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	options := &volumes.CreateOpts{
		Name:          volumeName,
		SnapshotID:    snapshotID,
		SourceReplica: volumeSourceID,
	}

	fields := eff(map[string]interface{}{
		"volumeName":     volumeName,
		"snapshotId":     snapshotID,
		"volumeSourceId": volumeSourceID,
	})

	if opts.Type != nil {
		options.VolumeType = *opts.Type
	}
	fields["volumeType"] = options.VolumeType

	if opts.AvailabilityZone != nil {
		options.AvailabilityZone = *opts.AvailabilityZone
	}
	if options.AvailabilityZone == "" {
		options.AvailabilityZone = d.availabilityZone
	}
	fields["availabilityZone"] = options.AvailabilityZone

	if opts.Size == nil {
		size := int64(minSizeGiB)
		opts.Size = &size
	}

	if *opts.Size < minSizeGiB {
		fields["minSize"] = minSizeGiB
		return nil, goof.WithFields(fields, "volume size too small")
	}

	options.Size = int(*opts.Size)
	fields["size"] = *opts.Size

	if d.clientBlockStoragev2 != nil {
		volume, err := volumes.Create(d.clientBlockStoragev2, options).Extract()
		if err != nil {
			return nil,
				goof.WithFieldsE(fields, "error creating volume", err)
		}

		fields["volumeId"] = volume.ID

		ctx.WithFields(fields).Info("waiting for volume creation to complete")
		err = volumes.WaitForStatus(d.clientBlockStoragev2, volume.ID, "available", int(d.createTimeout().Seconds()))
		if err != nil {
			return nil,
				goof.WithFieldsE(fields,
					"error waiting for volume creation to complete", err)
		}

		return translateVolume(volume, types.VolumeAttachmentsRequested), nil
	}

	volume, err := volumesv1.Create(d.clientBlockStorage, options).Extract()
	if err != nil {
		return nil,
			goof.WithFieldsE(fields, "error creating volume", err)
	}

	fields["volumeId"] = volume.ID

	ctx.WithFields(fields).Info("waiting for volume creation to complete")
	err = volumesv1.WaitForStatus(d.clientBlockStorage, volume.ID, "available", int(d.createTimeout().Seconds()))
	if err != nil {
		return nil,
			goof.WithFieldsE(fields,
				"error waiting for volume creation to complete", err)
	}

	return translateVolumeV1(volume, types.VolumeAttachmentsRequested), nil
}

func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeRemoveOpts) error {
	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})
	if volumeID == "" {
		return goof.WithFields(fields, "volumeId is required")
	}

	volumesClient := d.clientBlockStorage
	var res gophercloud.ErrResult
	if d.clientBlockStoragev2 != nil {
		volumesClient = d.clientBlockStoragev2
		res = volumes.Delete(d.clientBlockStoragev2, volumeID).ErrResult
	} else {
		res = volumesv1.Delete(d.clientBlockStorage, volumeID).ErrResult
	}

	if res.Err != nil {
		return goof.WithFieldsE(fields, "error removing volume", res.Err)
	}
	err := waitFor404(volumesClient, volumesClient.ServiceURL("volumes", volumeID), int(d.deleteTimeout().Seconds()))
	if err != nil {
		return goof.WithFieldsE(fields, "error waiting for volume removal", err)
	}

	return nil
}

func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return nil, nil
}

func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	iid := context.MustInstanceID(ctx)

	fields := eff(map[string]interface{}{
		"volumeId":   volumeID,
		"instanceId": iid.ID,
	})

	// Get the volume
	vol, err := d.VolumeInspect(
		ctx,
		volumeID,
		&types.VolumeInspectOpts{Attachments: types.VolAttReq})
	if err != nil {
		return nil, "", err
	}

	// Check if volume is already attached
	if len(vol.Attachments) > 0 {
		// Detach already attached volume if forced
		if !opts.Force {
			return nil, "", errVolAlreadyAttached
		}
		_, err := d.volumeDetach(ctx, vol)
		if err != nil {
			return nil, "", goof.WithError("error detaching volume", err)
		}
	}

	options := &volumeattach.CreateOpts{
		VolumeID: volumeID,
	}
	if opts.NextDevice != nil {
		options.Device = *opts.NextDevice
	}

	volumeAttach, err := volumeattach.Create(d.clientCompute, iid.ID, options).Extract()
	if err != nil {
		return nil, "", goof.WithFieldsE(
			fields, "error attaching volume", err)
	}

	ctx.WithFields(fields).Debug("waiting for volume to attach")
	volume, err := d.waitVolumeAttachStatus(ctx, volumeID, true, d.attachTimeout())
	if err != nil {
		return nil, "", goof.WithFieldsE(
			fields, "error waiting for volume to attach", err)
	}

	return volume, volumeAttach.Device, nil
}

func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})

	if volumeID == "" {
		return nil, goof.WithFields(fields, "volumeId is required")
	}

	// Get the volume
	vol, err := d.VolumeInspect(
		ctx,
		volumeID,
		&types.VolumeInspectOpts{Attachments: types.VolAttReq})
	if err != nil {
		return nil, err
	}

	return d.volumeDetach(ctx, vol)
}

func (d *driver) volumeDetach(
	ctx types.Context,
	vol *types.Volume) (*types.Volume, error) {

	fields := map[string]interface{}{
		"volumeID": vol.ID,
	}

	var (
		volume *types.Volume
		err    error
	)

	if len(vol.Attachments) == 0 {
		return nil, errVolAlreadyDetached
	}

	for _, att := range vol.Attachments {
		delResp := volumeattach.Delete(
			d.clientCompute, att.InstanceID.ID, vol.ID)
		if delResp.Err != nil {
			return nil, goof.WithFieldsE(
				fields, "error detaching volume", delResp.Err)
		}

		ctx.WithFields(fields).Debug("waiting for volume to detach")
		volume, err = d.waitVolumeAttachStatus(ctx, vol.ID, false,
			d.attachTimeout())
		if err == nil {
			continue
		}

		// If an error occured, try the v2 client if present
		if d.clientBlockStoragev2 == nil {
			return nil, goof.WithFieldsE(
				fields, "error waiting for volume to detach", err)
		}

		detResp := volumeactions.Detach(
			d.clientBlockStoragev2, vol.ID, volumeactions.DetachOpts{})
		if detResp.Err != nil {
			return nil, goof.WithFieldsE(
				fields, "error detaching volume", detResp.Err)
		}

		volume, err = d.waitVolumeAttachStatus(ctx, vol.ID, false,
			d.attachTimeout())
		if err != nil {
			return nil, goof.WithFieldsE(
				fields, "error waiting for volume to detach", err)
		}
	}
	return volume, nil
}

func (d *driver) waitVolumeAttachStatus(
	ctx types.Context, volumeID string,
	attachmentNeeded bool, timeout time.Duration) (*types.Volume, error) {

	fields := eff(map[string]interface{}{
		"volumeId": volumeID,
	})

	if volumeID == "" {
		return nil, goof.WithFields(fields, "volumeId is required")
	}
	begin := time.Now()
	for time.Now().Sub(begin) < timeout {
		volume, err := d.VolumeInspect(
			ctx, volumeID, &types.VolumeInspectOpts{
				Attachments: types.VolumeAttachmentsRequested})
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

	return nil, goof.WithFields(fields, "timeout reached")
}

func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO return nil, nil ?
	return nil, types.ErrNotImplemented
}

func (d *driver) authURL() string {
	return d.config.GetString(cinder.ConfigAuthURL)
}

func (d *driver) userID() string {
	return d.config.GetString(cinder.ConfigUserID)
}

func (d *driver) userName() string {
	return d.config.GetString(cinder.ConfigUserName)
}

func (d *driver) password() string {
	return d.config.GetString(cinder.ConfigPassword)
}

func (d *driver) tokenID() string {
	return d.config.GetString(cinder.ConfigTokenID)
}

func (d *driver) tenantID() string {
	return d.config.GetString(cinder.ConfigTenantID)
}

func (d *driver) tenantName() string {
	return d.config.GetString(cinder.ConfigTenantName)
}

func (d *driver) domainID() string {
	return d.config.GetString(cinder.ConfigDomainID)
}

func (d *driver) domainName() string {
	return d.config.GetString(cinder.ConfigDomainName)
}

func (d *driver) regionName() string {
	return d.config.GetString(cinder.ConfigRegionName)
}

func (d *driver) availabilityZoneName() string {
	return d.config.GetString(cinder.ConfigAvailabilityZoneName)
}

func (d *driver) trustID() string {
	return d.config.GetString(cinder.ConfigTrustID)
}

func (d *driver) attachTimeout() time.Duration {
	strVal := d.config.GetString(cinder.ConfigAttachTimeout)
	val, err := time.ParseDuration(strVal)

	if err != nil || val <= 0 {
		val = 1 * time.Minute
	}
	return val
}

func (d *driver) deleteTimeout() time.Duration {
	strVal := d.config.GetString(cinder.ConfigDeleteTimeout)
	val, err := time.ParseDuration(strVal)

	if err != nil || val <= 0 {
		val = 10 * time.Minute
	}
	return val
}

func (d *driver) createTimeout() time.Duration {
	strVal := d.config.GetString(cinder.ConfigCreateTimeout)
	val, err := time.ParseDuration(strVal)

	if err != nil || val <= 0 {
		val = 10 * time.Minute
	}
	return val
}

func (d *driver) snapshotTimeout() time.Duration {
	strVal := d.config.GetString(cinder.ConfigSnapshotTimeout)
	val, err := time.ParseDuration(strVal)

	if err != nil || val <= 0 {
		val = 10 * time.Minute
	}
	return val
}

func (d *driver) caCert() string {
	return d.config.GetString(cinder.ConfigCACert)
}

func (d *driver) insecure() bool {
	return d.config.GetBool(cinder.ConfigInsecure)
}
