package service

import (
	"net/http"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
)

func (r *router) servicesList(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	reply := map[string]*types.ServiceInfo{}
	for service := range services.StorageServices(ctx) {
		si := toServiceInfo(ctx, service)
		reply[si.Name] = si
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) serviceInspect(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}
	httputils.WriteJSON(w, http.StatusOK, toServiceInfo(ctx, service))
	return nil
}

func toServiceInfo(
	ctx types.Context,
	service types.StorageService) *types.ServiceInfo {

	d := service.Driver()
	dn := service.Driver().Name()

	return &types.ServiceInfo{
		Name: service.Name(),
		Driver: &types.DriverInfo{
			Name:       dn,
			Type:       d.Type(ctx),
			NextDevice: d.NextDeviceInfo(ctx),
		},
	}
}
