package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/akutz/goof"
	"golang.org/x/net/context/ctxhttp"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/api/utils"
)

func (c *Client) httpDo(
	ctx context.Context,
	method, path string,
	payload, reply interface{}) (*http.Response, error) {

	reqBody, err := encPayload(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%s%s", c.Host, path)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	for k, v := range c.Headers {
		req.Header[k] = v
	}

	c.logRequest(req)

	res, err := ctxhttp.Do(ctx, c.Client, req)
	if err != nil {
		return nil, err
	}
	defer c.setServerName(res)

	c.logResponse(res)

	if res.StatusCode > 299 {
		je := &types.JSONError{}
		if err := json.NewDecoder(res.Body).Decode(je); err != nil {
			return res, goof.WithField("status", res.StatusCode, "http error")
		}
		return res, je
	}

	if req.Method != http.MethodHead && reply != nil {
		if err := decRes(res.Body, reply); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (c *Client) setServerName(res *http.Response) {
	snh := utils.GetHeader(res.Header, apihttp.ServerNameHeader)
	if len(snh) > 0 {
		c.ServerName = snh[0]
	}
}

func (c *Client) httpGet(
	ctx context.Context,
	path string,
	reply interface{}) (*http.Response, error) {

	return c.httpDo(ctx, "GET", path, nil, reply)
}

func (c *Client) httpHead(
	ctx context.Context,
	path string) (*http.Response, error) {

	return c.httpDo(ctx, "HEAD", path, nil, nil)
}

func (c *Client) httpPost(
	ctx context.Context,
	path string,
	payload interface{},
	reply interface{}) (*http.Response, error) {

	return c.httpDo(ctx, "POST", path, payload, reply)
}

func (c *Client) httpDelete(
	ctx context.Context,
	path string,
	reply interface{}) (*http.Response, error) {

	return c.httpDo(ctx, "DELETE", path, nil, reply)
}

func encPayload(payload interface{}) (io.Reader, error) {
	if payload == nil {
		return nil, nil
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf), nil
}

func decRes(body io.Reader, reply interface{}) error {
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(buf, reply); err != nil {
		return err
	}
	return nil
}
