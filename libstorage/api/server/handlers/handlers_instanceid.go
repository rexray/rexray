package handlers

import (
	"net/http"
	"strings"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

// instanceIDHandler is a global HTTP filter for grokking the InstanceIDs
// from the headers
type instanceIDHandler struct {
	handler types.APIFunc
	s2d     map[string]string
}

// NewInstanceIDHandler returns a new global HTTP filter for grokking the
// InstanceIDs from the headers
func NewInstanceIDHandler(svcs <-chan types.StorageService) types.Middleware {
	iidh := &instanceIDHandler{
		s2d: map[string]string{},
	}
	for s := range svcs {
		iidh.s2d[strings.ToLower(s.Name())] = strings.ToLower(s.Driver().Name())
	}
	return iidh
}

func (h *instanceIDHandler) Name() string {
	return "instanceIDs-handler"
}

func (h *instanceIDHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&instanceIDHandler{m, h.s2d}).Handle
}

// Handle is the type's Handler function.
func (h *instanceIDHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	headers := req.Header[types.InstanceIDHeader]
	ctx.WithField(types.InstanceIDHeader, headers).Debug("http header")

	// this function has been updated to account for
	// https://github.com/AVENTER-UG/libstorage/pull/420 and
	// https://github.com/AVENTER-UG/rexray/issues/685.
	//
	// this handler now inspects each instance ID header and stores the
	// unmarshaled object in one of two maps -- a map keyed by the
	// service name and a map keyed by the driver name. the service
	// name-keyed map is only used if the instance ID has a service
	// name present.
	//
	// after the two maps are populated, all of the server's configured
	// services are iterated. first we check the service name-keyed map
	// for a service's instance ID, and only then if not present we
	// check the driver name-keyed map for the service's instance ID.

	valMap := types.InstanceIDMap{}
	d2i := map[string]*types.InstanceID{}
	s2i := map[string]*types.InstanceID{}

	for _, h := range headers {
		val := &types.InstanceID{}
		if err := val.UnmarshalText([]byte(h)); err != nil {
			return err
		}
		if len(val.Service) > 0 {
			s2i[strings.ToLower(val.Service)] = val
		} else {
			d2i[strings.ToLower(val.Driver)] = val
		}
	}

	for s, d := range h.s2d {
		if iid, ok := s2i[s]; ok {
			valMap[s] = iid
		} else if iid, ok := d2i[d]; ok {
			valMap[s] = iid
		}
	}

	ctx = ctx.WithValue(context.AllInstanceIDsKey, valMap)
	return h.handler(ctx, w, req, store)
}
