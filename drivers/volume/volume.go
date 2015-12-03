package volume

import (
	// loads the volume drivers
	"github.com/akutz/gofig"
	_ "github.com/emccode/rexray/drivers/volume/docker"
)

func init() {
	gofig.Register(configRegistration())
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("Volume")
	r.Key(gofig.Bool, "", false, "", "rexray.volume.mount.preempt", "preempt")
	return r
}
