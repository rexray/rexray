package core

import (
	"fmt"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	"github.com/emccode/rexray/util"
)

func init() {
	initDrivers()
	gofig.SetGlobalConfigPath(util.EtcDirPath())
	gofig.SetUserConfigPath(fmt.Sprintf("%s/.rexray", gotil.HomeDir()))
	gofig.Register(globalRegistration())
	gofig.Register(driverRegistration())
}

func globalRegistration() *gofig.Registration {
	r := gofig.NewRegistration("Global")
	r.Yaml(`
rexray:
    host: tcp://:7979
    logLevel: warn
`)
	r.Key(gofig.String, "h", "tcp://:7979",
		"The REX-Ray host", "rexray.host")
	r.Key(gofig.String, "l", "warn",
		"The log level (error, warn, info, debug)", "rexray.logLevel")
	return r
}

func driverRegistration() *gofig.Registration {
	r := gofig.NewRegistration("Driver")
	r.Yaml(`
rexray:
    osDrivers:
    - linux
    storageDrivers:
    - libstorage
    volumeDrivers:
    - docker
`)
	r.Key(gofig.String, "", "linux",
		"The OS drivers to consider", "rexray.osDrivers")
	r.Key(gofig.String, "", "",
		"The storage drivers to consider", "rexray.storageDrivers")
	r.Key(gofig.String, "", "docker",
		"The volume drivers to consider", "rexray.volumeDrivers")
	return r
}
