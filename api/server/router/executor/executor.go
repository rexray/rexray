package executor

import (
	gofig "github.com/akutz/gofig/types"

	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/server/httputils"
	"github.com/codedellemc/libstorage/api/types"
)

func init() {
	registry.RegisterRouter(&router{})
}

type router struct {
	routes []types.Route
}

func (r *router) Name() string {
	return "executor-router"
}

func (r *router) Init(config gofig.Config) {
	r.initRoutes()
}

// Routes returns the available routes.
func (r *router) Routes() []types.Route {
	return r.routes
}

func (r *router) initRoutes() {

	r.routes = []types.Route{

		// GET
		httputils.NewGetRoute(
			"executors",
			"/executors",
			r.executors),

		// GET
		httputils.NewGetRoute(
			"executorInspect",
			"/executors/{executor}",
			r.executorInspect),

		// HEAD
		httputils.NewHeadRoute(
			"executorHead",
			"/executors/{executor}",
			r.executorHead),
	}
}
