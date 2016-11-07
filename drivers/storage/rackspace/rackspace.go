package rackspace

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

// Name is the provider's name.
const Name string = "rackspace"

func init() {
	r := gofigCore.NewRegistration("Rackspace")
	r.Key(gofig.String, "", "", "", "rackspace.authURL")
	r.Key(gofig.String, "", "", "", "rackspace.userID")
	r.Key(gofig.String, "", "", "", "rackspace.userName")
	r.Key(gofig.String, "", "", "", "rackspace.password")
	r.Key(gofig.String, "", "", "", "rackspace.tenantID")
	r.Key(gofig.String, "", "", "", "rackspace.tenantName")
	r.Key(gofig.String, "", "", "", "rackspace.domainID")
	r.Key(gofig.String, "", "", "", "rackspace.domainName")
	gofigCore.Register(r)
}
