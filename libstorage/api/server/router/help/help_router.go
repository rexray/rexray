package help

import (
	"fmt"
	"net/http"
	"os"

	"github.com/AVENTER-UG/rexray/core"
	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/httputils"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
)

func (r *router) helpInspect(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	proto := "http"
	if req.TLS != nil {
		proto = "https"
	}
	rootURL := fmt.Sprintf("%s://%s", proto, req.Host)

	reply := []string{
		fmt.Sprintf("%s/help/config", rootURL),
		fmt.Sprintf("%s/help/env", rootURL),
		fmt.Sprintf("%s/help/version", rootURL),
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) versionInspect(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	httputils.WriteJSON(w, http.StatusOK, core.SemVer)
	return nil
}

func (r *router) configInspect(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	expectedToken, ok := ctx.Value(context.AdminTokenKey).(string)
	if !ok {
		return utils.NewBadAdminTokenError("missing")
	}

	actualToken := store.GetString("admin")
	if expectedToken != actualToken {
		return utils.NewBadAdminTokenError(actualToken)
	}

	httputils.WriteJSON(w, http.StatusOK, r.config.AllSettings())
	return nil
}

func (r *router) envInspect(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	expectedToken, ok := ctx.Value(context.AdminTokenKey).(string)
	if !ok {
		return utils.NewBadAdminTokenError("missing")
	}

	actualToken := store.GetString("admin")
	if expectedToken != actualToken {
		return utils.NewBadAdminTokenError(actualToken)
	}

	httputils.WriteJSON(w, http.StatusOK, os.Environ())
	return nil
}
