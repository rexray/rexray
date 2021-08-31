// +build !agent
// +build !controller

package scripts

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/akutz/goof"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
)

const (
	gistAPIUrl = "https://api.github.com/gists"
)

type gist struct {
	URL   string               `json:"url"`
	ID    string               `json:"id"`
	Files map[string]*gistFile `json:"files,omitempty"`
}

type gistFile struct {
	FileName string `json:"filename"`
	RawURL   string `json:"raw_url"`
}

func (gf *gistFile) getRaw(ctx apitypes.Context) (io.ReadCloser, error) {
	ctx.WithField("url", gf.RawURL).Debug("getting raw gist")
	req, err := http.NewRequest(http.MethodGet, gf.RawURL, nil)
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
			"url":    gf.RawURL,
		}, "http error getting raw gist")
	}
	return res.Body, nil
}

// GetGist retrieves a gist.
func GetGist(
	ctx apitypes.Context, id, fileName string) (string, io.ReadCloser, error) {

	url := fmt.Sprintf("%s/%s", gistAPIUrl, id)
	ctx.WithField("url", url).Debug("getting gist")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", nil, err
	}

	res, err := doRequest(ctx, req)
	if err != nil {
		return "", nil, err
	}

	if res.StatusCode > 299 {
		return "", nil, goof.WithFields(goof.Fields{
			"status": res.StatusCode,
			"url":    url,
		}, "http error getting gist info")
	}

	defer res.Body.Close()

	g := gist{}
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&g); err != nil {
		return "", nil, err
	}

	if fileName == "" {
		for _, v := range g.Files {
			rdr, err := v.getRaw(ctx)
			return v.FileName, rdr, err
		}
	}

	for _, v := range g.Files {
		if strings.EqualFold(fileName, v.FileName) {
			rdr, err := v.getRaw(ctx)
			return v.FileName, rdr, err
		}
	}

	return "", nil, goof.WithFields(goof.Fields{
		"id":       id,
		"fileName": fileName,
	}, "no gist found")
}
