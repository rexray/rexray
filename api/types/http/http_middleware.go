package http

import (
	"net/http"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
)

// APIFunc is an adapter to allow the use of ordinary functions as API
// endpoints. Any function that has the appropriate signature can be register
// as an API endpoint.
type APIFunc func(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	store types.Store) error

// MiddlewareFunc is an adapter to allow the use of ordinary functions as API
// filters. Any function that has the appropriate signature can be register as
// a middleware.
type MiddlewareFunc func(handler APIFunc) APIFunc

// Middleware is middleware for a route.
type Middleware interface {

	// Name returns the name of the middlware.
	Name() string

	// Handler enables the chaining of middlware.
	Handler(handler APIFunc) APIFunc

	// Handle is for processing an incoming request.
	Handle(
		ctx context.Context,
		w http.ResponseWriter,
		r *http.Request,
		store types.Store) error
}
