package httputils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

var (
	serviceTypeName = utils.GetTypePkgPathAndName(
		(*types.StorageService)(nil))

	remoteStorageDriverTypeName = utils.GetTypePkgPathAndName(
		(*types.StorageDriver)(nil))
)

// GetService gets the Service instance from a types.
func GetService(ctx types.Context) (types.StorageService, error) {
	serviceObj := ctx.Value(types.CtxKeyService)
	if serviceObj == nil {
		return nil, utils.NewCtxKeyErr(types.CtxKeyService)
	}
	service, ok := serviceObj.(types.StorageService)
	if !ok {
		return nil, utils.NewContextTypeErr(
			types.CtxKeyService,
			serviceTypeName, utils.GetTypePkgPathAndName(serviceObj))
	}
	return service, nil
}

// GetInstanceIDForService gets the instance ID for a service using the
// context's instance IDs map.
func GetInstanceIDForService(
	ctx types.Context, service types.StorageService) *types.InstanceID {
	sm := ctx.InstanceIDsByService()
	if len(sm) == 0 {
		return nil
	}
	if val, ok := sm[strings.ToLower(service.Driver().Name())]; ok {
		return val
	}
	return nil
}

// GetLocalDevicesForService gets the local devices for a service using the
// context's local devices map.
func GetLocalDevicesForService(
	ctx types.Context, service types.StorageService) map[string]string {
	sm := ctx.LocalDevicesByService()
	if len(sm) == 0 {
		return nil
	}
	if val, ok := sm[strings.ToLower(service.Driver().Name())]; ok {
		return val
	}
	return nil
}

// GetServiceContext gets the service types.
func GetServiceContext(
	ctx *types.Context,
	service types.StorageService) error {

	sctx := *ctx

	if iid := GetInstanceIDForService(sctx, service); iid != nil {
		sctx = sctx.WithInstanceID(iid)
		sctx = sctx.WithContextSID(types.CtxKeyInstanceID, iid.ID)
	}

	localDevices := GetLocalDevicesForService(sctx, service)
	if localDevices != nil {
		sctx = sctx.WithLocalDevices(localDevices)
	}

	sctx = sctx.WithValue(types.CtxKeyService, service)
	sctx = sctx.WithValue(types.CtxKeyServiceName, service.Name())
	sctx = sctx.WithContextSID(types.CtxKeyService, service.Name())
	sctx = sctx.WithContextSID(types.CtxKeyDriver, service.Driver().Name())

	*ctx = sctx

	sctx.Debug("set service context")
	return nil
}

// GetStorageDriver gets the RemoteStorageDriver instance from a types.
func GetStorageDriver(ctx types.Context) (types.StorageDriver, error) {

	service, err := GetService(ctx)
	if err != nil {
		return nil, err
	}
	return service.Driver(), nil
}

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
