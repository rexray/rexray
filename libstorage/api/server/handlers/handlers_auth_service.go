package handlers

import (
	"net/http"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/auth"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

// authSvcHandler is an HTTP filter for validating the JWT.
type authSvcHandler struct {
	handler types.APIFunc
}

// NewAuthSvcHandler returns a new authSvcHandler.
func NewAuthSvcHandler() types.Middleware {
	return &authSvcHandler{}
}

func (h *authSvcHandler) Name() string {
	return "auth-svc-handler"
}

func (h *authSvcHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&authSvcHandler{m}).Handle
}

// Handle is the type's Handler function.
func (h *authSvcHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	svc, ok := context.Service(ctx)
	if !ok {
		return types.ErrMissingStorageService
	}

	if svc.AuthConfig() == nil {
		ctx.Debug("skipping service auth handler; empty auth config")
		return h.handler(ctx, w, req, store)
	}

	if len(svc.AuthConfig().Allow) == 0 && len(svc.AuthConfig().Deny) == 0 {
		ctx.Debug("skipping svc auth handler; empty allow & deny lists")
		return h.handler(ctx, w, req, store)
	}

	tok, err := auth.ValidateAuthTokenWithCtxOrReq(ctx, svc.AuthConfig(), req)
	if err != nil {
		return err
	}

	if tok == nil {
		panic("token should never be nil here")
	}

	ctx.Debug("validated service security token")

	return h.handler(
		ctx.WithValue(context.AuthTokenKey, tok), w, req, store)
}
