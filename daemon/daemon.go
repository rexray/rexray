// +build !rexray_build_type_client

package daemon

import (
	"os"

	gofig "github.com/akutz/gofig/types"
	apitypes "github.com/codedellemc/rexray/libstorage/api/types"

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
		daemonErrChan <-chan error
	)

	if daemonErrChan, err = start(ctx, config, host, stop); err != nil {
		ctx.WithError(err).Error("daemon failed to initialize")
		return nil, err
	}

	ctx.Info("service successfully initialized, waiting on stop signal")

	go func() {
		sig := <-stop
		ctx.WithField("signal", sig).Info("service received stop signal")
		util.WaitUntilLibStorageStopped(ctx, daemonErrChan)
		close(errs)
	}()

	return errs, nil
}
