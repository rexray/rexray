package executor

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/ebs"
)

// driver is the storage executor for the ec2 storage driver.
type driver struct {
	config         gofig.Config
	nextDeviceInfo *types.NextDeviceInfo
}

func init() {
	registry.RegisterStorageExecutor(ebs.Name, newDriver)
	// Backwards compatibility for ec2 executor
	registry.RegisterStorageExecutor(ebs.OldName, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	// Ensure backwards compatibility with ebs and ec2 in config
	ebs.BackCompat(config)

	d.config = config
	// EBS suggests to use /dev/sd[f-p] for Linux EC2 instances.
	// Also on Linux EC2 instances, although the device path may show up
	// as /dev/sd* on the EC2 side, it will appear locally as /dev/xvd*
	d.nextDeviceInfo = &types.NextDeviceInfo{
		Prefix:  "xvd",
		Pattern: "[f-p]",
		Ignore:  false,
	}

	return nil
}

func (d *driver) Name() string {
	return ebs.Name
}

// InstanceID returns the local instance ID for the test
func InstanceID() (*types.InstanceID, error) {
	return newDriver().InstanceID(nil, nil)
}

const (
	iidURL = "http://169.254.169.254/latest/meta-data/instance-id/"
)

// InstanceID returns the instance ID from the current instance from metadata
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {
	// Retrieve instance ID from metadata
	res, err := http.Get(iidURL)
	if err != nil {
		return nil, goof.WithError("ec2 instance id lookup failed", err)
	}
	instanceID, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, goof.WithError("error reading ec2 instance id", err)
	}

	iid := &types.InstanceID{Driver: d.Name()}
	if err := iid.MarshalMetadata(string(instanceID)); err != nil {
		return nil, goof.WithError("error marshalling instance id", err)
	}

	return iid, nil
}

var errNoAvaiDevice = goof.New("No available device")

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	// All possible device paths on Linux EC2 instances are /dev/xvd[f-p]
	letters := []string{
		"f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"}

	// Find which letters are used for local devices
	localDeviceNames := make(map[string]bool)

	localDevices, err := d.LocalDevices(
		ctx, &types.LocalDevicesOpts{Opts: opts})
	if err != nil {
		return "", goof.WithError("error getting local devices", err)
	}
	localDeviceMapping := localDevices.DeviceMap

	for localDevice := range localDeviceMapping {
		re, _ := regexp.Compile(`^/dev/` +
			d.nextDeviceInfo.Prefix +
			`(` + d.nextDeviceInfo.Pattern + `)`)
		res := re.FindStringSubmatch(localDevice)
		if len(res) > 0 {
			localDeviceNames[res[1]] = true
		}
	}

	// Find which letters are used for ephemeral devices
	ephemeralDevices, err := d.getEphemeralDevices()
	if err != nil {
		return "", goof.WithError("error getting ephemeral devices", err)
	}

	for _, ephemeralDevice := range ephemeralDevices {
		re, _ := regexp.Compile(`^` +
			d.nextDeviceInfo.Prefix +
			`(` + d.nextDeviceInfo.Pattern + `)`)
		res := re.FindStringSubmatch(ephemeralDevice)
		if len(res) > 0 {
			localDeviceNames[res[1]] = true
		}
	}

	// Find next available letter for device path
	for _, letter := range letters {
		if !localDeviceNames[letter] {
			nextDeviceName := "/dev/" +
				d.nextDeviceInfo.Prefix + letter
			return nextDeviceName, nil
		}
	}
	return "", errNoAvaiDevice
}

// Retrieve device paths currently attached and/or mounted
func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {
	// Read from /proc/partitions
	localDevices := make(map[string]string)
	file := "/proc/partitions"
	contentBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, goof.WithError(
			"error reading /proc/partitions", err)
	}

	content := string(contentBytes)

	// Parse device names
	var deviceName string
	lines := strings.Split(content, "\n")
	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) == 4 {
			deviceName = "/dev/" + fields[3]
			// Device ID is also device path for EBS, since it
			// can be obtained both locally and remotely
			// (remotely being from the AWS API side)
			localDevices[deviceName] = deviceName
		}
	}

	return &types.LocalDevices{
		Driver:    d.Name(),
		DeviceMap: localDevices,
	}, nil

}

const bdmURL = "http://169.254.169.254/latest/meta-data/block-device-mapping/"

// Find ephemeral devices from metadata
func (d *driver) getEphemeralDevices() (deviceNames []string, err error) {
	// Get list of all block devices
	res, err := http.Get(bdmURL)
	if err != nil {
		return nil, goof.WithError("ec2 block device mapping lookup failed", err)
	}
	blockDeviceMappings, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, goof.WithError("error reading ec2 block device mappings", err)
	}

	// Filter list of all block devices for ephemeral devices
	re, _ := regexp.Compile(`ephemeral([0-9]|1[0-9]|2[0-3])$`)

	scanner := bufio.NewScanner(strings.NewReader(string(blockDeviceMappings)))
	scanner.Split(bufio.ScanWords)

	var input string
	for scanner.Scan() {
		input = scanner.Text()
		if re.MatchString(input) {
			// Find device name for ephemeral device
			res, err := http.Get(fmt.Sprintf("%s%s", bdmURL, input))
			if err != nil {
				return nil, goof.WithError("ec2 block device mapping lookup failed", err)
			}
			deviceName, err := ioutil.ReadAll(res.Body)
			// Compensate for kernel volume mapping i.e. change "/dev/sda" to "/dev/xvda"
			deviceNameStr := strings.Replace(string(deviceName), "sd", d.nextDeviceInfo.Prefix, 1)
			res.Body.Close()
			if err != nil {
				return nil, goof.WithError("error reading ec2 block device mappings", err)
			}

			deviceNames = append(deviceNames, deviceNameStr)
		}
	}

	return deviceNames, nil
}
