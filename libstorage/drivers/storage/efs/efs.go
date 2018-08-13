package efs

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the provider's name.
	Name = "efs"

	defaultStatusMaxAttempts = 6
	defaultStatusInitDelay   = "1s"

	/* This is hard deadline when waiting for the volume status to change to
	a desired state. At minimum is has to be more than the expontential
	backoff of sum 1*2^x, x=0 to 5 == 63s, but should also account for
	RTT of API requests, and how many API requests would be made to
	exhaust retries */
	defaultStatusTimeout = "2m"

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

	// InstanceIDFieldSubnetID is the key of the subnet where the instance lives from the
	// InstanceID Field map
	InstanceIDFieldSubnetID = "subnet"

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

	// Endpoint is a key constant.
	EC2Endpoint = "ec2endpoint"

	// EndpointFormat is a key constant.
	EC2EndpointFormat = "ec2endpointFormat"

	// MaxRetries is a key constant.
	MaxRetries = "maxRetries"

	// Tag is a key constant.
	Tag = "tag"

	// DisableSessionCache is a key constant.
	DisableSessionCache = "disableSessionCache"

	statusMaxAttempts  = "statusMaxAttempts"
	statusInitialDelay = "statusInitialDelay"
	statusTimeout      = "statusTimeout"
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

	// ConfigEC2Endpoint is a config key.
	ConfigEC2Endpoint = ConfigEFS + "." + EC2Endpoint

	// ConfigEC2EndpointFormat is a config key.
	ConfigEC2EndpointFormat = ConfigEFS + "." + EC2EndpointFormat

	// ConfigEFSMaxRetries is a config key.
	ConfigEFSMaxRetries = ConfigEFS + "." + MaxRetries

	// ConfigEFSTag is a config key.
	ConfigEFSTag = ConfigEFS + "." + Tag

	// ConfigEFSDisableSessionCache is a config key.
	ConfigEFSDisableSessionCache = ConfigEFS + "." + DisableSessionCache

	// ConfigStatusMaxAttempts is the key for the maximum number of times
	// a volume status will be queried when waiting for an action to finish
	ConfigStatusMaxAttempts = ConfigEFS + "." + statusMaxAttempts

	// ConfigStatusInitDelay is the key for the initial time duration
	// for exponential backoff
	ConfigStatusInitDelay = ConfigEFS + "." + statusInitialDelay

	// ConfigStatusTimeout is the key for the time duration for a timeout
	// on how long to wait for a desired volume status to appears
	ConfigStatusTimeout = ConfigEFS + "." + statusTimeout
)

func init() {
	r := gofigCore.NewRegistration("EFS")
	r.Key(gofig.String, "", "", "AWS access key", ConfigEFSAccessKey)
	r.Key(gofig.String, "", "", "AWS secret key", ConfigEFSSecretKey)
	r.Key(gofig.String, "", "", "List of security groups",
		ConfigEFSSecGroups)
	r.Key(gofig.String, "", "", "AWS region", ConfigEFSRegion)
	r.Key(gofig.String, "", "", "AWS EFS endpoint", ConfigEFSEndpoint)
	r.Key(gofig.String, "", `elasticfilesystem.%s.amazonaws.com`,
		"AWS EFS endpoint format", ConfigEFSEndpointFormat)
	r.Key(gofig.String, "", "", "AWS EC2 endpoint", ConfigEC2Endpoint)
	r.Key(gofig.String, "", `ec2.%s.amazonaws.com`,
		"AWS EC2 endpoint format", ConfigEC2EndpointFormat)
	r.Key(gofig.String, "", "", "Tag prefix for EFS naming", ConfigEFSTag)
	r.Key(gofig.Bool, "", false,
		"A flag that disables the session cache",
		ConfigEFSDisableSessionCache)
	r.Key(gofig.Int, "", defaultStatusMaxAttempts, "Max Status Attempts",
		ConfigStatusMaxAttempts)
	r.Key(gofig.String, "", defaultStatusInitDelay, "Status Initial Delay",
		ConfigStatusInitDelay)
	r.Key(gofig.String, "", defaultStatusTimeout, "Status Timeout",
		ConfigStatusTimeout)
	gofigCore.Register(r)
}
