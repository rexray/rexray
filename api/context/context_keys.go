package context

// Key is the type used as a context key.
type Key int

const (
	_ Key = -1 - iota

	// LoggerKey is a context key.
	LoggerKey

	// HTTPRequestKey is a context key.
	HTTPRequestKey

	// AllInstanceIDsKey is the key for the map[string]*types.InstanceID value
	// that maps all drivers to their instance IDs.
	AllInstanceIDsKey

	// LocalDevicesKey is a context key.
	LocalDevicesKey

	// AllLocalDevicesKey is the key for the map[string]*types.LocalDevices
	// value that maps all drivers to their instance IDs.
	AllLocalDevicesKey

	// keyLoggable is the minimum value from which the succeeding keys should
	// be checked when logging.
	keyLoggable

	// ClientKey is a context key.
	ClientKey

	// TaskKey is a context key.
	TaskKey

	// InstanceIDKey is a context key.
	InstanceIDKey

	// ProfileKey is a context key.
	ProfileKey

	// RouteKey is a context key.
	RouteKey

	// ServerKey is a context key.
	ServerKey

	// ServiceKey is an alias for StorageService.
	ServiceKey

	// StorageServiceKey is a context key.
	StorageServiceKey

	// TransactionKey is a context key.
	TransactionKey

	// DriverKey is an alias for StorageDriver.
	DriverKey

	// UserKey is a context key.
	UserKey

	// HostKey is a context key.
	HostKey

	// TLSKey is a context key.
	TLSKey

	// keyEOF should always be the final key
	keyEOF
)

// String returns the name of the context key.
func (k Key) String() string {
	if v, ok := keyNames[k]; ok {
		return v
	}
	return ""
}

var (
	keyNames = map[Key]string{
		TaskKey:           "task",
		InstanceIDKey:     "instanceID",
		ProfileKey:        "profile",
		RouteKey:          "route",
		ServerKey:         "server",
		ServiceKey:        "service",
		StorageServiceKey: "service",
		TransactionKey:    "tx",
		DriverKey:         "storageDriver",
		UserKey:           "user",
		HostKey:           "host",
		TLSKey:            "tls",
	}
)
