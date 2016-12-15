// +build rexray_build_type_controller

package util

import (
	gofig "github.com/akutz/gofig/types"
	apitypes "github.com/codedellemc/libstorage/api/types"
)

func newClient(apitypes.Context, gofig.Config) (apitypes.Client, error) {
	return nil, nil
}
