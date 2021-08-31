package help

import (
	gofig "github.com/akutz/gofig/types"

	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/httputils"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

func init() {
	registry.RegisterRouter(&router{})
}

type router struct {
	config gofig.Config
	routes []types.Route
}

func (r *router) Name() string {
	return "help-router"
}

func (r *router) Init(config gofig.Config) {
	r.config = config
	r.initRoutes()
}

// Routes returns the available routes.
func (r *router) Routes() []types.Route {
	return r.routes
}

func (r *router) initRoutes() {
	r.routes = []types.Route{
		// GET
		httputils.NewGetRoute("version", "/help", r.helpInspect),
		httputils.NewGetRoute("version", "/help/config", r.configInspect),
		httputils.NewGetRoute("version", "/help/env", r.envInspect),
		httputils.NewGetRoute("version", "/help/version", r.versionInspect),
	}
}
