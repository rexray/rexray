package utils

import (
	"os"

	gofig "github.com/akutz/gofig/types"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/s3fs"
)

// InstanceID returns the instance ID for the local host.
func InstanceID(
	ctx types.Context, config gofig.Config) (*types.InstanceID, error) {

	var hostName string
	if config == nil {
		hostName = config.GetString(s3fs.ConfigS3FSHostName)
	} else {
		hostName, _ = os.Hostname()
	}
	return &types.InstanceID{ID: hostName, Driver: s3fs.Name}, nil
}
