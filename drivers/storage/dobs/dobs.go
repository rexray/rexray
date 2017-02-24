// +build !libstorage_storage_driver libstorage_storage_driver_dobs

package dobs

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the name of the driver
	Name = "dobs"

	// InstanceIDFieldRegion is the key used to retrive the region from the
	// instance id map
	InstanceIDFieldRegion = "region"

	// InstanceIDFieldName is the key used to retrive the name from the instance
	// id map
	InstanceIDFieldName = "name"

	// VolumePrefix is the value that every DO volume appears with DigitalOcean
	// volumes are are found using disk/by-id, for example:
	//
	//     /dev/disk/by-id/scsi-0DO_Volume_volume-nyc1-01
	//
	// Please see https://goo.gl/MwReS6 for more information.
	VolumePrefix = "scsi-0DO_Volume_"

	// ConfigDOToken is the key for the token in the config file
	ConfigDOToken = Name + ".token"

	// ConfigDORegion is the key for the region in the config file
	ConfigDORegion = Name + ".region"
)

func init() {
	registerConfig()
}

func registerConfig() {
	r := gofigCore.NewRegistration("DigitalOcean Block Storage")
	r.Key(
		gofig.String,
		"",
		"",
		"The DigitalOcean access token",
		ConfigDOToken)
	r.Key(
		gofig.String,
		"",
		"",
		"The DigitalOcean region",
		ConfigDORegion)
	gofigCore.Register(r)
}
