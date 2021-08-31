package storage

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"hash"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	gofig "github.com/akutz/gofig/types"
	goof "github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	apiUtils "github.com/AVENTER-UG/rexray/libstorage/api/utils"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/gcepd"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/gcepd/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/googleapi"
)

const (
	cacheKeyC     = "cacheKey"
	tagKey        = "libstoragetag"
	minDiskSizeGB = 10
)

var (
	// GCE labels have to start with a lowercase letter, and have to end
	// with a lowercase letter or numeral. In between can be lowercase
	// letters, numbers or dashes
	tagRegex = regexp.MustCompile(`^[a-z](?:[a-z0-9\-]*[a-z0-9])?$`)
)

type driver struct {
	config          gofig.Config
	keyFile         string
	projectID       *string
	zone            string
	defaultDiskType string
	tag             string
	tokenSource     oauth2.TokenSource
	svcAccount      string
	maxAttempts     int
	statusDelay     int64
	statusTimeout   time.Duration
	convUnderscore  bool
}

func init() {
	registry.RegisterStorageDriver(gcepd.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{}
}

func (d *driver) Name() string {
	return gcepd.Name
}

// Init initializes the driver.
func (d *driver) Init(context types.Context, config gofig.Config) error {
	d.config = config

	d.keyFile = d.config.GetString(gcepd.ConfigKeyfile)
	if d.keyFile != "" {
		if !gotil.FileExists(d.keyFile) {
			return goof.Newf("keyfile at %s does not exist", d.keyFile)
		}
		pID, err := d.extractProjectID()
		if err != nil || pID == nil || *pID == "" {
			return goof.New("Unable to set project ID from keyfile")
		}
		d.projectID = pID
		context.Info("Will authenticate using local JSON credentials")
	} else {
		// We are using application default credentials
		defCreds, err := google.FindDefaultCredentials(
			context, compute.ComputeScope)
		if err != nil {
			return goof.WithError(
				"Unable to get application default credentials",
				err)
		}
		d.projectID = &defCreds.ProjectID
		if *d.projectID == "" {
			return goof.New(
				"Unable to get project ID from default creds")
		}
		d.tokenSource = defCreds.TokenSource
		context.Info("Will authenticate using app default credentials")
	}

	d.zone = d.config.GetString(gcepd.ConfigZone)
	if d.zone != "" {
		context.Infof("All access is restricted to zone: %s", d.zone)
	}

	d.defaultDiskType = config.GetString(gcepd.ConfigDefaultDiskType)

	switch d.defaultDiskType {
	case gcepd.DiskTypeSSD, gcepd.DiskTypeStandard:
		// noop
	case "":
		d.defaultDiskType = gcepd.DefaultDiskType
	default:
		return goof.Newf(
			"Invalid GCE disk type: %s", d.defaultDiskType)
	}

	d.tag = config.GetString(gcepd.ConfigTag)
	if d.tag != "" && !tagRegex.MatchString(d.tag) {
		return goof.New("Invalid GCE tag format")
	}

	d.maxAttempts = d.config.GetInt(gcepd.ConfigStatusMaxAttempts)

	statusDelayStr := d.config.GetString(gcepd.ConfigStatusInitDelay)
	statusDelay, err := time.ParseDuration(statusDelayStr)
	if err != nil {
		return err
	}
	d.statusDelay = statusDelay.Nanoseconds()

	statusTimeoutStr := d.config.GetString(gcepd.ConfigStatusTimeout)
	d.statusTimeout, err = time.ParseDuration(statusTimeoutStr)
	if err != nil {
		return err
	}

	d.convUnderscore = d.config.GetBool(gcepd.ConfigConvertUnderscores)

	context.Info("storage driver initialized")
	return nil
}

var (
	sessions  = map[string]*compute.Service{}
	sessionsL = &sync.Mutex{}
)

func writeHkey(h hash.Hash, ps *string) {
	if ps == nil {
		return
	}
	h.Write([]byte(*ps))
}

func (d *driver) Login(ctx types.Context) (interface{}, error) {
	sessionsL.Lock()
	defer sessionsL.Unlock()

	var (
		ckey   string
		hkey   = md5.New()
		client *http.Client
	)

	// Unique connections to google APIs are based on project ID
	// optionally there may be an additional service account
	writeHkey(hkey, d.projectID)
	if d.svcAccount != "" {
		writeHkey(hkey, &d.svcAccount)
	}
	ckey = fmt.Sprintf("%x", hkey.Sum(nil))

	// if the session is cached then return it
	if svc, ok := sessions[ckey]; ok {
		ctx.WithField(cacheKeyC, ckey).Debug("using cached gce service")
		return svc, nil
	}

	fields := map[string]interface{}{
		cacheKeyC:   ckey,
		"keyfile":   d.keyFile,
		"projectID": *d.projectID,
	}

	if d.keyFile != "" {
		serviceAccountJSON, err := d.getKeyFileJSON()
		if err != nil {
			ctx.WithFields(fields).Errorf(
				"Could not read service account credentials file: %s",
				err)
			return nil, err
		}

		config, err := google.JWTConfigFromJSON(
			serviceAccountJSON,
			compute.ComputeScope,
		)
		if err != nil {
			ctx.WithFields(fields).Errorf(
				"Could not create JWT Config From JSON: %s", err)
			return nil, err
		}
		d.svcAccount = config.Email
		writeHkey(hkey, &config.Email)
		ckey = fmt.Sprintf("%x", hkey.Sum(nil))
		client = config.Client(ctx)
	} else {
		// Using application default credentials
		if d.tokenSource == nil {
			return nil, goof.New("Token Source is nil")
		}

		client = oauth2.NewClient(ctx, d.tokenSource)
	}

	svc, err := compute.New(client)
	if err != nil {
		ctx.WithFields(fields).Errorf(
			"Could not create GCE service connection: %s", err)
		return nil, err

	}

	sessions[ckey] = svc
	ctx.Info("GCE service connection created and cached")
	return svc, nil
}

// NextDeviceInfo returns the information about the driver's next available
// device workflow.
func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return nil, nil
}

