package executor

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/emccode/libstorage/api/server/executors"
	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
)

func (r *router) executors(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	var reply apihttp.ExecutorsMap = map[string]*types.ExecutorInfo{}
	for ei := range executors.ExecutorInfos() {
		reply[ei.Name] = &ei.ExecutorInfo
	}

	return httputils.WriteJSON(w, http.StatusOK, reply)
}

func (r *router) executorInspect(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	ei, err := executors.ExecutorInfoInspect(store.GetString("executor"), true)
	if err != nil {
		return err
	}

	return writeFile(w, ei)
}

func (r *router) executorHead(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	ei, err := executors.ExecutorInfoInspect(store.GetString("executor"), false)
	if err != nil {
		return err
	}

	return writeFile(w, ei)
}

const lastModTimeFmt = "Mon, 02 Jan 2006 15:04:05 MST"

func writeFile(w http.ResponseWriter, ei *executors.ExecutorInfoEx) error {

	w.Header().Add("Accept-Ranges", "bytes")
	w.Header().Add("Content-Length", fmt.Sprintf("%d", ei.Size))
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Last-Modified",
		time.Unix(ei.LastModified, 0).Format(lastModTimeFmt))

	hexBuf, _ := hex.DecodeString(ei.MD5Checksum)
	b64str := base64.StdEncoding.EncodeToString(hexBuf)
	w.Header().Add("Content-MD5", b64str)

	if len(ei.Data) > 0 {
		if _, err := io.Copy(w, bytes.NewReader(ei.Data)); err != nil {
			return err
		}
	}

	return nil
}
