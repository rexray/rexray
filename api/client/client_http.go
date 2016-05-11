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

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/types"
)

func (c *client) httpDo(
	ctx types.Context,
	method, path string,
	payload, reply interface{}) (*http.Response, error) {

	reqBody, err := encPayload(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%s%s", c.host, path)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	ctx = context.RequireTX(ctx)
	tx := context.MustTransaction(ctx)
	req.Header.Set(types.TransactionHeader, tx.String())

	if iid, iidOK := context.InstanceID(ctx); iidOK {
		req.Header.Set(types.InstanceIDHeader, iid.String())
	} else if iidMap, ok := ctx.Value(
		context.AllInstanceIDsKey).(types.InstanceIDMap); ok {
		for _, v := range iidMap {
			req.Header.Add(types.InstanceIDHeader, v.String())
		}

	}

	if ld, ldOK := context.LocalDevices(ctx); ldOK {
		req.Header.Set(types.LocalDevicesHeader, ld.String())
	} else if ldMap, ok := ctx.Value(
		context.AllLocalDevicesKey).(types.LocalDevicesMap); ok {
		for _, v := range ldMap {
			req.Header.Add(types.LocalDevicesHeader, v.String())
		}
	}

	c.logRequest(req)

	res, err := ctxhttp.Do(ctx, &c.Client, req)
	if err != nil {
		return nil, err
	}
	defer c.setServerName(res)

	c.logResponse(res)

	if res.StatusCode > 299 {
		httpErr, err := goof.DecodeHTTPError(res.Body)
		if err != nil {
			return res, goof.WithField("status", res.StatusCode, "http error")
		}
		return res, httpErr
	}

	if req.Method != http.MethodHead && reply != nil {
		if err := decRes(res.Body, reply); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (c *client) setServerName(res *http.Response) {
	c.serverName = res.Header.Get(types.ServerNameHeader)
}

func (c *client) httpGet(
	ctx types.Context,
	path string,
	reply interface{}) (*http.Response, error) {

	return c.httpDo(ctx, "GET", path, nil, reply)
}

func (c *client) httpHead(
	ctx types.Context,
	path string) (*http.Response, error) {

	return c.httpDo(ctx, "HEAD", path, nil, nil)
}

func (c *client) httpPost(
	ctx types.Context,
	path string,
	payload interface{},
	reply interface{}) (*http.Response, error) {

	return c.httpDo(ctx, "POST", path, payload, reply)
}

func (c *client) httpDelete(
	ctx types.Context,
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
