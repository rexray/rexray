// +build !windows

package utils

import (
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

// NextDeviceInfo is the NextDeviceInfo object for S3FS.
//
var NextDeviceInfo = &types.NextDeviceInfo{
	Prefix:  "",
	Pattern: "",
	Ignore:  true,
}
