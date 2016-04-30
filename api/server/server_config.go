package server

import (
	"bytes"
	"fmt"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/akutz/gofig"
)

var (
	tlsPath = fmt.Sprintf(
		"%s/src/github.com/emccode/libstorage/.tls", os.Getenv("GOPATH"))
	serverCrt    = fmt.Sprintf("%s/libstorage-server.crt", tlsPath)
	serverKey    = fmt.Sprintf("%s/libstorage-server.key", tlsPath)
	clientCrt    = fmt.Sprintf("%s/libstorage-client.crt", tlsPath)
	clientKey    = fmt.Sprintf("%s/libstorage-client.key", tlsPath)
	trustedCerts = fmt.Sprintf("%s/libstorage-ca.crt", tlsPath)
)

const (
	/*
	   libStorageConfigBase is the base config for tests

	   01 - the host address to server and which the client uses
	   02 - the executors directory
	   03 - the client TLS section. use an empty string if TLS is disabled
	   04 - the server TLS section. use an empty string if TLS is disabled
	   05 - the services
	*/
	libStorageConfigBase = `
libstorage:
  host: %[1]s
  profiles:
    enabled: true
    groups:
    - local=127.0.0.1%[3]s
  server:
    endpoints:
      localhost:
        address: %[1]s%[4]s
    services:%[5]s
`

	libStorageConfigService = `
      %[1]s:
        libstorage:
          driver: %[2]s
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

// NewConfig returns a confiburation object from basic inputs
func NewConfig(
	host string, tls bool, driversAndServices ...string) gofig.Config {

	if host == "" {
		host = "tcp://127.0.0.1:7979"
	}
	config := gofig.New()

	if debug, _ := strconv.ParseBool(os.Getenv("LIBSTORAGE_DEBUG")); debug {
		log.SetLevel(log.DebugLevel)
		config.ReadConfig(bytes.NewReader([]byte(`libstorage:
  server:
    http:
      logging:
        enabled: true
        logrequest: true
        logresponse: true
  client:
    http:
      logging:
        enabled: true
        logrequest: true
        logresponse: true`)))
	}

	var clientTLS, serverTLS string
	if tls {
		clientTLS = fmt.Sprintf(
			libStorageConfigClientTLS,
			clientCrt, clientKey, trustedCerts)
		serverTLS = fmt.Sprintf(
			libStorageConfigServerTLS,
			serverCrt, serverKey, trustedCerts)
	}

	services := &bytes.Buffer{}

	for i := 0; i < len(driversAndServices); i = i + 2 {
		driverName := driversAndServices[i]
		serviceName := driverName
		if (i + 1) < len(driversAndServices) {
			serviceName = driversAndServices[i+1]
		}
		services.WriteString(
			fmt.Sprintf(libStorageConfigService, serviceName, driverName))
	}

	configYaml := fmt.Sprintf(
		libStorageConfigBase,
		host, "/tmp/libstorage/executors",
		clientTLS, serverTLS,
		services.String())

	log.Debug(configYaml)

	configYamlBuf := []byte(configYaml)
	if err := config.ReadConfig(bytes.NewReader(configYamlBuf)); err != nil {
		panic(err)
	}

	return config
}
