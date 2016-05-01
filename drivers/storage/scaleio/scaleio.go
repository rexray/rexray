package scaleio

import (
	"github.com/akutz/gofig"
)

const (
	// Name is the name of the storage driver
	Name = "scaleio"
)

func configRegistration() {
	r := gofig.NewRegistration("ScaleIO")
	r.Key(gofig.String, "", "", "", "scaleio.endpoint")
	r.Key(gofig.Bool, "", false, "", "scaleio.insecure")
	r.Key(gofig.Bool, "", false, "", "scaleio.useCerts")
	r.Key(gofig.String, "", "", "", "scaleio.userID")
	r.Key(gofig.String, "", "", "", "scaleio.userName")
	r.Key(gofig.String, "", "", "", "scaleio.password")
	r.Key(gofig.String, "", "", "", "scaleio.systemID")
	r.Key(gofig.String, "", "", "", "scaleio.systemName")
	r.Key(gofig.String, "", "", "", "scaleio.protectionDomainID")
	r.Key(gofig.String, "", "", "", "scaleio.protectionDomainName")
	r.Key(gofig.String, "", "", "", "scaleio.storagePoolID")
	r.Key(gofig.String, "", "", "", "scaleio.storagePoolName")
	r.Key(gofig.String, "", "", "", "scaleio.thinOrThick")
	r.Key(gofig.String, "", "", "", "scaleio.version")
	gofig.Register(r)
}
