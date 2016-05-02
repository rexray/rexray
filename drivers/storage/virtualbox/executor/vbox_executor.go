package executor

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/virtualbox/client"
)

const (
	// Name is the name of the storage executor and driver.
	Name = "virtualbox"
)

// Executor is the storage executor for the VFS storage driver.
type Executor struct {
	// Config is the executor's configuration instance.
	Config  gofig.Config
	vbox    *client.VirtualBox
	machine *client.Machine
}

func init() {
	gofig.Register(LoadConfig())
	registry.RegisterStorageExecutor(Name, newExecutor)
}

func newExecutor() types.StorageExecutor {
	return &Executor{}
}

//Init initializes the executor by connecting to the vbox endpoint
func (d *Executor) Init(ctx types.Context, config gofig.Config) error {
	log.Info("Initiializing executor ", Name)
	d.Config = config
	if err := d.vboxLogon(); err != nil {
		return err
	}
	return nil
}

// Name returns the human-readable name of the executor
func (d *Executor) Name() string {
	return Name
}

// InstanceID returns the local system's InstanceID.
func (d *Executor) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {
	if d.machine == nil {
		err := d.loadMachineInfo()
		if err != nil {
			return nil, err
		}
	}
	return &types.InstanceID{ID: d.machine.GetID()}, nil
}

// NextDevice returns the next available device (not implemented).
func (d *Executor) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "", types.ErrNotImplemented
}

// LocalDevices returns a map of the system's local devices.
func (d *Executor) LocalDevices(
	ctx types.Context,
	opts types.Store) (map[string]string, error) {
	mapDiskByID := make(map[string]string)
	diskIDPath := d.Config.GetString("virtualbox.diskIDPath")
	files, err := ioutil.ReadDir(diskIDPath)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if strings.Contains(f.Name(), "VBOX_HARDDISK_VB") {
			sid := d.getShortDeviceID(f.Name())
			if sid == "" {
				continue
			}
			devPath, _ := filepath.EvalSymlinks(fmt.Sprintf("%s/%s", diskIDPath, f.Name()))
			mapDiskByID[sid] = devPath
		}
	}
	return mapDiskByID, nil
}

func (d *Executor) vboxLogon() error {
	// connect to vbox
	uname := d.Config.GetString("virtualbox.username")
	pwd := d.Config.GetString("virtualbox.username")
	endpoint := d.Config.GetString("virtualbox.endpoint")

	d.vbox = client.NewVirtualBox(uname, pwd, endpoint)

	err := d.vbox.Logon()
	if err != nil {
		return err
	}
	return nil
}

func (d *Executor) loadMachineInfo() error {
	m, err := d.vbox.FindMachine(
		d.Config.GetString("virtualbox.localMachineNameOrId"))
	if err != nil {
		return err
	}
	d.machine = m

	err = d.vbox.PopulateMachineInfo(m)
	if err != nil {
		return err
	}
	return nil
}

func (d *Executor) getShortDeviceID(f string) string {
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

//LoadConfig loads configuration
func LoadConfig() *gofig.Registration {
	r := gofig.NewRegistration("virtualbox")
	r.Key(gofig.String, "", "http://127.0.0.1:18083", "", "virtualbox.endpoint")
	r.Key(gofig.String, "", "", "", "virtualbox.volumePath")
	r.Key(gofig.String, "", "default", "", "virtualbox.localMachineNameOrId")
	r.Key(gofig.String, "", "", "", "virtualbox.username")
	r.Key(gofig.String, "", "", "", "virtualbox.password")
	r.Key(gofig.Bool, "", false, "", "virtualbox.tls")
	r.Key(gofig.String, "", "", "", "virtualbox.controllerName")
	r.Key(gofig.String, "", "/dev/disk/by-id", "", "virtualbox.diskIDPath")
	r.Key(gofig.String, "", "/sys/class/scsi_host/", "", "virtualbox.scsiHostPath")
	return r
}
