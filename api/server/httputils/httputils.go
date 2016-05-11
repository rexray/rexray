package httputils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
)

// WriteJSON writes the value v to the http response stream as json with
// standard json encoding.
func WriteJSON(w http.ResponseWriter, code int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
	//return json.NewEncoder(w).Encode(v)
}

// WriteData writes the value v to the http response stream as binary.
func WriteData(w http.ResponseWriter, code int, v []byte) error {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(code)
	if _, err := w.Write(v); err != nil {
		return err
	}
	return nil
}

// WriteResponse writes a recorded response to a ResponseWriter.
func WriteResponse(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
	w.WriteHeader(rec.Code)
	for k, v := range rec.HeaderMap {
		w.Header()[k] = v
	}
	w.Write(rec.Body.Bytes())
}

// WriteTask writes a task to a ResponseWriter.
func WriteTask(
	ctx types.Context,
	w http.ResponseWriter,
	store types.Store,
	task *types.Task,
	okStatus int) error {

	if store.GetBool("async") {
		WriteJSON(w, http.StatusAccepted, task)
		return nil
	}

	timeout := time.NewTimer(time.Second * 60)

	select {
	case <-services.TaskWaitC(ctx, task.ID):
		if task.Error != nil {
			return task.Error
		}
		WriteJSON(w, okStatus, task.Result)
	case <-timeout.C:
		WriteJSON(w, http.StatusRequestTimeout, task)
	}

	return nil
}
