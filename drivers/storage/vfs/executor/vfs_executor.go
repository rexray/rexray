package executor

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sync"

	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
	"github.com/emccode/libstorage/api/utils/paths"
)

const (
	// Name is the name of the storage executor and driver.
	Name = "vfs"
)

var (
	devFileLocks    = map[string]*sync.RWMutex{}
	devFileLocksRWL = &sync.RWMutex{}
)

// Executor is the storage executor for the VFS storage driver.
type Executor struct {
	config      gofig.Config
	rootDir     string
	devFilePath string
	volDirPath  string
}

func init() {
	gofig.Register(configRegistration())
	registry.RegisterStorageExecutor(Name, newExecutor)
}

func newExecutor() drivers.StorageExecutor {
	return &Executor{}
}

func (d *Executor) Init(config gofig.Config) error {
	d.config = config

	d.rootDir = d.config.GetString("vfs.root")
	if err := os.MkdirAll(d.rootDir, 0755); err != nil {
		return err
	}

	d.volDirPath = fmt.Sprintf("%s/vol", d.rootDir)
	if err := os.MkdirAll(d.volDirPath, 0755); err != nil {
		return err
	}

	d.devFilePath = fmt.Sprintf("%s/dev", d.rootDir)
	if !gotil.FileExists(d.devFilePath) {
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

func (d *Executor) Name() string {
	return Name
}

// InstanceID returns the local system's InstanceID.
func (d *Executor) InstanceID(
	ctx context.Context,
	opts types.Store) (*types.InstanceID, error) {
	return GetInstanceID()
}

var (
	avaiDevRX = regexp.MustCompile(`^(/dev/xvd[a-z])$`)
)

// NextDevice returns the next available device.
func (d *Executor) NextDevice(
	ctx context.Context,
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
func (d *Executor) LocalDevices(
	ctx context.Context,
	opts types.Store) (map[string]string, error) {

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

	return localDevs, nil
}

// RootDir returns the path to the VFS root directory.
func (d *Executor) RootDir() string {
	return d.rootDir
}

// DeviceFilePath returns the path to the VFS devices file.
func (d *Executor) DeviceFilePath() string {
	return d.devFilePath
}

// VolumesDirPath returns the path to the VFS volumes directory.
func (d *Executor) VolumesDirPath() string {
	return d.volDirPath
}

func configRegistration() *gofig.Registration {
	defaultRootDir := fmt.Sprintf("%s/vfs", paths.UsrDirPath())
	r := gofig.NewRegistration("VFS")
	r.Key(gofig.String, "", defaultRootDir, "", "vfs.root")
	return r
}

// GetInstanceID returns the instance ID.
func GetInstanceID() (*types.InstanceID, error) {
	hostName, err := getHostName()
	if err != nil {
		return nil, err
	}
	return &types.InstanceID{ID: hostName}, nil
}

var initialDeviceFile = []byte(`/dev/xvda
/dev/xvdb
/dev/xvdc
/dev/xvdd`)
