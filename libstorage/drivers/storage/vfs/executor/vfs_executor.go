package executor

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/vfs"
)

var (
	devFileLocks    = map[string]*sync.RWMutex{}
	devFileLocksRWL = &sync.RWMutex{}
)

type driver struct {
	config      gofig.Config
	rootDir     string
	devFilePath string
}

func init() {
	registry.RegisterStorageExecutor(vfs.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Name() string {
	return vfs.Name
}

func (d *driver) Supported(
	ctx types.Context,
	opts types.Store) (bool, error) {

	return true, nil
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config

	d.rootDir = vfs.RootDir(config)
	d.devFilePath = vfs.DeviceFilePath(config)
	if !gotil.FileExists(d.devFilePath) {
		os.MkdirAll(path.Dir(d.devFilePath), 0755)
		err := ioutil.WriteFile(d.devFilePath, initialDeviceFile, 0644)
		if err != nil {
			return err
		}
	}

	devFileLocksRWL.Lock()
	defer devFileLocksRWL.Unlock()
	devFileLocks[d.devFilePath] = &sync.RWMutex{}

	return nil
}

// Mount mounts a device to a specified path.
func (d *driver) Mount(
	ctx types.Context,
	deviceName, mountPoint string,
	opts *types.DeviceMountOpts) error {

	return os.MkdirAll(mountPoint, 0755)
}

// Unmount unmounts the underlying device from the specified path.
func (d *driver) Unmount(
	ctx types.Context,
	mountPoint string,
	opts types.Store) error {

	if !gotil.FileExists(mountPoint) {
		return os.ErrNotExist
	}
	return os.RemoveAll(mountPoint)
}

// InstanceID returns the local system's InstanceID.
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	iid := &types.InstanceID{Driver: vfs.Name}
	hostName, err := utils.HostName()
	if err != nil {
		return nil, err
	}

	if ok, _ := strconv.ParseBool(os.Getenv("VFS_INSTANCEID_USE_FIELDS")); ok {
		iid.ID = hostName
		iid.Fields = map[string]string{"region": "east"}
		return iid, nil
	}

	if err := iid.MarshalMetadata(hostName); err != nil {
		return nil, err
	}
	return iid, nil
}

var (
	avaiDevRX = regexp.MustCompile(`^(/dev/xvd[a-z])$`)
)

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {

	var devFileRWL *sync.RWMutex
	func() {
		devFileLocksRWL.RLock()
		defer devFileLocksRWL.RUnlock()
		devFileRWL = devFileLocks[d.devFilePath]
	}()
	devFileRWL.Lock()
	defer devFileRWL.Unlock()

	if !gotil.FileExists(d.devFilePath) {
		return "", goof.New("device file missing")
	}

	f, err := os.Open(d.devFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scn := bufio.NewScanner(f)
	for {
		if !scn.Scan() {
			break
		}

		m := avaiDevRX.FindSubmatch(scn.Bytes())
		if len(m) == 0 {
			continue
		}

		return string(m[1]), nil
	}

	return "", goof.New("no available devices")
}

var (
	devRX = regexp.MustCompile(`^(/dev/xvd[a-z])(?:=(.+))?$`)
)

// LocalDevices returns a map of the system's local devices.
func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	ctx.WithFields(log.Fields{
		"vfs.root": d.rootDir,
		"dev.path": d.devFilePath,
	}).Debug("config info")

	var devFileRWL *sync.RWMutex
	func() {
		devFileLocksRWL.RLock()
		defer devFileLocksRWL.RUnlock()
		devFileRWL = devFileLocks[d.devFilePath]
	}()
	devFileRWL.Lock()
	defer devFileRWL.Unlock()

	if !gotil.FileExists(d.devFilePath) {
		return nil, goof.New("device file missing")
	}

	f, err := os.Open(d.devFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	localDevs := map[string]string{}

	scn := bufio.NewScanner(f)
	for {
		if !scn.Scan() {
			break
		}

		m := devRX.FindSubmatch(scn.Bytes())
		if len(m) == 0 {
			continue
		}

		dev := m[1]
		var mountPoint []byte
		if len(m) > 2 {
			mountPoint = m[2]
		}

		localDevs[string(dev)] = string(mountPoint)
	}

	return &types.LocalDevices{Driver: vfs.Name, DeviceMap: localDevs}, nil
}

var initialDeviceFile = []byte(`/dev/xvda
/dev/xvdb
/dev/xvdc
/dev/xvdd`)
