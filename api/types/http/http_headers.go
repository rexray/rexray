package http

const (
	// InstanceIDHeader is the HTTP header that contains an InstanceID.
	InstanceIDHeader = "libstorage-instanceid"

	// InstanceID64Header is the HTTP header that contains a base64-encoded
	// InstanceID.
	InstanceID64Header = "libstorage-instanceid64"

	// LocalDevicesHeader is the HTTP header that contains a local device pair.
	LocalDevicesHeader = "libstorage-localdevices"

	// ServerNameHeader is the HTTP header that contains the randomly generated
	// name the server creates for unique identification when the server starts
	// for the first time. This header is provided with every response sent
	// from the server.
	ServerNameHeader = "libstorage-servername"
)
