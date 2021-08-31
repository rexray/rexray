// +build !client
// +build !agent

package cli

import (
	gofig "github.com/akutz/gofig/types"

	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/util"
)

func init() {
	startFuncs = append(startFuncs, startLibStorageAsService)
}

func startLibStorageAsService(
	ctx apitypes.Context,
	config gofig.Config) (apitypes.Context, <-chan error, error) {

	var (
		err  error
		errs <-chan error
	)

	// Attempt to activate libStorage. If an ErrHostDetectionFailed
	// error occurs then just log it as a warning since modules may
	// define hosts directly
	ctx, _, errs, err = util.ActivateLibStorage(ctx, config)
	if err != nil {
		if err.Error() == util.ErrHostDetectionFailed.Error() {
			ctx.Warn(err)
		} else {
			return nil, nil, err
		}
	}

	if errs == nil {
		return ctx, nil, nil
	}

	var (
		waitErrs = make(chan error, 1)
		strtErrs = make(chan error, 1)
	)

	go func() {
		for err := range errs {
			if err != nil {
				waitErrs <- err
				strtErrs <- err
			}
		}
		close(waitErrs)
	}()

	// Stop libStorage if the context is cancelled.
	go func() {
		ctx.Info("libStorage context cancellation - waiting")
		<-ctx.Done()
		ctx.Info("libStorage context cancellation - received")
		util.WaitUntilLibStorageStopped(ctx, waitErrs)
		close(strtErrs)
	}()

	return ctx, strtErrs, nil
}
