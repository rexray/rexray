package storage

import (
	"crypto/md5"
	"fmt"
	"hash"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	awsefs "github.com/aws/aws-sdk-go/service/efs"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	apiUtils "github.com/AVENTER-UG/rexray/libstorage/api/utils"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/efs"
)

const (
	tagDelimiter = "/"
)

// Driver represents a EFS driver implementation of StorageDriver
type driver struct {
	config              gofig.Config
	region              *string
	endpoint            *string
	endpointFormat      string
	maxRetries          *int
	tag                 string
	accessKey           string
	secGroups           []string
	disableSessionCache bool
	maxAttempts         int
	statusDelay         int64
	statusTimeout       time.Duration
}

func init() {
	registry.RegisterStorageDriver(efs.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{}
}

// Name returns the name of the driver
func (d *driver) Name() string {
	return efs.Name
}

// Init initializes the driver.
func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config

	fields := log.Fields{}

	d.accessKey = d.getAccessKey()
	fields["accessKey"] = d.accessKey

	d.tag = d.getTag()
	fields["tag"] = d.tag

	d.secGroups = d.getSecurityGroups()
	fields["securityGroups"] = d.secGroups

	d.disableSessionCache = d.getDisableSessionCache()
	fields["disableSessionCache"] = d.disableSessionCache

	if v := d.getRegion(); v != "" {
		d.region = &v
		fields["region"] = v
	}
	if v := d.getEndpoint(); v != "" {
		d.endpoint = &v
		fields["endpoint"] = v
	}
	d.endpointFormat = d.getEndpointFormat()
	fields["endpointFormat"] = d.endpointFormat
	maxRetries := d.getMaxRetries()
	d.maxRetries = &maxRetries
	fields["maxRetries"] = maxRetries

	d.maxAttempts = d.config.GetInt(efs.ConfigStatusMaxAttempts)
	fields["maxStatusAttempts"] = d.maxAttempts

	statusDelayStr := d.config.GetString(efs.ConfigStatusInitDelay)
	statusDelay, err := time.ParseDuration(statusDelayStr)
	if err != nil {
		return err
	}
	d.statusDelay = statusDelay.Nanoseconds()
	fields["statusDelay"] = fmt.Sprintf(
		"%v", time.Duration(d.statusDelay)*time.Nanosecond)

	statusTimeoutStr := d.config.GetString(efs.ConfigStatusTimeout)
	d.statusTimeout, err = time.ParseDuration(statusTimeoutStr)
	if err != nil {
		return err
	}
	fields["statusTimeout"] = d.statusTimeout

	ctx.WithFields(fields).Info("storage driver initialized")
	return nil
}

const cacheKeyC = "cacheKey"

var (
	sessions  = map[string]*awsefs.EFS{}
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
		endpoint *string
		ckey     string
		hkey     = md5.New()
		akey     = d.accessKey
		region   = d.mustRegion(ctx)
	)

	if region != nil && d.endpointFormat != "" {
		szEndpoint := fmt.Sprintf(d.endpointFormat, *region)
		endpoint = &szEndpoint
	} else {
		endpoint = d.endpoint
	}

	if !d.disableSessionCache {
		writeHkey(hkey, region)
		writeHkey(hkey, endpoint)
		writeHkey(hkey, &akey)
		ckey = fmt.Sprintf("%x", hkey.Sum(nil))

		// if the session is cached then return it
		if svc, ok := sessions[ckey]; ok {
			ctx.WithField(cacheKeyC, ckey).Debug("using cached efs service")
			return svc, nil
		}
	}

	var (
		skey   = d.getSecretKey()
		fields = map[string]interface{}{
			efs.AccessKey: akey,
			efs.Tag:       d.tag,
			cacheKeyC:     ckey,
		}
	)

	if skey == "" {
		fields[efs.SecretKey] = ""
	} else {
		fields[efs.SecretKey] = "******"
	}
	if region != nil {
		fields[efs.Region] = *region
	}
	if endpoint != nil {
		fields[efs.Endpoint] = *endpoint
	}

	ctx.WithFields(fields).Debug("efs service connection attempt")
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

	svc := awsefs.New(sess, &aws.Config{
		Region:     region,
		Endpoint:   endpoint,
		MaxRetries: d.maxRetries,
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     akey,
						SecretAccessKey: skey,
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
	})

	ctx.WithFields(fields).Info("efs service connection created")

	if !d.disableSessionCache {
		sessions[ckey] = svc
		ctx.WithFields(fields).Info("efs service connection cached")
	}

	return svc, nil
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

