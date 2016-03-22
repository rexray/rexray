package service

import (
	"net/http"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	httptypes "github.com/emccode/libstorage/api/types/http"
)

func (r *router) servicesList(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	var reply httptypes.ServicesResponse = map[string]*types.ServiceInfo{}
	for _, service := range r.services {
		si := toServiceInfo(service)
		reply[si.Name] = si
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) serviceInspect(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}
	reply := toServiceInfo(service)
	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) executorsList(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	//driverName := service.Driver().Name()

	reply := toServiceInfo(service)
	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func toServiceInfo(service httputils.Service) *types.ServiceInfo {
	return &types.ServiceInfo{
		Name: service.Name(),
		Driver: &types.DriverInfo{
			Name:       service.Driver().Name(),
			Type:       service.Driver().Type(),
			NextDevice: service.Driver().NextDeviceInfo(),
		},
	}
}
