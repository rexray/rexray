package routers

import (
	// imports to load routers
	_ "github.com/codedellemc/libstorage/api/server/router/executor"
	_ "github.com/codedellemc/libstorage/api/server/router/help"
	_ "github.com/codedellemc/libstorage/api/server/router/root"
	_ "github.com/codedellemc/libstorage/api/server/router/service"
	_ "github.com/codedellemc/libstorage/api/server/router/snapshot"
	_ "github.com/codedellemc/libstorage/api/server/router/tasks"
	_ "github.com/codedellemc/libstorage/api/server/router/volume"
)
