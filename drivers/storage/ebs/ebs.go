package ebs

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the provider's name.
	Name = "ebs"

	// NameEC2 is the provider's old EC2 name.
	NameEC2 = "ec2"

	// NameAWS is the provider's old AWS name.
	NameAWS = "aws"

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
)

func init() {
	r := gofigCore.NewRegistration("EBS")
	r.Key(gofig.String, "", "", "", Name+"."+AccessKey)
	r.Key(gofig.String, "", "", "", Name+"."+SecretKey)
	r.Key(gofig.String, "", "", "", Name+"."+Region)
	r.Key(gofig.String, "", "", "", Name+"."+Endpoint)
	r.Key(gofig.Int, "", DefaultMaxRetries, "", Name+"."+MaxRetries)
	r.Key(gofig.String, "", "", "Tag prefix for EBS naming", Name+"."+Tag)
	r.Key(gofig.String, "", "", "", Name+"."+KmsKeyID)

	r.Key(gofig.String, "", "", "", NameEC2+"."+AccessKey)
	r.Key(gofig.String, "", "", "", NameEC2+"."+SecretKey)
	r.Key(gofig.String, "", "", "", NameEC2+"."+Region)
	r.Key(gofig.String, "", "", "", NameEC2+"."+Endpoint)
	r.Key(gofig.Int, "", DefaultMaxRetries, "", NameEC2+"."+MaxRetries)
	r.Key(gofig.String, "", "", "Tag prefix for EBS naming", NameEC2+"."+Tag)
	r.Key(gofig.String, "", "", "", NameEC2+"."+KmsKeyID)

	r.Key(gofig.String, "", "", "", NameAWS+"."+AccessKey)
	r.Key(gofig.String, "", "", "", NameAWS+"."+SecretKey)
	r.Key(gofig.String, "", "", "", NameAWS+"."+Region)
	r.Key(gofig.String, "", "", "", NameAWS+"."+Endpoint)
	r.Key(gofig.Int, "", DefaultMaxRetries, "", NameAWS+"."+MaxRetries)
	r.Key(gofig.String, "", "", "Tag prefix for EBS naming", NameAWS+"."+Tag)
	r.Key(gofig.String, "", "", "", NameAWS+"."+KmsKeyID)

	gofigCore.Register(r)
}