func mustSession(ctx types.Context) *awsefs.EFS {
	return context.MustSession(ctx).(*awsefs.EFS)
}

func mustInstanceIDID(ctx types.Context) *string {
	return &context.MustInstanceID(ctx).ID
}

func (d *driver) mustRegion(ctx types.Context) *string {
	if iid, ok := context.InstanceID(ctx); ok {
		if v, ok := iid.Fields[efs.InstanceIDFieldRegion]; ok && v != "" {
			return &v
		}
	}
	return d.region
}

func (d *driver) mustAvailabilityZone(ctx types.Context) *string {
	if iid, ok := context.InstanceID(ctx); ok {
		if v, ok := iid.Fields[efs.InstanceIDFieldAvailabilityZone]; ok {
			if v != "" {
				return &v
			}
		}
	}
	return nil
}

// InstanceInspect returns an instance.
func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	iid := context.MustInstanceID(ctx)
	return &types.Instance{
		Name:         iid.ID,
		Region:       iid.Fields[efs.InstanceIDFieldRegion],
		InstanceID:   iid,
		ProviderName: iid.Driver,
	}, nil
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

// Volumes returns all volumes or a filtered list of volumes.
func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	svc := mustSession(ctx)

	fileSystems, err := d.getAllFileSystems(svc)
	if err != nil {
		return nil, err
	}

	var volumesSD []*types.Volume
	for _, fileSystem := range fileSystems {
		// Make sure that name is popullated
		if fileSystem.Name == nil {
			ctx.WithFields(log.Fields{
				"filesystemid": *fileSystem.FileSystemId,
			}).Warn("missing EFS filesystem name")
			continue
		}

		// Only volumes with partition prefix
		if !strings.HasPrefix(*fileSystem.Name, d.tag+tagDelimiter) {
			continue
		}

		// Only volumes in "available" state
		if *fileSystem.LifeCycleState != awsefs.LifeCycleStateAvailable {
			continue
		}

		volumeSD := &types.Volume{
			Name:        d.getPrintableName(*fileSystem.Name),
			ID:          *fileSystem.FileSystemId,
			Size:        *fileSystem.SizeInBytes.Value,
			Attachments: nil,
		}

		var atts []*types.VolumeAttachment
		if opts.Attachments.Requested() {
			atts, err = d.getVolumeAttachments(
				ctx, *fileSystem.FileSystemId, opts.Attachments)
			if err != nil {
				return nil, err
			}
		}
		if len(atts) > 0 {
			volumeSD.Attachments = atts
		}
		volumesSD = append(volumesSD, volumeSD)
	}

	return volumesSD, nil
}

// VolumeInspect inspects a single volume.
func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	resp, err := mustSession(ctx).DescribeFileSystems(
		&awsefs.DescribeFileSystemsInput{FileSystemId: aws.String(volumeID)})
	if err != nil {
		return nil, err
	}

	if len(resp.FileSystems) == 0 {
		return nil, apiUtils.NewNotFoundError(volumeID)
	}

	fileSystem := resp.FileSystems[0]

	// Only volumes in "available" state
	if *fileSystem.LifeCycleState != awsefs.LifeCycleStateAvailable {
		return nil, goof.WithField("volumeID", volumeID,
			"Volume not available")
	}

	// Name is optional via tag so make sure it exists
	var fileSystemName string
	if fileSystem.Name != nil {
		fileSystemName = *fileSystem.Name
	} else {
		ctx.WithFields(log.Fields{
			"filesystemid": *fileSystem.FileSystemId,
		}).Warn("missing EFS filesystem name")
	}

	volume := &types.Volume{
		Name:        d.getPrintableName(fileSystemName),
		ID:          *fileSystem.FileSystemId,
		Size:        *fileSystem.SizeInBytes.Value,
		Attachments: nil,
	}

	var atts []*types.VolumeAttachment

	if opts.Attachments.Requested() {
		atts, err = d.getVolumeAttachments(
			ctx, *fileSystem.FileSystemId, opts.Attachments)
		if err != nil {
			return nil, err
		}
	}
	if len(atts) > 0 {
		volume.Attachments = atts
	}
	return volume, nil
}

