package volume

import (
	"net/http"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/httputils"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/services"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils/filters"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils/schema"
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
	if filter != nil {
		store.Set("filter", filter)
	}

	var (
		tasks   = map[string]*types.Task{}
		taskIDs []int
		opts    = &types.VolumesOpts{
			Attachments: store.GetAttachments(),
			Opts:        store,
		}
		reply = types.ServiceVolumeMap{}
	)

	for service := range services.StorageServices(ctx) {

		run := func(
			ctx types.Context,
			svc types.StorageService) (interface{}, error) {

			ctx = context.WithStorageService(ctx, svc)

			var err error
			if ctx, err = context.WithStorageSession(ctx); err != nil {
				return nil, err
			}

			return getFilteredVolumes(ctx, req, store, svc, opts, filter)
		}

		task := service.TaskEnqueue(ctx, run, schema.VolumeMapSchema)
		taskIDs = append(taskIDs, task.ID)
		tasks[service.Name()] = task
	}

	run := func(ctx types.Context) (interface{}, error) {

		services.TaskWaitAll(ctx, taskIDs...)

		for k, v := range tasks {
			if v.Error != nil {
				return nil, utils.NewBatchProcessErr(reply, v.Error)
			}

			objMap, ok := v.Result.(types.VolumeMap)
			if !ok {
				return nil, utils.NewBatchProcessErr(
					reply, goof.New("error casting to types.VolumeMap"))
			}
			reply[k] = objMap
		}

		return reply, nil
	}

	return httputils.WriteTask(
		ctx,
		r.config,
		w,
		store,
		services.TaskEnqueue(ctx, run, schema.ServiceVolumeMapSchema),
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
	if filter != nil {
		store.Set("filter", filter)
	}

	service := context.MustService(ctx)

	opts := &types.VolumesOpts{
		Attachments: store.GetAttachments(),
		Opts:        store,
	}

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		return getFilteredVolumes(ctx, req, store, svc, opts, filter)
	}

	return httputils.WriteTask(
		ctx,
		r.config,
		w,
		store,
		service.TaskEnqueue(ctx, run, schema.VolumeMapSchema),
		http.StatusOK)
}

func handleVolAttachments(
	ctx types.Context,
	lf log.Fields,
	iid *types.InstanceID,
	vol *types.Volume,
	attachments types.VolumeAttachmentsTypes) bool {

	if attachments == 0 {
		vol.Attachments = nil
		return true
	}

	if lf == nil {
		lf = log.Fields{}
	}

	f := func(s types.VolumeAttachmentStates) bool {
		lf["attachmentState"] = s
		// if the volume has no attachments and the mask indicates that
		// only attached volumes should be returned then omit this volume
		if s == types.VolumeAvailable &&
			attachments.Attached() &&
			!attachments.Unattached() {
			ctx.WithFields(lf).Debug("omitting unattached volume")
			return false
		}
		// if the volume has attachments and the mask indicates that
		// only unattached volumes should be returned then omit this volume
		if (s == types.VolumeAttached || s == types.VolumeUnavailable) &&
			!attachments.Attached() &&
			attachments.Unattached() {
			ctx.WithFields(lf).Debug("omitting attached volume")
			return false
		}
		// if the volume is attached to an instance other than the one provided
		// and only the instance's volumes should be returned, then omit this
		// volume
		if attachments.Mine() &&
			attachments.Attached() &&
			((s == types.VolumeAvailable && !attachments.Unattached()) ||
				s == types.VolumeUnavailable) {
			//!attachments.Unattached() {
			ctx.WithFields(lf).Debug("omitting unavailable volume")
			return false
		}
		ctx.WithFields(lf).Debug("including volume")
		return true
	}

	// if the attachment state has already been set by the driver then
	// use it to determine whether the volume should be omitted
	if vol.AttachmentState > 0 {
		ctx.WithFields(lf).Debug(
			"deferring to driver-specified attachment state")
		return f(vol.AttachmentState)
	}

	ctx.WithFields(lf).Debug("manually calculating attachment state")

	// determine a volume's attachment state
	if len(vol.Attachments) == 0 {
		vol.AttachmentState = types.VolumeAvailable
	} else {
		vol.AttachmentState = types.VolumeUnavailable
		if iid != nil {
			for _, a := range vol.Attachments {
				if a.InstanceID != nil &&
					strings.EqualFold(iid.ID, a.InstanceID.ID) {
					vol.AttachmentState = types.VolumeAttached
					break
				}
			}
		}
	}

	// use the ascertained attachment state to determine whether or not the
	// volume should be omitted
	return f(vol.AttachmentState)
}

