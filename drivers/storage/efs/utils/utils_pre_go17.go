// +build !go1.7

package utils

import (
	"net/http"

	"golang.org/x/net/context/ctxhttp"

	"github.com/emccode/libstorage/api/types"
)

func doRequest(ctx types.Context, req *http.Request) (*http.Response, error) {
	return doRequestWithClient(ctx, http.DefaultClient, req)
}

func doRequestWithClient(
	ctx types.Context,
	client *http.Client,
	req *http.Request) (*http.Response, error) {
	return ctxhttp.Do(ctx, client, req)
}