// VolumeCreate creates a new volume.
func (d *driver) VolumeCreate(
	ctx types.Context,
	name string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	// Token is limited to 64 ASCII characters so just create MD5 hash from full
	// tag/name identifier
	creationToken := fmt.Sprintf("%x", md5.Sum([]byte(d.getFullVolumeName(name))))
	request := &awsefs.CreateFileSystemInput{
		CreationToken:   aws.String(creationToken),
		PerformanceMode: aws.String(awsefs.PerformanceModeGeneralPurpose),
	}
	if opts.Type != nil && strings.ToLower(*opts.Type) == "maxio" {
		request.PerformanceMode = aws.String(awsefs.PerformanceModeMaxIo)
	}

	svc := mustSession(ctx)
	fileSystem, err := svc.CreateFileSystem(request)

	if err != nil {
		return nil, err
	}

	_, err = svc.CreateTags(&awsefs.CreateTagsInput{
		FileSystemId: fileSystem.FileSystemId,
		Tags: []*awsefs.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(d.getFullVolumeName(name)),
			},
		},
	})

	if err != nil {
		// To not leak the EFS instances remove the filesystem that couldn't
		// be tagged with correct name before returning error response.
		_, deleteErr := svc.DeleteFileSystem(
			&awsefs.DeleteFileSystemInput{
				FileSystemId: fileSystem.FileSystemId,
			})
		if deleteErr != nil {
			ctx.WithFields(log.Fields{
				"filesystemid": *fileSystem.FileSystemId,
			}).Error("failed to delete EFS")
		}

		return nil, err
	}

	f := func() (interface{}, error) {
		duration := d.statusDelay
		for i := 1; i <= d.maxAttempts; i++ {
			state, err := d.getFileSystemLifeCycleState(
				svc, *fileSystem.FileSystemId)
			if err != nil {
				return nil, goof.WithFieldE(
					"filesystemid",
					*fileSystem.FileSystemId,
					"failed to retreive EFS state",
					err)
			}
			if state == awsefs.LifeCycleStateAvailable {
				return nil, nil
			}
			ctx.WithFields(log.Fields{
				"state":        state,
				"filesystemid": *fileSystem.FileSystemId,
			}).Debug("EFS not ready")

			time.Sleep(time.Duration(duration) * time.Nanosecond)
			duration = int64(2) * duration
		}
		return nil, goof.WithField("maxAttempts", d.maxAttempts,
			"Status attempts exhausted")
	}

	// Wait until FS is in "available" state
	_, ok, err := apiUtils.WaitFor(f, d.statusTimeout)
	if !ok {
		return nil, goof.WithFields(goof.Fields{
			"filesystemid":  *fileSystem.FileSystemId,
			"statusTimeout": d.statusTimeout},
			"Timeout occured waiting for filesystem status")
	}
	if err != nil {
		return nil, goof.WithFieldE("filesystemid", *fileSystem.FileSystemId,
			"Error while waiting for storage action to finish", err)
	}

	return d.VolumeInspect(ctx, *fileSystem.FileSystemId,
		&types.VolumeInspectOpts{Attachments: 0})
}

