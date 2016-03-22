package httputils

import (
	"net/http"

	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
)

// APIFunc is an adapter to allow the use of ordinary functions as API
// endpoints. Any function that has the appropriate signature can be register
// as an API endpoint.
type APIFunc func(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	store types.Store) error

// NewRequestObjFunc is a function that creates a new instance of the type to
// which the request body is serialized.
type NewRequestObjFunc func() interface{}

// Service is information about a service.
type Service interface {
	// Name gets the name of the service.
	Name() string

	// Driver gets the service's driver.
	Driver() drivers.StorageDriver

	// Config gets the service's configuration.
	Config() gofig.Config
}
