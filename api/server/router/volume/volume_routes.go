package volume

import (
	"net/http"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
	httptypes "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/api/utils"
)

func newVolumesRoute(
	services map[string]httputils.Service,
	queryAttachments bool) *volumesRoute {
	return &volumesRoute{services, queryAttachments}
}

type volumesRoute struct {
	services         map[string]httputils.Service
	queryAttachments bool
}

func (r *volumesRoute) volumes(ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	attachments := false
	if r.queryAttachments {
		attachments = store.GetBool("attachments")
	}

	var reply httptypes.ServiceVolumeMap = map[string]httptypes.VolumeMap{}

	for _, service := range r.services {

		volumes, err := service.Driver().Volumes(
			ctx,
			&drivers.VolumesOpts{
				Attachments: attachments,
				Opts:        store,
			})
		if err != nil {
			return utils.NewBatchProcessErr(reply, err)
		}

		volumeMap := map[string]*types.Volume{}
		reply[service.Name()] = volumeMap
		for _, v := range volumes {
			volumeMap[v.ID] = v
		}
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *volumesRoute) volumesForService(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	d, err := httputils.GetStorageDriver(ctx)
	if err != nil {
		return err
	}

	var reply httptypes.VolumeMap = map[string]*types.Volume{}

	volumes, err := d.Volumes(
		ctx,
		&drivers.VolumesOpts{
			Attachments: store.GetBool("attachments"),
			Opts:        store,
		})
	if err != nil {
		return utils.NewBatchProcessErr(reply, err)
	}

	for _, v := range volumes {
		reply[v.ID] = v
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) volumeInspect(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	reply, err := getVolume(ctx)
	if err != nil {
		return err
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) volumeCreate(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	d, err := httputils.GetStorageDriver(ctx)
	if err != nil {
		return err
	}

	reply, err := d.VolumeCreate(
		ctx,
		store.GetString("name"),
		&drivers.VolumeCreateOpts{
			AvailabilityZone: store.GetStringPtr("availabilityZone"),
			IOPS:             store.GetInt64Ptr("iops"),
			Size:             store.GetInt64Ptr("size"),
			Type:             store.GetStringPtr("type"),
			Opts:             store,
		})
	if err != nil {
		return err
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) volumeCopy(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	d, err := httputils.GetStorageDriver(ctx)
	if err != nil {
		return err
	}

	reply, err := d.VolumeCopy(
		ctx,
		store.GetString("volumeID"),
		store.GetString("volumeName"),
		store)
	if err != nil {
		return err
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) volumeSnapshot(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	d, err := httputils.GetStorageDriver(ctx)
	if err != nil {
		return err
	}

	reply, err := d.VolumeSnapshot(
		ctx,
		store.GetString("volumeID"),
		store.GetString("snapshotName"),
		store)
	if err != nil {
		return err
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) volumeAttach(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	d, err := httputils.GetStorageDriver(ctx)
	if err != nil {
		return err
	}

	reply, err := d.VolumeAttach(
		ctx,
		store.GetString("volumeID"),
		&drivers.VolumeAttachByIDOpts{
			NextDevice: store.GetStringPtr("nextDeviceName"),
			Opts:       store,
		})
	if err != nil {
		return err
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) volumeDetach(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	d, err := httputils.GetStorageDriver(ctx)
	if err != nil {
		return err
	}

	volume, err := getVolume(ctx)
	if err != nil {
		return err
	}

	err = d.VolumeDetach(
		ctx,
		volume.ID,
		store)
	if err != nil {
		return err
	}

	httputils.WriteJSON(w, http.StatusResetContent, volume)
	return nil
}

func (r *router) volumeDetachAll(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	var reply httptypes.ServiceVolumeMap = map[string]httptypes.VolumeMap{}

	for _, service := range r.services {

		svcCtx := ctx.WithValue("serviceID", service.Name())

		d := service.Driver()

		volumes, err := d.Volumes(
			svcCtx,
			&drivers.VolumesOpts{
				Attachments: store.GetBool("attachments"),
				Opts:        store,
			})
		if err != nil {
			return utils.NewBatchProcessErr(reply, err)
		}

		volumeMap := map[string]*types.Volume{}
		reply[service.Name()] = volumeMap

		for _, volume := range volumes {
			if err := d.VolumeDetach(svcCtx, volume.ID, store); err != nil {
				return utils.NewBatchProcessErr(reply, err)
			}
			volumeMap[volume.ID] = volume
		}
	}

	httputils.WriteJSON(w, http.StatusResetContent, reply)
	return nil
}

func (r *router) volumeDetachAllForService(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	d, err := httputils.GetStorageDriver(ctx)
	if err != nil {
		return err
	}

	var reply httptypes.VolumeMap = map[string]*types.Volume{}

	volumes, err := d.Volumes(
		ctx,
		&drivers.VolumesOpts{
			Attachments: store.GetBool("attachments"),
			Opts:        store,
		})
	if err != nil {
		return err
	}

	for _, volume := range volumes {
		if err := d.VolumeDetach(ctx, volume.ID, store); err != nil {
			return utils.NewBatchProcessErr(reply, err)
		}
		reply[volume.ID] = volume
	}

	httputils.WriteJSON(w, http.StatusResetContent, reply)
	return nil
}

func (r *router) volumeRemove(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	d, err := httputils.GetStorageDriver(ctx)
	if err != nil {
		return err
	}

	err = d.VolumeRemove(
		ctx,
		store.GetString("volumeID"),
		store)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusResetContent)
	return nil
}

var (
	volumeTypeName = utils.GetTypePkgPathAndName(
		&types.Volume{})
)

func getVolume(
	ctx context.Context) (*types.Volume, error) {
	obj := ctx.Value("volume")
	if obj == nil {
		return nil, utils.NewContextKeyErr("volume")
	}
	typedObj, ok := obj.(*types.Volume)
	if !ok {
		return nil, utils.NewContextTypeErr(
			"volume",
			volumeTypeName,
			utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}
