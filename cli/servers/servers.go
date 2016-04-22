package servers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage"

	// load the drivers
	_ "github.com/emccode/libstorage/drivers"
)

// Run runs the server and blocks until a Kill signal is received by the
// owner process or the server returns an error via its error channel.
func Run(host string, tls bool, driversAndServices ...string) error {

	if runHost := os.Getenv("LIBSTORAGE_RUN_HOST"); runHost != "" {
		host = runHost
	}
	if runTLS, err := strconv.ParseBool(
		os.Getenv("LIBSTORAGE_RUN_TLS")); err != nil {
		tls = runTLS
	}

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

	return start(host, tls, driversAndServices...)
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

var (
	portLock = &sync.Mutex{}
)

func start(host string, tls bool, driversAndServices ...string) error {
	if host == "" {
		portLock.Lock()
		defer portLock.Unlock()

		port := 7979
		if !gotil.IsTCPPortAvailable(port) {
			port = gotil.RandomTCPPort()
		}
		host = fmt.Sprintf("tcp://localhost:%d", port)
	}

	config := getConfig(host, tls, driversAndServices...)
	server, errs := libstorage.Serve(config)

	if server != nil {
		servers = append(servers, server)
	}

	err := <-errs
	return err
}

func getConfig(
	host string, tls bool, driversAndServices ...string) gofig.Config {

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
