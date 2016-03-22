// +build ignore

package driver

import (
	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/server/router"
	"github.com/emccode/libstorage/api/types/drivers"
)

// driverRouter is a router to talk with the driver controller
type driverRouter struct {
	config  gofig.Config
	drivers map[string]drivers.StorageDriver
	routes  []router.Route
}

// NewRouter initializes a new driverRouter
func NewRouter(
	config gofig.Config,
	drivers map[string]drivers.StorageDriver) router.Router {

	r := &driverRouter{
		config:  config,
		drivers: drivers,
	}

	r.initRoutes()
	return r
}

// Routes returns the available routers to the driver controller
func (r *driverRouter) Routes() []router.Route {
	return r.routes
}

func (r *driverRouter) initRoutes() {
	r.routes = []router.Route{
		// GET
		router.NewGetRoute(
			"drivers",
			"/drivers",
			r.driversList),
		router.NewGetRoute(
			"driverInspect",
			"/drivers/{driver}",
			r.driverInspect),
		router.NewGetRoute(
			"executors",
			"/drivers/{driver}/executors",
			r.executorsList),
		router.NewGetRoute(
			"executorInspect",
			"/drivers/{driver}/executors/{name}",
			r.executorDownload),
	}
}
