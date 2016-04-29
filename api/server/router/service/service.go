package service

import (
	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/server/handlers"
	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils/schema"
)

func init() {
	registry.RegisterRouter(&router{})
}

type router struct {
	routes []types.Route
}

func (r *router) Name() string {
	return "service-router"
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
			"services",
			"/services",
			r.servicesList,
			handlers.NewSchemaValidator(nil, schema.ServiceInfoMapSchema, nil)),

		httputils.NewGetRoute(
			"serviceInspect",
			"/services/{service}",
			r.serviceInspect,
			handlers.NewServiceValidator(),
			handlers.NewSchemaValidator(nil, schema.ServiceInfoSchema, nil)),
	}
}
