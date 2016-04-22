package services

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/services"
)

var (
	servicesByServer    = map[string]*serviceContainer{}
	servicesByServerRWL = &sync.RWMutex{}
)

type serviceContainer struct {
	config          gofig.Config
	storageServices map[string]services.StorageService
	taskService     *globalTaskService
}

// Init initializes the services.
func Init(ctx context.Context, config gofig.Config) error {
	serverName := ctx.Value("server").(string)

	sc := &serviceContainer{
		taskService:     &globalTaskService{name: "global-task-service"},
		storageServices: map[string]services.StorageService{},
	}

	if err := sc.Init(config); err != nil {
		return err
	}

	servicesByServerRWL.Lock()
	defer servicesByServerRWL.Unlock()
	servicesByServer[serverName] = sc

	return nil
}

func (sc *serviceContainer) Init(config gofig.Config) error {
	sc.config = config

	if err := sc.taskService.Init(config); err != nil {
		return err
	}

	if err := sc.initStorageServices(); err != nil {
		return err
	}

	return nil
}

func getStorageServices(
	ctx context.Context) map[string]services.StorageService {

	return servicesByServer[ctx.Value("server").(string)].storageServices
}

// GetStorageService returns the storage service specified by the given name;
// otherwise a nil value is returned if no such service exists.
func GetStorageService(
	ctx context.Context, name string) services.StorageService {

	servicesByServerRWL.RLock()
	defer servicesByServerRWL.RUnlock()
	return getStorageServices(ctx)[strings.ToLower(name)]
}

// StorageServices returns a channel on which all the storage services are
// received.
func StorageServices(ctx context.Context) <-chan services.StorageService {
	c := make(chan services.StorageService)
	go func() {
		for _, v := range getStorageServices(ctx) {
			c <- v
		}
		close(c)
	}()
	return c
}

func (sc *serviceContainer) initStorageServices() error {

	cfgSvcs := sc.config.Get("libstorage.server.services")
	cfgSvcsMap, ok := cfgSvcs.(map[string]interface{})
	if !ok {
		return goof.New("invalid format libstorage.server.services")
	}
	log.WithField("count", len(cfgSvcsMap)).Debug("got services map")

	for serviceName := range cfgSvcsMap {
		serviceName = strings.ToLower(serviceName)

		log.WithField("service", serviceName).Debug("processing service config")

		scope := fmt.Sprintf("libstorage.server.services.%s", serviceName)
		log.WithField("scope", scope).Debug("getting scoped config for service")
		config := sc.config.Scope(scope)

		storSvc := &storageService{name: serviceName}
		if err := storSvc.Init(config); err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"service": storSvc.Name(),
			"driver":  storSvc.Driver().Name(),
		}).Info("created new service")

		sc.storageServices[serviceName] = storSvc
	}

	return nil
}

func getTaskService(ctx context.Context) *globalTaskService {
	servicesByServerRWL.RLock()
	defer servicesByServerRWL.RUnlock()
	return servicesByServer[ctx.Value("server").(string)].taskService
}

// Tasks returns a channel on which all tasks are received.
func Tasks(ctx context.Context) <-chan *types.Task {
	return getTaskService(ctx).Tasks()
}

// TaskTrack creates a new, trackable task.
func TaskTrack(ctx context.Context) *types.Task {
	return getTaskService(ctx).TaskTrack(ctx)
}

// TaskExecute enqueues a task for execution.
func TaskExecute(
	ctx context.Context,
	run services.TaskRunFunc,
	schema []byte) *types.Task {
	return getTaskService(ctx).TaskExecute(ctx, run, schema)
}

// TaskInspect returns the task with the specified ID.
func TaskInspect(ctx context.Context, taskID int) *types.Task {
	return getTaskService(ctx).TaskInspect(taskID)
}

// TaskWait blocks until the specified task is completed.
func TaskWait(ctx context.Context, taskID int) {
	getTaskService(ctx).TaskWait(taskID)
}

// TaskWaitC returns a channel that is closed only when the specified task is
// completed.
func TaskWaitC(ctx context.Context, taskID int) <-chan int {
	return getTaskService(ctx).TaskWaitC(taskID)
}

// TaskWaitAll blocks until all the specified task are complete.
func TaskWaitAll(ctx context.Context, taskIDs ...int) {
	getTaskService(ctx).TaskWaitAll(taskIDs...)
}

// TaskWaitAllC returns a channel that is closed only when all the specified
// tasks are completed.
func TaskWaitAllC(ctx context.Context, taskIDs ...int) <-chan int {
	return getTaskService(ctx).TaskWaitAllC(taskIDs...)
}
