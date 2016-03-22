package root

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	httptypes "github.com/emccode/libstorage/api/types/http"
)

var (
	rootURLRx = regexp.MustCompile(`^(.+)/[^/]*$`)
)

func (r *router) root(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	proto := "http"
	if req.TLS != nil {
		proto = "https"
	}
	rootURL := fmt.Sprintf("%s://%s", proto, req.Host)

	var reply httptypes.RootResponse = []string{
		fmt.Sprintf("%s/services", rootURL),
		fmt.Sprintf("%s/snapshots", rootURL),
		fmt.Sprintf("%s/volumes", rootURL),
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}
