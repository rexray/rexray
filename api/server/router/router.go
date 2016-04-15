package router

import (
	// imports to load packges
	_ "github.com/emccode/libstorage/api/server/router/executor"
	_ "github.com/emccode/libstorage/api/server/router/root"
	_ "github.com/emccode/libstorage/api/server/router/service"
	_ "github.com/emccode/libstorage/api/server/router/snapshot"
	_ "github.com/emccode/libstorage/api/server/router/tasks"
	_ "github.com/emccode/libstorage/api/server/router/volume"
)
