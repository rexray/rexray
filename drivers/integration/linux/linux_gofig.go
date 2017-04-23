// +build gofig

package linux

import (
	"path"

	gofig "github.com/akutz/gofig/types"

	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"
)

func init() {
	registry.RegisterConfigReg(
		"Integration",
		func(ctx types.Context, r gofig.ConfigRegistration) {

			r.Key(
				gofig.String,
				"", "ext4", "",
				types.ConfigIgVolOpsCreateDefaultFsType)

			r.Key(
				gofig.String,
				"", "", "", types.ConfigIgVolOpsCreateDefaultType)

			r.Key(
				gofig.String,
				"", "", "",
				types.ConfigIgVolOpsCreateDefaultIOPS)

			r.Key(
				gofig.String,
				"", "16", "",
				types.ConfigIgVolOpsCreateDefaultSize)

			r.Key(
				gofig.String,
				"", "", "",
				types.ConfigIgVolOpsCreateDefaultAZ)

			r.Key(
				gofig.String,
				"",
				path.Join(context.MustPathConfig(ctx).Lib, "volumes"),
				"",
				types.ConfigIgVolOpsMountPath)

			r.Key(
				gofig.String,
				"", "/data", "",
				types.ConfigIgVolOpsMountRootPath)

			r.Key(
				gofig.Bool,
				"", true, "",
				types.ConfigIgVolOpsCreateImplicit)

			r.Key(
				gofig.Bool,
				"", false, "",
				types.ConfigIgVolOpsMountPreempt)
		})
}
