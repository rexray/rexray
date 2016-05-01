/*
Package server is the homeo of the reference implementation of the  libStorage
API server. There are five functions that will start a server instance:

    1. Run
    2. Start
    3. RunWithConfig
    4. StartWithConfig
    5. Serve

The Run functions start a server and block until a signal is received by the
owner process. The Start functions will start a server and return a channel
on which any server errors are returned. This channel can be used to block or
ignored to start a server asynchronously.

The Serve function is ultimately what the above functions invoke, but with an
important distinction. The above functions 1-4 all track the servers that are
started inside a single process and upon the process's abrupt termination will
enable the graceful shutdown of all running/started server instances. However,
the Server function is the low-level method for creating and running a server,
and it's up to the end-user to do any type of resource tracking in order to
enable graceful shutdowns if that method is used directly.
*/
package server

import (
	"io"
	"os"
	"strconv"

	"github.com/akutz/gofig"

	// imported to load routers
	_ "github.com/emccode/libstorage/imports/routers"

	// imported to load remote storage drivers
	_ "github.com/emccode/libstorage/imports/remote"
)

// IsNil returns a flag indicating whether a server returned by the Serve
// function is nil.
func IsNil(closer io.Closer) bool {
	return closer == (*server)(nil)
}

// Run runs the server and blocks until a Kill signal is received by the
// owner process or the server returns an error via its error channel.
func Run(host string, tls bool, driversAndServices ...string) error {
	_, _, err, errs := Start(host, tls, driversAndServices...)
	if err != nil {
		return err
	}
	err = <-errs
	return err
}

// Start starts the server and returns a channel when errors occur runs until a
// Kill signal is received by the owner process or the server returns an error
// via its error channel.
func Start(host string, tls bool, driversAndServices ...string) (
	gofig.Config, io.Closer, error, <-chan error) {

	if runHost := os.Getenv("LIBSTORAGE_RUN_HOST"); runHost != "" {
		host = runHost
	}
	if runTLS, err := strconv.ParseBool(
		os.Getenv("LIBSTORAGE_RUN_TLS")); err != nil {
		tls = runTLS
	}

	return start(host, tls, driversAndServices...)
}

// RunWithConfig runs the server by specifying a configuration object
// and blocks until a Kill signal is received by the owner process or the
// server returns an error via its error channel.
func RunWithConfig(config gofig.Config) error {
	_, err, errs := startWithConfig(config)
	if err != nil {
		return err
	}
	err = <-errs
	return err
}

// StartWithConfig starts the server by specifying a configuration object and
// returns a channel when errors occur runs until a Kill signal is received
// by the owner process or the server returns an error via its error channel.
func StartWithConfig(config gofig.Config) (io.Closer, error, <-chan error) {
	return startWithConfig(config)
}
