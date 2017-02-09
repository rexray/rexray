// +build !windows
// +build !libstorage_storage_driver libstorage_storage_driver_s3fs

package utils

import (
	"github.com/codedellemc/libstorage/api/types"
)

// NextDeviceInfo is the NextDeviceInfo object for S3FS.
//
var NextDeviceInfo = &types.NextDeviceInfo{
	Prefix:  "",
	Pattern: "",
	Ignore:  true,
}
