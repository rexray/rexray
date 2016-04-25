package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"io"
	"os"
	"strconv"

	// imported to load routers
	_ "github.com/emccode/libstorage/api/server/router"

	// imported to load drivers
	_ "github.com/emccode/libstorage/drivers"
)

// RunSync runs the server and blocks until a Kill signal is received by the
// owner process or the server returns an error via its error channel.
func RunSync(host string, tls bool, driversAndServices ...string) error {
	_, _, errs := Run(host, tls, driversAndServices...)
	err := <-errs
	return err
}

// Run runs the server and returns a channel when errors occur runs until a Kill
// signal is received by the owner process or the server returns an error via
// its error channel.
func Run(host string, tls bool, driversAndServices ...string) (
	gofig.Config, io.Closer, <-chan error) {

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

// RunWithConfigSync runs the server by specifying a configuration object
// and blocks until a Kill signal is received by the owner process or the
// server returns an error via its error channel.
func RunWithConfigSync(config gofig.Config) error {
	_, errs := RunWithConfig(config)
	err := <-errs
	return err
}

// RunWithConfig runs the server by specifying a configuration object and
// returns a channel when errors occur runs until a Kill signal is received
// by the owner process or the server returns an error via its error channel.
func RunWithConfig(config gofig.Config) (io.Closer, <-chan error) {
	return startWithConfig(config)
}
