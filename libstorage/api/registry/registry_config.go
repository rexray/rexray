package registry

import (
	"github.com/akutz/gofig"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

// NewConfig is a function that returns a new Config object.
var NewConfig = gofig.New

// NewConfigReg is a function that returns a new ConfigRegistration object.
var NewConfigReg = gofig.NewRegistration

// ProcessRegisteredConfigs processes the registered configuration requests.
func ProcessRegisteredConfigs(ctx types.Context) {
	for r := range ConfigRegs(ctx) {
		gofig.Register(r)
	}
}
