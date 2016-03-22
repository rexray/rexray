// +build ignore

package driver

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"

	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
	httptypes "github.com/emccode/libstorage/api/types/http"
)

func (r *driverRouter) driversList(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	var reply httptypes.DriversResponse = map[string]*types.DriverInfo{}

	for _, d := range r.drivers {
		reply[d.Name()] = &types.DriverInfo{
			Name:       d.Name(),
			Type:       d.Type(),
			NextDevice: d.NextDevice(),
			Executors:  getExecutors(ctx, d),
		}
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *driverRouter) driverInspect(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	dname := store.GetString("driver")

	d, ok := r.drivers[strings.ToLower(dname)]
	if !ok {
		return goof.New("invalid driver")
	}

	reply := &types.DriverInfo{
		Name:       d.Name(),
		Type:       d.Type(),
		NextDevice: d.NextDevice(),
		Executors:  getExecutors(ctx, d),
	}

	return httputils.WriteJSON(w, http.StatusOK, reply)
}

func getExecutors(
	ctx context.Context, d drivers.StorageDriver) []*types.ExecutorInfo {

	executors := d.Executors()
	hashExecutors(ctx, d, executors)
	return executors
}

func (r *driverRouter) executorsList(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	dname := store.GetString("driver")

	d, ok := r.drivers[strings.ToLower(dname)]
	if !ok {
		return goof.New("invalid driver")
	}

	var reply types.ExecutorsListResponse = getExecutors(ctx, d)

	return httputils.WriteJSON(w, http.StatusOK, reply)
}

func (r *driverRouter) executorDownload(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	dname := store.GetString("driver")

	d, ok := r.drivers[strings.ToLower(dname)]
	if !ok {
		return goof.New("invalid driver")
	}

	executor, err := getExecutor(ctx, d, store.GetString("name"))
	if err != nil {
		return err
	}

	return httputils.WriteData(w, http.StatusOK, executor)
}

func getExecutor(
	ctx context.Context,
	d drivers.StorageDriver,
	goos, goarch string,
	store types.Store) ([]byte, error) {

	tool, err := d.ExecutorInspect(ctx, goos, goarch, store)
	if err != nil {
		return nil, err
	}

	return tool, nil
}

func hashExecutors(
	ctx context.Context,
	d drivers.StorageDriver,
	tinfos []*types.ExecutorInfo) error {

	for _, ti := range tinfos {
		if err := hashExecutor(ctx, d, ti); err != nil {
			return err
		}
	}
	return nil
}

func hashExecutor(
	ctx context.Context,
	d drivers.StorageDriver,
	tinfo *types.ExecutorInfo) error {

	if tinfo.MD5Checksum == "" {
		return nil
	}

	tool, err := getExecutor(ctx, d, tinfo.Name)
	if err != nil {
		return err
	}

	tinfo.MD5Checksum = fmt.Sprintf("%x", md5.Sum(tool))
	return nil
}
