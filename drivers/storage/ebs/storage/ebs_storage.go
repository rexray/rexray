package storage

import (
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/ebs"
)

const (
	// waitVolumeCreate signifies to wait for volume creation to complete
	waitVolumeCreate = "create"
	// waitVolumeAttach signifies to wait for volume attachment to complete
	waitVolumeAttach = "attach"
	// waitVolumeDetach signifies to wait for volume detachment to complete
	waitVolumeDetach = "detach"
)

type driver struct {
	name     string
	config   gofig.Config
	awsCreds *credentials.Credentials
}

var (
	nextDeviceInfo = &types.NextDeviceInfo{
		Prefix: "xvd",
		// EBS suggests to use /dev/sd[f-p] for Linux EC2 instances.
		// Also on Linux EC2 instances, although the device path may show up
		// as /dev/sd* on the EC2 side, it will appear locally as /dev/xvd*
		Pattern: "[f-p]",
		Ignore:  false,
	}
)

func init() {
	registry.RegisterStorageDriver(ebs.Name, newDriver)
	// Backwards compatibility for ec2 driver
	registry.RegisterStorageDriver(ebs.OldName, newOldDriver)
}

func newDriver() types.StorageDriver {
	return &driver{name: ebs.Name}
}

func newOldDriver() types.StorageDriver {
	return &driver{name: ebs.OldName}
}

func (d *driver) Name() string {
	return d.name
}

// Init initializes the driver.
func (d *driver) Init(context types.Context, config gofig.Config) error {
	// Ensure backwards compatibility with ebs and ec2 in config
	ebs.BackCompat(config)
	d.config = config
	log.Info("storage driver initialized")
	return nil
}

var (
	sessions    = map[string]interface{}{}
	sessionsRWL = &sync.RWMutex{}
)

func (d *driver) Login(ctx types.Context) (interface{}, error) {

	fields := map[string]interface{}{
		"accessKey": d.accessKey(),
		"tag":       d.tag(),
	}

	log.WithFields(fields).Debug("beginning ebs login attempt")

	if d.secretKey() == "" {
		fields["secretKey"] = ""
	} else {
		fields["secretKey"] = "******"
	}

	var (
		region     string
		endpoint   string
		iid        *types.InstanceID
		sess       = session.New()
		maxRetries = d.maxRetries()
	)

	iid, _ = context.InstanceID(ctx)

	if iid != nil {
		region = iid.Fields[ebs.InstanceIDFieldRegion]
	} else {
		region = d.region()
	}

	if region != "" {
		endpoint = fmt.Sprintf("ec2.%s.amazonaws.com", region)
	} else {
		endpoint = d.endpoint()
	}

	fields["region"] = region
	fields["endpoint"] = endpoint

	log.WithFields(fields).Debug("ebs service connetion attempt")

	svc := awsec2.New(
		sess,
		&aws.Config{
			Region:     &region,
			Endpoint:   &endpoint,
			MaxRetries: &maxRetries,
			Credentials: credentials.NewChainCredentials(
				[]credentials.Provider{
					&credentials.StaticProvider{
						Value: credentials.Value{
							AccessKeyID:     d.accessKey(),
							SecretAccessKey: d.secretKey(),
						},
					},
					&credentials.EnvProvider{},
					&credentials.SharedCredentialsProvider{},
					&ec2rolecreds.EC2RoleProvider{
						Client: ec2metadata.New(sess),
					},
				},
			),
		},
	)

	log.WithFields(fields).Info("ebs service connetion created")
	return svc, nil
}

func mustSession(ctx types.Context) *awsec2.EC2 {
	return context.MustSession(ctx).(*awsec2.EC2)
}

func mustInstanceIDID(ctx types.Context) *string {
	return &context.MustInstanceID(ctx).ID
}

// NextDeviceInfo returns the information about the driver's next available
// device workflow.
func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return nextDeviceInfo, nil
}

// Type returns the type of storage the driver provides.
func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	//Example: Block storage
	return types.Block, nil
}

// InstanceInspect returns an instance.
func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {
	// get instance ID
	iid := context.MustInstanceID(ctx)

	// If no instance ID, return blank instance
	if iid.ID != "" {
		return &types.Instance{InstanceID: iid}, nil
	}

	// Decode metadata from instance ID
	var awsInstanceID string
	if err := iid.UnmarshalMetadata(&awsInstanceID); err != nil {
		return nil, goof.WithError(
			"error unmarshalling instance id metadata", err)
	}
	instanceID := &types.InstanceID{ID: awsInstanceID, Driver: d.Name()}

	return &types.Instance{InstanceID: instanceID}, nil
}

