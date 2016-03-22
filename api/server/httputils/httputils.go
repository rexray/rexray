package httputils

import (
	"encoding/json"
	"net/http"

	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
	"github.com/emccode/libstorage/api/utils"
)

var (
	serviceTypeName       = utils.GetTypePkgPathAndName((*Service)(nil))
	storageDriverTypeName = utils.GetTypePkgPathAndName((*drivers.StorageDriver)(nil))
)

// GetService gets the Service instance from a context.
func GetService(ctx context.Context) (Service, error) {
	serviceObj := ctx.Value("service")
	if serviceObj == nil {
		return nil, utils.NewContextKeyErr("service")
	}
	service, ok := serviceObj.(Service)
	if !ok {
		return nil, utils.NewContextTypeErr(
			"service", serviceTypeName, utils.GetTypePkgPathAndName(serviceObj))
	}
	return service, nil
}

// GetStorageDriver gets the StorageDriver instance from a context.
func GetStorageDriver(ctx context.Context) (drivers.StorageDriver, error) {
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
