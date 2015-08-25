// +build !exclude_admin

package daemon

import (
	_ "github.com/emccode/rexray/daemon/module/admin"
	_ "github.com/emccode/rexray/daemon/module/docker/volumedriver"
)
