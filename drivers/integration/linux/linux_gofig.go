// +build gofig

package linux

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
	"github.com/codedellemc/libstorage/api/types"
)

func init() {
	r := gofigCore.NewRegistration("Integration")
	r.Key(gofig.String, "", "ext4", "",
		types.ConfigIgVolOpsCreateDefaultFsType)
	r.Key(gofig.String, "", "", "", types.ConfigIgVolOpsCreateDefaultType)
	r.Key(gofig.String, "", "", "", types.ConfigIgVolOpsCreateDefaultIOPS)
	r.Key(gofig.String, "", "16", "", types.ConfigIgVolOpsCreateDefaultSize)
	r.Key(gofig.String, "", "", "", types.ConfigIgVolOpsCreateDefaultAZ)
	r.Key(gofig.String, "", types.Lib.Join("volumes"), "",
		types.ConfigIgVolOpsMountPath)
	r.Key(gofig.String, "", "/data", "", types.ConfigIgVolOpsMountRootPath)
	r.Key(gofig.Bool, "", true, "", types.ConfigIgVolOpsCreateImplicit)
	r.Key(gofig.Bool, "", false, "", types.ConfigIgVolOpsMountPreempt)
	gofigCore.Register(r)
}
