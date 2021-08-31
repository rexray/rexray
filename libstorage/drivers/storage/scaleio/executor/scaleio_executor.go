package executor

import (
	"io/ioutil"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/scaleio"
)

// driver is the storage executor for the VFS storage driver.
type driver struct {
	guid   string
	drvCfg string
}

func init() {
	registry.RegisterStorageExecutor(scaleio.Name, newdriver)
}

func newdriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(context types.Context, config gofig.Config) error {
	if d.guid = config.GetString("scaleio.guid"); d.guid == "" {
		d.drvCfg = config.GetString("scaleio.drvCfg")
	}
	return nil
}

func (d *driver) Name() string {
	return scaleio.Name
}

// Supported returns a flag indicating whether or not the platform
// implementing the executor is valid for the host on which the executor
// resides.
func (d *driver) Supported(
	ctx types.Context,
	opts types.Store) (bool, error) {

	if d.guid != "" {
		return true, nil
	}

	return gotil.FileExists(d.drvCfg), nil
}

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {

	return "", types.ErrNotImplemented
}

// LocalDevices returns a map of the system's local devices.
func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	lvm, err := getLocalVolumeMap()
	if err != nil {
		return nil, err
	}

	return &types.LocalDevices{
		Driver:    scaleio.Name,
		DeviceMap: lvm,
	}, nil
}

const diskIDPath = "/dev/disk/by-id"

func getLocalVolumeMap() (map[string]string, error) {
	volMap := map[string]string{}

	if !gotil.FileExists(diskIDPath) {
		// the diskIDPath does not exist -- therefore no vols
		return volMap, nil
	}
	files, err := ioutil.ReadDir(diskIDPath)
	if err != nil {
		return nil, err
	}
	diskIDRX := regexp.MustCompile(`(?i)emc-vol-[^-].+-(.+)$`)
	for _, f := range files {
		m := diskIDRX.FindStringSubmatch(f.Name())
		if len(m) == 0 {
			continue
		}
		devPath, err := filepath.EvalSymlinks(path.Join(diskIDPath, f.Name()))
		if err != nil {
			return nil, err
		}
		volMap[m[1]] = devPath
	}
	return volMap, nil
}

// InstanceID returns the local system's InstanceID.
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	return d.getInstanceID()
}

// GetInstanceID returns the instance ID object
func GetInstanceID(
	ctx types.Context, config gofig.Config) (*types.InstanceID, error) {

	d := &driver{}
	d.Init(ctx, config)
	return d.getInstanceID()
}

func (d *driver) getInstanceID() (*types.InstanceID, error) {

	if d.guid != "" {
		iid := &types.InstanceID{Driver: scaleio.Name}
		if err := iid.MarshalMetadata(d.guid); err != nil {
			return nil, err
		}
		return iid, nil
	}

	out, err := exec.Command(d.drvCfg, "--query_guid").CombinedOutput()
	if err != nil {
		return nil, goof.WithError("error getting sdc guid", err)
	}

	sdcGUID := strings.Replace(string(out), "\n", "", -1)
	iid := &types.InstanceID{Driver: scaleio.Name}
	if err := iid.MarshalMetadata(sdcGUID); err != nil {
		return nil, err
	}
	return iid, nil
}
