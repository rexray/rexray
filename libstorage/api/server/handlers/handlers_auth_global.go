package handlers

import (
	"net/http"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/auth"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

// authGlobalHandler is an HTTP filter for validating the JWT.
type authGlobalHandler struct {
	handler types.APIFunc
	config  *types.AuthConfig
}

// NewAuthGlobalHandler returns a new authGlobalHandler.
func NewAuthGlobalHandler(
	config *types.AuthConfig) types.Middleware {
	return &authGlobalHandler{config: config}
}

func (h *authGlobalHandler) Name() string {
	return "auth-global-handler"
}

func (h *authGlobalHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&authGlobalHandler{m, h.config}).Handle
}

// Handle is the type's Handler function.
func (h *authGlobalHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	if h.config == nil {
		ctx.Debug("skipping global auth handler; empty auth config")
		return h.handler(ctx, w, req, store)
	}

	if len(h.config.Allow) == 0 && len(h.config.Deny) == 0 {
		ctx.Debug("skipping global auth handler; empty allow & deny lists")
		return h.handler(ctx, w, req, store)
	}

	tok, err := auth.ValidateAuthTokenWithReq(ctx, h.config, req)
	if err != nil {
		return err
	}

	if tok == nil {
		panic("token should never be nil here")
	}

	ctx.Debug("validated global security token")

	return h.handler(
		ctx.WithValue(context.AuthTokenKey, tok), w, req, store)
}
