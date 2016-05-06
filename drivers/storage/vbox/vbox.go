package vbox

import (
	"github.com/akutz/gofig"
)

const (
	// Name is the provider's name.
	Name = "virtualbox"
)

func init() {
	registerConfig()
}

func registerConfig() {
	r := gofig.NewRegistration("Oracle VM VirtualBox")
	r.Key(gofig.String, "", "", "", "virtualbox.username")
	r.Key(gofig.String, "", "", "", "virtualbox.password")
	r.Key(gofig.String, "", "http://127.0.0.1:18083", "", "virtualbox.endpoint")
	r.Key(gofig.String, "", "", "", "virtualbox.volumePath")
	r.Key(gofig.String, "", "", "", "virtualbox.localMachineNameOrId")
	r.Key(gofig.Bool, "", false, "", "virtualbox.tls")
	r.Key(gofig.String, "", "", "", "virtualbox.controllerName")
	r.Key(gofig.String, "", "/dev/disk/by-id", "", "virtualbox.diskIDPath")
	r.Key(gofig.String,
		"", "/sys/class/scsi_host/", "", "virtualbox.scsiHostPath")
	gofig.Register(r)
}
