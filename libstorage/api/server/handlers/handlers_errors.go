package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/httputils"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
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

	gerr := goof.Newe(err)
	ctx.WithError(gerr).Error("error: api call failed")

	httpErr := goof.NewHTTPError(gerr, getStatus(err))
	if isLogAPICallErrJSON(ctx) {
		buf, err := json.Marshal(httpErr)
		if err != nil {
			ctx.WithError(err).Error(
				"error marshalling api call err to json")
		} else {
			ctx.WithField("apiErr", string(buf)).Debug("api call error json")
		}
	}

	httputils.WriteJSON(w, httpErr.Status(), httpErr)
	return nil
}

func getStatus(err error) int {
	if err == types.ErrMissingStorageService {
		return http.StatusInternalServerError
	}
	switch err.(type) {
	case *types.ErrBadAdminToken,
		*types.ErrSecTokInvalid:
		return http.StatusUnauthorized
	case *types.ErrNotFound:
		return http.StatusNotFound
	case *types.ErrMissingInstanceID,
		*types.ErrMissingLocalDevices:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func isLogAPICallErrJSON(ctx types.Context) bool {
	if types.Debug {
		return true
	}
	if lvl, ok := context.GetLogLevel(ctx); ok && lvl == log.DebugLevel {
		return true
	}
	return false
}
