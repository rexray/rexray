package services

import (
	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
)

// Service is the base type for services.
type Service interface {

	// Name gets the name of the service.
	Name() string

	// Init initializes the service.
	Init(config gofig.Config) error
}

// TaskRunFunc is a function responsible for a task's execution.
type TaskRunFunc func(ctx context.Context) (interface{}, error)

// StorageTaskRunFunc is a function responsible for a storage-service task's
// execution.
type StorageTaskRunFunc func(
	ctx context.Context,
	service StorageService) (interface{}, error)

// StorageService is a service that provides the interaction with
// StorageDrivers.
type StorageService interface {
	Service

	// Driver returns the service's StorageDriver.
	Driver() drivers.StorageDriver

	// TaskExecute enqueues a task for execution.
	TaskExecute(
		ctx context.Context,
		run StorageTaskRunFunc,
		schema []byte) *types.Task
}

// TaskTrackingService a service for tracking tasks.
type TaskTrackingService interface {
	Service

	// Tasks returns a channel on which all tasks tracked via TrackTasks are
	// received.
	Tasks() <-chan *types.Task

	// TaskTrack creates a new, trackable task.
	TaskTrack(ctx context.Context) *types.Task

	// TaskExecute enqueues a task for execution.
	TaskExecute(
		ctx context.Context,
		run TaskRunFunc,
		schema []byte) *types.Task

	// TaskInspect returns the task with the specified ID.
	TaskInspect(taskID int) *types.Task

	// TaskWait blocks until the specified task completes.
	TaskWait(taskID int) <-chan int

	// TaskWaitAll blocks until all the specified tasks complete.
	TaskWaitAll(taskIDs ...int) <-chan int

	// TaskWaitC returns a channel that is closed when the specified task
	// completes.
	TaskWaitC(taskID int) <-chan int

	// TaskWaitAll returns a channel that is closed when the specified task
	// completes.
	TaskWaitAllC(taskIDs ...int) <-chan int
}

// TaskExecutionService is a service for executing tasks.
type TaskExecutionService interface {
	Service

	// TaskExecute enqueues a task for execution.
	TaskExecute(
		ctx context.Context,
		run TaskRunFunc,
		schema []byte) *types.Task
}
