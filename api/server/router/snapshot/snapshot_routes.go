package snapshot

import (
	"net/http"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	httptypes "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/api/utils"
)

func (r *router) snapshots(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	var reply httptypes.ServiceSnapshotMap = map[string]httptypes.SnapshotMap{}

	for _, service := range r.services {

		snapshots, err := service.Driver().Snapshots(ctx, store)
		if err != nil {
			return utils.NewBatchProcessErr(reply, err)
		}

		snapshotMap := map[string]*types.Snapshot{}
		reply[service.Name()] = snapshotMap
		for _, s := range snapshots {
			snapshotMap[s.ID] = s
		}
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) snapshotsForService(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	d, err := httputils.GetStorageDriver(ctx)
	if err != nil {
		return err
	}

	var reply httptypes.SnapshotMap = map[string]*types.Snapshot{}

	snapshots, err := d.Snapshots(ctx, store)
	if err != nil {
		return utils.NewBatchProcessErr(reply, err)
	}

	for _, s := range snapshots {
		reply[s.ID] = s
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) snapshotInspect(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	reply, err := getSnapshot(ctx)
	if err != nil {
		return err
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) snapshotRemove(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	d, err := httputils.GetStorageDriver(ctx)
	if err != nil {
		return err
	}

	err = d.SnapshotRemove(
		ctx,
		store.GetString("snapshotID"),
		store)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusResetContent)
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

	reply, err := d.VolumeCreateFromSnapshot(
		ctx,
		store.GetString("snapshotID"),
		store.GetString("volumeName"),
		store)
	if err != nil {
		return err
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

func (r *router) snapshotCopy(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	d, err := httputils.GetStorageDriver(ctx)
	if err != nil {
		return err
	}

	reply, err := d.SnapshotCopy(
		ctx,
		store.GetString("snapshotID"),
		store.GetString("snapshotName"),
		store.GetString("destinationID"),
		store)
	if err != nil {
		return err
	}

	httputils.WriteJSON(w, http.StatusOK, reply)
	return nil
}

var (
	snapshotTypeName = utils.GetTypePkgPathAndName(
		&types.Snapshot{})
)

func getSnapshot(
	ctx context.Context) (*types.Snapshot, error) {
	obj := ctx.Value("snapshot")
	if obj == nil {
		return nil, utils.NewContextKeyErr("snapshot")
	}
	typedObj, ok := obj.(*types.Snapshot)
	if !ok {
		return nil, utils.NewContextTypeErr(
			"snapshot",
			snapshotTypeName,
			utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}