// Volumes returns all volumes or a filtered list of volumes.
func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {
	// Get all volumes via EC2 API
	ec2vols, err := d.getVolume(ctx, "", "")
	if err != nil {
		return nil, goof.WithError("error getting volume", err)
	}
	if len(ec2vols) == 0 {
		return nil, errNoVolReturned
	}
	// Convert retrieved volumes to libStorage types.Volume
	vols, convErr := d.toTypesVolume(ctx, ec2vols, opts.Attachments)
	if convErr != nil {
		return nil, goof.WithError("error converting to types.Volume", convErr)
	}
	return vols, nil
}

// VolumeInspect inspects a single volume.
func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {
	// Get volume corresponding to volume ID via EC2 API
	ec2vols, err := d.getVolume(ctx, volumeID, "")
	if err != nil {
		return nil, goof.WithError("error getting volume", err)
	}
	if len(ec2vols) == 0 {
		return nil, errNoVolReturned
	}
	vols, convErr := d.toTypesVolume(ctx, ec2vols, opts.Attachments)
	if convErr != nil {
		return nil, goof.WithError("error converting to types.Volume", convErr)
	}

	// Because getVolume returns an array
	// and we only expect the 1st element to be a match, return 1st element
	return vols[0], nil
}

// VolumeCreate creates a new volume.
func (d *driver) VolumeCreate(ctx types.Context, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	// Initialize for logging
	fields := map[string]interface{}{
		"driverName": d.Name(),
		"volumeName": volumeName,
		"opts":       opts,
	}

	log.WithFields(fields).Debug("creating volume")

	// Check if volume with same name exists
	ec2vols, err := d.getVolume(ctx, "", volumeName)
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error getting volume", err)
	}
	volumes, convErr := d.toTypesVolume(ctx, ec2vols, false)
	if convErr != nil {
		return nil, goof.WithFieldsE(
			fields, "error converting to types.Volume", convErr)
	}

	if len(volumes) > 0 {
		return nil, goof.WithFields(fields, "volume name already exists")
	}

	volume := &types.Volume{}

	// Pass arguments into libStorage types.Volume
	if opts.AvailabilityZone != nil {
		volume.AvailabilityZone = *opts.AvailabilityZone
	}
	if opts.Type != nil {
		volume.Type = *opts.Type
	}
	if opts.Size != nil {
		volume.Size = *opts.Size
	}
	if opts.IOPS != nil {
		volume.IOPS = *opts.IOPS
	}
	if opts.Encrypted != nil {
		volume.Encrypted = *opts.Encrypted
	}

	// Pass libStorage types.Volume to helper function which calls EC2 API
	vol, err := d.createVolume(ctx, volumeName, "", volume)
	if err != nil {
		return nil, goof.WithFieldsE(fields, "error creating volume", err)
	}
	// Return the volume created
	return d.VolumeInspect(ctx, *vol.VolumeId, &types.VolumeInspectOpts{
		Attachments: true,
	})
}

// VolumeCreateFromSnapshot creates a new volume from an existing snapshot.
func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
	/*
		// Initialize for logging
		fields := map[string]interface{}{
			"driverName": d.Name(),
			"snapshotID": snapshotID,
			"volumeName": volumeName,
			"opts":       opts,
		}

		log.WithFields(fields).Debug("creating volume from snapshot")

		// Check if volume with same name exists
		ec2vols, err := d.getVolume(ctx, "", volumeName)
		if err != nil {
			return nil, goof.WithFieldsE(fields, "Error getting volume", err)
		}
		volumes, convErr := d.toTypesVolume(ctx, ec2vols, false)
		if convErr != nil {
			return nil, goof.WithFieldsE(fields,
				"Error converting to types.Volume", convErr)
		}

		if len(volumes) > 0 {
			return nil, goof.WithFields(fields, "volume name already exists")
		}

		volume := &types.Volume{}

		// Pass arguments into libStorage types.Volume
		if opts.AvailabilityZone != nil {
			volume.AvailabilityZone = *opts.AvailabilityZone
		}
		if opts.Type != nil {
			volume.Type = *opts.Type
		}
		if opts.Size != nil {
			volume.Size = *opts.Size
		}
		if opts.IOPS != nil {
			volume.IOPS = *opts.IOPS
		}
		if *opts.Encrypted == false {
			// Volume must be encrypted if snapshot is encrypted
			snapshot, err := d.SnapshotInspect(ctx, snapshotID, nil)
			if err != nil {
				return &types.Volume{}, goof.WithFieldsE(fields,
					"Error getting snapshot", err)
			}
			volume.Encrypted = snapshot.Encrypted
		} else {
			volume.Encrypted = *opts.Encrypted
		}

		// Pass libStorage types.Volume to helper function which calls EC2 API
		vol, err := d.createVolume(ctx, volumeName, snapshotID, volume)
		if err != nil {
			return &types.Volume{}, goof.WithFieldsE(fields,
				"error creating volume", err)
		}
		// Return the volume created
		return d.VolumeInspect(ctx, *vol.VolumeId, &types.VolumeInspectOpts{
			Attachments: true,
		})
	*/
}

