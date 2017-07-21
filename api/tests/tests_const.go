package tests

const (
	// configFileFormat is the default configuration file format used
	// by most the test specs. some specs do override this format with
	// their own; the tls tests for example
	configFileFormat = `
libstorage:
  host: %[1]s://%[2]s
  server:
    endpoints:
      localhost:
        address: %[1]s://%[2]s
    services:
      %[3]s:
        driver: %[3]s
`

	// protoUnix is the string that identifies the golang net unix protocol
	protoUnix = "unix"

	// localhostFingerprint is the fingerprint for the libStorage development
	// and test certificate at .tls/libstorage.crt.
	localhostFingerprint = "52:C7:5D:00:1B:E7:33:66:14:3C:47:07:77:" +
		"59:9C:94:F1:EA:76:00:41:B1:9D:71:0B:80:05:1F:F7:2D:6B:69"
)
