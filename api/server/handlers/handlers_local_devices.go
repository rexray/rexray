package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/types"
)

// localDevicesHandler is a global HTTP filter for grokking the local devices
// from the headers
type localDevicesHandler struct {
	handler types.APIFunc
}

// NewLocalDevicesHandler returns a new global HTTP filter for grokking the
// local devices from the headers
func NewLocalDevicesHandler() types.Middleware {
	return &localDevicesHandler{}
}

func (h *localDevicesHandler) Name() string {
	return "local-devices-handler"
}

func (h *localDevicesHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&localDevicesHandler{m}).Handle
}

var (
	locDevRX = regexp.MustCompile(`^(.+?)=((?:(?:.+?)=(?:.+?)(?:,\s*)?)+)$`)
)

// Handle is the type's Handler function.
func (h *localDevicesHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	headers := req.Header[types.LocalDevicesHeader]
	ctx.WithField(types.LocalDevicesHeader, headers).Debug("http header")

	valMap := types.LocalDevicesMap{}
	for _, h := range headers {
		val := &types.LocalDevices{}
		if err := val.UnmarshalText([]byte(h)); err != nil {
			return err
		}
		valMap[strings.ToLower(val.Driver)] = val
	}

	ctx = ctx.WithValue(context.AllLocalDevicesKey, valMap)
	return h.handler(ctx, w, req, store)
}
