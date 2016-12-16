// +build !libstorage_storage_driver libstorage_storage_driver_rbd

package rbd

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the name of the storage driver
	Name = "rbd"
)

func init() {
	registerConfig()
}

func registerConfig() {
	r := gofigCore.NewRegistration("RBD")
	r.Key(gofig.String, "", "rbd", "", "rbd.defaultPool")
	gofigCore.Register(r)
}
