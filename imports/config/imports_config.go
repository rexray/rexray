package config

import (
	"os"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils/paths"
)

const (
	logStdoutDesc = "The file to which to log os.Stdout"
	logStderrDesc = "The file to which to log os.Stderr"
)

func init() {
	gofig.LogGetAndSet = false
	gofig.LogSecureKey = false
	gofig.LogFlattenEnvVars = false

	logLevelSz := os.Getenv("LIBSTORAGE_LOGGING_LEVEL")
	logLevel, err := log.ParseLevel(logLevelSz)
	if err != nil {
		logLevel = log.WarnLevel
	}
	log.SetLevel(logLevel)

	paths.Init()

	lsxBinPath := paths.Lib.Join(types.LSX)

	r := gofig.NewRegistration("libStorage")

	rk := func(
		keyType gofig.KeyType,
		defaultVal interface{},
		description string,
		keyVal types.ConfigKey,
		args ...string) {

		if args == nil {
			args = []string{string(keyVal)}
		} else {
			args = append([]string{string(keyVal)}, args...)
		}

		r.Key(keyType, "", defaultVal, description, args...)
	}

	rk(gofig.String, "", "", types.ConfigHost)
	rk(gofig.String, "", "", types.ConfigService)
	rk(gofig.String, runtime.GOOS, "", types.ConfigOSDriver)
	rk(gofig.String, types.LibStorageDriverName, "", types.ConfigStorageDriver)
	rk(gofig.String, "docker", "", types.ConfigIntegrationDriver)
	rk(gofig.String, logLevel.String(), "", types.ConfigLogLevel)
	rk(gofig.String, "", logStdoutDesc, types.ConfigLogStderr)
	rk(gofig.String, "", logStderrDesc, types.ConfigLogStdout)
	rk(gofig.Bool, false, "", types.ConfigLogHTTPRequests)
	rk(gofig.Bool, false, "", types.ConfigLogHTTPResponses)
	rk(gofig.Bool, false, "", types.ConfigHTTPDisableKeepAlive)
	rk(gofig.Int, 300, "", types.ConfigHTTPWriteTimeout)
	rk(gofig.Int, 300, "", types.ConfigHTTPReadTimeout)
	rk(gofig.String, lsxBinPath, "", types.ConfigExecutorPath)
	rk(gofig.Bool, false, "", types.ConfigExecutorNoDownload)
	rk(gofig.Bool, false, "", types.ConfigVolMountPreempt)
	rk(gofig.Bool, false, "", types.ConfigVolCreateDisable)
	rk(gofig.Bool, false, "", types.ConfigVolRemoveDisable)
	rk(gofig.Bool, false, "", types.ConfigVolUnmountIgnoreUsed)
	rk(gofig.Bool, false, "", types.ConfigVolPathCache)
	rk(gofig.String, "30m", "", types.ConfigClientCacheInstanceID)
	rk(gofig.String, "30s", "", types.ConfigDeviceAttachTimeout)
	rk(gofig.Int, 0, "", types.ConfigDeviceScanType)
	rk(gofig.Bool, false, "", types.ConfigEmbedded)

	gofig.Register(r)
}
