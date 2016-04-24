package server

import (
	"github.com/emccode/libstorage/api/server/handlers"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
)

func (s *server) initGlobalMiddleware() {

	s.addGlobalMiddleware(handlers.NewQueryParamsHandler())

	if s.logHTTPEnabled {
		s.addGlobalMiddleware(handlers.NewLoggingHandler(
			s.stdOut,
			s.logHTTPRequests,
			s.logHTTPResponses))
	}

	s.addGlobalMiddleware(handlers.NewErrorHandler())
	s.addGlobalMiddleware(handlers.NewInstanceIDHandler())
	s.addGlobalMiddleware(handlers.NewLocalDevicesHandler())
}

func (s *server) initRouteMiddleware() {
	// add the route-specific middleware for all the existing routes. it's
	// also possible to add route-specific middleware that is not defined as
	// part of a route's Middlewares collection.
	s.routeHandlers = map[string][]apihttp.Middleware{}
	for _, router := range s.routers {
		for _, r := range router.Routes() {
			s.addRouterMiddleware(r, r.GetMiddlewares()...)
		}
	}
}

func (s *server) addRouterMiddleware(
	r apihttp.Route, middlewares ...apihttp.Middleware) {

	middlewaresForRouteName, ok := s.routeHandlers[r.GetName()]
	if !ok {
		middlewaresForRouteName = []apihttp.Middleware{}
	}
	middlewaresForRouteName = append(middlewaresForRouteName, middlewares...)
	s.routeHandlers[r.GetName()] = middlewaresForRouteName
}

func (s *server) addGlobalMiddleware(m apihttp.Middleware) {
	s.globalHandlers = append(s.globalHandlers, m)
}

func (s *server) handleWithMiddleware(
	ctx context.Context,
	route apihttp.Route) apihttp.APIFunc {

	/*if route.GetMethod() == "HEAD" {
		return route.GetHandler()
	}*/

	handler := route.GetHandler()

	middlewaresForRouteName, ok := s.routeHandlers[route.GetName()]
	if !ok {
		ctx.Log().Warn("no middlewares for route")
	} else {
		for h := range reverse(middlewaresForRouteName) {
			handler = h.Handler(handler)
			ctx.Log().WithField(
				"middleware", h.Name()).Debug("added route middleware")
		}
	}

	// add the global handlers
	for h := range reverse(s.globalHandlers) {
		handler = h.Handler(handler)
		ctx.Log().WithField(
			"middleware", h.Name()).Debug("added global middleware")
	}

	return handler
}

func reverse(middlewares []apihttp.Middleware) chan apihttp.Middleware {
	ret := make(chan apihttp.Middleware)
	go func() {
		for i := range middlewares {
			ret <- middlewares[len(middlewares)-1-i]
		}
		close(ret)
	}()
	return ret
}
