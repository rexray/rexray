package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
)

const (
	libStorageInstanceIDHeaderKey = "libstorage-instanceid"
)

// instanceIDHandler is a global HTTP filter for grokking the InstanceIDs
// from the headers
type instanceIDHandler struct {
	handler httputils.APIFunc
}

// NewInstanceIDHandler returns a new global HTTP filter for grokking the
// InstanceIDs from the headers
func NewInstanceIDHandler() httputils.Middleware {
	return &instanceIDHandler{}
}

func (h *instanceIDHandler) Name() string {
	return "instanceIDs-handler"
}

func (h *instanceIDHandler) Handler(m httputils.APIFunc) httputils.APIFunc {
	h.handler = m
	return h.Handle
}

// Handle is the type's Handler function.
func (h *instanceIDHandler) Handle(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	var iidHeader string
	for k, v := range req.Header {
		if strings.ToLower(k) == libStorageInstanceIDHeaderKey {
			iidHeader = v[0]
			break
		}
	}

	ctx.Log().WithField(
		libStorageInstanceIDHeaderKey, iidHeader).Debug("http header")

	var iidPairs []string
	if len(iidHeader) > 0 {
		iidPairs = strings.Split(iidHeader, ",")
	}

	iidMap := map[string]*types.InstanceID{}

	for _, iidPair := range iidPairs {
		iidParts := strings.Split(iidPair, ":")
		iidService := iidParts[0]
		iidBase64 := iidParts[1]

		iid := &types.InstanceID{}
		decoded, err := base64.StdEncoding.DecodeString(iidBase64)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(decoded, iid); err != nil {
			return err
		}

		iidMap[strings.ToLower(iidService)] = iid
	}

	if len(iidMap) > 0 {
		ctx = ctx.WithValue("instanceIDs", iidMap)
	}

	return h.handler(ctx, w, req, store)
}
