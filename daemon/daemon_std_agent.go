// +build !rexray_build_type_client
// +build !rexray_build_type_controller

package daemon

import (
	"os"

	gofig "github.com/akutz/gofig/types"
	apitypes "github.com/codedellemc/libstorage/api/types"

	"github.com/codedellemc/rexray/daemon/module"
)

func start(
	ctx apitypes.Context,
	config gofig.Config,
	host string,
	stop <-chan os.Signal) (<-chan error, error) {

	moduleErrChan, err := module.InitializeDefaultModules(ctx, config)
	if err != nil {
		ctx.WithError(err).Error("default module(s) failed to initialize")
		return nil, err
	}

	if err = module.StartDefaultModules(ctx, config); err != nil {
		ctx.WithError(err).Error("default module(s) failed to start")
		return nil, err
	}

	return moduleErrChan, nil
}
