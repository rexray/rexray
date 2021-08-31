package services

import (
	"fmt"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
)

type storageService struct {
	name          string
	driver        types.StorageDriver
	config        gofig.Config
	authConfig    *types.AuthConfig
	taskExecQueue chan *task
}

func (s *storageService) Init(ctx types.Context, config gofig.Config) error {
	s.config = config

	if err := s.initStorageDriver(ctx); err != nil {
		return err
	}

	s.taskExecQueue = make(chan *task)
	go func() {
		for t := range s.taskExecQueue {
			execTask(t)
		}
	}()

	authFields := map[string]interface{}{}
	authConfig, err := utils.ParseAuthConfig(
		ctx, config, authFields,
		fmt.Sprintf("libstorage.server.services.%s", s.name))
	if err != nil {
		return err
	}
	s.authConfig = authConfig
	if s.authConfig != nil {
		ctx.WithFields(authFields).Info("configured service auth")
	}

	return nil
}

func (s *storageService) initStorageDriver(ctx types.Context) error {
	driverName := s.config.GetString("driver")
	if driverName == "" {
		driverName = s.config.GetString("libstorage.driver")
		if driverName == "" {
			driverName = s.config.GetString("libstorage.storage.driver")
			if driverName == "" {
				return goof.WithField(
					"service", s.name, "error getting driver name")
			}
		}
	}

	ctx.WithField("driverName", driverName).Debug("got driver name")
	driver, err := registry.NewStorageDriver(driverName)
	if err != nil {
		return err
	}

	ctx = ctx.WithValue(context.DriverKey, driver)

	if err := driver.Init(ctx, s.config); err != nil {
		return err
	}

	s.driver = driver
	return nil
}

func (s *storageService) Config() gofig.Config {
	return s.config
}

func (s *storageService) AuthConfig() *types.AuthConfig {
	return s.authConfig
}

func (s *storageService) Driver() types.StorageDriver {
	return s.driver
}

func (s *storageService) TaskEnqueue(
	ctx types.Context,
	run types.StorageTaskRunFunc,
	schema []byte) *types.Task {

	t := newStorageServiceTask(ctx, run, s, schema)
	go func() { s.taskExecQueue <- t }()
	return &t.Task
}

func (s *storageService) Name() string {
	return s.name
}
