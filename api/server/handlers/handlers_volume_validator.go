package handlers

import (
	"net/http"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
	"github.com/emccode/libstorage/api/utils"
)

// volumeValidator is an HTTP filter for validating that the volume
// specified as part of the path is valid.
type volumeValidator struct {
	handler httputils.APIFunc
}

// NewVolumeValidator returns a new filter for validating that the volume
// specified as part of the path is valid.
func NewVolumeValidator() httputils.Middleware {
	return &volumeValidator{}
}

func (h *volumeValidator) Name() string {
	return "volume-validator"
}

func (h *volumeValidator) Handler(m httputils.APIFunc) httputils.APIFunc {
	h.handler = m
	return h.Handle
}

// Handle is the type's Handler function.
func (h *volumeValidator) Handle(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	if !store.IsSet("volumeID") {
		return utils.NewStoreKeyErr("volumeID")
	}

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	volumeID := store.GetString("volumeID")

	volume, err := service.Driver().VolumeInspect(
		ctx,
		volumeID,
		&drivers.VolumeInspectOpts{
			Attachments: store.GetBool("attachments"),
			Opts:        store,
		})
	if err != nil {
		return err
	}

	ctx = ctx.WithValue("volume", volume)
	ctx = ctx.WithContextID("volume", volume.ID)
	ctx.Log().Debug("set volume context")

	return h.handler(ctx, w, req, store)
}
