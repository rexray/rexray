// +build !rexray_build_type_client
// +build !rexray_build_type_agent
// +build !rexray_build_type_controller

package util

import (
	gofig "github.com/akutz/gofig/types"

	apitypes "github.com/codedellemc/rexray/libstorage/api/types"
	apiclient "github.com/codedellemc/rexray/libstorage/client"
)

func newClient(
	ctx apitypes.Context, config gofig.Config) (apitypes.Client, error) {
	return apiclient.New(ctx, config)
}
