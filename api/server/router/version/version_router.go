package version

import (
	"net/http"

	"github.com/emccode/libstorage/api"
	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
)

func (r *router) versionInspect(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	httputils.WriteJSON(w, http.StatusOK, api.Version)
	return nil
}
