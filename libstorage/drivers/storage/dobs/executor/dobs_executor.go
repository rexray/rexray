package executor

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	gofig "github.com/akutz/gofig/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	do "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/dobs"
	doUtils "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/dobs/utils"
)

var (
	diskPrefix = regexp.MustCompile(`^` + do.VolumePrefix + `(.*)`)
	diskSuffix = regexp.MustCompile("part-.*$")
)

type driver struct {
	config gofig.Config
}

func init() {
	registry.RegisterStorageExecutor(do.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Name() string {
	return do.Name
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	return nil
}

func (d *driver) InstanceID(
	ctx types.Context, opts types.Store) (*types.InstanceID, error) {
	return doUtils.InstanceID(ctx)
}

func (d *driver) NextDevice(
	ctx types.Context, opts types.Store) (string, error) {
	return "", types.ErrNotImplemented
}

func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {
	deviceMap := map[string]string{}
	diskIDPath := "/dev/disk/by-id"

	dir, _ := ioutil.ReadDir(diskIDPath)
	for _, device := range dir {
		switch {
		case !diskPrefix.MatchString(device.Name()):
			continue
		case diskSuffix.MatchString(device.Name()):
			continue
		case diskPrefix.MatchString(device.Name()):
			volumeName := diskPrefix.FindStringSubmatch(device.Name())[1]
			devPath, err := filepath.EvalSymlinks(
				fmt.Sprintf("%s/%s", diskIDPath, device.Name()))
			if err != nil {
				return nil, err
			}
			deviceMap[volumeName] = devPath
		}
	}

	ld := &types.LocalDevices{Driver: d.Name()}
	if len(deviceMap) > 0 {
		ld.DeviceMap = deviceMap
	}

	return ld, nil
}

func (d *driver) Supported(ctx types.Context, opts types.Store) (bool, error) {
	return doUtils.IsDroplet(ctx)
}
