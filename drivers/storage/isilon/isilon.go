// +build !libstorage_storage_driver libstorage_storage_driver_isilon

package isilon

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the provider's name.
	Name = "isilon"
)

func init() {
	r := gofigCore.NewRegistration("Isilon")
	r.Key(gofig.String, "", "", "", "isilon.endpoint")
	r.Key(gofig.Bool, "", false, "", "isilon.insecure")
	r.Key(gofig.String, "", "", "", "isilon.userName")
	r.Key(gofig.String, "", "", "", "isilon.group")
	r.Key(gofig.String, "", "", "", "isilon.password")
	r.Key(gofig.String, "", "", "", "isilon.volumePath")
	r.Key(gofig.String, "", "", "", "isilon.nfsHost")
	r.Key(gofig.String, "", "", "", "isilon.dataSubnet")
	r.Key(gofig.Bool, "", false, "", "isilon.quotas")
	r.Key(gofig.Bool, "", false, "", "isilon.sharedMounts")
	gofigCore.Register(r)
}
