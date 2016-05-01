package services

import (
	"fmt"
	"strings"
	"sync"

	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/types"
)

var (
	servicesByServer    = map[string]*serviceContainer{}
	servicesByServerRWL = &sync.RWMutex{}
)

type serviceContainer struct {
	config          gofig.Config
	storageServices map[string]types.StorageService
	taskService     *globalTaskService
}

// Init initializes the types.
func Init(ctx types.Context, config gofig.Config) error {
	serverName := ctx.ServerName()
	ctx.Info("initializing server services")

	sc := &serviceContainer{
		taskService:     &globalTaskService{name: "global-task-service"},
		storageServices: map[string]types.StorageService{},
	}

	if err := sc.Init(ctx, config); err != nil {
		return err
	}

	servicesByServerRWL.Lock()
	defer servicesByServerRWL.Unlock()
	servicesByServer[serverName] = sc

	return nil
}

func (sc *serviceContainer) Init(ctx types.Context, config gofig.Config) error {
	sc.config = config

	if err := sc.taskService.Init(config); err != nil {
		return err
	}

	if err := sc.initStorageServices(ctx); err != nil {
		return err
	}

	return nil
}

func getStorageServices(
	ctx types.Context) map[string]types.StorageService {
	return servicesByServer[ctx.ServerName()].storageServices
}

// GetStorageService returns the storage service specified by the given name;
// otherwise a nil value is returned if no such service exists.
func GetStorageService(
	ctx types.Context, name string) types.StorageService {
	servicesByServerRWL.RLock()
	defer servicesByServerRWL.RUnlock()
	name = strings.ToLower(name)
	ctx.WithField("service", name).Debug("getting storage service")
	return getStorageServices(ctx)[name]
}

// StorageServices returns a channel on which all the storage services are
// received.
func StorageServices(ctx types.Context) <-chan types.StorageService {
	c := make(chan types.StorageService)
	go func() {
		for _, v := range getStorageServices(ctx) {
			c <- v
		}
		close(c)
	}()
	return c
}

func (sc *serviceContainer) initStorageServices(ctx types.Context) error {
	if ctx == nil {
		panic("ctx is nil")
	}
	if sc.config == nil {
		panic("sc.config is nil")
	}
	cfgSvcs := sc.config.Get(types.ConfigServices)
	cfgSvcsMap, ok := cfgSvcs.(map[string]interface{})
	if !ok {
		err := goof.WithFields(goof.Fields{
			"configKey": types.ConfigServices,
			"obj":       cfgSvcs,
		}, "invalid format")
		err.IncludeFieldsInString(true)
		err.IncludeFieldsInError(true)
		err.IncludeFieldsInFormat(true)
		return err
	}
	ctx.WithField("count", len(cfgSvcsMap)).Debug("got services map")

	for serviceName := range cfgSvcsMap {
		serviceName = strings.ToLower(serviceName)

		ctx = ctx.WithContextSID(types.ContextServiceName, serviceName)
		ctx.Debug("processing service config")

		scope := fmt.Sprintf("libstorage.server.services.%s", serviceName)
		ctx.WithField("scope", scope).Debug("getting scoped config for service")
		config := sc.config.Scope(scope)

		storSvc := &storageService{name: serviceName}
		if err := storSvc.Init(ctx, config); err != nil {
			return err
		}

		ctx = ctx.WithContextSID(
			types.ContextDriverName, storSvc.Driver().Name())
		ctx.Info("created new service")

		sc.storageServices[serviceName] = storSvc
	}

	return nil
}

func getTaskService(ctx types.Context) *globalTaskService {
	servicesByServerRWL.RLock()
	defer servicesByServerRWL.RUnlock()
	ctx.Debug("getting task service")
	return servicesByServer[ctx.ServerName()].taskService
}

// Tasks returns a channel on which all tasks are received.
func Tasks(ctx types.Context) <-chan *types.Task {
	return getTaskService(ctx).Tasks()
}

// TaskTrack creates a new, trackable task.
func TaskTrack(ctx types.Context) *types.Task {
	return getTaskService(ctx).TaskTrack(ctx)
}

// TaskExecute enqueues a task for execution.
func TaskExecute(
	ctx types.Context,
	run types.TaskRunFunc,
	schema []byte) *types.Task {
	return getTaskService(ctx).TaskExecute(ctx, run, schema)
}

// TaskInspect returns the task with the specified ID.
func TaskInspect(ctx types.Context, taskID int) *types.Task {
	return getTaskService(ctx).TaskInspect(taskID)
}

// TaskWait blocks until the specified task is completed.
func TaskWait(ctx types.Context, taskID int) {
	getTaskService(ctx).TaskWait(taskID)
}

// TaskWaitC returns a channel that is closed only when the specified task is
// completed.
func TaskWaitC(ctx types.Context, taskID int) <-chan int {
	return getTaskService(ctx).TaskWaitC(taskID)
}

// TaskWaitAll blocks until all the specified task are complete.
func TaskWaitAll(ctx types.Context, taskIDs ...int) {
	getTaskService(ctx).TaskWaitAll(taskIDs...)
}

// TaskWaitAllC returns a channel that is closed only when all the specified
// tasks are completed.
func TaskWaitAllC(ctx types.Context, taskIDs ...int) <-chan int {
	return getTaskService(ctx).TaskWaitAllC(taskIDs...)
}
