package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/akutz/goof"
	"golang.org/x/net/context/ctxhttp"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

func (c *client) httpDo(
	ctx types.Context,
	method, path string,
	payload, reply interface{}) (*http.Response, error) {

	txID := ctx.TransactionID()
	if txID == "" {
		txIDUUID, _ := utils.NewUUID()
		txID = txIDUUID.String()
		ctx = ctx.WithTransactionID(txID).WithContextSID(
			types.ContextTransactionID, txID)
	}
	txCR := ctx.TransactionCreated()
	if txCR.IsZero() {
		txCR = time.Now().UTC()
		ctx = ctx.WithTransactionCreated(txCR).WithContextSID(
			types.ContextTransactionCreated,
			fmt.Sprintf("%d", txCR.Unix()))
	}

	reqBody, err := encPayload(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%s%s", c.host, path)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	liidh := strings.ToLower(types.InstanceIDHeader)
	liid64h := strings.ToLower(types.InstanceID64Header)
	ldh := strings.ToLower(types.LocalDevicesHeader)

	for k, v := range c.headers {
		lk := strings.ToLower(k)
		switch lk {
		case liidh, liid64h:
			if c.enableInstanceIDHeaders {
				req.Header[k] = v
			}
		case ldh:
			if c.enableLocalDevicesHeaders {
				req.Header[k] = v
			}
		default:
			req.Header[k] = v
		}

	}

	req.Header.Set(types.TransactionIDHeader, txID)
	req.Header.Set(
		types.TransactionCreatedHeader, fmt.Sprintf("%d", txCR.Unix()))
	c.logRequest(req)

	res, err := ctxhttp.Do(ctx, &c.Client, req)
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
