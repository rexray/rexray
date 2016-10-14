package executor

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/scaleio"
)

const (
	sioBinPath = "/opt/emc/scaleio/sdc/bin/drv_cfg"
)

// driver is the storage executor for the VFS storage driver.
type driver struct{}

func init() {
	registry.RegisterStorageExecutor(scaleio.Name, newdriver)
}

func newdriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(context types.Context, config gofig.Config) error {
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

	if _, err := os.Stat(sioBinPath); os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
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

type sdcMappedVolume struct {
	mdmID       string
	volumeID    string
	mdmVolumeID string
	sdcDevice   string
}

func getLocalVolumeMap() (map[string]string, error) {
	mappedVolumesMap := make(map[string]*sdcMappedVolume)
	volumeMap := make(map[string]string)

	out, err := exec.Command(sioBinPath, "--query_vols").Output()
	if err != nil {
		return nil, goof.WithError("error querying volumes", err)
	}

	result := string(out)
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		split := strings.Split(line, " ")
		if split[0] == "VOL-ID" {
			mappedVolume := &sdcMappedVolume{
				mdmID:    split[3],
				volumeID: split[1],
			}
			mappedVolume.mdmVolumeID = fmt.Sprintf(
				"%s-%s", mappedVolume.mdmID, mappedVolume.volumeID)
			mappedVolumesMap[mappedVolume.mdmVolumeID] = mappedVolume
		}
	}

	diskIDPath := "/dev/disk/by-id"
	files, _ := ioutil.ReadDir(diskIDPath)
	r, _ := regexp.Compile(`^emc-vol-\w*-\w*$`)
	for _, f := range files {
		matched := r.MatchString(f.Name())
		if matched {
			mdmVolumeID := strings.Replace(f.Name(), "emc-vol-", "", 1)
			devPath, _ := filepath.EvalSymlinks(
				fmt.Sprintf("%s/%s", diskIDPath, f.Name()))
			if _, ok := mappedVolumesMap[mdmVolumeID]; ok {
				volumeID := mappedVolumesMap[mdmVolumeID].volumeID
				volumeMap[volumeID] = devPath
			}
		}
	}

	return volumeMap, nil
}

// InstanceID returns the local system's InstanceID.
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	return GetInstanceID()
}

// GetInstanceID returns the instance ID object
func GetInstanceID() (*types.InstanceID, error) {
	sg, err := getSdcLocalGUID()
	if err != nil {
		return nil, err
	}
	iid := &types.InstanceID{Driver: scaleio.Name}
	if err := iid.MarshalMetadata(sg); err != nil {
		return nil, err
	}
	return iid, nil
}

func getSdcLocalGUID() (sdcGUID string, err error) {
	out, err := exec.Command(sioBinPath, "--query_guid").Output()
	if err != nil {
		return "", goof.WithError("problem getting sdc guid", err)
	}

	sdcGUID = strings.Replace(string(out), "\n", "", -1)

	return sdcGUID, nil
}
