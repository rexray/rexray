package types

// All header names below follow the Golang canonical format for header keys.
// Please do not alter their casing to your liking or you will break stuff.
const (
	// InstanceIDHeader is the HTTP header that contains an InstanceID.
	InstanceIDHeader = "Libstorage-Instanceid"

	// InstanceID64Header is the HTTP header that contains a base64-encoded
	// InstanceID.
	InstanceID64Header = "Libstorage-Instanceid64"

	// LocalDevicesHeader is the HTTP header that contains a local device pair.
	LocalDevicesHeader = "Libstorage-Localdevices"

	// TransactionIDHeader is the HTTP header that contains the transaction ID
	// sent from the client.
	TransactionIDHeader = "Libstorage-Txid"

	// TransactionCreatedHeader is the HTTP header that contains the UTC
	// epoch of the time that the transaction was created.
	TransactionCreatedHeader = "Libstorage-Txcr"

	// ServerNameHeader is the HTTP header that contains the randomly generated
	// name the server creates for unique identification when the server starts
	// for the first time. This header is provided with every response sent
	// from the server.
	ServerNameHeader = "Libstorage-Servername"
)
