package storage

import (
	"sync"

	log "github.com/sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	apiUtils "github.com/AVENTER-UG/rexray/libstorage/api/utils"

	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/s3fs"
	s3fsUtils "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/s3fs/utils"
)

type driver struct {
	config           gofig.Config
	tag              string
	region           string
	accessKey        string
	secretKey        string
	endpoint         string
	maxRetries       int
	disablePathStyle bool
	svcs             map[string]*awss3.S3
	svcsRWL          *sync.RWMutex
}

func init() {
	registry.RegisterStorageDriver(s3fs.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{}
}

func (d *driver) Name() string {
	return s3fs.Name
}

// Init initializes the driver.
func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	d.svcs = map[string]*awss3.S3{}
	d.svcsRWL = &sync.RWMutex{}

	fields := log.Fields{}

	d.tag = d.config.GetString(s3fs.ConfigS3FSTag)
	fields[s3fs.Tag] = d.tag

	d.accessKey = d.config.GetString(s3fs.ConfigS3FSAccessKey)
	fields[s3fs.AccessKey] = d.accessKey

	d.secretKey = d.config.GetString(s3fs.ConfigS3FSSecretKey)
	if d.secretKey != "" {
		fields[s3fs.SecretKey] = "******"
	}

	d.endpoint = d.config.GetString(s3fs.ConfigS3FSEndpoint)
	fields[s3fs.Endpoint] = d.endpoint

	d.region = d.config.GetString(s3fs.ConfigS3FSRegion)
	fields[s3fs.Region] = d.region

	d.maxRetries = d.config.GetInt(s3fs.ConfigS3FSMaxRetries)
	fields[s3fs.MaxRetries] = d.maxRetries

	d.disablePathStyle = d.config.GetBool(s3fs.ConfigS3FSDisablePathStyle)
	fields[s3fs.DisablePathStyle] = d.disablePathStyle

	if _, err := d.getService(ctx, d.region); err != nil {
		return err
	}

	ctx.WithFields(fields).Info("storage driver initialized")
	return nil
}

func (d *driver) getService(
	ctx types.Context,
	region string) (*awss3.S3, error) {

	if region == "" {
		region = d.region
	}

	ctx.WithField("region", region).Debug("getting s3 service connection")

	svc := d.serviceExists(region)
	if svc != nil {
		ctx.WithField("region", region).Debug(
			"got existing s3 service connection")
		return svc, nil
	}

	ctx.WithField("region", region).Debug("creating new s3 service connection")

	d.svcsRWL.Lock()
	defer d.svcsRWL.Unlock()

	sess := session.New()

	var (
		awsLogger   = &awsLogger{ctx: ctx}
		awsLogLevel = aws.LogOff
	)
	if ll, ok := context.GetLogLevel(ctx); ok {
		switch ll {
		case log.DebugLevel:
			awsLogger.lvl = log.DebugLevel
			awsLogLevel = aws.LogDebugWithHTTPBody
		case log.InfoLevel:
			awsLogger.lvl = log.InfoLevel
			awsLogLevel = aws.LogDebug
		}
	}

	config := &aws.Config{
		Region:           &region,
		MaxRetries:       &d.maxRetries,
		S3ForcePathStyle: aws.Bool(!d.disablePathStyle),
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     d.accessKey,
						SecretAccessKey: d.secretKey,
					},
				},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				&ec2rolecreds.EC2RoleProvider{
					Client: ec2metadata.New(sess),
				},
			},
		),
		Logger:   awsLogger,
		LogLevel: aws.LogLevel(awsLogLevel),
	}

	if d.endpoint != "" {
		config.Endpoint = &d.endpoint
	}

	svc = awss3.New(sess, config)
	ctx.WithField("region", region).Debug("s3 connection created")

	if _, err := svc.ListBuckets(&awss3.ListBucketsInput{}); err != nil {
		ctx.WithField("region", region).WithError(err).Error(
			"s3 connection failed")
		return nil, err
	}

	d.svcs[region] = svc
	ctx.WithField("region", region).Debug("s3 connection successful")

	return svc, nil
}

func (d *driver) serviceExists(region string) *awss3.S3 {
	d.svcsRWL.RLock()
	defer d.svcsRWL.RUnlock()
	if svc, ok := d.svcs[region]; ok {
		return svc
	}
	return nil
}

type awsLogger struct {
	lvl log.Level
	ctx types.Context
}

func (a *awsLogger) Log(args ...interface{}) {
	switch a.lvl {
	case log.DebugLevel:
		a.ctx.Debugln(args...)
	case log.InfoLevel:
		a.ctx.Infoln(args...)
	}
}

func mustInstanceIDID(ctx types.Context) *string {
	return &context.MustInstanceID(ctx).ID
}

// NextDeviceInfo returns the information about the driver's next available
func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return s3fsUtils.NextDeviceInfo, nil
}

