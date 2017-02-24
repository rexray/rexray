// +build go1.7
// +build !libstorage_storage_driver libstorage_storage_driver_fittedcloud

package utils

import (
	"net/http"

	"github.com/codedellemc/libstorage/api/types"
)

func doRequest(ctx types.Context, req *http.Request) (*http.Response, error) {
	return doRequestWithClient(ctx, http.DefaultClient, req)
}

func doRequestWithClient(
	ctx types.Context,
	client *http.Client,
	req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return client.Do(req)
}
