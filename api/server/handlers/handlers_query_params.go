package handlers

import (
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
)

// queryParamsHandler is an HTTP filter for injecting the store with query
// parameters
type queryParamsHandler struct {
	handler apihttp.APIFunc
}

func (h *queryParamsHandler) Name() string {
	return "query-params-handler"
}

// NewQueryParamsHandler returns a new filter for injecting the store with query
// parameters
func NewQueryParamsHandler() apihttp.Middleware {
	return &queryParamsHandler{}
}

func (h *queryParamsHandler) Handler(m apihttp.APIFunc) apihttp.APIFunc {
	return (&queryParamsHandler{m}).Handle
}

// Handle is the type's Handler function.
func (h *queryParamsHandler) Handle(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	for k, v := range req.URL.Query() {
		ctx.Log().WithFields(log.Fields{
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
				if b, err := strconv.ParseBool(v[0]); err != nil {
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