// Type returns the type of storage the driver provides.
func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	return types.Object, nil
}

// InstanceInspect returns an instance.
func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	iid := context.MustInstanceID(ctx)
	return &types.Instance{
		Name:         iid.ID,
		InstanceID:   iid,
		ProviderName: iid.Driver,
	}, nil
}

// Volumes returns all volumes or a filtered list of volumes.
func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	vols, err := d.toTypeVolumes(ctx, opts.Attachments)
	if err != nil {
		return nil, goof.WithError("error getting s3fs volumes", err)
	}
	return vols, nil
}

// VolumeInspect inspects a single volume.
func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	return d.getVolume(ctx, volumeID, opts.Attachments)
}

// VolumeCreate creates a new volume.
func (d *driver) VolumeCreate(ctx types.Context, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	if opts.Encrypted != nil && *opts.Encrypted {
		return nil, types.ErrNotImplemented
	}

	var cbc *awss3.CreateBucketConfiguration
	if d.region != "us-east-1" {
		cbc = &awss3.CreateBucketConfiguration{LocationConstraint: &d.region}
	}

	svc, _ := d.getService(ctx, "")

	_, err := svc.CreateBucket(
		&awss3.CreateBucketInput{
			Bucket: &volumeName,
			CreateBucketConfiguration: cbc,
		})
	if err != nil {
		ctx.WithField(
			"volumeName", volumeName).WithError(err).Error(
			"error creating s3 bucket")
		return nil, goof.WithFieldE(
			"volumeName", volumeName, "error creating s3 bucket", err)
	}

	return d.toTypeVolume(ctx, volumeName, types.VolAttNone), nil
}

// VolumeCreateFromSnapshot creates a new volume from an existing snapshot.
func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// VolumeCopy copies an existing volume.
func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// VolumeSnapshot snapshots a volume.
func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// VolumeRemove removes a volume.
func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeRemoveOpts) error {

	var svc *awss3.S3

	{
		var err error
		if svc, err = d.getServiceForBucket(ctx, volumeID); err != nil {
			return err
		}
	}

	if opts.Force {

		{
			// delete objects
			res, err := svc.ListObjects(
				&awss3.ListObjectsInput{Bucket: &volumeID})
			if err != nil {
				ctx.WithField(
					"volumeID", volumeID).WithError(err).Error(
					"error listing objects")
				return goof.WithFieldE(
					"volumeID", volumeID,
					"error listing objects", err)
			}
			if res.Contents != nil {
				for {
					var key *string
					for _, obj := range res.Contents {
						if obj.Key == nil {
							continue
						}
						key = obj.Key
						_, err := svc.DeleteObject(&awss3.DeleteObjectInput{
							Bucket: &volumeID,
							Key:    key,
						})
						if err != nil {
							ctx.WithFields(map[string]interface{}{
								"volumeID": volumeID,
								"key":      key,
							}).WithError(err).Error("error deleting object")
							return goof.WithFieldsE(map[string]interface{}{
								"volumeID": volumeID,
								"key":      key,
							}, "error deleting object", err)
						}
					}

					if res.IsTruncated == nil || !*res.IsTruncated {
						break
					}

					var err error
					keyMarker := res.NextMarker
					if keyMarker == nil {
						keyMarker = key
					}
					res, err = svc.ListObjects(
						&awss3.ListObjectsInput{
							Bucket: &volumeID,
							Marker: keyMarker,
						})
					if err != nil {
						ctx.WithFields(map[string]interface{}{
							"volumeID":  volumeID,
							"keyMarker": keyMarker,
						}).WithError(err).Error("error listing next objects")
						return goof.WithFieldsE(map[string]interface{}{
							"volumeID":  volumeID,
							"keyMarker": keyMarker,
						}, "error listing next objects", err)
					}
				}
			}
		}

		{
			// delete versions
			res, err := svc.ListObjectVersions(
				&awss3.ListObjectVersionsInput{Bucket: &volumeID})
			if err != nil {
				ctx.WithField(
					"volumeID", volumeID).WithError(err).Error(
					"error listing object versions")
				return goof.WithFieldE(
					"volumeID", volumeID,
					"error listing object versions", err)
			}
			if res.Versions != nil {
				for {
					var key *string
					var ver *string
					for _, obj := range res.Versions {
						if obj.Key == nil {
							continue
						}
						if obj.VersionId == nil {
							continue
						}
						key = obj.Key
						ver = obj.VersionId
						_, err := svc.DeleteObject(&awss3.DeleteObjectInput{
							Bucket:    &volumeID,
							Key:       key,
							VersionId: ver,
						})
						if err != nil {
							ctx.WithFields(map[string]interface{}{
								"volumeID":  volumeID,
								"key":       key,
								"versionID": ver,
							}).WithError(err).Error("error deleting object ver")
							return goof.WithFieldsE(map[string]interface{}{
								"volumeID":  volumeID,
								"key":       key,
								"versionID": ver,
							}, "error deleting object ver", err)
						}
					}

					if res.IsTruncated == nil || !*res.IsTruncated {
						break
					}

					var err error
					keyMarker := res.NextKeyMarker
					verMarker := res.NextVersionIdMarker
					if keyMarker == nil {
						keyMarker = key
					}
					if verMarker == nil {
						verMarker = ver
					}
					res, err = svc.ListObjectVersions(
						&awss3.ListObjectVersionsInput{
							Bucket:          &volumeID,
							KeyMarker:       keyMarker,
							VersionIdMarker: verMarker,
						})
					if err != nil {
						ctx.WithFields(map[string]interface{}{
							"volumeID":  volumeID,
							"keyMarker": keyMarker,
							"verMarkre": verMarker,
						}).WithError(err).Error("error listing next objects")
						return goof.WithFieldsE(map[string]interface{}{
							"volumeID":  volumeID,
							"keyMarker": keyMarker,
							"verMarkre": verMarker,
						}, "error listing next objects", err)
					}
				}
			}
		}
	}

	_, err := svc.DeleteBucket(&awss3.DeleteBucketInput{Bucket: &volumeID})
	if err != nil {
		ctx.WithField(
			"volumeID", volumeID).WithError(err).Error(
			"error deleting s3 bucket")
		return goof.WithFieldE(
			"volumeID", volumeID, "error deleting s3 bucket", err)
	}

	return nil
}

