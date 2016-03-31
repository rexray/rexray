package servers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage"
)

// Run runs the server and blocks until a Kill signal is received by the
// owner process or the server returns an error via its error channel.
func Run(driverName, host string, tls bool) {

	// make sure all servers get closed even if the test is abrubptly aborted
	trapAbort()

	if debug, _ := strconv.ParseBool(os.Getenv("LIBSTORAGE_DEBUG")); debug {
		log.SetLevel(log.DebugLevel)
		os.Setenv("LIBSTORAGE_SERVER_HTTP_LOGGING_ENABLED", "true")
		os.Setenv("LIBSTORAGE_SERVER_HTTP_LOGGING_LOGREQUEST", "true")
		os.Setenv("LIBSTORAGE_SERVER_HTTP_LOGGING_LOGRESPONSE", "true")
		os.Setenv("LIBSTORAGE_CLIENT_HTTP_LOGGING_ENABLED", "true")
		os.Setenv("LIBSTORAGE_CLIENT_HTTP_LOGGING_LOGREQUEST", "true")
		os.Setenv("LIBSTORAGE_CLIENT_HTTP_LOGGING_LOGRESPONSE", "true")
	}

	serve(driverName, host, tls)
}

func trapAbort() {
	// make sure all servers get closed even if the test is abrubptly aborted
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		fmt.Println("received abort signal")
		closeAllServers()
		fmt.Println("all servers closed")
		os.Exit(1)
	}()
}

var servers []io.Closer

func closeAllServers() bool {
	noErrors := true
	for _, server := range servers {
		if err := server.Close(); err != nil {
			noErrors = false
			fmt.Printf("error closing server: %v\n", err)
		}
	}
	return noErrors
}

func serve(driverName, host string, tls bool) {

	if host == "" {
		host = fmt.Sprintf("tcp://localhost:%d", gotil.RandomTCPPort())
	}
	config := getConfig(driverName, host, tls)
	server, errs := libstorage.Serve(config)
	if server != nil {
		servers = append(servers, server)
	}
	<-errs
}

func getConfig(driverName, host string, tls bool) gofig.Config {
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
		host, "/tmp/libstorage/executors",
		clientTLS, serverTLS,
		driverName, driverName)

	configYamlBuf := []byte(configYaml)
	if err := config.ReadConfig(bytes.NewReader(configYamlBuf)); err != nil {
		panic(err)
	}
	return config
}

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
	   05 - the first service name
	   06 - the first service's driver type
	*/
	libStorageConfigBase = `
libstorage:
  host: %[1]s
  driver: invalidDriverName
  executorsDir: %[2]s
  profiles:
    enabled: true
    groups:
    - local=127.0.0.1%[3]s
  server:
    endpoints:
      localhost:
        address: %[1]s%[4]s
    services:
      %[5]s:
        libstorage:
          driver: %[6]s
          profiles:
            enabled: true
            groups:
            - remote=127.0.0.1
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