// VolumeCopy copies an existing volume.
func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
	/*
		// Creates a temp snapshot of an existing volume,
		// and creates a volume from that snapshot.
		var (
			ec2vols  []*awsec2.Volume
			err      error
			snapshot *types.Snapshot
			vol      *types.Volume
		)

		// Initialize for logging
		fields := map[string]interface{}{
			"driverName": d.Name(),
			"volumeID":   volumeID,
			"volumeName": volumeName,
			"opts":       opts,
		}

		log.WithFields(fields).Debug("creating volume from snapshot")

		// Check if volume with same name exists
		ec2VolsToCheck, err := d.getVolume(ctx, "", volumeName)
		if err != nil {
			return nil, goof.WithFieldsE(fields, "Error getting volume", err)
		}
		volsToCheck, convErr := d.toTypesVolume(ctx, ec2VolsToCheck, false)
		if convErr != nil {
			return nil, goof.WithFieldsE(fields, "Error converting to types.Volume",
				convErr)
		}

		if len(volsToCheck) > 0 {
			return nil, goof.WithFields(fields, "volume name already exists")
		}

		// Get volume to copy using volumeID
		ec2vols, err = d.getVolume(ctx, volumeID, "")
		if err != nil {
			return &types.Volume{}, goof.WithFieldsE(fields,
				"error getting volume", err)
		}
		volumes, convErr2 := d.toTypesVolume(ctx, ec2vols, false)
		if convErr2 != nil {
			return nil, goof.WithFieldsE(fields,
				"Error converting to types.Volume", convErr2)
		}
		if len(volumes) > 1 {
			return &types.Volume{},
				goof.WithFields(fields, "multiple volumes returned")
		} else if len(volumes) == 0 {
			return &types.Volume{}, goof.WithFields(fields, "no volumes returned")
		}

		// Create temporary snapshot
		snapshotName := fmt.Sprintf("temp-%s-%d", volumeID, time.Now().UnixNano())
		fields["snapshotName"] = snapshotName
		snapshot, err = d.VolumeSnapshot(ctx, volumeID, snapshotName, opts)
		if err != nil {
			return &types.Volume{}, goof.WithFieldsE(fields,
				"error creating temporary snapshot", err)
		}

		// Use temporary snapshot to create volume
		vol, err = d.VolumeCreateFromSnapshot(ctx, snapshot.ID,
			volumeName, &types.VolumeCreateOpts{Encrypted: &snapshot.Encrypted,
				Opts: opts})
		if err != nil {
			return &types.Volume{}, goof.WithFieldsE(fields,
				"error creating volume copy from snapshot", err)
		}

		// Remove temporary snapshot created
		if err = d.SnapshotRemove(ctx, snapshot.ID, opts); err != nil {
			return &types.Volume{}, goof.WithFieldsE(fields,
				"error removing temporary snapshot", err)
		}

		log.Println("Created volume " + vol.ID + " from volume " + volumeID)
		return vol, nil
	*/
}

// VolumeSnapshot snapshots a volume.
func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
	/*
		// Create snapshot with EC2 API call
		csInput := &awsec2.CreateSnapshotInput{
			VolumeId: &volumeID,
		}

		resp, err := d.ec2Instance.CreateSnapshot(csInput)
		if err != nil {
			return nil, goof.WithError("Error creating snapshot", err)
		}

		// Add tags to EC2 snapshot
		if err = d.createTags(*resp.SnapshotId, snapshotName); err != nil {
			return &types.Snapshot{}, goof.WithError(
				"Error creating tags", err)
		}

		log.Println("Waiting for snapshot to complete")
		err = d.waitSnapshotComplete(ctx, *resp.SnapshotId)
		if err != nil {
			return &types.Snapshot{}, goof.WithError(
				"Error waiting for snapshot creation", err)
		}

		// Check if successful snapshot
		snapshot, err := d.SnapshotInspect(ctx, *resp.SnapshotId, nil)
		if err != nil {
			return &types.Snapshot{}, goof.WithError(
				"Error getting snapshot", err)
		}

		log.Println("Created Snapshot: " + snapshot.ID)
		return snapshot, nil
	*/
}

// VolumeRemove removes a volume.
func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts types.Store) error {
	// Initialize for logging
	fields := map[string]interface{}{
		"provider": d.Name(),
		"volumeID": volumeID,
	}

	//TODO check if volume is attached? if so fail

	// Delete volume via EC2 API call
	dvInput := &awsec2.DeleteVolumeInput{
		VolumeId: &volumeID,
	}
	_, err := mustSession(ctx).DeleteVolume(dvInput)
	if err != nil {
		return goof.WithFieldsE(fields, "error deleting volume", err)
	}

	return nil
}

