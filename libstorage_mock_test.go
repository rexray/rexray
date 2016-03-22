// +build mock

package libstorage

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/drivers/storage/mock"
)

func getConfig(host string, tls bool, t *testing.T) gofig.Config {
	if host == "" {
		host = "tcp://127.0.0.1:7979"
	}
	config := gofig.New()

	var clientTLS, serverTLS string
	if tls {
		clientTLS = fmt.Sprintf(
			libStorageConfigClientTLS,
			clientCrt, clientKey, trustedCerts)
		serverTLS = fmt.Sprintf(
			libStorageConfigServerTLS,
			serverCrt, serverKey, trustedCerts)
	}
	configYaml := fmt.Sprintf(
		libStorageConfigBase,
		host,
		clientToolDir, localDevicesFile,
		clientTLS, serverTLS,
		testServer1Name, mock.Driver1Name,
		testServer2Name, mock.Driver2Name,
		testServer3Name, mock.Driver3Name)

	configYamlBuf := []byte(configYaml)
	if err := config.ReadConfig(bytes.NewReader(configYamlBuf)); err != nil {
		panic(err)
	}
	return config
}

const (
	/*
	   libStorageConfigBase is the base config for tests

	   01 - the host address to server and which the client uses
	   02 - the executors directory
	   03 - the local devices file
	   04 - the client TLS section. use an empty string if TLS is disabled
	   05 - the server TLS section. use an empty string if TLS is disabled
	   06 - the first service name
	   07 - the first service's driver type
	   08 - the second service name
	   09 - the second service's driver type
	   10 - the third service name
	   11 - the third service's driver type
	*/
	libStorageConfigBase = `
libstorage:
  host: %[1]s
  driver: invalidDriverName
  executorsDir: %[2]s
  profiles:
    enabled: true
    groups:
    - local=127.0.0.1
  client:
    localdevicesfile: %[3]s%[4]s
  server:
    endpoints:
      localhost:
        address: %[1]s%[5]s
    services:
      %[6]s:
        libstorage:
          driver: %[7]s
          profiles:
            enabled: true
            groups:
            - remote=127.0.0.1
      %[8]s:
        libstorage:
          driver: %[9]s
      %[10]s:
        libstorage:
          driver: %[11]s
`

	libStorageConfigClientTLS = `
    tls:
      serverName: libstorage-server
      certFile: %s
      keyFile: %s
      trustedCertsFile: %s
`

	libStorageConfigServerTLS = `
        tls:
          serverName: libstorage-server
          certFile: %s
          keyFile: %s
          trustedCertsFile: %s
          clientCertRequired: true
`
)
