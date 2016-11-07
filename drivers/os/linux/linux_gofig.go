// +build gofig

package linux

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

func init() {
	r := gofigCore.NewRegistration("Linux")
	r.Key(gofig.Int, "", 0700, "", "linux.volume.filemode")
	r.Key(gofig.String, "", "/data", "", "linux.volume.rootpath")
	gofigCore.Register(r)
}
