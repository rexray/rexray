package handlers

import (
	"net/http"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
)

// errorHandler is a global HTTP filter for handlling errors
type errorHandler struct {
	handler httputils.APIFunc
}

// NewErrorHandler returns a new global HTTP filter for handling errors.
func NewErrorHandler() httputils.Middleware {
	return &errorHandler{}
}

func (h *errorHandler) Name() string {
	return "error-handler"
}

func (h *errorHandler) Handler(m httputils.APIFunc) httputils.APIFunc {
	h.handler = m
	return h.Handle
}

// Handle is the type's Handler function.
func (h *errorHandler) Handle(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	err := h.handler(ctx, w, req, store)
	if err == nil {
		return nil
	}

	ctx.Log().Error(err)

	jsonError := types.JSONError{
		Status:  getStatus(err),
		Message: err.Error(),
		Error:   err,
	}

	httputils.WriteJSON(w, jsonError.Status, jsonError)
	return nil
}

func getStatus(err error) int {
	switch err.(type) {
	case *types.ErrNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
