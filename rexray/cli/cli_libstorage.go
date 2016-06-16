package cli

import (
	"bytes"
	"fmt"

	"github.com/akutz/gofig"
	apitypes "github.com/emccode/libstorage/api/types"
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
		return nil
	}

	serviceName := config.GetString(apitypes.ConfigService)
	if serviceName == "" {
		return nil
	}

	ctx.WithField("driver", serviceName).Info(
		"initializing default libStorage services")

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, defaultServiceConfigFormat, serviceName)

	if err := config.ReadConfig(buf); err != nil {
		return err
	}

	return nil
}
