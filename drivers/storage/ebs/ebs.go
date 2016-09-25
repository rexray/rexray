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
)

func init() {
	registerConfig()
}

func registerConfig() {
	r := gofig.NewRegistration("EBS")
	r.Key(gofig.String, "", "", "", "ebs.accessKey")
	r.Key(gofig.String, "", "", "", "ebs.secretKey")
	r.Key(gofig.String, "", "", "", "ebs.region")
	r.Key(gofig.String, "", "", "", "ebs.endpoint")
	r.Key(gofig.String, "", "", "", "ebs.maxRetries")
	r.Key(gofig.String, "", "", "Tag prefix for EBS naming", "ebs.tag")
	r.Key(gofig.String, "", "", "", "ec2.accessKey")
	r.Key(gofig.String, "", "", "", "ec2.secretKey")
	r.Key(gofig.String, "", "", "", "ec2.region")
	r.Key(gofig.String, "", "", "", "ec2.endpoint")
	r.Key(gofig.String, "", "", "", "ec2.maxRetries")
	r.Key(gofig.String, "", "", "Tag prefix for EBS naming", "ec2.tag")
	gofig.Register(r)
}
