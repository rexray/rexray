package tasks

import (
	"fmt"
	"net/http"

	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/utils"
)

func (r *router) tasks(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	tasks := map[string]*types.Task{}
	for t := range services.Tasks(ctx) {
		tasks[fmt.Sprintf("%d", t.ID)] = t
	}
	httputils.WriteJSON(w, http.StatusOK, tasks)
	return nil
}

func (r *router) taskInspect(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	task := services.TaskInspect(ctx, store.GetInt("taskID"))
	if task == nil {
		return utils.NewNotFoundError(store.GetString("taskID"))
	}

	httputils.WriteJSON(w, http.StatusOK, task)
	return nil
}
