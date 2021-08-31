package tests

import (
	"fmt"
	"os"
	"path"

	apiserver "github.com/AVENTER-UG/rexray/libstorage/api/server"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

var (
	// v0ID is the expected ID of the volume used in the test operations.
	//
	// This value is configurable externally via the environment variable
	// LSX_TESTS_V0ID.
	v0ID = "vfs-000"

	// v0Name is the execpted name of the volume used in the test operations.
	//
	// This value is configurable externally via the environment variable
	// LSX_TESTS_V0NAME.
	v0Name = "v0"

	// v0NextDev is the expected name of the next available devie.
	//
	// This value is configurable externally via the environment variable
	// LSX_TESTS_V0NEXTDEVICE.
	v0NextDev = "/dev/xvda"

	tlsPath = path.Join(
		os.Getenv("GOPATH"),
		"/src/github.com/AVENTER-UG/rexray/libstorage/.tls")

	suiteServerCrt    = path.Join(tlsPath, "libstorage-server.crt")
	suiteServerKey    = path.Join(tlsPath, "libstorage-server.key")
	suiteClientCrt    = path.Join(tlsPath, "libstorage-client.crt")
	suiteClientKey    = path.Join(tlsPath, "libstorage-client.key")
	suiteTrustedCerts = path.Join(tlsPath, "libstorage-ca.crt")
	suiteKnownHosts   = path.Join(tlsPath, "known_hosts")
)

func init() {
	if v := os.Getenv("LSX_TESTS_V0ID"); v != "" {
		v0ID = v
	}
	if v := os.Getenv("LSX_TESTS_V0NAME"); v != "" {
		v0Name = v
	}
	if v := os.Getenv("LSX_TESTS_V0NEXTDEVICE"); v != "" {
		v0NextDev = v
	}

	//types.Stdout = GinkgoWriter
	apiserver.DisableStartupInfo = true

	if !types.Debug {
		if v := os.Getenv("LIBSTORAGE_LOGGING_LEVEL"); v == "" {
			os.Setenv("LIBSTORAGE_LOGGING_LEVEL", "panic")
		}
	}
}

func (t *testRunner) initConfigData() {
	configFileData := fmt.Sprintf(
		configFileFormat, t.proto, t.laddr, t.driverName)
	t.configFileData = []byte(configFileData)
}
