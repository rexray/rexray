// +build rexray_build_type_client rexray_build_type_agent

package util

import (
	gofig "github.com/akutz/gofig/types"

	apitypes "github.com/codedellemc/rexray/libstorage/api/types"
	apiclient "github.com/codedellemc/rexray/libstorage/client"
)

func activateLibStorage(
	ctx apitypes.Context,
	config gofig.Config) (apitypes.Context, gofig.Config, <-chan error, error) {
	var (
		host      string
		isRunning bool
	)
	if host = config.GetString(apitypes.ConfigHost); host == "" {
		if host, isRunning = IsLocalServerActive(ctx, config); isRunning {
			ctx = setHost(ctx, config, host)
		}
	}
	if host == "" && !isRunning {
		return ctx, config, nil, ErrHostDetectionFailed
	}
	return ctx, config, nil, nil
}

func waitUntilLibStorageStopped(apitypes.Context, <-chan error) {}

func newClient(
	ctx apitypes.Context, config gofig.Config) (apitypes.Client, error) {
	return apiclient.New(ctx, config)
}