// Type returns the type of storage the driver provides.
func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Block, nil
}

// InstanceInspect returns an instance.
func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	iid := context.MustInstanceID(ctx)
	return &types.Instance{
		InstanceID: iid,
	}, nil
}

// Volumes returns all volumes or a filtered list of volumes.
func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	var gceDisks []*compute.Disk
	var err error
	vols := []*types.Volume{}

	zone, err := d.validZone(ctx)
	if err != nil {
		return nil, err
	}

	if zone != nil && *zone != "" {
		// get list of disks in zone from GCE
		gceDisks, err = d.getDisks(ctx, zone)
	} else {
		// without a zone, get disks in all zones
		gceDisks, err = d.getAggregatedDisks(ctx)
	}

	if err != nil {
		ctx.Errorf("Unable to get disks from GCE API")
		return nil, goof.WithError(
			"Unable to get disks from GCE API", err)
	}

	// shortcut early if nothing is returned
	// TODO: is it an error if there are no volumes? EBS driver returns an
	// error, but other drivers (ScaleIO, RBD) return an empty slice
	if len(gceDisks) == 0 {
		return vols, nil
	}

	// convert GCE disks to libstorage types.Volume
	vols, err = d.toTypeVolume(ctx, gceDisks, opts.Attachments, zone)
	if err != nil {
		return nil, goof.WithError("error converting to types.Volume",
			err)
	}

	return vols, nil
}

