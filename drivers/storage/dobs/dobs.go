// +build !libstorage_storage_driver libstorage_storage_driver_dobs

package dobs

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the name of the driver
	Name = "dobs"

	defaultStatusMaxAttempts = 10
	defaultStatusInitDelay   = "100ms"

	/* This is hard deadline when waiting for the volume status to change to
	a desired state. At minimum is has to be more than the expontential
	backoff of sum 100*2^x, x=0 to 9 == 102s3ms, but should also account for
	RTT of API requests, and how many API requests would be made to
	exhaust retries */
	defaultStatusTimeout = "2m"

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

	// ConfigToken is the key for the token in the config file
	ConfigToken = Name + ".token"

	// ConfigRegion is the key for the region in the config file
	ConfigRegion = Name + ".region"

	// ConfigStatusMaxAttempts is the key for the maximum number of times
	// a volume status will be queried when waiting for an action to finish
	ConfigStatusMaxAttempts = Name + ".statusMaxAttempts"

	// ConfigStatusInitDelay is the key for the initial time duration
	// for exponential backoff
	ConfigStatusInitDelay = Name + ".statusInitialDelay"

	// ConfigStatusTimeout is the key for the time duration for a timeout
	// on how long to wait for a desired volume status to appears
	ConfigStatusTimeout = Name + ".statusTimeout"
)

func init() {
	registerConfig()
}

func registerConfig() {
	r := gofigCore.NewRegistration("DigitalOcean Block Storage")
	r.Key(gofig.String, "", "", "The DigitalOcean access token",
		ConfigToken)
	r.Key(gofig.String, "", "", "The DigitalOcean region",
		ConfigRegion)
	r.Key(gofig.Int, "", defaultStatusMaxAttempts, "Max Status Attempts",
		ConfigStatusMaxAttempts)
	r.Key(gofig.String, "", defaultStatusInitDelay, "Status Initial Delay",
		ConfigStatusInitDelay)
	r.Key(gofig.String, "", defaultStatusTimeout, "Status Timeout",
		ConfigStatusTimeout)
	gofigCore.Register(r)
}
