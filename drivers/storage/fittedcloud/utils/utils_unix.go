// +build !windows
// +build !libstorage_storage_driver libstorage_storage_driver_fittedcloud

package utils

import (
	"github.com/codedellemc/libstorage/api/types"
)

// NextDeviceInfo is the NextDeviceInfo object for EBS.
//
// EBS suggests to use /dev/sd[f-p] for Linux EC2 instances. Also on Linux EC2
// instances, although the device path may show up as /dev/sd* on the EC2 side,
// it will appear locally as /dev/xvd*
var NextDeviceInfo = &types.NextDeviceInfo{
	Prefix:  "xvd",
	Pattern: "[f-p]",
	Ignore:  false,
}
