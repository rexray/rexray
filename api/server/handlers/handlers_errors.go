package handlers

import (
	"net/http"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
)

// errorHandler is a global HTTP filter for handlling errors
type errorHandler struct {
	handler types.APIFunc
}

// NewErrorHandler returns a new global HTTP filter for handling errors.
func NewErrorHandler() types.Middleware {
	return &errorHandler{}
}

func (h *errorHandler) Name() string {
	return "error-handler"
}

func (h *errorHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&errorHandler{m}).Handle
}

// Handle is the type's Handler function.
func (h *errorHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	err := h.handler(ctx, w, req, store)
	if err == nil {
		return nil
	}

	ctx.Error(err)

	jsonError := types.JSONError{
		Status:     getStatus(err),
		Message:    err.Error(),
		InnerError: err,
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
