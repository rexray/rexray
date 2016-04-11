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
	once            sync.Once
	config          gofig.Config
	storageServices map[string]services.StorageService
	taskService     *globalTaskService
)

// Init initializes the services.
func Init(cfg gofig.Config) (err error) {
	once.Do(func() {
		config = cfg
		if err = initGlobalTaskService(); err != nil {
			return
		}
		if err = initStorageServices(); err != nil {
			return
		}
	})
	return
}

// GetStorageService returns the storage service specified by the given name;
// otherwise a nil value is returned if no such service exists.
func GetStorageService(name string) services.StorageService {
	return storageServices[strings.ToLower(name)]
}

// StorageServices returns a channel on which all the storage services are
// received.
func StorageServices() <-chan services.StorageService {
	c := make(chan services.StorageService)
	go func() {
		for _, v := range storageServices {
			c <- v
		}
		close(c)
	}()
	return c
}

func initStorageServices() error {

	cfgSvcs := config.Get("libstorage.server.services")
	cfgSvcsMap, ok := cfgSvcs.(map[string]interface{})
	if !ok {
		return goof.New("invalid format libstorage.server.services")
	}
	log.WithField("count", len(cfgSvcsMap)).Debug("got services map")

	storageServices = map[string]services.StorageService{}

	for serviceName := range cfgSvcsMap {
		serviceName = strings.ToLower(serviceName)

		log.WithField("service", serviceName).Debug("processing service config")

		scope := fmt.Sprintf("libstorage.server.services.%s", serviceName)
		log.WithField("scope", scope).Debug("getting scoped config for service")
		config := config.Scope(scope)

		storSvc := &storageService{name: serviceName}
		if err := storSvc.Init(config); err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"service": storSvc.Name(),
			"driver":  storSvc.Driver().Name(),
		}).Info("created new service")

		storageServices[serviceName] = storSvc
	}

	return nil
}

func initGlobalTaskService() error {
	taskService = &globalTaskService{name: "global-task-service"}
	return taskService.Init(config)
}

// Tasks returns a channel on which all tasks are received.
func Tasks() <-chan *types.Task {
	return taskService.Tasks()
}

// TaskTrack creates a new, trackable task.
func TaskTrack(ctx context.Context) *types.Task {
	return taskService.TaskTrack(ctx)
}

// TaskExecute enqueues a task for execution.
func TaskExecute(
	ctx context.Context,
	run services.TaskRunFunc,
	schema []byte) *types.Task {
	return taskService.TaskExecute(ctx, run, schema)
}

// TaskInspect returns the task with the specified ID.
func TaskInspect(taskID int) *types.Task {
	return taskService.TaskInspect(taskID)
}

// TaskWait blocks until the specified task is completed.
func TaskWait(taskID int) {
	taskService.TaskWait(taskID)
}

// TaskWaitC returns a channel that is closed only when the specified task is
// completed.
func TaskWaitC(taskID int) <-chan int {
	return taskService.TaskWaitC(taskID)
}

// TaskWaitAll blocks until all the specified task are complete.
func TaskWaitAll(taskIDs ...int) {
	taskService.TaskWaitAll(taskIDs...)
}

// TaskWaitAllC returns a channel that is closed only when all the specified
// tasks are completed.
func TaskWaitAllC(taskIDs ...int) <-chan int {
	return taskService.TaskWaitAllC(taskIDs...)
}
