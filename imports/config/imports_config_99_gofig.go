// +build gofig

package config

import (
	"os"
	"runtime"

	log "github.com/Sirupsen/logrus"
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
	"github.com/codedellemc/libstorage/api/types"
)

const (
	logStdoutDesc = "The file to which to log os.Stdout"
	logStderrDesc = "The file to which to log os.Stderr"
)

func init() {
	gofigCore.LogGetAndSet = false
	gofigCore.LogSecureKey = false
	gofigCore.LogFlattenEnvVars = false

	logLevelSz := os.Getenv("LIBSTORAGE_LOGGING_LEVEL")
	logLevel, err := log.ParseLevel(logLevelSz)
	if err != nil {
		logLevel = log.WarnLevel
	}
	log.SetLevel(logLevel)

	r := gofigCore.NewRegistration("libStorage")

	rk := func(
		keyType gofig.ConfigKeyTypes,
		defaultVal interface{},
		description string,
		keyVal types.ConfigKey,
		args ...interface{}) {

		if args == nil {
			args = []interface{}{keyVal}
		} else {
			args = append([]interface{}{keyVal}, args...)
		}

		r.Key(keyType, "", defaultVal, description, args...)
	}

	defaultAEM := types.UnixEndpoint.String()
	defaultStorageDriver := types.LibStorageDriverName
	defaultLogLevel := logLevel.String()
	defaultClientType := types.IntegrationClient.String()

	rk(gofig.String, "", "", types.ConfigHost)
	rk(gofig.String, "", "", types.ConfigService)
	rk(gofig.String, defaultAEM, "", types.ConfigServerAutoEndpointMode)
	rk(gofig.String, runtime.GOOS, "", types.ConfigOSDriver)
	rk(gofig.String, defaultStorageDriver, "", types.ConfigStorageDriver)
	rk(gofig.String, defaultIntDriver, "", types.ConfigIntegrationDriver)
	rk(gofig.String, defaultClientType, "", types.ConfigClientType)
	rk(gofig.String, defaultLogLevel, "", types.ConfigLogLevel)
	rk(gofig.String, "", logStdoutDesc, types.ConfigLogStderr)
	rk(gofig.String, "", logStderrDesc, types.ConfigLogStdout)
	rk(gofig.Bool, false, "", types.ConfigLogHTTPRequests)
	rk(gofig.Bool, false, "", types.ConfigLogHTTPResponses)
	rk(gofig.Bool, false, "", types.ConfigHTTPDisableKeepAlive)
	rk(gofig.Int, 300, "", types.ConfigHTTPWriteTimeout)
	rk(gofig.Int, 300, "", types.ConfigHTTPReadTimeout)
	rk(gofig.String, types.LSX.String(), "", types.ConfigExecutorPath)
	rk(gofig.Bool, false, "", types.ConfigExecutorNoDownload)
	rk(gofig.Bool, false, "", types.ConfigIgVolOpsMountPreempt)
	rk(gofig.Bool, false, "", types.ConfigIgVolOpsCreateDisable)
	rk(gofig.Bool, false, "", types.ConfigIgVolOpsRemoveDisable)
	rk(gofig.Bool, false, "", types.ConfigIgVolOpsUnmountIgnoreUsed)
	rk(gofig.Bool, true, "", types.ConfigIgVolOpsPathCacheEnabled)
	rk(gofig.Bool, true, "", types.ConfigIgVolOpsPathCacheAsync)
	rk(gofig.String, "30m", "", types.ConfigClientCacheInstanceID)
	rk(gofig.String, "30s", "", types.ConfigDeviceAttachTimeout)
	rk(gofig.Int, 0, "", types.ConfigDeviceScanType)
	rk(gofig.Bool, false, "", types.ConfigEmbedded)
	rk(gofig.String, "1m", "", types.ConfigServerTasksExeTimeout)
	rk(gofig.String, "0s", "", types.ConfigServerTasksLogTimeout)
	rk(gofig.Bool, false, "", types.ConfigServerParseRequestOpts)

	// auth config - client
	rk(gofig.String, "", "", types.ConfigClientAuthToken)

	// auth config - server
	rk(gofig.String, "", "", types.ConfigServerAuthKey)
	rk(gofig.String, "HS256", "", types.ConfigServerAuthAlg)
	rk(gofig.String, "", "", types.ConfigServerAuthAllow)
	rk(gofig.String, "", "", types.ConfigServerAuthDeny)
	rk(gofig.Bool, false, "", types.ConfigServerAuthDisabled)

	gofigCore.Register(r)
}
