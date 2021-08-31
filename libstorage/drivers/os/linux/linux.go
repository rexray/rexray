// +build linux

/*
Package linux is the OS driver for linux. In order to reduce external
dependencies, this package borrows the following packages:

  - github.com/docker/docker/pkg/mount
  - github.com/opencontainers/runc/libcontainer/label
*/
package linux

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"

	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

const driverName = "linux"

var (
	errUnknownOS             = goof.New("unknown OS")
	errUnknownFileSystem     = goof.New("unknown file system")
	errUnsupportedFileSystem = goof.New("unsupported file system")
)

func init() {
	registry.RegisterOSDriver(driverName, newDriver)

	r := gofigCore.NewRegistration("Linux")
	r.Key(gofig.Int, "", 0700, "", "linux.volume.filemode")
	r.Key(gofig.String, "", "/data", "", "linux.volume.rootpath")
	gofigCore.Register(r)
}

type driver struct {
	config gofig.Config
}

func newDriver() types.OSDriver {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	if runtime.GOOS != "linux" {
		return errUnknownOS
	}
	d.config = config
	return nil
}

func (d *driver) Name() string {
	return driverName
}

func (d *driver) Mounts(
	ctx types.Context,
	deviceName, mountPoint string,
	opts types.Store) ([]*types.MountInfo, error) {

	mounts, err := d.getMounts(ctx, opts)
	if err != nil {
		return nil, err
	}

	if mountPoint == "" && deviceName == "" {
		return mounts, nil
	} else if mountPoint != "" && deviceName != "" {
		return nil, goof.New("cannot specify mountPoint and deviceName")
	}

	matchedMounts := []*types.MountInfo{}
	for _, m := range mounts {
		if m.MountPoint == mountPoint || m.Source == deviceName {
			matchedMounts = append(matchedMounts, m)
		}
	}
	return matchedMounts, nil
}

func (d *driver) getMounts(
	ctx types.Context,
	opts types.Store) ([]*types.MountInfo, error) {

	// see if we should use the executor?
	if client, ok := context.Client(ctx); ok {
		if _, ok := context.ServiceName(ctx); ok {
			if client.Executor() != nil {
				lsxSO, err := client.Executor().Supported(ctx, opts)
				if err != nil {
					return nil, err
				}
				if lsxSO.Mounts() {
					return client.Executor().Mounts(ctx, opts)
				}
			}
		}
	}

	return getMounts()
}

func (d *driver) Mount(
	ctx types.Context,
	deviceName, mountPoint string,
	opts *types.DeviceMountOpts) error {

	// see if we should use the executor?
	if client, ok := context.Client(ctx); ok {
		if _, ok := context.ServiceName(ctx); ok {
			if client.Executor() != nil {
				lsxSO, err := client.Executor().Supported(ctx, opts.Opts)
				if err != nil {
					return err
				}
				if lsxSO.Mount() {
					if err := client.Executor().Mount(
						ctx, deviceName, mountPoint, opts); err != nil {
						return err
					}
					mp := d.volumeMountPath(mountPoint)
					fm := d.fileModeMountPath()
					os.MkdirAll(mp, fm)
					os.Chmod(mp, fm)
					return nil
				}
			}
		}
	}

	if d.isNfsDevice(deviceName) {
		if err := d.nfsMount(deviceName, mountPoint); err != nil {
			return err
		}
		os.MkdirAll(d.volumeMountPath(mountPoint), d.fileModeMountPath())
		os.Chmod(d.volumeMountPath(mountPoint), d.fileModeMountPath())
		return nil
	}

	fsType := opts.FsType
	if opts.FsType == "" {
		var err error
		fsType, err = probeFsType(deviceName)
		if err != nil {
			return err
		}
	}

	options := fmt.Sprintf("%s,%s", opts.MountOptions, opts.MountLabel)
	if fsType == "xfs" {
		options = fmt.Sprintf("%s,nouuid", opts.MountLabel)
	}

	if err := mount(deviceName, mountPoint, fsType, options); err != nil {
		return goof.WithFieldsE(goof.Fields{
			"deviceName": deviceName,
			"mountPoint": mountPoint,
		}, "error mounting directory", err)
	}

	os.MkdirAll(d.volumeMountPath(mountPoint), d.fileModeMountPath())
	os.Chmod(d.volumeMountPath(mountPoint), d.fileModeMountPath())

	return nil
}

