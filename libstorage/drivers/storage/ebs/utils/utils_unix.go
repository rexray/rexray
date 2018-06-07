// +build !windows

package utils

import (
	"regexp"
	"strings"

	"github.com/rexray/rexray/libstorage/api/types"
	log "github.com/sirupsen/logrus"
)

// DeviceRange holds slices for device namespace iteration
// and patterns for matching device nodes to specific namespaces.
//
// EBS suggests to use /dev/sd[f-p] for Linux EC2 instances. Also on Linux EC2
// instances, although the device path may show up as /dev/sd* on the EC2
// side, it will appear locally as /dev/xvd*
//
// The broadest device path namespace available for Linux EC2 instances is
// /dev/xvd[b-c][a-z]
// See http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/device_naming.htm
type DeviceRange struct {
	ParentLetters  []string
	ChildLetters   []string
	NextDeviceInfo *types.NextDeviceInfo
	DeviceRE       *regexp.Regexp
}

var (
	largeDeviceRange = &DeviceRange{
		ParentLetters: []string{"b", "c"},
		ChildLetters: []string{
			"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
			"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
		NextDeviceInfo: &types.NextDeviceInfo{
			Prefix:  "xvd",
			Pattern: "[b-c][a-z]",
			Ignore:  false,
		},
		DeviceRE: regexp.MustCompile(`^xvd[b-c][a-z]$`),
	}
	defaultDeviceRange = &DeviceRange{
		ParentLetters: []string{"d"},
		ChildLetters: []string{
			"f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"},
		NextDeviceInfo: &types.NextDeviceInfo{
			Prefix:  "xvd",
			Pattern: "[f-p]",
			Ignore:  false,
		},
		DeviceRE: regexp.MustCompile(`^xvd[f-p]$`),
	}
	nvmeDeviceRange = &DeviceRange{
		ParentLetters: []string{"0"},
		ChildLetters: []string{
			"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13",
			"14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26"},
		NextDeviceInfo: &types.NextDeviceInfo{
			Prefix:  "nvme",
			Pattern: "[0-26]n.*",
			Ignore:  false,
		},
		DeviceRE: regexp.MustCompile(`^nvme[0-26]n.*`),
	}
)

// GetDeviceRange returns a specified DeviceRange object
func GetDeviceRange(useLargeDeviceRange bool, instanceType string) *DeviceRange {
	log.Debug("InstanceType: ", instanceType)
	if strings.HasPrefix(instanceType, "c5") || strings.HasPrefix(instanceType, "m5") {
		log.Debug("nvme device")
		return nvmeDeviceRange
	}
	if useLargeDeviceRange {
		return largeDeviceRange
	}
	return defaultDeviceRange
}
