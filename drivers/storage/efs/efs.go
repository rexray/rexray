// +build !libstorage_storage_driver libstorage_storage_driver_efs

package efs

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the provider's name.
	Name = "efs"

	// TagDelimiter separates tags from volume or snapshot names
	TagDelimiter = "/"

	// DefaultMaxRetries is the max number of times to retry failed operations
	DefaultMaxRetries = 10

	// InstanceIDFieldRegion is the key to retrieve the region value from the
	// InstanceID Field map.
	InstanceIDFieldRegion = "region"

	// InstanceIDFieldAvailabilityZone is the key to retrieve the availability
	// zone value from the InstanceID Field map.
	InstanceIDFieldAvailabilityZone = "availabilityZone"

	// InstanceIDFieldSecurityGroups is the key to retrieve the default security
	// group value from the InstanceID Field map.
	InstanceIDFieldSecurityGroups = "securityGroups"

	// AccessKey is a key constant.
	AccessKey = "accessKey"

	// SecretKey is a key constant.
	SecretKey = "secretKey"

	// Region is a key constant.
	Region = "region"

	// SecurityGroups is a key constant.
	SecurityGroups = "securityGroups"

	// Endpoint is a key constant.
	Endpoint = "endpoint"

	// EndpointFormat is a key constant.
	EndpointFormat = "endpointFormat"

	// MaxRetries is a key constant.
	MaxRetries = "maxRetries"

	// Tag is a key constant.
	Tag = "tag"

	// DisableSessionCache is a key constant.
	DisableSessionCache = "disableSessionCache"
)

const (
	// ConfigEFS is a config key.
	ConfigEFS = Name

	// ConfigEFSAccessKey is a config key.
	ConfigEFSAccessKey = ConfigEFS + "." + AccessKey

	// ConfigEFSSecretKey is a config key.
	ConfigEFSSecretKey = ConfigEFS + "." + SecretKey

	// ConfigEFSRegion is a config key.
	ConfigEFSRegion = ConfigEFS + "." + Region

	// ConfigEFSSecGroups is a config key.
	ConfigEFSSecGroups = ConfigEFS + "." + SecurityGroups

	// ConfigEFSEndpoint is a config key.
	ConfigEFSEndpoint = ConfigEFS + "." + Endpoint

	// ConfigEFSEndpointFormat is a config key.
	ConfigEFSEndpointFormat = ConfigEFS + "." + EndpointFormat

	// ConfigEFSMaxRetries is a config key.
	ConfigEFSMaxRetries = ConfigEFS + "." + MaxRetries

	// ConfigEFSTag is a config key.
	ConfigEFSTag = ConfigEFS + "." + Tag

	// ConfigEFSDisableSessionCache is a config key.
	ConfigEFSDisableSessionCache = ConfigEFS + "." + DisableSessionCache
)

func init() {
	r := gofigCore.NewRegistration("EFS")
	r.Key(gofig.String, "", "", "AWS access key", ConfigEFSAccessKey)
	r.Key(gofig.String, "", "", "AWS secret key", ConfigEFSSecretKey)
	r.Key(gofig.String, "", "", "List of security groups", ConfigEFSSecGroups)
	r.Key(gofig.String, "", "", "AWS region", ConfigEFSRegion)
	r.Key(gofig.String, "", "", "AWS EFS endpoint", ConfigEFSEndpoint)
	r.Key(gofig.String, "", `elasticfilesystem.%s.amazonaws.com`,
		"AWS EFS endpoint format", ConfigEFSEndpointFormat)
	r.Key(gofig.String, "", "", "Tag prefix for EFS naming", ConfigEFSTag)
	r.Key(gofig.Bool, "", false,
		"A flag that disables the session cache", ConfigEFSDisableSessionCache)
	gofigCore.Register(r)
}
