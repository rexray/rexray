package root

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/AVENTER-UG/rexray/libstorage/api/server/httputils"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

var (
	rootURLRx = regexp.MustCompile(`^(.+)/[^/]*$`)
)

func (r *router) root(
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
		fmt.Sprintf("%s/services", rootURL),
		fmt.Sprintf("%s/snapshots", rootURL),
		fmt.Sprintf("%s/tasks", rootURL),
		fmt.Sprintf("%s/help", rootURL),
		fmt.Sprintf("%s/volumes", rootURL),
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}
