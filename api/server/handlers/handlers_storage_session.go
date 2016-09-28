package handlers

import (
	"net/http"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/types"
)

// storageSessionHandler is an HTTP filter for ensuring that a storage session
// is injected for routes that request it.
type storageSessionHandler struct {
	handler types.APIFunc
}

// NewStorageSessionHandler returns a new filter for ensuring that a storage
// session is injected for routes that request it.
func NewStorageSessionHandler() types.Middleware {
	return &storageSessionHandler{}
}

func (h *storageSessionHandler) Name() string {
	return "storage-session-handler"
}

func (h *storageSessionHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&storageSessionHandler{m}).Handle
}

// Handle is the type's Handler function.
func (h *storageSessionHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	var err error
	if ctx, err = context.WithStorageSession(ctx); err != nil {
		return err
	}
	return h.handler(ctx, w, req, store)
}
