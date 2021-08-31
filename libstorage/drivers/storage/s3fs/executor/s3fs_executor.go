package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	log "github.com/sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"

	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/s3fs"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/s3fs/utils"
)

// driver is the storage executor for the s3fs storage driver.
type driver struct {
	config gofig.Config
	cmd    string
	opts   []string
	szOpts string
}

func init() {
	registry.RegisterStorageExecutor(s3fs.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config

	fields := log.Fields{"driver": s3fs.Name}
	d.cmd = d.config.GetString(s3fs.ConfigS3FSCmd)
	fields["cmd"] = d.cmd

	if v := d.config.GetStringSlice(s3fs.ConfigS3FSOptions); len(v) > 0 {
		d.opts = v
		fields["opts"] = d.opts
	} else {
		d.szOpts = d.config.GetString(s3fs.ConfigS3FSOptions)
		fields["opts"] = d.szOpts
	}

	ctx.WithFields(fields).Debug("storage executor initialized")
	return nil
}

func (d *driver) Name() string {
	return s3fs.Name
}

// Supported returns a flag indicating whether or not the platform
// implementing the executor is valid for the host on which the executor
// resides.
func (d *driver) Supported(
	ctx types.Context,
	opts types.Store) (bool, error) {

	return gotil.FileExistsInPath(d.cmd), nil
}

// InstanceID
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {
	return utils.InstanceID(ctx, d.config)
}

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "", types.ErrNotImplemented
}

// Return list of local devices
func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	buckets, err := d.getMountedBuckets(ctx)
	if err != nil {
		return nil, err
	}
	return &types.LocalDevices{Driver: s3fs.Name, DeviceMap: buckets}, nil
}

func (d *driver) Mount(
	ctx types.Context,
	deviceName, mountPoint string,
	opts *types.DeviceMountOpts) error {

	if mp, ok := d.findMountPoint(ctx, deviceName); ok {
		ctx.WithFields(log.Fields{
			"bucket":     deviceName,
			"mountPoint": mp,
		}).Debug("bucket is already mounted")
		if mp == mountPoint {
			// bucket is mounted to the required target => ok
			return nil
		}
		// bucket is mounted to another target => error
		return goof.WithFields(goof.Fields{
			"bucket":      deviceName,
			"mountPointt": mp,
		}, "bucket is already mounted")
	}
	return d.s3fsMount(ctx, deviceName, mountPoint, opts)
}

// Mounts get a list of mount points.
func (d *driver) Mounts(
	ctx types.Context,
	opts types.Store) ([]*types.MountInfo, error) {

	ld, err := d.LocalDevices(ctx, &types.LocalDevicesOpts{Opts: opts})
	if err != nil {
		return nil, err
	}

	if len(ld.DeviceMap) == 0 {
		return nil, nil
	}

	mounts := []*types.MountInfo{}
	for k, v := range ld.DeviceMap {
		mounts = append(mounts, &types.MountInfo{Source: k, MountPoint: v})
	}

	return mounts, nil
}

func (d *driver) s3fsMount(
	ctx types.Context,
	bucket, mountPoint string,
	opts *types.DeviceMountOpts) error {

	args := []string{bucket, mountPoint}
	if len(d.opts) > 0 {
		for _, o := range d.opts {
			args = append(args, fmt.Sprintf("-o%s", o))
		}
	} else if d.szOpts != "" {
		args = append(args, d.szOpts)
	}

	fields := map[string]interface{}{
		"bucket":           bucket,
		"mountPoint":       mountPoint,
		"cmd":              d.cmd,
		"args":             args,
		"isAWSAuthEnvVars": false,
	}

	cmd := exec.Command(d.cmd, args...)
	if ak := d.getAccessKey(); ak != "" {
		if sk := d.getSecretKey(); sk != "" {
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, fmt.Sprintf("AWSACCESSKEYID=%s", ak))
			cmd.Env = append(cmd.Env, fmt.Sprintf("AWSSECRETACCESSKEY=%s", sk))
			fields["isAWSAuthEnvVars"] = true
		}
	}

	ctx.WithFields(fields).Debug("attempting s3fs mount")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fields["output"] = string(out)
		return goof.WithFieldsE(fields, "error mounting s3fs bucket", err)
	}

	return nil
}

func (d *driver) findMountPoint(
	ctx types.Context,
	bucket string) (string, bool) {

	if buckets, err := d.getMountedBuckets(ctx); err == nil {
		b, ok := buckets[bucket]
		return b, ok
	}
	return "", false
}

func (d *driver) getMountedBuckets(
	ctx types.Context) (map[string]string, error) {

	return getMountedBuckets(ctx, path.Base(d.cmd))
}

func (d *driver) getAccessKey() string {
	return d.config.GetString(s3fs.ConfigS3FSAccessKey)
}

func (d *driver) getSecretKey() string {
	return d.config.GetString(s3fs.ConfigS3FSSecretKey)
}