var (
	errMissingNextDevice  = goof.New("missing next device")
	errVolAlreadyAttached = goof.New("volume already attached to a host")
)

// VolumeAttach attaches a volume and provides a token clients can use
// to validate that device has appeared locally.
func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {
	// review volume with attachments to any host
	ec2vols, err := d.getVolume(ctx, volumeID, "")
	if err != nil {
		return nil, "", goof.WithError("error getting volume", err)
	}
	volumes, convErr := d.toTypesVolume(ctx, ec2vols, true)
	if convErr != nil {
		return nil, "", goof.WithError(
			"error converting to types.Volume", convErr)
	}

	// Check if there a volume to attach
	if len(volumes) == 0 {
		return nil, "", goof.New("no volume found")
	}
	// Check if volume is already attached
	if len(volumes[0].Attachments) > 0 {
		// Detach already attached volume if forced
		if !opts.Force {
			return nil, "", errVolAlreadyAttached
		}
		_, err := d.VolumeDetach(
			ctx,
			volumeID,
			&types.VolumeDetachOpts{
				Force: opts.Force,
				Opts:  opts.Opts,
			})
		if err != nil {
			return nil, "", goof.WithError("error detaching volume", err)
		}
	}

	if opts.NextDevice == nil {
		return nil, "", errMissingNextDevice
	}

	// Attach volume via helper function which uses EC2 API call
	err = d.attachVolume(ctx, volumeID, volumes[0].Name, *opts.NextDevice)
	if err != nil {
		return nil, "", goof.WithFieldsE(
			log.Fields{
				"provider": d.Name(),
				"volumeID": volumeID},
			"error attaching volume",
			err,
		)
	}

	// Wait for volume's status to update
	if err = d.waitVolumeComplete(ctx, volumeID, waitVolumeAttach); err != nil {
		return nil, "", goof.WithError("error waiting for volume attach", err)
	}

	// Check if successful attach
	attachedVol, err := d.VolumeInspect(
		ctx, volumeID, &types.VolumeInspectOpts{
			Attachments: true,
			Opts:        opts.Opts,
		})
	if err != nil {
		return nil, "", goof.WithError("error getting volume", err)
	}

	// Token is the attachment's device name, which will be matched
	// to the executor's device ID
	return attachedVol, *opts.NextDevice, nil
}

var errVolAlreadyDetached = goof.New("volume already detached")

// VolumeDetach detaches a volume.
func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {
	// review volume with attachments to any host
	ec2vols, err := d.getVolume(ctx, volumeID, "")
	if err != nil {
		return nil, goof.WithError("error getting volume", err)
	}
	volumes, convErr := d.toTypesVolume(ctx, ec2vols, true)
	if convErr != nil {
		return nil, goof.WithError("error converting to types.Volume", convErr)
	}

	// no volumes to detach
	if len(volumes) == 0 {
		return nil, errNoVolReturned
	}

	// volume has no attachments
	if len(volumes[0].Attachments) == 0 {
		return nil, errVolAlreadyDetached
	}

	dvInput := &awsec2.DetachVolumeInput{
		VolumeId: &volumeID,
		Force:    &opts.Force,
	}

	// Detach volume using EC2 API call
	if _, err = mustSession(ctx).DetachVolume(dvInput); err != nil {
		return nil, goof.WithFieldsE(
			log.Fields{
				"provider": d.Name(),
				"volumeID": volumeID}, "error detaching volume", err)
	}

	if err = d.waitVolumeComplete(ctx, volumeID, waitVolumeDetach); err != nil {
		return nil, goof.WithError("error waiting for volume detach", err)
	}

	ctx.Info("detached volume", volumeID)

	// check if successful detach
	detachedVol, err := d.VolumeInspect(
		ctx, volumeID, &types.VolumeInspectOpts{
			Attachments: true,
			Opts:        opts.Opts,
		})
	if err != nil {
		return nil, goof.WithError("error getting volume", err)
	}

	return detachedVol, nil
}

// Snapshots returns all volumes or a filtered list of snapshots.
func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
	/*
		// Get all snapshots
		ec2snapshots, err := d.getSnapshot(ctx, "", "", "")
		if err != nil {
			return nil, goof.WithError("error getting snapshot", err)
		}
		if len(ec2snapshots) == 0 {
			return nil, goof.New("no snapshots returned")
		}
		// Convert to libStorage types.Snapshot
		snapshots := d.toTypesSnapshot(ec2snapshots)
		return snapshots, nil
	*/
}