// VolumeRemove removes a volume.
func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeRemoveOpts) error {

	svc := mustSession(ctx)

	// Remove MountTarget(s)
	resp, err := svc.DescribeMountTargets(
		&awsefs.DescribeMountTargetsInput{
			FileSystemId: aws.String(volumeID),
		})
	if err != nil {
		return err
	}

	for _, mountTarget := range resp.MountTargets {
		_, err = svc.DeleteMountTarget(
			&awsefs.DeleteMountTargetInput{
				MountTargetId: aws.String(*mountTarget.MountTargetId),
			})

		if err != nil {
			return err
		}
	}

	// FileSystem can be deleted only after all mountpoints are deleted (
	// just in "deleting" life cycle state). Here code will wait until all
	// mountpoints are deleted.
	f := func() (interface{}, error) {
		duration := d.statusDelay
		for i := 1; i <= d.maxAttempts; i++ {
			resp, err := svc.DescribeMountTargets(
				&awsefs.DescribeMountTargetsInput{
					FileSystemId: aws.String(volumeID),
				})
			if err != nil {
				return nil, err
			}

			if len(resp.MountTargets) == 0 {
				return nil, nil
			}
			ctx.WithFields(log.Fields{
				"mounttargets": resp.MountTargets,
				"filesystemid": volumeID,
			}).Debug("waiting for MountTargets deletion")

			time.Sleep(time.Duration(duration) * time.Nanosecond)
			duration = int64(2) * duration
		}
		return nil, goof.WithField("maxAttempts", d.maxAttempts,
			"Status attempts exhausted")
	}
	_, ok, err := apiUtils.WaitFor(f, d.statusTimeout)
	if !ok {
		return goof.WithFields(goof.Fields{
			"filesystemid":  volumeID,
			"statusTimeout": d.statusTimeout},
			"Timeout occured waiting for MountTargets deletion")
	}
	if err != nil {
		return goof.WithFieldE("filesystemid", volumeID,
			"Error while waiting for MountTargets deletion", err)
	}

	// Remove FileSystem
	_, err = svc.DeleteFileSystem(
		&awsefs.DeleteFileSystemInput{
			FileSystemId: aws.String(volumeID),
		})
	if err != nil {
		return err
	}

	f = func() (interface{}, error) {
		duration := d.statusDelay
		for i := 1; i <= d.maxAttempts; i++ {
			ctx.WithFields(log.Fields{
				"filesystemid": volumeID,
			}).Info("waiting for FileSystem deletion")

			_, err := svc.DescribeFileSystems(
				&awsefs.DescribeFileSystemsInput{
					FileSystemId: aws.String(volumeID),
				})
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					if awsErr.Code() == "FileSystemNotFound" {
						return nil, nil
					}
					return nil, err
				}
				return nil, err
			}
			time.Sleep(time.Duration(duration) * time.Nanosecond)
			duration = int64(2) * duration
		}
		return nil, goof.WithField("maxAttempts", d.maxAttempts,
			"Status attempts exhausted")
	}
	_, ok, err = apiUtils.WaitFor(f, d.statusTimeout)
	if !ok {
		return goof.WithFields(goof.Fields{
			"filesystemid":  volumeID,
			"statusTimeout": d.statusTimeout},
			"Timeout occured waiting for FileSystem deletion")
	}
	if err != nil {
		return goof.WithFieldE("filesystemid", volumeID,
			"Error while waiting for FileSystem deletion", err)
	}

	return nil
}

var errInvalidSecGroups = goof.New("security groups required")

// VolumeAttach attaches a volume and provides a token clients can use
// to validate that device has appeared locally.
func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	svc := mustSession(ctx)

	vol, err := d.VolumeInspect(ctx, volumeID,
		&types.VolumeInspectOpts{Attachments: types.VolAttReqTrue})
	if err != nil {
		return nil, "", err
	}

	iid := context.MustInstanceID(ctx)

	var ma *types.VolumeAttachment
	for _, att := range vol.Attachments {
		if att.InstanceID.ID == iid.ID {
			ma = att
			break
		}
	}

	// No mount targets were found
	if ma == nil {

		var iSecGrpIDs []string
		secGrpIDs := d.secGroups
		if v, ok := iid.Fields[efs.InstanceIDFieldSecurityGroups]; ok {
			iSecGrpIDs = strings.Split(v, ";")
			if len(iSecGrpIDs) == 1 {
				ctx.WithField("secGrpIDs", iSecGrpIDs).Debug(
					"using instance security group IDs")
				secGrpIDs = iSecGrpIDs
			}
		}

		if len(secGrpIDs) == 0 {
			return nil, "", errInvalidSecGroups
		}

		// make sure all of the request security groups
		// are available on the instance
		var missingSecGrpIDs []string
		for _, csg := range secGrpIDs {
			var found bool
			for _, isg := range iSecGrpIDs {
				if csg == isg {
					found = true
					break
				}
			}
			if !found {
				missingSecGrpIDs = append(missingSecGrpIDs, csg)
			}
		}

		// log a warning if any of the server-side defined SGs
		// are not present in the list sent by the client instance
		if len(missingSecGrpIDs) > 0 {
			log.WithField("missingStorageGroups", missingSecGrpIDs).Warn(
				"configured sec grps not present on instance")
		}

		request := &awsefs.CreateMountTargetInput{
			FileSystemId:   aws.String(vol.ID),
			SubnetId:       aws.String(iid.ID),
			SecurityGroups: aws.StringSlice(secGrpIDs),
		}
		// TODO(mhrabovcin): Should we block here until MountTarget is in
		// "available" LifeCycleState? Otherwise mount could fail until creation
		//  is completed.
		_, err = svc.CreateMountTarget(request)
		// Failed to create mount target
		if err != nil {
			return nil, "", err
		}
	}

	return vol, "", err
}

