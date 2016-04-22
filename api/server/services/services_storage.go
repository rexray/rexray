package services

import (
	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
	"github.com/emccode/libstorage/api/types/services"
)

type storageService struct {
	name          string
	driver        drivers.RemoteStorageDriver
	config        gofig.Config
	taskExecQueue chan *task
}

func (s *storageService) Init(config gofig.Config) error {
	s.config = config

	if err := s.initStorageDriver(); err != nil {
		return err
	}

	s.taskExecQueue = make(chan *task)
	go func() {
		for t := range s.taskExecQueue {
			execTask(t)
		}
	}()
	return nil
}

func (s *storageService) initStorageDriver() error {
	driverName := s.config.GetString("driver")
	if driverName == "" {
		driverName = s.config.GetString("libstorage.driver")
		if driverName == "" {
			return goof.WithField(
				"service", s.name, "error getting driver name")
		}
	}

	driver, err := registry.NewRemoteStorageDriver(driverName)
	if err != nil {
		return err
	}

	if err := driver.Init(s.config); err != nil {
		return err
	}

	s.driver = driver
	return nil
}

func (s *storageService) Config() gofig.Config {
	return s.config
}

func (s *storageService) Driver() drivers.RemoteStorageDriver {
	return s.driver
}

func (s *storageService) TaskExecute(
	ctx context.Context,
	run services.StorageTaskRunFunc,
	schema []byte) *types.Task {

	t := newStorageServiceTask(ctx, run, s, schema)
	go func() { s.taskExecQueue <- t }()
	return &t.Task
}

func (s *storageService) Name() string {
	return s.name
}
