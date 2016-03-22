package handlers

import (
	"net/http"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/utils"
)

// snapshotValidator is an HTTP filter for validating that the snapshot
// specified as part of the path is valid.
type snapshotValidator struct {
	handler httputils.APIFunc
}

// NewSnapshotValidator returns a new filter for validating that the snapshot
// specified as part of the path is valid.
func NewSnapshotValidator() httputils.Middleware {
	return &snapshotValidator{}
}

func (h *snapshotValidator) Name() string {
	return "snapshot-validator"
}

func (h *snapshotValidator) Handler(m httputils.APIFunc) httputils.APIFunc {
	h.handler = m
	return h.Handle
}

// Handle is the type's Handler function.
func (h *snapshotValidator) Handle(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	if !store.IsSet("snapshotID") {
		return utils.NewStoreKeyErr("snapshotID")
	}

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	snapshotID := store.GetString("snapshotID")

	snapshot, err := service.Driver().SnapshotInspect(
		ctx,
		snapshotID,
		store)
	if err != nil {
		return err
	}

	ctx = ctx.WithValue("snapshot", snapshot)
	ctx = ctx.WithContextID("snapshot", snapshot.ID)
	ctx.Log().Debug("set snapshot context")

	return h.handler(ctx, w, req, store)
}
