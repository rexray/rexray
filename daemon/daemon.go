// +build !exclude_module

package daemon

import (
	"os"

	"github.com/akutz/gofig"
	apitypes "github.com/codedellemc/libstorage/api/types"

	"github.com/codedellemc/rexray/daemon/module"
	"github.com/codedellemc/rexray/util"
)

// Start starts the daemon.
func Start(
	ctx apitypes.Context,
	config gofig.Config,
	host string,
	stop <-chan os.Signal) (<-chan error, error) {

	var (
		err           error
		errs          = make(chan error)
		serverErrChan <-chan error
	)

	if serverErrChan, err = module.InitializeDefaultModules(
		ctx, config); err != nil {
		ctx.WithError(err).Error("default module(s) failed to initialize")
		return nil, err
	}

	if err = module.StartDefaultModules(ctx, config); err != nil {
		ctx.WithError(err).Error("default module(s) failed to start")
		return nil, err
	}

	ctx.Info("service successfully initialized, waiting on stop signal")

	go func() {
		sig := <-stop
		ctx.WithField("signal", sig).Info("service received stop signal")
		util.WaitUntilLibStorageStopped(ctx, serverErrChan)
		close(errs)
	}()

	return errs, nil
}