// VolumeAttach attaches a volume and provides a token clients can use
// to validate that device has appeared locally.
func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	vol, err := d.getVolume(ctx, volumeID, types.VolAttReq)
	if err != nil {
		return nil, "", err
	}
	return vol, "", nil
}

// VolumeDetach detaches a volume.
func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	vol, err := d.getVolume(ctx, volumeID, types.VolAttReq)
	if err != nil {
		return nil, err
	}
	return vol, nil
}

// Snapshots returns all volumes or a filtered list of snapshots.
func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// SnapshotInspect inspects a single snapshot.
func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// SnapshotCopy copies an existing snapshot.
func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// SnapshotRemove removes a snapshot.
func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {
	// TODO Snapshots are not implemented yet
	return types.ErrNotImplemented
}

var errGetLocDevs = goof.New("error getting local devices from context")

func (d *driver) toTypeVolumes(
	ctx types.Context,
	attachments types.VolumeAttachmentsTypes) ([]*types.Volume, error) {

	svc, _ := d.getService(ctx, "")

	res, err := svc.ListBuckets(&awss3.ListBucketsInput{})
	if err != nil {
		ctx.WithError(err).Error("error listing s3 buckets")
		return nil, goof.WithError("error listing s3 buckets", err)
	}

	var vols []*types.Volume
	for _, b := range res.Buckets {

		vols = append(vols, d.toTypeVolume(ctx, *b.Name, attachments))
	}
	return vols, nil
}

func (d *driver) toTypeVolume(
	ctx types.Context,
	bucket string,
	attachments types.VolumeAttachmentsTypes) *types.Volume {

	vol := &types.Volume{
		Name: bucket,
		ID:   bucket,
	}

	iid, iidOK := context.InstanceID(ctx)
	if iidOK && attachments.Requested() {
		vatt := &types.VolumeAttachment{
			VolumeID:   bucket,
			DeviceName: bucket,
			InstanceID: iid,
		}
		if attachments.Devices() {
			if ld, ldOK := context.LocalDevices(ctx); ldOK {
				if mp, mpOK := ld.DeviceMap[bucket]; mpOK {
					vatt.MountPoint = mp
				}
			}
		}
		vol.Attachments = []*types.VolumeAttachment{vatt}
	}

	return vol
}

func (d *driver) getVolume(
	ctx types.Context,
	volumeID string,
	attachments types.VolumeAttachmentsTypes) (*types.Volume, error) {

	svc, _ := d.getService(ctx, "")
	req, _ := svc.HeadBucketRequest(&awss3.HeadBucketInput{Bucket: &volumeID})
	if err := req.Send(); err != nil && req.HTTPResponse.StatusCode != 301 {
		return nil, apiUtils.NewNotFoundError(volumeID)
	}
	return d.toTypeVolume(ctx, volumeID, attachments), nil
}

func (d *driver) getServiceForBucket(
	ctx types.Context,
	bucket string) (*awss3.S3, error) {

	svc, _ := d.getService(ctx, "")
	res, err := svc.GetBucketLocation(
		&awss3.GetBucketLocationInput{Bucket: &bucket})
	if err != nil {
		return nil, apiUtils.NewNotFoundError(bucket)
	}
	var region string
	if res.LocationConstraint != nil {
		region = *res.LocationConstraint
	}
	if svc, err = d.getService(ctx, region); err != nil {
		return nil, err
	}
	return svc, nil
}
