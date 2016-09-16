package executor

import (
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"

	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/rackspace"
)

// driver is the storage executor for the VFS storage driver.
type driver struct {
	config gofig.Config
}

func init() {
	registry.RegisterStorageExecutor(rackspace.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	return nil
}

func (d *driver) Name() string {
	return rackspace.Name
}

// InstanceID returns the local instance ID for the test
func InstanceID(config gofig.Config) (*types.InstanceID, error) {
	d := newDriver()
	d.Init(nil, config)
	return d.InstanceID(nil, nil)
}

// InstanceID returns the aws instance configuration
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {
	cmd := exec.Command("xenstore-read", "name")
	cmd.Env = d.config.EnvVars()
	cmdOut, err := cmd.Output()
	if err != nil {
		return nil,
			goof.WithError("problem getting InstanceID", err)
	}

	instanceID := strings.Replace(string(cmdOut), "\n", "", -1)

	validInstanceID := regexp.MustCompile(`^instance-`)
	valid := validInstanceID.MatchString(instanceID)
	if !valid {
		return nil, goof.WithError("InstanceID not valid", err)
	}
	instanceID = strings.Replace(instanceID, "instance-", "", 1)
	iid := &types.InstanceID{Driver: rackspace.Name}
	if err := iid.MarshalMetadata(instanceID); err != nil {
		return nil, err
	}
	return iid, nil
}

func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "", types.ErrNotImplemented
}

func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {
	// Read from /proc/partitions
	localDevices := make(map[string]string)
	file := "/proc/partitions"
	contentBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, goof.WithError(
			"Error reading /proc/partitions", err)
	}

	content := string(contentBytes)

	// Parse device names
	var deviceName string
	lines := strings.Split(content, "\n")
	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) == 4 {
			deviceName = "/dev/" + fields[3]
			localDevices[deviceName] = deviceName
		}
	}

	return &types.LocalDevices{
		Driver:    rackspace.Name,
		DeviceMap: localDevices,
	}, nil

}