// VolumeInspect inspects a single volume by ID.
func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	zone, err := d.validZone(ctx)
	if err != nil {
		return nil, err
	}

	if zone == nil || *zone == "" {
		return nil, goof.New("Zone is required for VolumeInspect")
	}

	gceDisk, err := d.getDisk(ctx, zone, &volumeID)
	if err != nil {
		ctx.Errorf("Unable to get disk from GCE API")
		return nil, goof.WithError(
			"Unable to get disk from GCE API", err)
	}
	if gceDisk == nil {
		return nil, apiUtils.NewNotFoundError(volumeID)
	}

	gceDisks := []*compute.Disk{gceDisk}
	vols, err := d.toTypeVolume(ctx, gceDisks, opts.Attachments, zone)
	if err != nil {
		return nil, goof.WithError("error converting to types.Volume",
			err)
	}

	return vols[0], nil
}

// VolumeInspectByName inspects a single volume by name.
func (d *driver) VolumeInspectByName(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	volumeName = d.convUnderscores(volumeName)

	// For GCE, name and ID are the same
	return d.VolumeInspect(
		ctx,
		volumeName,
		opts,
	)
}

// VolumeCreate creates a new volume.
func (d *driver) VolumeCreate(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	fields := map[string]interface{}{
		"driverName": d.Name(),
		"volumeName": volumeName,
		"opts":       opts,
	}

	zone, err := d.validZone(ctx)
	if err != nil {
		return nil, err
	}

	// No zone information from request, driver, or IID
	if (zone == nil || *zone == "") && (opts.AvailabilityZone == nil || *opts.AvailabilityZone == "") {
		return nil, goof.New("Zone is required for VolumeCreate")
	}

	if zone != nil && *zone != "" {
		if opts.AvailabilityZone != nil && *opts.AvailabilityZone != "" {
			// If request and driver/IID have zone, they must match
			if *zone != *opts.AvailabilityZone {
				return nil, goof.WithFields(fields,
					"Cannot create volume in given zone")
			}
		} else {
			// Set the zone to the driver/IID config
			opts.AvailabilityZone = zone
		}

	}

	volumeName = d.convUnderscores(volumeName)
	fields["volumeName"] = volumeName
	if !utils.IsValidDiskName(&volumeName) {
		return nil, goof.WithFields(fields,
			"Volume name does not meet GCE naming requirements")
	}

	if opts.Size == nil {
		size := int64(minDiskSizeGB)
		opts.Size = &size
	}

	fields["size"] = *opts.Size

	if *opts.Size < minDiskSizeGB {
		fields["minSize"] = minDiskSizeGB
		return nil, goof.WithFields(fields, "volume size too small")
	}

	ctx.WithFields(fields).Debug("creating volume")

	// Check if volume with same name exists
	gceDisk, err := d.getDisk(ctx, opts.AvailabilityZone, &volumeName)
	if err != nil {
		return nil, goof.WithFieldsE(fields,
			"error querying for existing volume", err)
	}
	if gceDisk != nil {
		return nil, goof.WithFields(fields,
			"volume name already exists")
	}

	err = d.createVolume(ctx, &volumeName, opts)
	if err != nil {
		return nil, goof.WithFieldsE(
			fields, "error creating volume", err)
	}

	// Return the volume created
	return d.VolumeInspect(ctx, volumeName,
		&types.VolumeInspectOpts{
			Attachments: types.VolAttNone,
		},
	)
}

// VolumeCreateFromSnapshot creates a new volume from an existing snapshot.
func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	return nil, types.ErrNotImplemented
}

// VolumeCopy copies an existing volume.
func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {

	return nil, types.ErrNotImplemented
}

// VolumeSnapshot snapshots a volume.
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
	opts *types.VolumeRemoveOpts) error {

	zone, err := d.validZone(ctx)
	if err != nil {
		return err
	}

	if zone == nil || *zone == "" {
		return goof.New("Zone is required for VolumeRemove")
	}

	// TODO: check if disk is still attached first
	asyncOp, err := mustSession(ctx).Disks.Delete(
		*d.projectID, *zone, volumeID).Do()
	if err != nil {
		return goof.WithError("Failed to initiate disk deletion", err)
	}

	err = d.waitUntilOperationIsFinished(
		ctx, zone, asyncOp)
	if err != nil {
		return err
	}

	return nil
}

