// +build !windows

package utils

import (
	"github.com/thecodeteam/rexray/libstorage/api/types"
)

// NextDeviceInfo is the NextDeviceInfo object for Azure.
//
// On Azure Linux instance /dev/sda is the boot volume,
// /dev/sdb is a temporary disk.
// Other letters can be used for data volumes.
var NextDeviceInfo = &types.NextDeviceInfo{
	Prefix:  "sd",
	Pattern: "[c-z]",
	Ignore:  false,
}
