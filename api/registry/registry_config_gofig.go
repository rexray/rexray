// +build gofig

package registry

import (
	"github.com/akutz/gofig"
	"github.com/codedellemc/libstorage/api/types"
)

func init() {
	NewConfig = gofig.New
	NewConfigReg = gofig.NewRegistration
}

// ProcessRegisteredConfigs processes the registered configuration requests.
func ProcessRegisteredConfigs(ctx types.Context) {
	for r := range ConfigRegs(ctx) {
		gofig.Register(r)
	}
}
