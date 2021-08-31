package config

import (
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

const (
	logStdoutDesc = "The file to which to log os.Stdout"
	logStderrDesc = "The file to which to log os.Stderr"
)

func init() {
	var defaultIntDriver string
	switch runtime.GOOS {
	case "linux":
		defaultIntDriver = runtime.GOOS
	}

	gofigCore.LogGetAndSet = false
	gofigCore.LogSecureKey = false
	gofigCore.LogFlattenEnvVars = false

	registry.RegisterConfigReg(
		"libStorage",
		func(ctx types.Context, r gofig.ConfigRegistration) {
			pathConfig := context.MustPathConfig(ctx)

			var lvl log.Level
			if types.Debug {
				lvl = log.DebugLevel
			} else {
				ll, err := log.ParseLevel(os.Getenv("LIBSTORAGE_LOGGING_LEVEL"))
				if err != nil {
					ll = log.WarnLevel
				}
				lvl = ll
			}

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
			defaultLogLevel := lvl.String()
			defaultClientType := types.IntegrationClient.String()

			rk(gofig.String, "", "", types.ConfigHost)
			rk(gofig.String, "", "", types.ConfigService)
			rk(gofig.String, defaultAEM, "", types.ConfigServerAutoEndpointMode)
			rk(gofig.String, runtime.GOOS, "", types.ConfigOSDriver)
			rk(gofig.String, defaultStorageDriver, "",
				types.ConfigStorageDriver)
			rk(gofig.String, defaultIntDriver, "",
				types.ConfigIntegrationDriver)
			rk(gofig.String, defaultClientType, "", types.ConfigClientType)
			rk(gofig.String, defaultLogLevel, "", types.ConfigLogLevel)
			rk(gofig.String, "", logStdoutDesc, types.ConfigLogStderr)
			rk(gofig.String, "", logStderrDesc, types.ConfigLogStdout)
			rk(gofig.Bool, types.Debug, "", types.ConfigLogHTTPRequests)
			rk(gofig.Bool, types.Debug, "", types.ConfigLogHTTPResponses)
			rk(gofig.Bool, false, "", types.ConfigHTTPDisableKeepAlive)
			rk(gofig.Int, 300, "", types.ConfigHTTPWriteTimeout)
			rk(gofig.Int, 300, "", types.ConfigHTTPReadTimeout)

			rk(gofig.Bool, false, "", types.ConfigExecutorNoDownload)
			rk(gofig.Bool, false, "", types.ConfigIgVolOpsMountPreempt)
			rk(gofig.Int, 0, "", types.ConfigIgVolOpsMountRetryCount)
			rk(gofig.String, "5s", "", types.ConfigIgVolOpsMountRetryWait)
			rk(gofig.Bool, false, "", types.ConfigIgVolOpsCreateDisable)
			rk(gofig.Bool, false, "", types.ConfigIgVolOpsRemoveDisable)
			rk(gofig.Bool, false, "", types.ConfigIgVolOpsRemoveForce)
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

			// tls config
			rk(
				gofig.String,
				pathConfig.DefaultTLSCertFile,
				"",
				types.ConfigTLSCertFile)
			rk(
				gofig.String,
				pathConfig.DefaultTLSKeyFile,
				"",
				types.ConfigTLSKeyFile)
			rk(
				gofig.String,
				pathConfig.DefaultTLSTrustedRootsFile,
				"",
				types.ConfigTLSTrustedCertsFile)
			rk(
				gofig.String,
				pathConfig.DefaultTLSKnownHosts,
				"",
				types.ConfigTLSKnownHosts)
			rk(gofig.String, "", "", types.ConfigTLSServerName)
			rk(gofig.String, "", "", types.ConfigTLSDisabled)
			rk(gofig.String, "", "", types.ConfigTLSInsecure)
			rk(gofig.String, "", "", types.ConfigTLSClientCertRequired)

			// auth config - client
			rk(gofig.String, "", "", types.ConfigClientAuthToken)

			// auth config - server
			rk(gofig.String, "", "", types.ConfigServerAuthKey)
			rk(gofig.String, "HS256", "", types.ConfigServerAuthAlg)
			rk(gofig.String, "", "", types.ConfigServerAuthAllow)
			rk(gofig.String, "", "", types.ConfigServerAuthDeny)
			rk(gofig.Bool, false, "", types.ConfigServerAuthDisabled)
		})
}
