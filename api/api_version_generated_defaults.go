package api

import (
	"fmt"
	"runtime"
	"time"

	"github.com/codedellemc/libstorage/api/types"
)

// Version of the current REST API. If not initialized the Arch and
// BuildTimestamp will reflect the target GOOS-GOARCH and current time
// respectively. The remaining fields will be empty strings.
var Version = &types.VersionInfo{
	Arch:           fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH),
	BuildTimestamp: time.Now().UTC(),
}
