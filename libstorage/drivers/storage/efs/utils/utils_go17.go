// +build go1.7

package utils

import (
	"net/http"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
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
