package handlers

import (
	"net/http"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

// serviceValidator is an HTTP filter for validating that the service
// specified as part of the path is valid.
type serviceValidator struct {
	handler types.APIFunc
}

// NewServiceValidator returns a new filter for validating that the service
// specified as part of the path is valid.
func NewServiceValidator() types.Middleware {
	return &serviceValidator{}
}

func (h *serviceValidator) Name() string {
	return "service-validator"
}

func (h *serviceValidator) Handler(m types.APIFunc) types.APIFunc {
	return (&serviceValidator{m}).Handle
}

// Handle is the type's Handler function.
func (h *serviceValidator) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	if !store.IsSet("service") {
		return utils.NewStoreKeyErr("service")
	}

	serviceName := store.GetString("service")
	service := services.GetStorageService(ctx, serviceName)
	if service == nil {
		return utils.NewNotFoundError(serviceName)
	}

	ctx = context.WithStorageService(ctx, service)
	return h.handler(ctx, w, req, store)
}
