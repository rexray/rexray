package volume

import (
	"net/http"
	"strings"
	"sync"

	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	"github.com/emccode/libstorage/api/utils/filters"
	"github.com/emccode/libstorage/api/utils/schema"
)

func (r *router) volumes(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	filter, err := parseFilter(store)
	if err != nil {
		return err
	}
	lcaseFLeft := ""
	lcaseFRight := ""
	if filter != nil {
		store.Set("filter", filter)
		lcaseFLeft = strings.ToLower(filter.Left)
		lcaseFRight = strings.ToLower(filter.Right)
	}

	attachments := store.GetBool("attachments")

	var (
		tasks   = map[string]*types.Task{}
		taskIDs []int
		opts    = &types.VolumesOpts{
			Attachments: attachments,
			Opts:        store,
		}
		reply = types.ServiceVolumeMap{}
	)

	for service := range services.StorageServices(ctx) {

		run := func(
			ctx types.Context,
			svc types.StorageService) (interface{}, error) {

			ctx, err := httputils.WithServiceContext(ctx, svc)
			if err != nil {
				return nil, err
			}

			if attachments && ctx.InstanceID() == nil {
				return nil, utils.NewMissingInstanceIDError(service.Name())
			}

			objs, err := svc.Driver().Volumes(ctx, opts)
			if err != nil {
				return nil, err
			}

			objMap := map[string]*types.Volume{}
			for _, obj := range objs {
				if lcaseFLeft == "name" {
					if strings.ToLower(obj.Name) == lcaseFRight {
						objMap[obj.ID] = obj
					}
				} else {
					objMap[obj.ID] = obj
				}
			}
			return objMap, nil
		}

		task := service.TaskExecute(ctx, run, schema.VolumeMapSchema)
		taskIDs = append(taskIDs, task.ID)
		tasks[service.Name()] = task
	}

	run := func(ctx types.Context) (interface{}, error) {

		services.TaskWaitAll(ctx, taskIDs...)

		for k, v := range tasks {
			if v.Error != nil {
				return nil, utils.NewBatchProcessErr(reply, v.Error)
			}

			objMap, ok := v.Result.(map[string]*types.Volume)
			if !ok {
				return nil, utils.NewBatchProcessErr(
					reply, goof.New("error casting to []*types.Volume"))
			}
			reply[k] = objMap
		}

		return reply, nil
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		services.TaskExecute(ctx, run, schema.ServiceVolumeMapSchema),
		http.StatusOK)
}

func (r *router) volumesForService(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	filter, err := parseFilter(store)
	if err != nil {
		return err
	}
	lcaseFLeft := ""
	lcaseFRight := ""
	if filter != nil {
		store.Set("filter", filter)
		lcaseFLeft = strings.ToLower(filter.Left)
		lcaseFRight = strings.ToLower(filter.Right)
	}

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	attachments := store.GetBool("attachments")
	if attachments && ctx.InstanceID() == nil {
		return utils.NewMissingInstanceIDError(service.Name())
	}

	opts := &types.VolumesOpts{
		Attachments: attachments,
		Opts:        store,
	}

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		var reply types.VolumeMap = map[string]*types.Volume{}

		objs, err := svc.Driver().Volumes(ctx, opts)
		if err != nil {
			return nil, err
		}

		for _, obj := range objs {
			if lcaseFLeft == "name" {
				if strings.ToLower(obj.Name) == lcaseFRight {
					reply[obj.ID] = obj
				}
			} else {
				reply[obj.ID] = obj
			}
		}
		return reply, nil
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, schema.VolumeMapSchema),
		http.StatusOK)
}

func (r *router) volumeInspect(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	attachments := store.GetBool("attachments")
	if attachments && ctx.InstanceID() == nil {
		return utils.NewMissingInstanceIDError(service.Name())
	}

	opts := &types.VolumeInspectOpts{
		Attachments: attachments,
		Opts:        store,
	}

	var run types.StorageTaskRunFunc
	if store.IsSet("byName") {
		run = func(
			ctx types.Context,
			svc types.StorageService) (interface{}, error) {

			vols, err := svc.Driver().Volumes(
				ctx,
				&types.VolumesOpts{
					Attachments: attachments,
					Opts:        store,
				})

			if err != nil {
				return nil, err
			}

			volID := strings.ToLower(store.GetString("volumeID"))
			for _, v := range vols {
				if strings.ToLower(v.Name) == volID {
					return v, nil
				}
			}

			return nil, utils.NewNotFoundError(volID)
		}

	} else {

		run = func(
			ctx types.Context,
			svc types.StorageService) (interface{}, error) {
			return svc.Driver().VolumeInspect(
				ctx, store.GetString("volumeID"), opts)
		}
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, schema.VolumeSchema),
		http.StatusOK)
}

func (r *router) volumeCreate(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		return svc.Driver().VolumeCreate(
			ctx,
			store.GetString("name"),
			&types.VolumeCreateOpts{
				AvailabilityZone: store.GetStringPtr("availabilityZone"),
				IOPS:             store.GetInt64Ptr("iops"),
				Size:             store.GetInt64Ptr("size"),
				Type:             store.GetStringPtr("type"),
				Opts:             store,
			})
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, schema.VolumeSchema),
		http.StatusCreated)
}

