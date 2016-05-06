package types

// ConfigKey is a configuration key.
type ConfigKey string

const (
	// ConfigRoot is a config key.
	ConfigRoot = "libstorage"

	// ConfigServer is a config key.
	ConfigServer = ConfigRoot + ".server"

	// ConfigClient is a config key.
	ConfigClient = ConfigRoot + ".client"

	// ConfigHost is a config key.
	ConfigHost = ConfigRoot + ".host"

	// ConfigEmbedded is a config key.
	ConfigEmbedded = ConfigRoot + ".embedded"

	// ConfigService is a config key.
	ConfigService = ConfigRoot + ".service"

	// ConfigOSDriver is a config key.
	ConfigOSDriver = ConfigRoot + ".os.driver"

	// ConfigStorageDriver is a config key.
	ConfigStorageDriver = ConfigRoot + ".storage.driver"

	// ConfigIntegrationDriver is a config key.
	ConfigIntegrationDriver = ConfigRoot + ".integration.driver"

	// ConfigLogLevel is a config key.
	ConfigLogLevel = ConfigRoot + ".logging.level"

	// ConfigLogStdout is a config key.
	ConfigLogStdout = ConfigRoot + ".logging.stdout"

	// ConfigLogStderr is a config key.
	ConfigLogStderr = ConfigRoot + ".logging.stderr"

	// ConfigLogHTTPRequests is a config key.
	ConfigLogHTTPRequests = ConfigRoot + ".logging.httpRequests"

	// ConfigLogHTTPResponses is a config key.
	ConfigLogHTTPResponses = ConfigRoot + ".logging.httpResponses"

	// ConfigHTTPDisableKeepAlive is a config key.
	ConfigHTTPDisableKeepAlive = ConfigRoot + ".http.disableKeepAlive"

	// ConfigHTTPWriteTimeout is a config key.
	ConfigHTTPWriteTimeout = ConfigRoot + ".http.writeTimeout"

	// ConfigHTTPReadTimeout is a config key.
	ConfigHTTPReadTimeout = ConfigRoot + ".http.readTimeout"

	// ConfigServices is a config key.
	ConfigServices = ConfigServer + ".services"

	// ConfigEndpoints is a config key.
	ConfigEndpoints = ConfigServer + ".endpoints"

	// ConfigExecutorPath is a config key.
	ConfigExecutorPath = ConfigRoot + ".executor.path"

	// ConfigExecutorNoDownload is a config key.
	ConfigExecutorNoDownload = ConfigRoot + ".executor.disableDownload"

	// ConfigVolMountPreempt is a config key.
	ConfigVolMountPreempt = ConfigRoot + ".volume.mount.preempt"

	// ConfigVolCreateDisable is a config key.
	ConfigVolCreateDisable = ConfigRoot + ".volume.mount.disable"

	// ConfigVolRemoveDisable is a config key.
	ConfigVolRemoveDisable = ConfigRoot + ".volume.remove.disable"

	// ConfigVolUnmountIgnoreUsed is a config key.
	ConfigVolUnmountIgnoreUsed = ConfigRoot + ".volume.unmount.ignoreusedcount"

	// ConfigVolPathCache is a config key.
	ConfigVolPathCache = ConfigRoot + ".volume.path.cache"

	// ConfigClientCacheInstanceID is a config key.
	ConfigClientCacheInstanceID = ConfigClient + ".cache.instanceID"

	// ConfigTLS is a config key.
	ConfigTLS = ConfigRoot + ".tls"

	// ConfigTLSDisabled is a config key.
	ConfigTLSDisabled = ConfigTLS + ".disabled"

	// ConfigTLSServerName is a config key.
	ConfigTLSServerName = ConfigTLS + ".serverName"

	// ConfigTLSClientCertRequired is a config key.
	ConfigTLSClientCertRequired = ConfigTLS + ".clientCertRequired"

	// ConfigTLSTrustedCertsFile is a config key.
	ConfigTLSTrustedCertsFile = ConfigTLS + ".trustedCertsFile"

	// ConfigTLSCertFile is a config key.
	ConfigTLSCertFile = ConfigTLS + ".certFile"

	// ConfigTLSKeyFile is a config key.
	ConfigTLSKeyFile = ConfigTLS + ".keyFile"

	// ConfigDeviceAttachTimeout is a config key.
	ConfigDeviceAttachTimeout = ConfigRoot + ".device.attachTimeout"

	// ConfigDeviceScanType is a config key.
	ConfigDeviceScanType = ConfigRoot + ".device.scanType"
)
