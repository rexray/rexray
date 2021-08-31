package executor

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/ebs"
	ebsUtils "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/ebs/utils"
)

// driver is the storage executor for the ec2 storage driver.
type driver struct {
	name        string
	config      gofig.Config
	deviceRange *ebsUtils.DeviceRange
	nvmeBinPath string
}

func init() {
	registry.RegisterStorageExecutor(ebs.Name, newDriver)
	// backwards compatibility for ec2 executor
	registry.RegisterStorageExecutor(ebs.NameEC2, newEC2Driver)
}

func newDriver() types.StorageExecutor {
	return &driver{name: ebs.Name}
}

func newEC2Driver() types.StorageExecutor {
	return &driver{name: ebs.NameEC2}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	// ensure backwards compatibility with ebs and ec2 in config
	ebs.BackCompat(config)
	d.config = config
	// initialize device range config
	useLargeDeviceRange := d.config.GetBool(ebs.ConfigUseLargeDeviceRange)
	ctx.WithValue("deviceRange", useLargeDeviceRange).Debug(
		"executor using large device range")
	d.deviceRange = ebsUtils.GetDeviceRange(useLargeDeviceRange)
	d.nvmeBinPath = d.config.GetString(ebs.ConfigNvmeBinPath)
	return nil
}

func (d *driver) Name() string {
	return d.name
}

// Supported returns a flag indicating whether or not the platform
// implementing the executor is valid for the host on which the executor
// resides.
func (d *driver) Supported(
	ctx types.Context,
	opts types.Store) (bool, error) {

	return ebsUtils.IsEC2Instance(ctx)
}

// InstanceID returns the instance ID from the current instance from metadata
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {
	return ebsUtils.InstanceID(ctx, d.Name())
}

var errNoAvaiDevice = goof.New("no available device")

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {

	// Find which letters are used for local devices
	localDeviceNames := make(map[string]bool)

	// Get device range
	ns := d.deviceRange

	localDevices, err := d.LocalDevices(
		ctx, &types.LocalDevicesOpts{Opts: opts})
	if err != nil {
		return "", goof.WithError("error getting local devices", err)
	}
	localDeviceMapping := localDevices.DeviceMap

	for localDevice := range localDeviceMapping {
		re, _ := regexp.Compile(`^/dev/` +
			ns.NextDeviceInfo.Prefix +
			`(` + ns.NextDeviceInfo.Pattern + `)`)
		res := re.FindStringSubmatch(localDevice)
		if len(res) > 0 {
			localDeviceNames[res[1]] = true
		}
	}

	// Find which letters are used for ephemeral devices
	ephemeralDevices, err := d.getEphemeralDevices(ctx)
	if err != nil {
		return "", goof.WithError("error getting ephemeral devices", err)
	}

	for _, ephemeralDevice := range ephemeralDevices {
		re, _ := regexp.Compile(`^` +
			ns.NextDeviceInfo.Prefix +
			`(` + ns.NextDeviceInfo.Pattern + `)`)
		res := re.FindStringSubmatch(ephemeralDevice)
		if len(res) > 0 {
			localDeviceNames[res[1]] = true
		}
	}

	// Find next available letter for device path.
	// Device namespace is iterated in random order
	// to mitigate ghost device issues.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	parentLength := len(ns.ParentLetters)
	for _, pIndex := range r.Perm(parentLength) {
		parentSuffix := ""
		if parentLength > 1 {
			parentSuffix = ns.ParentLetters[pIndex]
		}
		for _, cIndex := range r.Perm(len(ns.ChildLetters)) {
			suffix := parentSuffix + ns.ChildLetters[cIndex]
			if localDeviceNames[suffix] {
				continue
			}
			return fmt.Sprintf(
				"/dev/%s%s", ns.NextDeviceInfo.Prefix, suffix), nil
		}
	}
	return "", errNoAvaiDevice
}

const procPartitions = "/proc/partitions"

// fileExists returns a flag indicating whether or not a file
// path exists.
func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Retrieve device paths currently attached and/or mounted
func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	f, err := os.Open(procPartitions)
	if err != nil {
		return nil, goof.WithError("error reading "+procPartitions, err)
	}
	defer f.Close()

	devMap := map[string]string{}
	ns := d.deviceRange

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 4 {
			continue
		}
		devName := fields[3]

		devPath := path.Join("/dev/", devName)
		devAlias := devPath

		// NVMe support
		if strings.Contains(devName, "nvme") {
			// find the EBS device name that we *think* we will mount
			// the device as (nvme ignore this)
			if out, err := exec.Command(
				d.nvmeBinPath,
				"id-ctrl",
				"--raw-binary", devPath).Output(); err == nil {

				// read the binary output slice and trim it
				dev := strings.TrimSpace(string(out[3072:3104]))

				// if the result contains a /dev/ then we got a match
				if strings.Contains(dev, "/dev/") {
					ctx.WithFields(map[string]interface{}{
						"deviceName": devName,
						"device":     dev,
					}).Debug("found symlink")
					// if the alias / udev path exist, its a match
					if ok, err := fileExists(dev); !ok {
						if err != nil {
							ctx.WithField("devicePath", dev).WithError(err).Error(
								"error checking if device exists")
							return nil, err
						}
					} else {
						devName = strings.TrimLeft(dev, "/dev/")
						devPath = dev
					}
				}
			}
		}

		if !ns.DeviceRE.MatchString(devName) {
			ctx.WithFields(map[string]interface{}{
				"deviceName": devName,
				"deviceRX":   ns.DeviceRE,
			}).Warn("device does not match")
			continue
		}

		devMap[devPath] = devAlias
	}

	ld := &types.LocalDevices{Driver: d.Name()}
	if len(devMap) > 0 {
		ld.DeviceMap = devMap
	}

	return ld, nil
}

var ephemDevRX = regexp.MustCompile(`ephemeral([0-9]|1[0-9]|2[0-3])$`)

// Find ephemeral devices from metadata
func (d *driver) getEphemeralDevices(
	ctx types.Context) (deviceNames []string, err error) {

	buf, err := ebsUtils.BlockDevices(ctx)
	if err != nil {
		return nil, err
	}

	// Filter list of all block devices for ephemeral devices
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Bytes()
		if !ephemDevRX.Match(word) {
			continue
		}

		name, err := ebsUtils.BlockDeviceName(ctx, string(word))
		if err != nil {
			return nil, goof.WithError(
				"ec2 block device mapping lookup failed", err)
		}

		// compensate for kernel volume mapping i.e. change "/dev/sda" to
		// "/dev/xvda"
		deviceNameStr := strings.Replace(
			string(name),
			"sd",
			d.deviceRange.NextDeviceInfo.Prefix, 1)

		deviceNames = append(deviceNames, deviceNameStr)
	}
	return deviceNames, nil
}
