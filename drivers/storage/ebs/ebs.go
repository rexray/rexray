package ebs

import (
	"github.com/akutz/gofig"
)

const (
	// Name is the provider's name.
	Name = "ebs"

	// OldName is the provider's old name.
	OldName = "ec2"

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
)

func init() {
	registerConfig()
}

func registerConfig() {
	r := gofig.NewRegistration("EBS")
	r.Key(gofig.String, "", "", "", Name+"."+AccessKey)
	r.Key(gofig.String, "", "", "", Name+"."+SecretKey)
	r.Key(gofig.String, "", "", "", Name+"."+Region)
	r.Key(gofig.String, "", "", "", Name+"."+Endpoint)
	r.Key(gofig.Int, "", DefaultMaxRetries, "", Name+"."+MaxRetries)
	r.Key(gofig.String, "", "", "Tag prefix for EBS naming", Name+"."+Tag)
	r.Key(gofig.String, "", "", "", OldName+"."+AccessKey)
	r.Key(gofig.String, "", "", "", OldName+"."+SecretKey)
	r.Key(gofig.String, "", "", "", OldName+"."+Region)
	r.Key(gofig.String, "", "", "", OldName+"."+Endpoint)
	r.Key(gofig.Int, "", DefaultMaxRetries, "", OldName+"."+MaxRetries)
	r.Key(gofig.String, "", "", "Tag prefix for EBS naming", OldName+"."+Tag)
	gofig.Register(r)
}
