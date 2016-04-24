package httputils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
	apisvcs "github.com/emccode/libstorage/api/types/services"
	"github.com/emccode/libstorage/api/utils"
)

var (
	serviceTypeName = utils.GetTypePkgPathAndName(
		(*apisvcs.StorageService)(nil))

	remoteStorageDriverTypeName = utils.GetTypePkgPathAndName(
		(*drivers.RemoteStorageDriver)(nil))
)

// GetHeader is a case-insensitive way to retrieve a header's value.
func GetHeader(headers http.Header, name string) []string {
	for k, v := range headers {
		if strings.ToLower(k) == strings.ToLower(name) {
			return v
		}
	}
	return nil
}

// GetService gets the Service instance from a context.
func GetService(ctx context.Context) (apisvcs.StorageService, error) {
	serviceObj := ctx.Value(context.ContextKeyService)
	if serviceObj == nil {
		return nil, utils.NewContextKeyErr(context.ContextKeyService)
	}
	service, ok := serviceObj.(apisvcs.StorageService)
	if !ok {
		return nil, utils.NewContextTypeErr(
			context.ContextKeyService,
			serviceTypeName, utils.GetTypePkgPathAndName(serviceObj))
	}
	return service, nil
}

// GetInstanceIDForService gets the instance ID for a service using the
// context's instance IDs map.
func GetInstanceIDForService(
	ctx context.Context, service apisvcs.StorageService) *types.InstanceID {
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
	ctx context.Context, service apisvcs.StorageService) map[string]string {
	sm := ctx.LocalDevicesByService()
	if len(sm) == 0 {
		return nil
	}
	if val, ok := sm[strings.ToLower(service.Driver().Name())]; ok {
		return val
	}
	return nil
}

// GetServiceContext gets the service context.
func GetServiceContext(
	ctx *context.Context,
	service apisvcs.StorageService) error {

	iid := GetInstanceIDForService(*ctx, service)
	if iid == nil {
		return utils.NewMissingInstanceIDError(service.Name())
	}

	sctx := context.WithInstanceID(*ctx, iid)
	sctx = sctx.WithContextID(context.ContextKeyInstanceID, iid.ID)

	localDevices := GetLocalDevicesForService(sctx, service)
	if localDevices != nil {
		sctx = sctx.WithLocalDevices(localDevices)
	}

	sctx = sctx.WithValue(context.ContextKeyService, service)
	sctx = sctx.WithValue(context.ContextKeyServiceName, service.Name())
	sctx = sctx.WithContextID(context.ContextKeyService, service.Name())
	sctx = sctx.WithContextID(context.ContextKeyDriver, service.Driver().Name())

	*ctx = sctx
	return nil
}

// GetStorageDriver gets the RemoteStorageDriver instance from a context.
func GetStorageDriver(
	ctx context.Context) (drivers.RemoteStorageDriver, error) {

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
	ctx context.Context,
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
