// +build go1.7

package libstorage

import (
	"context"

	gofig "github.com/akutz/gofig/types"

	apictx "github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/server"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	"github.com/AVENTER-UG/rexray/libstorage/client"
)

// New starts an embedded libStorage server and returns both the server
// instnace as well as a client connected to said instnace.
//
// While a new server may be launched, it's still up to the caller to provide
// a config instance with the correct properties to specify service
// information for a libStorage server.
func New(
	goCtx context.Context,
	config gofig.Config) (types.Client, types.Server, <-chan error, error) {

	ctx := apictx.New(goCtx)

	if _, ok := apictx.PathConfig(ctx); !ok {
		pathConfig := utils.NewPathConfig()
		ctx = ctx.WithValue(apictx.PathConfigKey, pathConfig)
	}

	s, errs, err := server.Serve(ctx, config)
	if err != nil {
		return nil, nil, nil, err
	}

	if h := config.GetString(types.ConfigHost); h == "" {
		config.Set(types.ConfigHost, s.Addrs()[0])
	}

	c, err := client.New(ctx, config)
	if err != nil {
		return nil, nil, nil, err
	}

	return c, s, errs, nil
}