// SnapshotInspect inspects a single snapshot.
func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
	/*
		// Get snapshot corresponding to snapshot ID
		ec2snapshots, err := d.getSnapshot(ctx, "", snapshotID, "")
		if err != nil {
			return nil, goof.WithError("error getting snapshot", err)
		}
		if len(ec2snapshots) == 0 {
			return nil, goof.New("no snapshots returned")
		}
		// Convert to libStorage types.Snapshot
		snapshots := d.toTypesSnapshot(ec2snapshots)

		// Because getSnapshot returns an array
		// and we only expect the 1st element to be a match, return 1st element
		return snapshots[0], nil
	*/
}

// SnapshotCopy copies an existing snapshot.
func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
	/*
		// no snapshot id inputted
		if snapshotID == "" {
			return &types.Snapshot{}, goof.New("Missing snapshotID")
		}

		// Get snapshot to copy
		origSnapshots, err := d.getSnapshot(ctx, "", snapshotID, "")
		if err != nil {
			return &types.Snapshot{},
				goof.WithError("Error getting snapshot", err)
		}

		if len(origSnapshots) > 1 {
			return &types.Snapshot{},
				goof.New("multiple snapshots returned")
		} else if len(origSnapshots) == 0 {
			return &types.Snapshot{}, goof.New("no snapshots returned")
		}

		// Copy snapshot with EC2 API call
		snapshotID = *(origSnapshots[0]).SnapshotId
		snapshotName = d.getName(origSnapshots[0].Tags)

		options := &awsec2.CopySnapshotInput{
			SourceSnapshotId: &snapshotID,
			SourceRegion:     &d.instanceDocument.Region,
			Description:      aws.String(fmt.Sprintf("Copy of %s", snapshotID)),
		}
		resp := &awsec2.CopySnapshotOutput{}

		resp, err = d.ec2Instance.CopySnapshot(options)
		if err != nil {
			return nil, goof.WithError("error copying snapshot", err)
		}

		// Add tags to copied snapshot
		if err = d.createTags(*resp.SnapshotId, snapshotName); err != nil {
			return &types.Snapshot{}, goof.WithError(
				"Error creating tags", err)
		}

		log.WithFields(log.Fields{
			"moduleName":      d.Name(),
			"driverName":      d.Name(),
			"snapshotName":    snapshotName,
			"resp.SnapshotId": *resp.SnapshotId}).Info("waiting for snapshot to complete")

		// Wait for snapshot status to update
		err = d.waitSnapshotComplete(ctx, *resp.SnapshotId)
		if err != nil {
			return &types.Snapshot{}, goof.WithError(
				"Error waiting for snapshot creation", err)
		}

		// Check if successful snapshot
		snapshotCopy, err := d.SnapshotInspect(ctx, *resp.SnapshotId, nil)
		if err != nil {
			return &types.Snapshot{}, goof.WithError(
				"Error getting snapshot copy", err)
		}
		destinationID = snapshotCopy.ID

		log.Println("Copied Snapshot: " + destinationID)
		return snapshotCopy, nil
	*/
}

// SnapshotRemove removes a snapshot.
func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {
	// TODO Snapshots are not implemented yet
	return types.ErrNotImplemented
	/*
		// Initialize for logging
		fields := map[string]interface{}{
			"provider":   d.Name(),
			"snapshotID": snapshotID,
		}

		// no snapshot ID inputted
		if snapshotID == "" {
			return goof.New("missing snapshot id")
		}

		// Delete snapshot using EC2 API call
		dsInput := &awsec2.DeleteSnapshotInput{
			SnapshotId: &snapshotID,
		}
		_, err := d.ec2Instance.DeleteSnapshot(dsInput)
		if err != nil {
			return goof.WithFieldsE(fields, "error deleting snapshot", err)
		}

		return nil
	*/
}

///////////////////////////////////////////////////////////////////////
/////////        HELPER FUNCTIONS SPECIFIC TO PROVIDER        /////////
///////////////////////////////////////////////////////////////////////
// getVolume searches for and returns volumes matching criteria
func (d *driver) getVolume(
	ctx types.Context,
	volumeID, volumeName string) ([]*awsec2.Volume, error) {

	iid := context.MustInstanceID(ctx)
	avaiZone := iid.Fields[ebs.InstanceIDFieldAvailabilityZone]

	// prepare filters
	filters := []*awsec2.Filter{
		&awsec2.Filter{
			Name:   aws.String("availability-zone"),
			Values: []*string{&avaiZone},
		},
	}

	if volumeName != "" {
		filters = append(filters, &awsec2.Filter{
			Name: aws.String("tag:Name"), Values: []*string{&volumeName}})
	}

	if volumeID != "" {
		filters = append(filters, &awsec2.Filter{
			Name: aws.String("volume-id"), Values: []*string{&volumeID}})
	}

	// TODO rexrayTag
	/*	if d.ec2Tag != "" {
			filters = append(filters, &awsec2.Filter{
				Name:   aws.String(fmt.Sprintf("tag:%s", d.rexrayTag())),
				Values: []*string{&d.ec2Tag}})
		}
	*/
	// Prepare input
	dvInput := &awsec2.DescribeVolumesInput{}

	// Apply filters if arguments are specified
	if len(filters) > 0 {
		dvInput.Filters = filters
	}

	if volumeID != "" {
		dvInput.VolumeIds = []*string{&volumeID}
	}

	// Retrieve filtered volumes through EC2 API call
	resp, err := mustSession(ctx).DescribeVolumes(dvInput)
	if err != nil {
		return []*awsec2.Volume{}, err
	}

	return resp.Volumes, nil
}

