package handlers

import (
	"net/http"
	"strings"

	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/api/utils"
)

// serviceValidator is an HTTP filter for validating that the service
// specified as part of the path is valid.
type serviceValidator struct {
	handler apihttp.APIFunc
}

// NewServiceValidator returns a new filter for validating that the service
// specified as part of the path is valid.
func NewServiceValidator() apihttp.Middleware {
	return &serviceValidator{}
}

func (h *serviceValidator) Name() string {
	return "service-validator"
}

func (h *serviceValidator) Handler(m apihttp.APIFunc) apihttp.APIFunc {
	return (&serviceValidator{m}).Handle
}

// Handle is the type's Handler function.
func (h *serviceValidator) Handle(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	if !store.IsSet("service") {
		return utils.NewStoreKeyErr("service")
	}

	serviceName := store.GetString("service")
	service := services.GetStorageService(serviceName)
	if service == nil {
		return utils.NewNotFoundError(serviceName)
	}

	instanceIDs, ok := ctx.Value("instanceIDs").(map[string]*types.InstanceID)
	if ok {
		if iid, ok := instanceIDs[strings.ToLower(serviceName)]; ok {
			ctx = ctx.WithInstanceID(iid)
			ctx = ctx.WithContextID("instanceID", iid.ID)
			ctx.Log().Debug("set instanceID")
		}
	}

	ctx = ctx.WithValue("service", service)
	ctx = ctx.WithValue("serviceID", service.Name())
	ctx = ctx.WithContextID("service", service.Name())
	ctx = ctx.WithContextID("driver", service.Driver().Name())

	ctx.Log().Debug("set service context")
	return h.handler(ctx, w, req, store)
}