// VolumeAttach attaches a volume and provides a token clients can use
// to validate that device has appeared locally.
func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	zone, err := d.validZone(ctx)
	if err != nil {
		return nil, "", err
	}

	if zone == nil || *zone == "" {
		return nil, "", goof.New("Zone is required for VolumeAttach")
	}

	instanceName := context.MustInstanceID(ctx).ID
	gceInst, err := d.getInstance(ctx, zone, &instanceName)
	if err != nil {
		return nil, "", err
	}
	if gceInst == nil {
		return nil, "", goof.New("Instance to attach to not found")
	}

	// Check if volume is already attached somewhere, if so, force detach?
	gceDisk, err := d.getDisk(ctx, zone, &volumeID)
	if err != nil {
		return nil, "", err
	}
	if gceDisk == nil {
		return nil, "", apiUtils.NewNotFoundError(volumeID)
	}

	if len(gceDisk.Users) > 0 {
		if !opts.Force {
			return nil, "", goof.New(
				"Volume already attached to different host")
		}
		ctx.Info("Automatically detaching volume from other instance")
		err = d.detachVolume(ctx, gceDisk)
		if err != nil {
			return nil, "", goof.WithError(
				"Error detaching volume during force attach",
				err)
		}
	}

	err = d.attachVolume(ctx, &instanceName, zone, &volumeID)
	if err != nil {
		return nil, "", err
	}

	vol, err := d.VolumeInspect(
		ctx, volumeID, &types.VolumeInspectOpts{
			Attachments: types.VolAttReq,
			Opts:        opts.Opts,
		},
	)
	if err != nil {
		return nil, "", goof.WithError("Error getting volume", err)
	}

	return vol, volumeID, nil
}

// VolumeDetach detaches a volume.
func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	zone, err := d.validZone(ctx)
	if err != nil {
		return nil, err
	}

	if zone == nil || *zone == "" {
		return nil, goof.New("Zone is required for VolumeDetach")
	}

	// Check if volume is attached at all
	gceDisk, err := d.getDisk(ctx, zone, &volumeID)
	if err != nil {
		return nil, err
	}
	if gceDisk == nil {
		return nil, apiUtils.NewNotFoundError(volumeID)
	}

	if len(gceDisk.Users) == 0 {
		return nil, goof.New("Volume already detached")
	}

	err = d.detachVolume(ctx, gceDisk)
	if err != nil {
		return nil, goof.WithError("Error detaching disk", err)
	}

	vol, err := d.VolumeInspect(
		ctx, volumeID, &types.VolumeInspectOpts{
			Attachments: types.VolAttReq,
			Opts:        opts.Opts,
		},
	)
	if err != nil {
		return nil, goof.WithError("Error getting volume", err)
	}

	return vol, nil
}

// Snapshots returns all volumes or a filtered list of snapshots.
func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {

	return nil, types.ErrNotImplemented
}

// SnapshotInspect inspects a single snapshot.
func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {

	return nil, types.ErrNotImplemented
}

// SnapshotCopy copies an existing snapshot.
func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {

	return nil, types.ErrNotImplemented
}

// SnapshotRemove removes a snapshot.
func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {

	return types.ErrNotImplemented
}

///////////////////////////////////////////////////////////////////////
/////////        HELPER FUNCTIONS SPECIFIC TO PROVIDER        /////////
///////////////////////////////////////////////////////////////////////

func (d *driver) getKeyFileJSON() ([]byte, error) {
	serviceAccountJSON, err := ioutil.ReadFile(d.keyFile)
	if err != nil {
		log.Errorf(
			"Could not read credentials file: %s, %s",
			d.keyFile, err)
		return nil, err
	}
	return serviceAccountJSON, nil
}

type keyData struct {
	ProjectID string `json:"project_id"`
}

