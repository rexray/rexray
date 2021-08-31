package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"

	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/fittedcloud"
)

var errNoAvaiDevice = goof.New("no available device")

// NextDevice returns the next available device.
func NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	// All possible device paths on Linux EC2 instances are /dev/xvd[f-p]
	letters := []string{
		"f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"}

	// Find which letters are used for local devices
	localDeviceNames := make(map[string]bool)

	localDevices, err := LocalDevices(
		ctx, &types.LocalDevicesOpts{Opts: opts})
	if err != nil {
		return "", goof.WithError("error getting local devices", err)
	}
	localDeviceMapping := localDevices.DeviceMap

	for localDevice := range localDeviceMapping {
		re, _ := regexp.Compile(`^/dev/` +
			NextDeviceInfo.Prefix +
			`(` + NextDeviceInfo.Pattern + `)`)
		res := re.FindStringSubmatch(localDevice)
		if len(res) > 0 {
			localDeviceNames[res[1]] = true
		}
	}

	// Find which letters are used for ephemeral devices
	ephemeralDevices, err := getEphemeralDevices(ctx)
	if err != nil {
		return "", goof.WithError("error getting ephemeral devices", err)
	}

	for _, ephemeralDevice := range ephemeralDevices {
		re, _ := regexp.Compile(`^` +
			NextDeviceInfo.Prefix +
			`(` + NextDeviceInfo.Pattern + `)`)
		res := re.FindStringSubmatch(ephemeralDevice)
		if len(res) > 0 {
			localDeviceNames[res[1]] = true
		}
	}

	// Find next available letter for device path
	for _, letter := range letters {
		if localDeviceNames[letter] {
			continue
		}
		return fmt.Sprintf(
			"/dev/%s%s", NextDeviceInfo.Prefix, letter), nil
	}
	return "", errNoAvaiDevice
}

const procPartitions = "/proc/partitions"

var xvdRX = regexp.MustCompile(`^xvd[a-z]$`)

// LocalDevices retrieves device paths currently attached and/or mounted
func LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	f, err := os.Open(procPartitions)
	if err != nil {
		return nil, goof.WithError("error reading "+procPartitions, err)
	}
	defer f.Close()

	devMap := map[string]string{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 4 {
			continue
		}
		devName := fields[3]
		if !xvdRX.MatchString(devName) {
			continue
		}
		devPath := path.Join("/dev/", devName)
		devMap[devPath] = devPath
	}

	ld := &types.LocalDevices{Driver: fittedcloud.Name}
	if len(devMap) > 0 {
		ld.DeviceMap = devMap
	}

	return ld, nil
}

var ephemDevRX = regexp.MustCompile(`ephemeral([0-9]|1[0-9]|2[0-3])$`)

// Find ephemeral devices from metadata
func getEphemeralDevices(
	ctx types.Context) (deviceNames []string, err error) {

	buf, err := BlockDevices(ctx)
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

		name, err := BlockDeviceName(ctx, string(word))
		if err != nil {
			return nil, goof.WithError(
				"ec2 block device mapping lookup failed", err)
		}

		// compensate for kernel volume mapping i.e. change "/dev/sda" to
		// "/dev/xvda"
		deviceNameStr := strings.Replace(
			string(name),
			"sd",
			NextDeviceInfo.Prefix, 1)

		deviceNames = append(deviceNames, deviceNameStr)
	}
	return deviceNames, nil
}
