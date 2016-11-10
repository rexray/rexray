package util

import (
	"bytes"
	"fmt"

	gofig "github.com/akutz/gofig/types"
	apitypes "github.com/codedellemc/libstorage/api/types"
)

const defaultServiceConfigFormat = `
rexray:
  libstorage:
    service: %[1]s
    server:
      services:
        %[1]s:
          driver: %[1]s
`

// initDefaultLibStorageServices initializes the config object with a default
// libStorage service if one is not present.
//
// TODO Move this into libStorage in libStorage 0.1.2
func initDefaultLibStorageServices(
	ctx apitypes.Context, config gofig.Config) error {

	if config.IsSet(apitypes.ConfigServices) {
		ctx.Debug(
			"libStorage auto service mode disabled; services defined")
		return nil
	}

	serviceName := config.GetString(apitypes.ConfigService)
	if serviceName == "" {
		ctx.Debug(
			"libStorage auto service mode disabled; service name empty")
		return nil
	}

	ctx.WithField("driver", serviceName).Info(
		"libStorage auto service mode enabled")

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, defaultServiceConfigFormat, serviceName)

	if err := config.ReadConfig(buf); err != nil {
		return err
	}

	return nil
}