func (d *driver) extractProjectID() (*string, error) {
	keyJSON, err := d.getKeyFileJSON()
	if err != nil {
		return nil, err
	}

	data := keyData{}

	err = json.Unmarshal(keyJSON, &data)
	if err != nil {
		return nil, err
	}

	return &data.ProjectID, nil
}

func getClientProjectID(ctx types.Context) (*string, bool) {
	if iid, ok := context.InstanceID(ctx); ok {
		if v, ok := iid.Fields[gcepd.InstanceIDFieldProjectID]; ok {
			return &v, v != ""
		}
	}
	return nil, false
}

func getClientZone(ctx types.Context) (*string, bool) {
	if iid, ok := context.InstanceID(ctx); ok {
		if v, ok := iid.Fields[gcepd.InstanceIDFieldZone]; ok {
			return &v, v != ""
		}
	}
	return nil, false

}

func mustSession(ctx types.Context) *compute.Service {
	return context.MustSession(ctx).(*compute.Service)
}

func (d *driver) getDisks(
	ctx types.Context,
	zone *string) ([]*compute.Disk, error) {

	diskListQ := mustSession(ctx).Disks.List(*d.projectID, *zone)
	if d.tag != "" {
		filter := fmt.Sprintf("labels.%s eq %s", tagKey, d.tag)
		ctx.Debugf("query filter: %s", filter)
		diskListQ.Filter(filter)
	}

	diskList, err := diskListQ.Do()
	if err != nil {
		ctx.Errorf("Error listing disks: %s", err)
		return nil, err
	}

	return diskList.Items, nil
}

func (d *driver) getAggregatedDisks(
	ctx types.Context) ([]*compute.Disk, error) {

	aggListQ := mustSession(ctx).Disks.AggregatedList(*d.projectID)
	if d.tag != "" {
		filter := fmt.Sprintf("labels.%s eq %s", tagKey, d.tag)
		ctx.Debugf("query filter: %s", filter)
		aggListQ.Filter(filter)
	}

	aggList, err := aggListQ.Do()
	if err != nil {
		ctx.Errorf("Error listing aggregated disks: %s", err)
		return nil, err
	}

	disks := []*compute.Disk{}

	for _, diskList := range aggList.Items {
		if diskList.Disks != nil && len(diskList.Disks) > 0 {
			disks = append(disks, diskList.Disks...)
		}
	}

	return disks, nil
}

func (d *driver) getDisk(
	ctx types.Context,
	zone *string,
	name *string) (*compute.Disk, error) {

	disk, err := mustSession(ctx).Disks.Get(*d.projectID, *zone, *name).Do()
	if err != nil {
		if apiE, ok := err.(*googleapi.Error); ok {
			if apiE.Code == 404 {
				return nil, nil
			}
		}
		ctx.Errorf("Error getting disk: %s", err)
		return nil, err
	}

	return disk, nil
}

func (d *driver) getInstance(
	ctx types.Context,
	zone *string,
	name *string) (*compute.Instance, error) {

	inst, err := mustSession(ctx).Instances.Get(*d.projectID, *zone, *name).Do()
	if err != nil {
		if apiE, ok := err.(*googleapi.Error); ok {
			if apiE.Code == 404 {
				return nil, nil
			}
		}
		ctx.Errorf("Error getting instance: %s", err)
		return nil, err
	}

	return inst, nil
}

func (d *driver) toTypeVolume(
	ctx types.Context,
	disks []*compute.Disk,
	attachments types.VolumeAttachmentsTypes,
	zone *string) ([]*types.Volume, error) {

	var (
		ld   *types.LocalDevices
		ldOK bool
	)

	if attachments.Devices() {
		// Get local devices map from context
		// Check for presence because this is required by the API, even
		// though we don't actually need this data
		if ld, ldOK = context.LocalDevices(ctx); !ldOK {
			return nil, goof.New(
				"error getting local devices from context")
		}
	}

	lsVolumes := make([]*types.Volume, len(disks))

	for i, disk := range disks {
		volume := &types.Volume{
			Name:             disk.Name,
			ID:               disk.Name,
			AvailabilityZone: utils.GetIndex(disk.Zone),
			Status:           disk.Status,
			Type:             utils.GetIndex(disk.Type),
			Size:             disk.SizeGb,
		}

		if attachments.Requested() {
			attachment := getAttachment(disk, attachments, ld)
			if attachment != nil {
				volume.Attachments = attachment
			}

		}

		lsVolumes[i] = volume
	}

	return lsVolumes, nil
}

