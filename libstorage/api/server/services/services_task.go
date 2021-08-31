package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"

	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils/schema"
)

type task struct {
	types.Task
	ctx                           types.Context
	runFunc                       types.TaskRunFunc
	storRunFunc                   types.StorageTaskRunFunc
	storService                   types.StorageService
	resultSchema                  []byte
	resultSchemaValidationEnabled bool
	done                          chan int
}

func newTask(ctx types.Context, schema []byte) *task {
	t := getTaskService(ctx).taskTrack(ctx)
	t.resultSchema = schema
	t.done = make(chan int)
	return t
}

func newGenericTask(
	ctx types.Context,
	run types.TaskRunFunc,
	schema []byte) *task {

	t := newTask(ctx, schema)
	t.runFunc = run
	return t
}

func newStorageServiceTask(
	ctx types.Context,
	run types.StorageTaskRunFunc,
	svc types.StorageService,
	schema []byte) *task {

	t := newTask(ctx, schema)
	t.storRunFunc = run
	t.storService = svc
	return t
}

func execTask(t *task) {
	defer func() {
		t.CompleteTime = time.Now().Unix()
		if t.Error != nil {
			t.ctx.Error(t.Error)
			t.State = types.TaskStateError
		} else {
			t.State = types.TaskStateSuccess
		}
		close(t.done)
		t.ctx.Debug("task completed")
	}()

	t.State = types.TaskStateRunning
	t.StartTime = time.Now().Unix()

	t.ctx.Info("executing task")

	if t.storRunFunc != nil && t.storService != nil {
		t.Result, t.Error = t.storRunFunc(t.ctx, t.storService)
	} else if t.runFunc != nil {
		t.Result, t.Error = t.runFunc(t.ctx)
	} else {
		t.Error = goof.New("invalid task")
	}

	if t.Error != nil {
		return
	}

	if t.Result == nil {
		t.ctx.Debug("skipping response schema validation; result == nil")
		return
	}

	if t.resultSchema == nil {
		t.ctx.Debug("skipping response schema validation; schema == nil")
		return
	}

	if !t.resultSchemaValidationEnabled {
		t.ctx.Debug("skipping response schema validation; disabled")
		return
	}

	var buf []byte
	if buf, t.Error = json.Marshal(t.Result); t.Error != nil {
		return
	}

	t.Error = schema.Validate(t.ctx, t.resultSchema, buf)
	if t.Error != nil {
		return
	}
}

type globalTaskService struct {
	sync.RWMutex
	name                          string
	config                        gofig.Config
	tasks                         map[int]*task
	resultSchemaValidationEnabled bool
}

// Init initializes the service.
func (s *globalTaskService) Init(ctx types.Context, config gofig.Config) error {
	s.tasks = map[int]*task{}
	s.config = config

	s.resultSchemaValidationEnabled = config.GetBool(
		types.ConfigSchemaResponseValidationEnabled)
	ctx.WithField("enabled", s.resultSchemaValidationEnabled).Debug(
		"configured result schema validation")

	return nil
}

func (s *globalTaskService) Name() string {
	return s.name
}

// Tasks returns a channel on which all tasks are received.
func (s *globalTaskService) Tasks() <-chan *types.Task {
	tasks := []*types.Task{}
	s.RLock()
	for _, v := range s.tasks {
		tasks = append(tasks, &v.Task)
	}
	s.RUnlock()

	c := make(chan *types.Task)
	go func() {
		for _, t := range tasks {
			c <- t
		}
		close(c)
	}()
	return c
}

// TaskTrack creates a new, trackable task.
func (s *globalTaskService) TaskTrack(ctx types.Context) *types.Task {
	return &s.taskTrack(ctx).Task
}
func (s *globalTaskService) taskTrack(ctx types.Context) *task {

	now := time.Now().Unix()
	s.RLock()
	taskID := len(s.tasks)
	s.RUnlock()

	t := &task{
		Task: types.Task{
			ID:        taskID,
			QueueTime: now,
		},
		resultSchemaValidationEnabled: s.resultSchemaValidationEnabled,
		ctx: ctx.WithValue(context.TaskKey, fmt.Sprintf("%d", taskID)),
	}

	s.Lock()
	s.tasks[taskID] = t
	s.Unlock()

	return t
}

// TaskExecute enqueues a task for execution.
func (s *globalTaskService) TaskEnqueue(
	ctx types.Context,
	run types.TaskRunFunc,
	schema []byte) *types.Task {

	t := newGenericTask(ctx, run, schema)
	go func() { execTask(t) }()
	return &t.Task
}

// TaskInspect returns the task with the specified ID.
func (s *globalTaskService) TaskInspect(taskID int) *types.Task {
	s.RLock()
	defer s.RUnlock()
	if t, ok := s.tasks[taskID]; ok {
		return &t.Task
	}
	return nil
}

// TaskWait blocks until the specified task is completed.
func (s *globalTaskService) TaskWait(taskID int) {
	<-s.TaskWaitC(taskID)
}

// TaskWait returns a channel that is closed only when the specified task is
// completed.
func (s *globalTaskService) TaskWaitC(taskID int) <-chan int {
	c := make(chan int)

	go func() {
		defer close(c)

		s.RLock()
		t, ok := s.tasks[taskID]
		s.RUnlock()

		if !ok {
			return
		}

		// remove the task from the queue after a configured about of time
		defer s.taskRemoveAfter(t)

		// signal that the task is complete
		<-t.done
	}()

	return c
}

// taskRemoveAfter tells the task service to remove the task after the duration
// specified by `libstorage.server.tasks.logTimeout`.
func (s *globalTaskService) taskRemoveAfter(t *task) {
	go func() {
		logTimeoutDur, err := time.ParseDuration(
			s.config.GetString(types.ConfigServerTasksLogTimeout))
		if err != nil {
			logTimeoutDur = time.Duration(time.Second * 60)
		}

		// wait to remove the task
		time.Sleep(logTimeoutDur)

		// sync access to the task map for querying its size before and after
		// executing the delete operation on it
		s.Lock()
		defer s.Unlock()

		t.ctx.WithFields(log.Fields{
			"removedAfter": logTimeoutDur,
			"tasksLen":     len(s.tasks),
		}).Debug("removing task")

		// delete the task
		delete(s.tasks, t.ID)

		t.ctx.WithField("tasksLen", len(s.tasks)).Debug("removed task")
	}()
}

// TaskWaitAll blocks until all the specified task are complete.
func (s *globalTaskService) TaskWaitAll(taskIDs ...int) {
	<-s.TaskWaitAllC(taskIDs...)
}

// TaskWaitAllC returns a channel that is closed only when all the specified
// tasks are completed.
func (s *globalTaskService) TaskWaitAllC(taskIDs ...int) <-chan int {
	c := make(chan int)

	go func() {
		defer close(c)
		wg := &sync.WaitGroup{}
		for _, tid := range taskIDs {
			wg.Add(1)
			go func(tid int) {
				s.TaskWait(tid)
				wg.Done()
			}(tid)
		}
		wg.Wait()
	}()

	return c
}
