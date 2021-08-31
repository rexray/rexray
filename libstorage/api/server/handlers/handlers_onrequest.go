package handlers

import (
	"net/http"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

// OnRequest is a handler to which an external provider can attach that is
// invoked for every incoming HTTP request.
var OnRequest types.APIFunc

type onRequestHandler struct {
	handler types.APIFunc
}

// NewOnRequestHandler is a handler.
func NewOnRequestHandler() types.Middleware {
	return &onRequestHandler{}
}

func (h *onRequestHandler) Name() string {
	return "transaction-handler"
}

func (h *onRequestHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&onRequestHandler{m}).Handle
}

// Handle is the type's Handler function.
func (h *onRequestHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	if OnRequest != nil {
		if err := OnRequest(ctx, w, req, store); err != nil {
			return err
		}
	}

	return h.handler(ctx, w, req, store)
}
