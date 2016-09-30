// +build go1.7

package utils

import (
	"net/http"

	"github.com/emccode/libstorage/api/types"
)

func doRequest(ctx types.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return http.DefaultClient.Do(req)
}
