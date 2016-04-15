package service

import (
	"net/http"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
	apisvcs "github.com/emccode/libstorage/api/types/services"
)

func (r *router) servicesList(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	var reply apihttp.ServicesResponse = map[string]*types.ServiceInfo{}
	for service := range services.StorageServices() {
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
	httputils.WriteJSON(w, http.StatusOK, toServiceInfo(service))
	return nil
}

func toServiceInfo(service apisvcs.StorageService) *types.ServiceInfo {
	d := service.Driver()
	dn := service.Driver().Name()

	return &types.ServiceInfo{
		Name: service.Name(),
		Driver: &types.DriverInfo{
			Name:       dn,
			Type:       d.Type(),
			NextDevice: d.NextDeviceInfo(),
		},
	}
}
