package rackspace

import	"github.com/akutz/gofig";

// Name is the provider's name.
const	Name string  = "rackspace";

func init() {
	registerConfig()
}

func registerConfig() {
	r := gofig.NewRegistration("Rackspace")
	r.Key(gofig.String, "", "", "", "rackspace.authURL")
	r.Key(gofig.String, "", "", "", "rackspace.userID")
	r.Key(gofig.String, "", "", "", "rackspace.userName")
	r.Key(gofig.String, "", "", "", "rackspace.password")
	r.Key(gofig.String, "", "", "", "rackspace.tenantID")
	r.Key(gofig.String, "", "", "", "rackspace.tenantName")
	r.Key(gofig.String, "", "", "", "rackspace.domainID")
	r.Key(gofig.String, "", "", "", "rackspace.domainName")
	gofig.Register(r)
}