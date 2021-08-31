package handlers

import (
	"net/http"

	"github.com/AVENTER-UG/rexray/libstorage/api/server/auth"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/services"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

// authAllSvcsHandler is an HTTP filter for validating the JWT.
type authAllSvcsHandler struct {
	handler types.APIFunc
}

// NewAuthAllSvcsHandler returns a new authAllSvcsHandler.
func NewAuthAllSvcsHandler() types.Middleware {
	return &authAllSvcsHandler{}
}

func (h *authAllSvcsHandler) Name() string {
	return "auth-svc-handler"
}

func (h *authAllSvcsHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&authAllSvcsHandler{m}).Handle
}

// Handle is the type's Handler function.
func (h *authAllSvcsHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	for svc := range services.StorageServices(ctx) {
		if svc.AuthConfig() == nil {
			ctx.WithField("service", svc.Name()).Debug(
				"skipping service auth handler; empty auth config")
			continue
		}
		if len(svc.AuthConfig().Allow) == 0 && len(svc.AuthConfig().Deny) == 0 {
			ctx.Debug("skipping svc auth handler; empty allow & deny lists")
			continue
		}
		_, err := auth.ValidateAuthTokenWithCtxOrReq(
			ctx, svc.AuthConfig(), req)
		if err != nil {
			return err
		}
	}

	ctx.Debug("validated all services access")

	return h.handler(ctx, w, req, store)
}
