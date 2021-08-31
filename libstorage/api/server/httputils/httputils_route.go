package httputils

import (
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

// route defines an individual API route.
type route struct {
	name        string
	method      string
	path        string
	queries     []string
	handler     types.APIFunc
	middlewares []types.Middleware
}

func (r *route) ContextLoggerField() (string, interface{}) {
	return "route", r.name
}

// Method specifies the method for the route.
func (r *route) Method(method string) types.Route {
	r.method = method
	return r
}

// Path specifies the path for the route.
func (r *route) Path(path string) types.Route {
	r.path = path
	return r
}

// Queries add query strings that must match for a route.
func (r *route) Queries(queries ...string) types.Route {
	r.queries = append(r.queries, queries...)
	if len(queries) == 1 {
		r.queries = append(r.queries, "")
	}
	return r
}

// Handler specifies the handler for the route.
func (r *route) Handler(handler types.APIFunc) types.Route {
	r.handler = handler
	return r
}

// Middlewares adds middleware to the route.
func (r *route) Middlewares(
	middlewares ...types.Middleware) types.Route {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// Name returns the name of the route.
func (r *route) GetName() string {
	return r.name
}

// GetHandler returns the APIFunc to let the server wrap it in middlewares
func (r *route) GetHandler() types.APIFunc {
	return r.handler
}

// GetMethod returns the http method that the route responds to.
func (r *route) GetMethod() string {
	return r.method
}

// GetPath returns the subpath where the route responds to.
func (r *route) GetPath() string {
	return r.path
}

// GetQueries returns the query strings for which the route should respond.
func (r *route) GetQueries() []string {
	return r.queries
}

// GetMiddlwares returns a list of route-specific middleware.
func (r *route) GetMiddlewares() []types.Middleware {
	return r.middlewares
}

// NewRoute initialies a new local route for the reouter
func NewRoute(
	name, method, path string,
	handler types.APIFunc,
	middlewares ...types.Middleware) types.Route {

	return &route{
		name:        name,
		method:      method,
		path:        path,
		handler:     handler,
		middlewares: middlewares,
	}
}

// NewGetRoute initializes a new route with the http method GET.
func NewGetRoute(
	name, path string,
	handler types.APIFunc,
	middlewares ...types.Middleware) types.Route {
	return NewRoute(name, "GET", path, handler, middlewares...)
}

// NewPostRoute initializes a new route with the http method POST.
func NewPostRoute(
	name, path string,
	handler types.APIFunc,
	middlewares ...types.Middleware) types.Route {
	return NewRoute(name, "POST", path, handler, middlewares...)
}

// NewPutRoute initializes a new route with the http method PUT.
func NewPutRoute(
	name, path string,
	handler types.APIFunc,
	middlewares ...types.Middleware) types.Route {
	return NewRoute(name, "PUT", path, handler, middlewares...)
}

// NewDeleteRoute initializes a new route with the http method DELETE.
func NewDeleteRoute(
	name, path string,
	handler types.APIFunc,
	middlewares ...types.Middleware) types.Route {
	return NewRoute(name, "DELETE", path, handler, middlewares...)
}

// NewOptionsRoute initializes a new route with the http method OPTIONS
func NewOptionsRoute(
	name, path string,
	handler types.APIFunc,
	middlewares ...types.Middleware) types.Route {
	return NewRoute(name, "OPTIONS", path, handler, middlewares...)
}

// NewHeadRoute initializes a new route with the http method HEAD.
func NewHeadRoute(
	name, path string,
	handler types.APIFunc,
	middlewares ...types.Middleware) types.Route {
	return NewRoute(name, "HEAD", path, handler, middlewares...)
}