var errGetLocDevs = goof.New("error getting local devices from context")

// Converts EC2 API volumes to libStorage types.Volume
func (d *driver) toTypesVolume(
	ctx types.Context,
	ec2vols []*awsec2.Volume,
	attachments bool) ([]*types.Volume, error) {
	// Get local devices map from context
	ld, ldOK := context.LocalDevices(ctx)
	if !ldOK {
		return nil, errGetLocDevs
	}

	var volumesSD []*types.Volume
	for _, volume := range ec2vols {
		var attachmentsSD []*types.VolumeAttachment
		// Leave attachment's device name blank if attachments is false
		for _, attachment := range volume.Attachments {
			deviceName := ""
			if attachments {
				// Compensate for kernel volume mapping i.e. change "/dev/sda" to "/dev/xvda"
				deviceName = strings.Replace(
					*attachment.Device, "sd", nextDeviceInfo.Prefix, 1)
				// Keep device name if it is found in local devices
				if _, ok := ld.DeviceMap[deviceName]; !ok {
					deviceName = ""
				}
			}
			attachmentSD := &types.VolumeAttachment{
				VolumeID: *attachment.VolumeId,
				InstanceID: &types.InstanceID{
					ID:     *attachment.InstanceId,
					Driver: d.Name(),
				},
				DeviceName: deviceName,
				Status:     *attachment.State,
			}
			attachmentsSD = append(attachmentsSD, attachmentSD)
		}
		name := d.getName(volume.Tags)
		volumeSD := &types.Volume{
			Name:             name,
			ID:               *volume.VolumeId,
			AvailabilityZone: *volume.AvailabilityZone,
			Encrypted:        *volume.Encrypted,
			Status:           *volume.State,
			Type:             *volume.VolumeType,
			Size:             *volume.Size,
			Attachments:      attachmentsSD,
		}
		// Some volume types have no IOPS, so we get nil in volume.Iops
		if volume.Iops != nil {
			volumeSD.IOPS = *volume.Iops
		}
		volumesSD = append(volumesSD, volumeSD)
	}
	return volumesSD, nil
}

// getSnapshot searches for and returns snapshots matching criteria
// TODO Snapshots are not implemented yet
/*
func (d *driver) getSnapshot(
	ctx types.Context,
	volumeID, snapshotID, snapshotName string) ([]*awsec2.Snapshot, error) {
	// Prepare filters
	filters := []*awsec2.Filter{}
	if snapshotName != "" {
		filters = append(filters, &awsec2.Filter{
			Name: aws.String("tag:Name"), Values: []*string{&snapshotName}})
	}

	if volumeID != "" {
		filters = append(filters, &awsec2.Filter{
			Name: aws.String("volume-id"), Values: []*string{&volumeID}})
	}

	if snapshotID != "" {
		//using SnapshotIds in request is returning stale data
		filters = append(filters, &awsec2.Filter{
			Name: aws.String("snapshot-id"), Values: []*string{&snapshotID}})
	}

	// TODO rexrayTag?
	//	if d.ec2Tag != "" {
	//	filters = append(filters, &ec2.Filter{
	//		Name:   aws.String(fmt.Sprintf("tag:%s", rexrayTag)),
	//		Values: []*string{&d.ec2Tag}})
	//}

	// Prepare input
	dsInput := &awsec2.DescribeSnapshotsInput{}

	// Apply filters if arguments are specified
	if len(filters) > 0 {
		dsInput.Filters = filters
	}

	// Retrieve filtered volumes through EC2 API call
	resp, err := d.ec2Instance.DescribeSnapshots(dsInput)
	if err != nil {
		return nil, err
	}

	return resp.Snapshots, nil
}


// Converts EC2 API snapshots to libStorage types.Snapshot
func (d *driver) toTypesSnapshot(
	ec2snapshots []*awsec2.Snapshot) []*types.Snapshot {
	var snapshotsInt []*types.Snapshot
	for _, snapshot := range ec2snapshots {
		name := d.getName(snapshot.Tags)
		snapshotSD := &types.Snapshot{
			Name:        name,
			VolumeID:    *snapshot.VolumeId,
			ID:          *snapshot.SnapshotId,
			Encrypted:   *snapshot.Encrypted,
			VolumeSize:  *snapshot.VolumeSize,
			StartTime:   (*snapshot.StartTime).Unix(),
			Description: *snapshot.Description,
			Status:      *snapshot.State,
		}
		snapshotsInt = append(snapshotsInt, snapshotSD)
	}

	return snapshotsInt
}
*/

