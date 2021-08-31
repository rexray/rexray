// +build !agent
// +build !controller

package scripts

import (
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/akutz/goof"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
)

const (
	rawGitHubURL = "https://raw.githubusercontent.com"
)

// GetGitHubBlob retrieves a blob from GitHub.
func GetGitHubBlob(
	ctx apitypes.Context,
	user, repo, commit, name string) (io.ReadCloser, error) {

	if user == "" {
		user = "thecodeteam"
	}
	if repo == "" {
		repo = "rexray"
	}
	if commit == "" {
		commit = "master"
	}
	name = path.Join("scripts", "scripts", name)

	url := fmt.Sprintf(
		"%s/%s/%s/%s/%s",
		rawGitHubURL,
		user,
		repo,
		commit,
		name)

	ctx.WithField("url", url).Debug("getting github blob")

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
		}, "http error getting github blob")
	}

	return res.Body, nil
}
