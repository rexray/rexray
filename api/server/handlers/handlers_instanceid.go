package handlers

import (
	"net/http"
	"strings"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/types"
)

// instanceIDHandler is a global HTTP filter for grokking the InstanceIDs
// from the headers
type instanceIDHandler struct {
	handler types.APIFunc
}

// NewInstanceIDHandler returns a new global HTTP filter for grokking the
// InstanceIDs from the headers
func NewInstanceIDHandler() types.Middleware {
	return &instanceIDHandler{}
}

func (h *instanceIDHandler) Name() string {
	return "instanceIDs-handler"
}

func (h *instanceIDHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&instanceIDHandler{m}).Handle
}

// Handle is the type's Handler function.
func (h *instanceIDHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	headers := req.Header[types.InstanceIDHeader]
	ctx.WithField(types.InstanceIDHeader, headers).Debug("http header")

	valMap := types.InstanceIDMap{}
	for _, h := range headers {
		val := &types.InstanceID{}
		if err := val.UnmarshalText([]byte(h)); err != nil {
			return err
		}
		valMap[strings.ToLower(val.Driver)] = val
	}

	ctx = ctx.WithValue(context.AllInstanceIDsKey, valMap)
	return h.handler(ctx, w, req, store)
}