func (d *driver) Unmount(
	ctx types.Context,
	mountPoint string,
	opts types.Store) error {

	// see if we should use the executor?
	if client, ok := context.Client(ctx); ok {
		if _, ok := context.ServiceName(ctx); ok {
			if client.Executor() != nil {
				lsxSO, err := client.Executor().Supported(ctx, opts)
				if err != nil {
					return err
				}
				if lsxSO.Umount() {
					return client.Executor().Unmount(ctx, mountPoint, opts)
				}
			}
		}
	}

	return unmount(mountPoint)
}

func (d *driver) IsMounted(
	ctx types.Context,
	mountPoint string,
	opts types.Store) (bool, error) {

	mounts, err := d.getMounts(ctx, opts)
	if err != nil {
		return false, err
	}

	for _, m := range mounts {
		if m.MountPoint == mountPoint {
			return true, nil
		}
	}

	return false, nil
}

func (d *driver) Format(
	ctx types.Context,
	deviceName string,
	opts *types.DeviceFormatOpts) error {

	fsType, err := probeFsType(deviceName)
	if err != nil && err != errUnknownFileSystem {
		return err
	}
	fsDetected := fsType != ""

	ctx.WithFields(log.Fields{
		"fsDetected":  fsDetected,
		"fsType":      fsType,
		"deviceName":  deviceName,
		"overwriteFs": opts.OverwriteFS,
		"driverName":  driverName}).Info("probe information")

	if opts.OverwriteFS || !fsDetected {
		switch opts.NewFSType {
		case "ext4":
			if err := exec.Command(
				"mkfs.ext4", "-F", deviceName).Run(); err != nil {
				return goof.WithFieldE(
					"deviceName", deviceName,
					"error creating filesystem",
					err)
			}
		case "xfs":
			if err := exec.Command(
				"mkfs.xfs", "-f", deviceName).Run(); err != nil {
				return goof.WithFieldE(
					"deviceName", deviceName,
					"error creating filesystem",
					err)
			}
		default:
			return errUnsupportedFileSystem
		}
	}

	return nil
}

func (d *driver) isNfsDevice(device string) bool {
	return strings.Contains(device, ":")
}

func (d *driver) nfsMount(device, target string) error {
	command := exec.Command("mount", device, target)
	output, err := command.CombinedOutput()
	if err != nil {
		return goof.WithError(fmt.Sprintf("failed mounting: %s", output), err)
	}

	return nil
}

func (d *driver) fileModeMountPath() (fileMode os.FileMode) {
	return os.FileMode(d.volumeFileMode())
}

// from github.com/docker/docker/daemon/graphdriver/devmapper/
// this should be abstracted outside of graphdriver but within Docker package,
// here temporarily
type probeData struct {
	fsName string
	magic  string
	offset uint64
}

func probeFsType(device string) (string, error) {
	probes := []probeData{
		{"btrfs", "_BHRfS_M", 0x10040},
		{"ext4", "\123\357", 0x438},
		{"xfs", "XFSB", 0},
	}

	maxLen := uint64(0)
	for _, p := range probes {
		l := p.offset + uint64(len(p.magic))
		if l > maxLen {
			maxLen = l
		}
	}

	file, err := os.Open(device)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, maxLen)
	l, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	if uint64(l) != maxLen {
		return "", goof.WithField(
			"device", device, "error detecting filesystem")
	}

	for _, p := range probes {
		if bytes.Equal(
			[]byte(p.magic), buffer[p.offset:p.offset+uint64(len(p.magic))]) {
			return p.fsName, nil
		}
	}

	return "", errUnknownFileSystem
}

func (d *driver) volumeMountPath(target string) string {
	return fmt.Sprintf("%s%s", target, d.volumeRootPath())
}

func (d *driver) volumeFileMode() int {
	return d.config.GetInt("linux.volume.filemode")
}

func (d *driver) volumeRootPath() string {
	return d.config.GetString("linux.volume.rootpath")
}
