package s3fs

import (
	"os"

	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	defaultRegion   = "us-east-1"
	defaultEndpoint = "s3.amazonaws.com"
)

const (
	// Name is the provider's name.
	Name = "s3fs"

	// TagDelimiter separates tags from volume or snapshot names
	TagDelimiter = "/"

	// DefaultMaxRetries is the max number of times to retry failed operations
	DefaultMaxRetries = 10

	// Cmd is a key constant.
	Cmd = "cmd"

	// Options is a key constant.
	Options = "options"

	// HostName is a key constant.
	HostName = "hostName"

	// Endpoint is a key constant.
	Endpoint = "endpoint"

	// AccessKey is a key constant.
	AccessKey = "accessKey"

	// SecretKey is a key constant.
	SecretKey = "secretKey"

	// Region is a key constant.
	Region = "region"

	// MaxRetries is a key constant.
	MaxRetries = "maxRetries"

	// DisablePathStyle sets the S3 AWS config property S3ForcePathStyle
	// to false. Please see https://github.com/aws/aws-sdk-go/issues/168
	// for more information.
	DisablePathStyle = "disablePathStyle"

	// Tag is a key constant.
	Tag = "tag"
)

const (
	// ConfigS3FS is a config key
	ConfigS3FS = Name

	// ConfigS3FSAccessKey is a config key.
	ConfigS3FSAccessKey = ConfigS3FS + "." + AccessKey

	// ConfigS3FSSecretKey is a config key.
	ConfigS3FSSecretKey = ConfigS3FS + "." + SecretKey

	// ConfigS3FSMaxRetries is a config key.
	ConfigS3FSMaxRetries = ConfigS3FS + "." + MaxRetries

	// ConfigS3FSRegion is a config key.
	ConfigS3FSRegion = ConfigS3FS + "." + Region

	// ConfigS3FSCmd is a config key.
	ConfigS3FSCmd = ConfigS3FS + "." + Cmd

	// ConfigS3FSOptions is a config key
	ConfigS3FSOptions = ConfigS3FS + "." + Options

	// ConfigS3FSHostName is a config key
	ConfigS3FSHostName = ConfigS3FS + "." + HostName

	//ConfigS3FSEndpoint is a config key.
	ConfigS3FSEndpoint = ConfigS3FS + "." + Endpoint

	// ConfigS3FSTag is a config key.
	ConfigS3FSTag = ConfigS3FS + "." + Tag

	// ConfigS3FSDisablePathStyle is a config key.
	ConfigS3FSDisablePathStyle = ConfigS3FS + "." + DisablePathStyle
)

func init() {
	hostName, _ := os.Hostname()
	r := gofigCore.NewRegistration("S3FS")
	r.Key(gofig.String, "", "", "AWS access key", ConfigS3FSAccessKey)
	r.Key(gofig.String, "", "", "AWS secret key", ConfigS3FSSecretKey)
	r.Key(gofig.String,
		"",
		defaultRegion,
		"AWS region",
		ConfigS3FSRegion)
	r.Key(gofig.String,
		"",
		"s3fs",
		`The absolute path to the "s3fs" binary.`,
		ConfigS3FSCmd)
	r.Key(gofig.String,
		"",
		"",
		`The options to use with the "s3fs" command.`,
		ConfigS3FSOptions)
	r.Key(gofig.String,
		"",
		hostName,
		"The host name used as part of the instance ID.",
		ConfigS3FSHostName)
	r.Key(gofig.String,
		"",
		"",
		"Tag prefix for S3 naming",
		ConfigS3FSTag)
	r.Key(gofig.Int,
		"",
		DefaultMaxRetries,
		"Max number of times to retry failed operations",
		ConfigS3FSMaxRetries)
	r.Key(gofig.Bool,
		"",
		false,
		"A flag that disables the use of S3's path style for bucket endpoints",
		ConfigS3FSDisablePathStyle)
	r.Key(gofig.String,
		"",
		"",
		`Optional "s3fs" endpoint.`,
		ConfigS3FSEndpoint)
	gofigCore.Register(r)
}