func getFilteredVolumes(
	ctx types.Context,
	req *http.Request,
	store types.Store,
	storSvc types.StorageService,
	opts *types.VolumesOpts,
	filter *types.Filter) (types.VolumeMap, error) {

	var (
		filterOp    types.FilterOperator
		filterLeft  string
		filterRight string
		objMap      = types.VolumeMap{}
	)

	iid, iidOK := context.InstanceID(ctx)
	if opts.Attachments.RequiresInstanceID() && !iidOK {
		return nil, utils.NewMissingInstanceIDError(storSvc.Name())
	}

	ctx.WithField("attachments", opts.Attachments).Debug("querying volumes")

	objs, err := storSvc.Driver().Volumes(ctx, opts)
	if err != nil {
		return nil, err
	}

	if filter != nil {
		filterOp = filter.Op
		filterLeft = strings.ToLower(filter.Left)
		filterRight = strings.ToLower(filter.Right)
	}

	for _, obj := range objs {

		lf := log.Fields{
			"attachments": opts.Attachments,
			"volumeID":    obj.ID,
			"volumeName":  obj.Name,
		}

		if filterOp == types.FilterEqualityMatch && filterLeft == "name" {
			ctx.WithFields(lf).Debug("checking name filter")
			if !strings.EqualFold(obj.Name, filterRight) {
				ctx.WithFields(lf).Debug("omitted volume due to name filter")
				continue
			}
		}

		if !handleVolAttachments(ctx, lf, iid, obj, opts.Attachments) {
			continue
		}

		if OnVolume != nil {
			ctx.WithFields(lf).Debug("invoking OnVolume handler")
			ok, err := OnVolume(ctx, req, store, obj)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
		}

		objMap[obj.ID] = obj
	}

	return objMap, nil
}

func (r *router) volumeInspect(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	attachments := store.GetAttachments()

	service := context.MustService(ctx)
	iid, iidOK := context.InstanceID(ctx)
	if !iidOK && attachments.RequiresInstanceID() {
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

			var (
				vol *types.Volume
				err error
			)
			volID := store.GetString("volumeID")

			if sd, ok := svc.Driver().(types.StorageDriverVolInspectByName); ok {
				ctx.Debug("driver is StorageDriverVolInspectByName")
				vol, err = sd.VolumeInspectByName(
					ctx, volID, opts)
				if err != nil {
					return nil, err
				}
			} else {
				ctx.Debug("driver is not StorageDriverVolInspectByName")
				vols, err := svc.Driver().Volumes(
					ctx,
					&types.VolumesOpts{
						Attachments: attachments,
						Opts:        store,
					})
				if err != nil {
					return nil, err
				}
				for _, v := range vols {
					if strings.EqualFold(v.Name, volID) {
						vol = v
						break
					}
				}
			}

			if vol == nil {
				return nil, utils.NewNotFoundError(volID)
			}

			if !handleVolAttachments(
				ctx, nil, iid, vol, attachments) {
				return nil, utils.NewNotFoundError(volID)
			}
			if OnVolume != nil {
				ok, err := OnVolume(ctx, req, store, vol)
				if err != nil {
					return nil, err
				}
				if !ok {
					return nil,
						utils.NewNotFoundError(volID)
				}
			}

			return vol, nil
		}

	} else {

		run = func(
			ctx types.Context,
			svc types.StorageService) (interface{}, error) {

			v, err := svc.Driver().VolumeInspect(
				ctx, store.GetString("volumeID"), opts)

			if err != nil {
				return nil, err
			}

			if !handleVolAttachments(ctx, nil, iid, v, attachments) {
				return nil, utils.NewNotFoundError(v.ID)
			}

			if OnVolume != nil {
				ok, err := OnVolume(ctx, req, store, v)
				if err != nil {
					return nil, err
				}
				if !ok {
					return nil, utils.NewNotFoundError(v.ID)
				}
			}

			return v, nil
		}
	}

	return httputils.WriteTask(
		ctx,
		r.config,
		w,
		store,
		service.TaskEnqueue(ctx, run, schema.VolumeSchema),
		http.StatusOK)
}

