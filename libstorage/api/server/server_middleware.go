package server

import (
	"github.com/AVENTER-UG/rexray/libstorage/api/server/handlers"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/services"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

func (s *server) initGlobalMiddleware() {

	s.addGlobalMiddleware(handlers.NewQueryParamsHandler())
	if s.logHTTPEnabled {
		s.addGlobalMiddleware(handlers.NewLoggingHandler(
			s.stdOut,
			s.logHTTPRequests,
			s.logHTTPResponses))
	}
	s.addGlobalMiddleware(handlers.NewTransactionHandler())
	s.addGlobalMiddleware(handlers.NewErrorHandler())
	s.addGlobalMiddleware(handlers.NewAuthGlobalHandler(s.authConfig))
	s.addGlobalMiddleware(
		handlers.NewInstanceIDHandler(services.StorageServices(s.ctx)))
	s.addGlobalMiddleware(handlers.NewLocalDevicesHandler())
	s.addGlobalMiddleware(handlers.NewOnRequestHandler())
}

func (s *server) initRouteMiddleware() {
	// add the route-specific middleware for all the existing routes. it's
	// also possible to add route-specific middleware that is not defined as
	// part of a route's Middlewares collection.
	s.routeHandlers = map[string][]types.Middleware{}
	for _, router := range s.routers {
		for _, r := range router.Routes() {
			s.addRouterMiddleware(r, r.GetMiddlewares()...)
		}
	}
}

func (s *server) addRouterMiddleware(
	r types.Route, middlewares ...types.Middleware) {

	middlewaresForRouteName, ok := s.routeHandlers[r.GetName()]
	if !ok {
		middlewaresForRouteName = []types.Middleware{}
	}
	middlewaresForRouteName = append(middlewaresForRouteName, middlewares...)
	s.routeHandlers[r.GetName()] = middlewaresForRouteName
}

func (s *server) addGlobalMiddleware(m types.Middleware) {
	s.globalHandlers = append(s.globalHandlers, m)
}

func (s *server) handleWithMiddleware(
	ctx types.Context,
	route types.Route) types.APIFunc {

	/*if route.GetMethod() == "HEAD" {
		return route.GetHandler()
	}*/

	handler := route.GetHandler()

	middlewaresForRouteName, ok := s.routeHandlers[route.GetName()]
	if !ok {
		ctx.Warn("no middlewares for route")
	} else {
		for h := range reverse(middlewaresForRouteName) {
			handler = h.Handler(handler)
			ctx.WithField(
				"middleware", h.Name()).Debug("added route middleware")
		}
	}

	// add the global handlers
	for h := range reverse(s.globalHandlers) {
		handler = h.Handler(handler)
		ctx.WithField(
			"middleware", h.Name()).Debug("added global middleware")
	}

	return handler
}

func reverse(middlewares []types.Middleware) chan types.Middleware {
	ret := make(chan types.Middleware)
	go func() {
		for i := range middlewares {
			ret <- middlewares[len(middlewares)-1-i]
		}
		close(ret)
	}()
	return ret
}
