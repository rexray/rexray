// +build rexray_build_type_controller

package daemon

import (
	"os"

	gofig "github.com/akutz/gofig/types"
	apitypes "github.com/codedellemc/libstorage/api/types"

	"github.com/codedellemc/rexray/util"
)

func start(
	ctx apitypes.Context,
	config gofig.Config,
	host string,
	stop <-chan os.Signal) (<-chan error, error) {
	ctx, _, errs, err := util.ActivateLibStorage(ctx, config)
	if err != nil {
		ctx.WithError(err).Error(
			"error activing libStorage in server-only mode")
		return nil, err
	}
	return errs, nil
}