func (r *router) volumeCreate(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service := context.MustService(ctx)

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		volumeName := store.GetString("name")
		opts := &types.VolumeCreateOpts{
			AvailabilityZone: store.GetStringPtr("availabilityZone"),
			IOPS:             store.GetInt64Ptr("iops"),
			Size:             store.GetInt64Ptr("size"),
			Type:             store.GetStringPtr("type"),
			Encrypted:        store.GetBoolPtr("encrypted"),
			EncryptionKey:    store.GetStringPtr("encryptionKey"),
			Opts:             store,
		}
		fields := map[string]interface{}{
			"volumeName": store.GetString("name"),
		}
		if opts.AvailabilityZone != nil {
			fields["availabilityZone"] = &opts.AvailabilityZone
		}
		if opts.Encrypted != nil {
			fields["encrypted"] = &opts.Encrypted
		}
		if opts.EncryptionKey != nil {
			fields["encryptionKey"] = &opts.EncryptionKey
		}
		if opts.IOPS != nil {
			fields["iops"] = &opts.IOPS
		}
		if opts.Size != nil {
			fields["size"] = &opts.Size
		}
		if opts.Type != nil {
			fields["type"] = &opts.Type
		}
		ctx.WithFields(fields).Debug("creating volume")

		v, err := svc.Driver().VolumeCreate(ctx, volumeName, opts)
		if err != nil {
			ctx.WithFields(fields).WithError(err).Error("error creating volume")
			return nil, err
		}
		ctx.WithFields(fields).Debug("success creating volume")

		if OnVolume != nil {
			ok, err := OnVolume(ctx, req, store, v)
			if err != nil {
				ctx.WithFields(fields).WithError(err).Error(
					"error calling onvolume handler")
				return nil, err
			}
			if !ok {
				return nil, utils.NewNotFoundError(v.ID)
			}
		}

		if v.AttachmentState == 0 {
			v.AttachmentState = types.VolumeAvailable
		}

		return v, nil
	}

	return httputils.WriteTask(
		ctx,
		r.config,
		w,
		store,
		service.TaskEnqueue(ctx, run, schema.VolumeSchema),
		http.StatusCreated)
}

func (r *router) volumeCopy(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service := context.MustService(ctx)

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		v, err := svc.Driver().VolumeCopy(
			ctx,
			store.GetString("volumeID"),
			store.GetString("volumeName"),
			store)

		if err != nil {
			return nil, err
		}

		if OnVolume != nil {
			ok, err := OnVolume(ctx, req, store, v)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, utils.NewNotFoundError(v.ID)
			}
		}

		if v.AttachmentState == 0 {
			v.AttachmentState = types.VolumeAvailable
		}
		return v, nil
	}

	return httputils.WriteTask(
		ctx,
		r.config,
		w,
		store,
		service.TaskEnqueue(ctx, run, schema.VolumeSchema),
		http.StatusCreated)
}

func (r *router) volumeSnapshot(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service := context.MustService(ctx)

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
		r.config,
		w,
		store,
		service.TaskEnqueue(ctx, run, schema.SnapshotSchema),
		http.StatusCreated)
}

func (r *router) volumeAttach(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service := context.MustService(ctx)
	if _, ok := context.InstanceID(ctx); !ok {
		return utils.NewMissingInstanceIDError(service.Name())
	}

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		v, attTokn, err := svc.Driver().VolumeAttach(
			ctx,
			store.GetString("volumeID"),
			&types.VolumeAttachOpts{
				NextDevice: store.GetStringPtr("nextDeviceName"),
				Force:      store.GetBool("force"),
				Opts:       store,
			})

		if err != nil {
			return nil, err
		}

		if OnVolume != nil {
			ok, err := OnVolume(ctx, req, store, v)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, utils.NewNotFoundError(v.ID)
			}
		}

		if v.AttachmentState == 0 {
			v.AttachmentState = types.VolumeAttached
		}

		return &types.VolumeAttachResponse{
			Volume:      v,
			AttachToken: attTokn,
		}, nil
	}

	return httputils.WriteTask(
		ctx,
		r.config,
		w,
		store,
		service.TaskEnqueue(ctx, run, schema.VolumeAttachResponseSchema),
		http.StatusOK)
}

