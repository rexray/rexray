// +build !client
// +build !agent
// +build !controller

package util

import (
	gofig "github.com/akutz/gofig/types"

	apitypes "github.com/thecodeteam/rexray/libstorage/api/types"
	apiclient "github.com/thecodeteam/rexray/libstorage/client"
)

func newClient(
	ctx apitypes.Context, config gofig.Config) (apitypes.Client, error) {
	return apiclient.New(ctx, config)
}