var (
	errNoVolReturned       = goof.New("no volume returned")
	errTooManyVolsReturned = goof.New("too many volumes returned")
)

// Used in VolumeAttach
func (d *driver) attachVolume(
	ctx types.Context,
	volumeID, volumeName, deviceName string) error {

	// sanity check # of volumes to attach
	vol, err := d.getVolume(ctx, volumeID, volumeName)
	if err != nil {
		return goof.WithError("error getting volume", err)
	}

	if len(vol) == 0 {
		return errNoVolReturned
	}
	if len(vol) > 1 {
		return errTooManyVolsReturned
	}

	// Attach volume via EC2 API call
	avInput := &awsec2.AttachVolumeInput{
		Device:     &deviceName,
		InstanceId: mustInstanceIDID(ctx),
		VolumeId:   &volumeID,
	}

	if _, err := mustSession(ctx).AttachVolume(avInput); err != nil {
		return err
	}
	return nil
}

// Used in VolumeCreate
func (d *driver) createVolume(ctx types.Context, volumeName, snapshotID string,
	vol *types.Volume) (*awsec2.Volume, error) {
	var (
		err    error
		server awsec2.Instance
	)
	// Create volume using EC2 API call
	if server, err = d.getInstance(ctx); err != nil {
		return &awsec2.Volume{}, goof.WithError(
			"error creating volume with EC2 API call", err)
	}

	// Fill in Availability Zone if needed
	d.createVolumeEnsureAvailabilityZone(&vol.AvailabilityZone, &server)

	options := &awsec2.CreateVolumeInput{
		Size:             &vol.Size,
		AvailabilityZone: &vol.AvailabilityZone,
		Encrypted:        &vol.Encrypted,
		VolumeType:       &vol.Type,
	}
	if snapshotID != "" {
		options.SnapshotId = &snapshotID
	}

	if vol.IOPS > 0 {
		options.Iops = &vol.IOPS
	}
	var resp *awsec2.Volume

	if resp, err = mustSession(ctx).CreateVolume(options); err != nil {
		return &awsec2.Volume{}, goof.WithError(
			"error creating volume", err)
	}

	// Add tags to created volume
	if err = d.createTags(ctx, *resp.VolumeId, volumeName); err != nil {
		return &awsec2.Volume{}, goof.WithError(
			"error creating tags", err)
	}

	// Wait for volume status to change
	if err = d.waitVolumeComplete(
		ctx, *resp.VolumeId, waitVolumeCreate); err != nil {
		return &awsec2.Volume{}, goof.WithError(
			"error waiting for volume creation", err)
	}

	return resp, nil
}

// Make sure Availability Zone is non-empty and valid
func (d *driver) createVolumeEnsureAvailabilityZone(
	availabilityZone *string, server *awsec2.Instance) {
	if *availabilityZone == "" {
		*availabilityZone = *server.Placement.AvailabilityZone
	}
}

// Fill in tags for volume or snapshot
func (d *driver) createTags(ctx types.Context, id, name string) (err error) {
	var (
		ctInput   *awsec2.CreateTagsInput
		inputName string
	)
	initCTInput := func() {
		if ctInput != nil {
			return
		}
		ctInput = &awsec2.CreateTagsInput{
			Resources: []*string{&id},
			Tags:      []*awsec2.Tag{},
		}
		// Append config tag to name
		inputName = d.getFullName(d.getPrintableName(name))
	}

	initCTInput()
	ctInput.Tags = append(
		ctInput.Tags,
		&awsec2.Tag{
			Key:   aws.String("Name"),
			Value: &inputName,
		})

	// TODO rexrayTag
	/*	if d.ec2Tag != "" {
			initCTInput()
			ctInput.Tags = append(
				ctInput.Tags,
				&awsec2.Tag{
					Key:   aws.String(d.rexrayTag()),
					Value: &d.ec2Tag,
				})
		}
	*/
	_, err = mustSession(ctx).CreateTags(ctInput)
	if err != nil {
		return goof.WithError("error creating tags", err)
	}
	return nil
}

var errMissingVolID = goof.New("missing volume ID")

