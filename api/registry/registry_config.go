package registry

import (
	gofig "github.com/akutz/gofig/types"
)

// NewConfig is a function that returns a new Config object.
var NewConfig func() gofig.Config