func (r *router) volumeDetach(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service := context.MustService(ctx)
	if _, ok := context.InstanceID(ctx); !ok {
		return utils.NewMissingInstanceIDError(service.Name())
	}

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		v, err := svc.Driver().VolumeDetach(
			ctx,
			store.GetString("volumeID"),
			&types.VolumeDetachOpts{
				Force: store.GetBool("force"),
				Opts:  store,
			})

		if err != nil {
			return nil, err
		}

		if v == nil {
			return nil, nil
		}

		if OnVolume != nil {
			ok, err := OnVolume(ctx, req, store, v)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, utils.NewNotFoundError(v.ID)
			}
		}

		if v.AttachmentState == 0 {
			v.AttachmentState = types.VolumeAvailable
		}

		return v, nil
	}

	return httputils.WriteTask(
		ctx,
		r.config,
		w,
		store,
		service.TaskEnqueue(ctx, run, nil),
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

			ctx = context.WithStorageService(ctx, svc)

			if _, ok := context.InstanceID(ctx); !ok {
				return nil, utils.NewMissingInstanceIDError(service.Name())
			}

			var err error
			if ctx, err = context.WithStorageSession(ctx); err != nil {
				return nil, err
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
				v, err := driver.VolumeDetach(
					ctx,
					volume.ID,
					&types.VolumeDetachOpts{
						Force: store.GetBool("force"),
						Opts:  store,
					})
				if err != nil {
					return nil, err
				}

				if err != nil {
					return nil, err
				}

				if v == nil {
					continue
				}

				if OnVolume != nil {
					ok, err := OnVolume(ctx, req, store, v)
					if err != nil {
						return nil, err
					}
					if !ok {
						return nil, utils.NewNotFoundError(v.ID)
					}
				}

				if v.AttachmentState == 0 {
					v.AttachmentState = types.VolumeAvailable
				}

				volumeMap[v.ID] = v
			}

			return nil, nil
		}

		task := service.TaskEnqueue(ctx, run, nil)
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
		r.config,
		w,
		store,
		services.TaskEnqueue(ctx, run, schema.ServiceVolumeMapSchema),
		http.StatusResetContent)
}

func (r *router) volumeDetachAllForService(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service := context.MustService(ctx)
	if _, ok := context.InstanceID(ctx); !ok {
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
			v, err := driver.VolumeDetach(
				ctx,
				volume.ID,
				&types.VolumeDetachOpts{
					Force: store.GetBool("force"),
					Opts:  store,
				})
			if err != nil {
				return nil, utils.NewBatchProcessErr(reply, err)
			}

			if err != nil {
				return nil, err
			}

			if v == nil {
				continue
			}

			if OnVolume != nil {
				ok, err := OnVolume(ctx, req, store, v)
				if err != nil {
					return nil, err
				}
				if !ok {
					return nil, utils.NewNotFoundError(v.ID)
				}
			}

			if v.AttachmentState == 0 {
				v.AttachmentState = types.VolumeAvailable
			}

			reply[v.ID] = v
		}

		return reply, nil
	}

	return httputils.WriteTask(
		ctx,
		r.config,
		w,
		store,
		service.TaskEnqueue(ctx, run, schema.VolumeMapSchema),
		http.StatusResetContent)
}

func (r *router) volumeRemove(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	service := context.MustService(ctx)

	run := func(
		ctx types.Context,
		svc types.StorageService) (interface{}, error) {

		return nil, svc.Driver().VolumeRemove(
			ctx,
			store.GetString("volumeID"),
			&types.VolumeRemoveOpts{
				Force: store.GetBool("force"),
				Opts:  store,
			})
	}

	return httputils.WriteTask(
		ctx,
		r.config,
		w,
		store,
		service.TaskEnqueue(ctx, run, nil),
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
