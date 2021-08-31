// +build !windows

package utils

import (
	"regexp"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
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
)

// GetDeviceRange returns a specified DeviceRange object
func GetDeviceRange(useLargeDeviceRange bool) *DeviceRange {
	if useLargeDeviceRange {
		return largeDeviceRange
	}
	return defaultDeviceRange
}
