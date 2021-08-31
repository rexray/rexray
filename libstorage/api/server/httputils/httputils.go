package httputils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	gofig "github.com/akutz/gofig/types"

	"github.com/AVENTER-UG/rexray/libstorage/api/server/services"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
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
	config gofig.Config,
	w http.ResponseWriter,
	store types.Store,
	task *types.Task,
	okStatus int) error {

	if store.GetBool("async") {
		WriteJSON(w, http.StatusAccepted, task)
		return nil
	}

	exeTimeoutDur, err := time.ParseDuration(
		config.GetString(types.ConfigServerTasksExeTimeout))
	if err != nil {
		exeTimeoutDur = time.Duration(time.Second * 60)
	}
	exeTimeout := time.NewTimer(exeTimeoutDur)

	select {
	case <-services.TaskWaitC(ctx, task.ID):
		if task.Error != nil {
			return task.Error
		}
		WriteJSON(w, okStatus, task.Result)
	case <-exeTimeout.C:
		WriteJSON(w, http.StatusRequestTimeout, task)
	}

	return nil
}
