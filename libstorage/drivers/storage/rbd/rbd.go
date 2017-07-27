package rbd

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the name of the storage driver
	Name = "rbd"

	// ConfigDefaultPool is the config key for default pool
	ConfigDefaultPool = Name + ".defaultPool"

	// ConfigUserName is the config key for rbd username
	ConfigUserName = Name + ".userName"

	// ConfigTestModule is the config key for testing kernel module presence
	ConfigTestModule = Name + ".testModule"
)

func init() {
	registerConfig()
}

func registerConfig() {
	r := gofigCore.NewRegistration("RBD")
	r.Key(gofig.String, "", "rbd", "", ConfigDefaultPool)
	r.Key(gofig.String, "", "admin", "", ConfigUserName)
	r.Key(gofig.Bool, "", true, "", ConfigTestModule)
	gofigCore.Register(r)
}
