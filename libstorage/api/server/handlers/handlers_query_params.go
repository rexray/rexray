package handlers

import (
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

// queryParamsHandler is an HTTP filter for injecting the store with query
// parameters
type queryParamsHandler struct {
	handler types.APIFunc
}

func (h *queryParamsHandler) Name() string {
	return "query-params-handler"
}

// NewQueryParamsHandler returns a new filter for injecting the store with query
// parameters
func NewQueryParamsHandler() types.Middleware {
	return &queryParamsHandler{}
}

func (h *queryParamsHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&queryParamsHandler{m}).Handle
}

// Handle is the type's Handler function.
func (h *queryParamsHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	for k, v := range req.URL.Query() {
		ctx.WithFields(log.Fields{
			"key":        k,
			"value":      v,
			"len(value)": len(v),
		}).Debug("query param")
		switch len(v) {
		case 0:
			store.Set(k, true)
		case 1:
			if len(v[0]) == 0 {
				store.Set(k, true)
			} else {
				if i, err := strconv.ParseInt(v[0], 10, 64); err == nil {
					store.Set(k, i)
				} else if b, err := strconv.ParseBool(v[0]); err == nil {
					store.Set(k, b)
				} else {
					store.Set(k, v[0])
				}
			}
		default:
			store.Set(k, v)
		}
	}
	return h.handler(ctx, w, req, store)
}
