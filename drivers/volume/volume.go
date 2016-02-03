package volume

import (
	// loads the volume drivers
	_ "github.com/emccode/rexray/drivers/volume/docker"

	"github.com/akutz/gofig"
)

func init() {
	gofig.Register(configRegistration())
}

func configRegistration() *gofig.Registration {
	r := gofig.NewRegistration("Volume")
	r.Key(gofig.Bool, "", false, "", "rexray.volume.mount.preempt", "preempt")
	return r
}
