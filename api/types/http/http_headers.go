package http

const (
	// InstanceIDHeader is the HTTP header that contains an InstanceID.
	InstanceIDHeader = "libstorage-instanceid"

	// InstanceID64Header is the HTTP header that contains a base64-encoded
	// InstanceID.
	InstanceID64Header = "libstorage-instanceid64"

	// LocalDevicesHeader is the HTTP header that contains a local device pair.
	LocalDevicesHeader = "libstorage-localdevices"

	// TransactionIDHeader is the HTTP header that contains the transaction ID
	// sent from the client.
	TransactionIDHeader = "libstorage-txid"

	// TransactionCreatedHeader is the HTTP header that contains the UTC
	// epoch of the time that the transaction was created.
	TransactionCreatedHeader = "libstorage-txcr"

	// ServerNameHeader is the HTTP header that contains the randomly generated
	// name the server creates for unique identification when the server starts
	// for the first time. This header is provided with every response sent
	// from the server.
	ServerNameHeader = "libstorage-servername"
)