func (r *router) volumeCopy(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		return svc.Driver().VolumeCopy(
			ctx,
			store.GetString("volumeID"),
			store.GetString("volumeName"),
			store)
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, schema.VolumeSchema),
		http.StatusCreated)
}

func (r *router) volumeSnapshot(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		return svc.Driver().VolumeSnapshot(
			ctx,
			store.GetString("volumeID"),
			store.GetString("snapshotName"),
			store)
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, schema.SnapshotSchema),
		http.StatusCreated)
}

func (r *router) volumeAttach(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	if ctx.InstanceID() == nil {
		return utils.NewMissingInstanceIDError(service.Name())
	}

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		return svc.Driver().VolumeAttach(
			ctx,
			store.GetString("volumeID"),
			&types.VolumeAttachOpts{
				NextDevice: store.GetStringPtr("nextDeviceName"),
				Force:      store.GetBool("force"),
				Opts:       store,
			})
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, schema.VolumeSchema),
		http.StatusOK)
}

func (r *router) volumeDetach(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	if ctx.InstanceID() == nil {
		return utils.NewMissingInstanceIDError(service.Name())
	}

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		return svc.Driver().VolumeDetach(
			ctx,
			store.GetString("volumeID"),
			&types.VolumeDetachOpts{
				Force: store.GetBool("force"),
				Opts:  store,
			})
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, nil),
		http.StatusResetContent)
}

func (r *router) volumeDetachAll(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	var (
		taskIDs  []int
		tasks                           = map[string]*types.Task{}
		opts                            = &types.VolumesOpts{Opts: store}
		reply    types.ServiceVolumeMap = map[string]types.VolumeMap{}
		replyRWL                        = &sync.Mutex{}
	)

	for service := range services.StorageServices(ctx) {

		run := func(
			ctx types.Context,
			svc types.StorageService) (interface{}, error) {

			ctx, err := httputils.WithServiceContext(ctx, svc)
			if err != nil {
				return nil, err
			}

			if ctx.InstanceID() == nil {
				return nil, utils.NewMissingInstanceIDError(service.Name())
			}

			driver := svc.Driver()

			volumes, err := driver.Volumes(ctx, opts)
			if err != nil {
				return nil, err
			}

			// check here
			var volumeMap types.VolumeMap = map[string]*types.Volume{}
			defer func() {
				if len(volumeMap) > 0 {
					replyRWL.Lock()
					defer replyRWL.Unlock()
					reply[service.Name()] = volumeMap
				}
			}()

			for _, volume := range volumes {
				vol, err := driver.VolumeDetach(
					ctx,
					volume.ID,
					&types.VolumeDetachOpts{
						Force: store.GetBool("force"),
						Opts:  store,
					})
				if err != nil {
					return nil, err
				}
				volumeMap[volume.ID] = vol
			}

			return nil, nil
		}

		task := service.TaskExecute(ctx, run, nil)
		taskIDs = append(taskIDs, task.ID)
		tasks[service.Name()] = task
	}

	run := func(ctx types.Context) (interface{}, error) {
		services.TaskWaitAll(ctx, taskIDs...)
		for _, v := range tasks {
			if v.Error != nil {
				return nil, utils.NewBatchProcessErr(reply, v.Error)
			}
		}
		return reply, nil
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		services.TaskExecute(ctx, run, schema.ServiceVolumeMapSchema),
		http.StatusResetContent)
}

func (r *router) volumeDetachAllForService(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	if ctx.InstanceID() == nil {
		return utils.NewMissingInstanceIDError(service.Name())
	}

	var reply types.VolumeMap = map[string]*types.Volume{}

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		driver := svc.Driver()

		volumes, err := driver.Volumes(ctx, &types.VolumesOpts{Opts: store})
		if err != nil {
			return nil, err
		}

		for _, volume := range volumes {
			vol, err := driver.VolumeDetach(
				ctx,
				volume.ID,
				&types.VolumeDetachOpts{
					Force: store.GetBool("force"),
					Opts:  store,
				})
			if err != nil {
				return nil, utils.NewBatchProcessErr(reply, err)
			}
			reply[volume.ID] = vol
		}

		return reply, nil
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, schema.VolumeMapSchema),
		http.StatusResetContent)
}

func (r *router) volumeRemove(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service, err := httputils.GetService(ctx)
	if err != nil {
		return err
	}

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		return nil, svc.Driver().VolumeRemove(
			ctx,
			store.GetString("volumeID"),
			store)
	}

	return httputils.WriteTask(
		ctx,
		w,
		store,
		service.TaskExecute(ctx, run, nil),
		http.StatusNoContent)
}

func parseFilter(store types.Store) (*types.Filter, error) {
	if !store.IsSet("filter") {
		return nil, nil
	}
	fsz := store.GetString("filter")
	filter, err := filters.CompileFilter(fsz)
	if err != nil {
		return nil, utils.NewBadFilterErr(fsz, err)
	}
	return filter, nil
}
