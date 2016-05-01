package executor

import (
	"github.com/akutz/gofig"

	"fmt"
	"github.com/akutz/goof"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// Name is the name of the driver.
	Name = "scaleio"
)

// driver is the storage executor for the VFS storage driver.
type driver struct {
	Config     gofig.Config
	instanceID *types.InstanceID
}

func init() {
	registry.RegisterStorageExecutor(Name, newdriver)
}

func newdriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(context types.Context, config gofig.Config) error {
	d.Config = config
	id, err := GetInstanceID()
	if err != nil {
		return err
	}
	d.instanceID = id
	return nil
}

func (d *driver) Name() string {
	return Name
}

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "", nil
}

// LocalDevices returns a map of the system's local devices.
func (d *driver) LocalDevices(
	ctx types.Context,
	opts types.Store) (map[string]string, error) {
	return getLocalVolumeMap()
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

	out, err := exec.Command("/opt/emc/scaleio/sdc/bin/drv_cfg", "--query_vols").Output()
	if err != nil {
		return nil, goof.WithError("error querying volumes", err)
	}

	result := string(out)
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		split := strings.Split(line, " ")
		if split[0] == "VOL-ID" {
			mappedVolume := &sdcMappedVolume{mdmID: split[3], volumeID: split[1]}
			mappedVolume.mdmVolumeID = fmt.Sprintf("%s-%s", mappedVolume.mdmID, mappedVolume.volumeID)
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
			devPath, _ := filepath.EvalSymlinks(fmt.Sprintf("%s/%s", diskIDPath, f.Name()))
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
	return d.instanceID, nil
}

// GetInstanceID returns the instance ID object
func GetInstanceID() (*types.InstanceID, error) {
	sg, err := getSdcLocalGUID()
	if err != nil {
		return nil, err
	}
	return &types.InstanceID{
		ID: sg,
	}, nil
}

func getSdcLocalGUID() (sdcGUID string, err error) {
	out, err := exec.Command("/opt/emc/scaleio/sdc/bin/drv_cfg", "--query_guid").Output()
	if err != nil {
		return "", goof.WithError("problem getting sdc guid", err)
	}

	sdcGUID = strings.Replace(string(out), "\n", "", -1)

	return sdcGUID, nil
}