func (d *driver) validZone(ctx types.Context) (*string, error) {
	// Is there a zone in the IID header?
	zone, ok := getClientZone(ctx)
	if ok {
		// Since there is a zone in the IID header, we only allow access
		// to volumes from that zone. If driver has restricted
		// access to a specific zone, client zone must match
		if d.zone != "" && *zone != d.zone {
			return nil, goof.New("No access to given zone")
		}
		return zone, nil
	}
	// No zone in the header, so access depends on how the driver
	// is configured
	if d.zone != "" {
		return &d.zone, nil
	}
	return nil, nil
}

func getAttachment(
	disk *compute.Disk,
	attachments types.VolumeAttachmentsTypes,
	ld *types.LocalDevices) []*types.VolumeAttachment {

	var volAttachments []*types.VolumeAttachment

	for _, link := range disk.Users {
		att := &types.VolumeAttachment{
			VolumeID: disk.Name,
			InstanceID: &types.InstanceID{
				ID:     utils.GetIndex(link),
				Driver: gcepd.Name,
			},
		}
		if attachments.Devices() {
			if dev, ok := ld.DeviceMap[disk.Name]; ok {
				att.DeviceName = dev
				// TODO: Do we need to enforce that the zone
				// found in link matches the zone for the volume?
			}
		}
		volAttachments = append(volAttachments, att)
	}
	return volAttachments
}

func (d *driver) createVolume(
	ctx types.Context,
	volumeName *string,
	opts *types.VolumeCreateOpts) error {

	diskType := d.defaultDiskType
	if opts.Type != nil && *opts.Type != "" {
		if strings.EqualFold(gcepd.DiskTypeSSD, *opts.Type) {
			diskType = gcepd.DiskTypeSSD
		} else if strings.EqualFold(gcepd.DiskTypeStandard, *opts.Type) {
			diskType = gcepd.DiskTypeStandard
		}
	}
	diskTypeURI := fmt.Sprintf("zones/%s/diskTypes/%s",
		*opts.AvailabilityZone, diskType)

	createDisk := &compute.Disk{
		Name:   *volumeName,
		SizeGb: *opts.Size,
		Type:   diskTypeURI,
	}

	asyncOp, err := mustSession(ctx).Disks.Insert(
		*d.projectID, *opts.AvailabilityZone, createDisk).Do()
	if err != nil {
		return goof.WithError("Failed to initiate disk creation", err)
	}

	err = d.waitUntilOperationIsFinished(
		ctx, opts.AvailabilityZone, asyncOp)
	if err != nil {
		return err
	}

	if d.tag != "" {
		/* In order to set the labels on a disk, we have to query the
		   disk first in order to get the generated label fingerprint
		*/
		disk, err := d.getDisk(ctx, opts.AvailabilityZone, volumeName)
		if err != nil {
			ctx.WithError(err).Warn(
				"Unable to query disk for labeling")
			return nil
		}
		labels := getLabels(&d.tag)
		_, err = mustSession(ctx).Disks.SetLabels(
			*d.projectID, *opts.AvailabilityZone, *volumeName,
			&compute.ZoneSetLabelsRequest{
				Labels:           labels,
				LabelFingerprint: disk.LabelFingerprint,
			}).Do()
		if err != nil {
			ctx.WithError(err).Warn("Unable to label disk")
		}
	}

	return nil
}

