package handlers

import (
	"net/http"

	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
)

// instanceIDValidator is a global HTTP filter for validating that the
// InstanceID for a context is present when it's required
type instanceIDValidator struct {
	handler  apihttp.APIFunc
	required bool
}

// NewInstanceIDValidator returns a new global HTTP filter for validating that
// the InstanceID for a context is present when it's required
func NewInstanceIDValidator(required bool) apihttp.Middleware {
	return &instanceIDValidator{required: required}
}

func (h *instanceIDValidator) Name() string {
	return "instanceID-validator"
}

func (h *instanceIDValidator) Handler(m apihttp.APIFunc) apihttp.APIFunc {
	return (&instanceIDValidator{m, h.required}).Handle
}

// Handle is the type's Handler function.
func (h *instanceIDValidator) Handle(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	iid := ctx.InstanceID()

	if h.required && iid == nil {
		return goof.New("instanceID required")
	}

	if store.GetBool("attachments") && iid == nil {
		return goof.New("cannot get attachments without instance ID")
	}

	return h.handler(ctx, w, req, store)
}
