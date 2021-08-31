package service

import (
	gofig "github.com/akutz/gofig/types"

	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/handlers"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/httputils"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils/schema"
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
			handlers.NewAuthAllSvcsHandler(),
			handlers.NewSchemaValidator(nil, schema.ServiceInfoMapSchema, nil)),

		httputils.NewGetRoute(
			"serviceInspect",
			"/services/{service}",
			r.serviceInspect,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewSchemaValidator(nil, schema.ServiceInfoSchema, nil)),
	}
}