// Wait for volume action to complete (creation, attachment, detachment)
func (d *driver) waitVolumeComplete(
	ctx types.Context, volumeID, action string) error {
	// no volume id inputted
	if volumeID == "" {
		return errMissingVolID
	}

UpdateLoop:
	for {
		// update volume
		volumes, err := d.getVolume(ctx, volumeID, "")
		if err != nil {
			return goof.WithError("error getting volume", err)
		}

		// check retrieved volume
		switch action {
		case waitVolumeCreate:
			if *volumes[0].State == awsec2.VolumeStateAvailable {
				break UpdateLoop
			}
		case waitVolumeDetach:
			if len(volumes[0].Attachments) == 0 {
				break UpdateLoop
			}
		case waitVolumeAttach:
			if len(volumes[0].Attachments) == 1 &&
				*volumes[0].Attachments[0].State == awsec2.VolumeAttachmentStateAttached {
				break UpdateLoop
			}
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

// Wait for snapshot action to complete
// TODO Snapshots are not implemented yet
/*
func (d *driver) waitSnapshotComplete(
	ctx types.Context, snapshotID string) error {
	if snapshotID == "" {
		return goof.New("Missing snapshot ID")
	}

	for {
		// update snapshot
		snapshots, err := d.getSnapshot(ctx, "", snapshotID, "")
		if err != nil {
			return goof.WithError(
				"Error getting snapshot", err)
		}

		// check retrieved snapshot
		if len(snapshots) == 0 {
			return goof.New("No snapshots found")
		}
		snapshot := snapshots[0]
		if *snapshot.State == awsec2.SnapshotStateCompleted {
			break
		}
		if *snapshot.State == awsec2.SnapshotStateError {
			return goof.Newf("Snapshot state error: %s", *snapshot.StateMessage)
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}
*/

// Retrieve volume or snapshot name
func (d *driver) getName(tags []*awsec2.Tag) string {
	for _, tag := range tags {
		if *tag.Key == "Name" {
			return *tag.Value
		}
	}
	return ""
}

// Retrieve current instance using EC2 API call
func (d *driver) getInstance(ctx types.Context) (awsec2.Instance, error) {
	diInput := &awsec2.DescribeInstancesInput{
		InstanceIds: []*string{mustInstanceIDID(ctx)},
	}
	resp, err := mustSession(ctx).DescribeInstances(diInput)
	if err != nil {
		return awsec2.Instance{}, goof.WithError(
			"error retrieving instance with EC2 API call", err)
	}
	return *resp.Reservations[0].Instances[0], nil
}

// Get volume or snapshot name without config tag
func (d *driver) getPrintableName(name string) string {
	return strings.TrimPrefix(name, d.tag()+ebs.TagDelimiter)
}

// Prefix volume or snapshot name with config tag
func (d *driver) getFullName(name string) string {
	if d.tag() != "" {
		return d.tag() + ebs.TagDelimiter + name
	}
	return name
}

// Retrieve config arguments
func (d *driver) accessKey() string {
	if accessKey := d.config.GetString(
		ebs.ConfigEBSAccessKey); accessKey != "" {
		return accessKey
	}
	return d.config.GetString(ebs.ConfigOldEBSAccessKey)
}

func (d *driver) secretKey() string {
	if secretKey := d.config.GetString(
		ebs.ConfigEBSSecretKey); secretKey != "" {
		return secretKey
	}
	return d.config.GetString(ebs.ConfigOldEBSSecretKey)
}

func (d *driver) region() string {
	if region := d.config.GetString(ebs.ConfigEBSRegion); region != "" {
		return region
	}
	return d.config.GetString(ebs.ConfigOldEBSRegion)
}

func (d *driver) endpoint() string {
	if endpoint := d.config.GetString(ebs.ConfigEBSEndpoint); endpoint != "" {
		return endpoint
	}
	return d.config.GetString(ebs.ConfigOldEBSEndpoint)
}

func (d *driver) maxRetries() int {
	// if maxRetries in config is non-numeric or a negative number,
	// set it to the default number of max retries.
	if maxRetriesString := d.config.GetString(
		ebs.ConfigEBSMaxRetries); maxRetriesString != "" {
		if maxRetriesString == "0" {
			return 0
		} else if maxRetries := d.config.GetInt(
			ebs.ConfigEBSMaxRetries); maxRetries > 0 {
			return maxRetries
		}
	} else if maxRetriesString := d.config.GetString(
		ebs.ConfigOldEBSMaxRetries); maxRetriesString != "" {
		if maxRetriesString == "0" {
			return 0
		} else if maxRetries := d.config.GetInt(
			ebs.ConfigOldEBSMaxRetries); maxRetries > 0 {
			return maxRetries
		}
	}
	return ebs.DefaultMaxRetries
}

func (d *driver) tag() string {
	if tag := d.config.GetString(ebs.ConfigEBSTag); tag != "" {
		return tag
	}
	return d.config.GetString(ebs.ConfigOldEBSTag)
}

// TODO rexrayTag
/*func (d *driver) rexrayTag() string {
	if rexrayTag := d.config.GetString("ebs.rexrayTag"); rexrayTag != "" {
		return rexrayTag
	}
	return d.config.GetString("ec2.rexrayTag")
}*/
