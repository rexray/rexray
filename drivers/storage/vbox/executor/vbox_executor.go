package executor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/vbox"
)

// driver is the storage executor for the vbox storage driver.
type driver struct {
	config gofig.Config
}

func init() {
	registry.RegisterStorageExecutor(vbox.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	return nil
}

func (d *driver) Name() string {
	return vbox.Name
}

func getMacs() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var macs []string
	for _, intf := range interfaces {
		macUp := strings.ToUpper(strings.Replace(
			intf.HardwareAddr.String(), ":", "", -1))
		if macUp != "" {
			macs = append(macs, macUp)
		}
	}
	return macs, nil
}

func instanceIDJSON() ([]byte, error) {
	macs, err := getMacs()
	if err != nil {
		return nil, err
	}
	return json.Marshal(macs)
}

// InstanceID returns the local instance ID.
func InstanceID() (*types.InstanceID, error) {
	buf, err := instanceIDJSON()
	if err != nil {
		return nil, err
	}
	return &types.InstanceID{Metadata: buf}, nil
}

func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {
	return InstanceID()
}

func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "", types.ErrNotImplemented
}

func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (map[string]string, error) {

	mapDiskByID := make(map[string]string)
	files, err := ioutil.ReadDir(d.diskIDPath())
	if err != nil {
		return nil, err
	}

	d.rescanScsiHosts()
	for _, f := range files {
		if strings.Contains(f.Name(), "VBOX_HARDDISK_VB") {
			sid := d.getShortDeviceID(f.Name())
			if sid == "" {
				continue
			}
			devPath, _ := filepath.EvalSymlinks(
				fmt.Sprintf("%s/%s", d.diskIDPath(), f.Name()))
			mapDiskByID[sid] = devPath
		}
	}
	return mapDiskByID, nil
}

func (d *driver) rescanScsiHosts() {
	if dirs, err := ioutil.ReadDir(d.scsiHostPath()); err == nil {
		for _, f := range dirs {
			name := d.scsiHostPath() + f.Name() + "/scan"
			data := []byte("- - -")
			ioutil.WriteFile(name, data, 0666)
		}
	}
	time.Sleep(1 * time.Second)
}

func (d *driver) getShortDeviceID(f string) string {
	sid := strings.Split(f, "VBOX_HARDDISK_VB")
	if len(sid) < 1 {
		return ""
	}
	aid := strings.Split(sid[1], "-")
	if len(aid) < 1 {
		return ""
	}
	return aid[0]
}

func (d *driver) username() string {
	return d.config.GetString("virtualbox.username")
}

func (d *driver) password() string {
	return d.config.GetString("virtualbox.password")
}

func (d *driver) endpoint() string {
	return d.config.GetString("virtualbox.endpoint")
}

func (d *driver) volumePath() string {
	return d.config.GetString("virtualbox.volumePath")
}

func (d *driver) machineNameID() string {
	return d.config.GetString("virtualbox.localMachineNameOrId")
}

func (d *driver) tls() bool {
	return d.config.GetBool("virtualbox.tls")
}

func (d *driver) controllerName() string {
	return d.config.GetString("virtualbox.controllerName")
}

func (d *driver) diskIDPath() string {
	dip := d.config.GetString("virtualbox.diskIDPath")
	if dip != "" {
		return dip
	}
	return "/dev/disk/by-id"
}

func (d *driver) scsiHostPath() string {
	shp := d.config.GetString("virtualbox.scsiHostPath")
	if shp != "" {
		return shp
	}
	return "/sys/class/scsi_host/"
}
