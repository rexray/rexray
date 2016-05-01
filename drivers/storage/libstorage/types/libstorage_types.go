package types

import (
	apiclient "github.com/emccode/libstorage/api/client"
	"github.com/emccode/libstorage/api/types"
)

// Driver is the libStorage storage driver interface.
type Driver interface {
	types.StorageDriver

	// API returns the underlying API client.
	API() Client
}

// Client is the libStorage storage driver's client extensions.
type Client interface {
	apiclient.APIClient

	// EnableInstanceIDHeaders is a flag indicating whether or not the
	// client will automatically send the instance ID header(s) along with
	// storage-related API requests. The default is enabled.
	EnableInstanceIDHeaders(enabled bool)

	// EnableLocalDevicesHeaders is a flag indicating whether or not the
	// client will automatically send the local devices header(s) along with
	// storage-related API requests. The default is enabled.
	EnableLocalDevicesHeaders(enabled bool)

	// ServerName returns the name of the server to which the client is
	// connected. This is not the same as the host name, rather it's the
	// randomly generated name the server creates for unique identification
	// when the server starts for the first time.
	ServerName() string

	// InstanceID gets the client's instance ID.
	InstanceID(ctx types.Context, service string) (*types.InstanceID, error)

	// LocalDevices gets the client's local devices map.
	LocalDevices(ctx types.Context, service string) (map[string]string, error)

	// NextDevice gets the next available device ID.
	NextDevice(ctx types.Context, service string) (string, error)
}
