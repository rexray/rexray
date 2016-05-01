package registry

import (
	"strings"

	"github.com/emccode/libstorage/api/types"
)

type odm struct {
	types.OSDriver
	types.Context
}

// NewOSDriverManager returns a new OS driver manager.
func NewOSDriverManager(
	d types.OSDriver) types.OSDriver {
	return &odm{OSDriver: d}
}

func (d *odm) Format(
	ctx types.Context,
	deviceName string,
	opts *types.DeviceFormatOpts) error {
	if strings.Contains(deviceName, ":") {
		return nil
	}
	return d.OSDriver.Format(ctx, deviceName, opts)
}
