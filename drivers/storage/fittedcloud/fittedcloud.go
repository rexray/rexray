// +build !libstorage_storage_driver libstorage_storage_driver_fittedcloud

package fittedcloud

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the provider's name.
	Name = "fittedcloud"

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

	// AccessKey is a key constant.
	AccessKey = "accessKey"

	// SecretKey is a key constant.
	SecretKey = "secretKey"

	// Region is a key constant.
	Region = "region"

	// Endpoint is a key constant.
	Endpoint = "endpoint"

	// MaxRetries is a key constant.
	MaxRetries = "maxRetries"

	// Tag is a key constant.
	Tag = "tag"

	// KmsKeyID is the full ARN of the AWS Key Management Service (AWS KMS)
	// customer master key (CMK) to use when creating the encrypted volume.
	//
	// This parameter is only required if you want to use a non-default CMK;
	// if this parameter is not specified, the default CMK for EBS is used.
	// The ARN contains the arn:aws:kms namespace, followed by the region of
	// the CMK, the AWS account ID of the CMK owner, the key namespace, and
	// then the CMK ID. For example,
	// arn:aws:kms:us-east-1:012345678910:key/abcd1234-a123-456a-a12b-a123b4cd56ef.
	//
	// If a KmsKeyID is specified, all volumes will be created with their
	// Encrypted flag set to true.
	KmsKeyID = "kmsKeyID"

	// ConfigEBS is a config key.
	ConfigEBS = "ebs"

	// ConfigEBSAccessKey is a config key.
	ConfigEBSAccessKey = ConfigEBS + "." + AccessKey

	// ConfigEBSSecretKey is a config key.
	ConfigEBSSecretKey = ConfigEBS + "." + SecretKey

	// ConfigEBSRegion is a config key.
	ConfigEBSRegion = ConfigEBS + "." + Region

	// ConfigEBSEndpoint is a config key.
	ConfigEBSEndpoint = ConfigEBS + "." + Endpoint

	// ConfigEBSMaxRetries is a config key.
	ConfigEBSMaxRetries = ConfigEBS + "." + MaxRetries

	// ConfigEBSTag is a config key.
	ConfigEBSTag = ConfigEBS + "." + Tag

	// ConfigEBSKmsKeyID is a config key.
	ConfigEBSKmsKeyID = ConfigEBS + "." + KmsKeyID
)

func init() {
	r := gofigCore.NewRegistration("FittedCloud")
	r.Key(gofig.String, "", "", "", ConfigEBSAccessKey)
	r.Key(gofig.String, "", "", "", ConfigEBSSecretKey)
	r.Key(gofig.String, "", "", "", ConfigEBSRegion)
	r.Key(gofig.String, "", "", "", ConfigEBSEndpoint)
	r.Key(gofig.Int, "", DefaultMaxRetries, "", ConfigEBSMaxRetries)
	r.Key(gofig.String, "", "", "Tag prefix for EBS naming", ConfigEBSTag)
	r.Key(gofig.String, "", "", "", ConfigEBSKmsKeyID)

	gofigCore.Register(r)
}
