package executor

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	log "github.com/sirupsen/logrus"

	"github.com/rexray/rexray/libstorage/api/registry"
	"github.com/rexray/rexray/libstorage/api/types"
	"github.com/rexray/rexray/libstorage/drivers/storage/ebs"
	ebsUtils "github.com/rexray/rexray/libstorage/drivers/storage/ebs/utils"
)

// driver is the storage executor for the ec2 storage driver.
type driver struct {
	name        string
	config      gofig.Config
	deviceRange *ebsUtils.DeviceRange
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
	log.Debug("executor using large device range: ", useLargeDeviceRange)
	d.deviceRange = ebsUtils.GetDeviceRange(useLargeDeviceRange)
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
		if !ns.DeviceRE.MatchString(devName) {
			continue
		}
		devPath := path.Join("/dev/", devName)
		devMap[devPath] = devPath
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
