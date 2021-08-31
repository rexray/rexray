// +build !agent
// +build !controller

package scripts

import (
	"io"
	"net/http"

	"github.com/akutz/goof"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
)

// GetHTTP retrieves a URL.
func GetHTTP(ctx apitypes.Context, url string) (io.ReadCloser, error) {
	ctx.WithField("url", url).Debug("getting http url")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 299 {
		return nil, goof.WithFields(goof.Fields{
			"status": res.StatusCode,
			"url":    url,
		}, "http error getting url")
	}

	return res.Body, nil
}