// VolumeDetach detaches a volume.
func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	// TODO(kasisnu): Think about what to do here?
	// It is safe to remove the mount target
	// when it is no longer being used anywhere
	return nil, nil
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

// Retrieve all filesystems with tags from AWS API. This is very expensive
// operation as it issues AWS SDK call per filesystem to retrieve tags.
func (d *driver) getAllFileSystems(
	svc *awsefs.EFS) (filesystems []*awsefs.FileSystemDescription, err error) {

	resp, err := svc.DescribeFileSystems(&awsefs.DescribeFileSystemsInput{})
	if err != nil {
		return nil, err
	}
	filesystems = append(filesystems, resp.FileSystems...)

	for resp.NextMarker != nil {
		resp, err = svc.DescribeFileSystems(&awsefs.DescribeFileSystemsInput{
			Marker: resp.NextMarker,
		})
		if err != nil {
			return nil, err
		}
		filesystems = append(filesystems, resp.FileSystems...)
	}

	return filesystems, nil
}

func (d *driver) getFileSystemLifeCycleState(
	svc *awsefs.EFS,
	fileSystemID string) (string, error) {

	resp, err := svc.DescribeFileSystems(
		&awsefs.DescribeFileSystemsInput{
			FileSystemId: aws.String(fileSystemID)})
	if err != nil {
		return "", err
	}

	fileSystem := resp.FileSystems[0]
	return *fileSystem.LifeCycleState, nil
}

func (d *driver) getPrintableName(name string) string {
	return strings.TrimPrefix(name, d.tag+tagDelimiter)
}

func (d *driver) getFullVolumeName(name string) string {
	return d.tag + tagDelimiter + name
}

var errGetLocDevs = goof.New("error getting local devices from context")

func (d *driver) getVolumeAttachments(
	ctx types.Context,
	volumeID string,
	attachments types.VolumeAttachmentsTypes) (
	[]*types.VolumeAttachment, error) {

	if !attachments.Requested() {
		return nil, nil
	}

	if volumeID == "" {
		return nil, goof.New("missing volume ID")
	}

	resp, err := mustSession(ctx).DescribeMountTargets(
		&awsefs.DescribeMountTargetsInput{
			FileSystemId: aws.String(volumeID),
		})
	if err != nil {
		return nil, err
	}

	var (
		ld   *types.LocalDevices
		ldOK bool
	)

	if attachments.Devices() {
		// Get local devices map from context
		if ld, ldOK = context.LocalDevices(ctx); !ldOK {
			return nil, errGetLocDevs
		}
	}

	var atts []*types.VolumeAttachment
	for _, mountTarget := range resp.MountTargets {
		var (
			dev    string
			status string
		)
		if ldOK {
			// TODO(kasisnu): Check lifecycle state and build the path better
			dev = *mountTarget.IpAddress + ":" + "/"
			if _, ok := ld.DeviceMap[dev]; ok {
				status = "Exported and Mounted"
			} else {
				status = "Exported and Unmounted"
			}
		} else {
			status = "Exported"
		}
		attachmentSD := &types.VolumeAttachment{
			VolumeID: *mountTarget.FileSystemId,
			InstanceID: &types.InstanceID{
				ID:     *mountTarget.SubnetId,
				Driver: d.Name(),
			},
			DeviceName: dev,
			Status:     status,
		}
		atts = append(atts, attachmentSD)
	}

	return atts, nil
}

// Retrieve config arguments
func (d *driver) getAccessKey() string {
	return d.config.GetString(efs.ConfigEFSAccessKey)
}

func (d *driver) getSecretKey() string {
	return d.config.GetString(efs.ConfigEFSSecretKey)
}

func (d *driver) getRegion() string {
	return d.config.GetString(efs.ConfigEFSRegion)
}

func (d *driver) getEndpoint() string {
	return d.config.GetString(efs.ConfigEFSEndpoint)
}

func (d *driver) getEndpointFormat() string {
	return d.config.GetString(efs.ConfigEFSEndpointFormat)
}

func (d *driver) getMaxRetries() int {
	return d.config.GetInt(efs.ConfigEFSMaxRetries)
}

func (d *driver) getTag() string {
	return d.config.GetString(efs.ConfigEFSTag)
}

func (d *driver) getSecurityGroups() []string {
	return d.config.GetStringSlice(efs.ConfigEFSSecGroups)
}

func (d *driver) getDisableSessionCache() bool {
	return d.config.GetBool(efs.ConfigEFSDisableSessionCache)
}
