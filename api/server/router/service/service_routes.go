package service

import (
	"net/http"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

func (r *router) servicesList(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	reply := map[string]*types.ServiceInfo{}
	for service := range services.StorageServices(ctx) {
		si, err := toServiceInfo(ctx, service, store)
		if err != nil {
			return err
		}
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

	si, err := toServiceInfo(ctx, service, store)
	if err != nil {
		return err
	}

	httputils.WriteJSON(w, http.StatusOK, si)
	return nil
}

func toServiceInfo(
	ctx types.Context,
	service types.StorageService,
	store types.Store) (*types.ServiceInfo, error) {

	d := service.Driver()

	var instance *types.Instance
	if store.GetBool("instance") {
		if ctx.InstanceID() == nil {
			return nil, utils.NewMissingInstanceIDError(service.Name())
		}
		var err error
		instance, err = d.InstanceInspect(ctx, store)
		if err != nil {
			return nil, err
		}
	}

	st, err := d.Type(ctx)
	if err != nil {
		return nil, err
	}
	nd, err := d.NextDeviceInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &types.ServiceInfo{
		Name:     service.Name(),
		Instance: instance,
		Driver: &types.DriverInfo{
			Name:       d.Name(),
			Type:       st,
			NextDevice: nd,
		},
	}, nil
}