func (d *driver) waitUntilOperationIsFinished(
	ctx types.Context,
	zone *string,
	operation *compute.Operation) error {

	opName := operation.Name
	f := func() (interface{}, error) {
		duration := d.statusDelay
		for i := 1; i <= d.maxAttempts; i++ {

			op, err := mustSession(ctx).ZoneOperations.Get(
				*d.projectID, *zone, opName).Do()
			if err != nil {
				return nil, err
			}

			switch op.Status {
			case "PENDING", "RUNNING":
				ctx.WithField("status", op.Status).Debug(
					"still waiting for operation",
				)
				time.Sleep(time.Duration(duration) *
					time.Nanosecond)
				duration = int64(2) * duration
			case "DONE":
				if op.Error != nil {
					bytea, _ := op.Error.MarshalJSON()
					return nil, goof.New(string(bytea))
				}
				return nil, nil
			default:
				return nil, goof.Newf("Unknown status %q: %+v",
					op.Status, op)
			}
		}
		return nil, goof.WithField("maxAttempts", d.maxAttempts,
			"Status attempts exhausted")
	}

	_, ok, err := apiUtils.WaitFor(f, d.statusTimeout)
	if !ok {
		return goof.WithFields(goof.Fields{
			"statusTimeout": d.statusTimeout},
			"Timeout occured waiting for storage action")
	}
	if err != nil {
		return goof.WithError(
			"Error while waiting for storage action to finish", err)
	}
	return nil
}

func (d *driver) attachVolume(
	ctx types.Context,
	instanceID *string,
	zone *string,
	volumeName *string) error {

	disk := &compute.AttachedDisk{
		AutoDelete: false,
		Boot:       false,
		Source:     fmt.Sprintf("zones/%s/disks/%s", *zone, *volumeName),
		DeviceName: *volumeName,
	}

	asyncOp, err := mustSession(ctx).Instances.AttachDisk(
		*d.projectID, *zone, *instanceID, disk).Do()
	if err != nil {
		return err
	}

	err = d.waitUntilOperationIsFinished(ctx, zone, asyncOp)
	if err != nil {
		return err
	}

	return nil
}

func getBlockDevice(volumeID *string) string {
	return fmt.Sprintf("/dev/disk/by-id/google-%s", *volumeID)
}

func (d *driver) getAttachedDeviceName(
	ctx types.Context,
	zone *string,
	instanceName *string,
	diskLink *string) (string, error) {

	gceInstance, err := d.getInstance(ctx, zone, instanceName)
	if err != nil {
		return "", err
	}

	for _, disk := range gceInstance.Disks {
		if disk.Source == *diskLink {
			return disk.DeviceName, nil
		}
	}

	return "", goof.New("Unable to find attached volume on instance")
}

func (d *driver) detachVolume(
	ctx types.Context,
	gceDisk *compute.Disk) error {

	var ops = make([]*compute.Operation, 0)
	var asyncErr error

	zone := utils.GetIndex(gceDisk.Zone)

	for _, user := range gceDisk.Users {
		instanceName := utils.GetIndex(user)
		devName, err := d.getAttachedDeviceName(ctx, &zone, &instanceName,
			&gceDisk.SelfLink)
		if err != nil {
			return goof.WithError(
				"Unable to get device name from instance", err)
		}
		asyncOp, err := mustSession(ctx).Instances.DetachDisk(
			*d.projectID, zone, instanceName, devName).Do()
		if err != nil {
			asyncErr = goof.WithError("Error detaching disk", err)
			continue
		}
		ops = append(ops, asyncOp)
	}

	if len(ops) > 0 {
		for _, op := range ops {
			err := d.waitUntilOperationIsFinished(ctx,
				&zone, op)
			if err != nil {
				return err
			}
		}
	}

	return asyncErr
}

func (d *driver) convUnderscores(name string) string {
	if d.convUnderscore {
		name = strings.Replace(name, "_", "-", -1)
	}
	return name
}

func getLabels(tag *string) map[string]string {
	labels := map[string]string{
		tagKey: *tag,
	}

	return labels
}
