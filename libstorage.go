/*
Package libstorage provides a vendor agnostic storage orchestration model, API,
and reference client and server implementations. libStorage enables storage
consumption by leveraging methods commonly available, locally and/or externally,
to an operating system (OS).

The Past

The libStorage project and its architecture represents a culmination of
experience gained from the project authors' building of
several (http://bit.ly/1HIAet6) different storage (http://bit.ly/1Ya9Uft)
orchestration tools (https://github.com/emccode/rexray). While created using
different languages and targeting disparate storage platforms, all the tools
were architecturally aligned and embedded functionality directly inside the
tools and affected storage platforms.

This shared design goal enabled tools that natively consumed storage, sans
external dependencies.

The Present

Today libStorage focuses on adding value to container runtimes and storage
orchestration tools such as Docker and Mesos, however the libStorage
framework is available abstractly for more general usage across:

  * Operating systems
  * Storage platforms
  * Hardware platforms
  * Virtualization platforms

The client side implementation, focused on operating system activities,
has a minimal set of dependencies in order to avoid a large, runtime footprint.
*/
package libstorage

import (
	"io"

	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/server"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/client"
)

func init() {
	registerConfig()
}

// RegisterStorageDriver registers a new StorageDriver with the
// libStorage service.
func RegisterStorageDriver(
	name string, ctor types.NewStorageDriver) {
	registry.RegisterStorageDriver(name, ctor)
}

// RegisterOSDriver registers a new StorageDriver with the libStorage
// service.
func RegisterOSDriver(name string, ctor types.NewOSDriver) {
	registry.RegisterOSDriver(name, ctor)
}

// RegisterIntegrationDriver registers a new IntegrationDriver with the
// libStorage service.
func RegisterIntegrationDriver(name string, ctor types.NewIntegrationDriver) {
	registry.RegisterIntegrationDriver(name, ctor)
}

/*
Serve starts the reference implementation of a server hosting an
HTTP/JSON service that implements the libStorage API endpoint.

If the config parameter is nil a default instance is created. The
libStorage service is served at the address specified by the configuration
property libstorage.host.
*/
func Serve(config gofig.Config) (io.Closer, error, <-chan error) {
	return server.Serve(config)
}

// Dial opens a connection to a remote libStorage serice and returns the client
// that can be used to communicate with said endpoint.
//
// If the config parameter is nil a default instance is created. The
// function dials the libStorage service specified by the configuration
// property libstorage.host.
func Dial(config gofig.Config) (client.Client, error) {
	return client.New(config)
}

func registerConfig() {
	r := gofig.NewRegistration("libStorage")
	r.Key(gofig.String, "", "", "", "libstorage.host")
	r.Key(gofig.String, "", "", "", "libstorage.service")
	r.Key(gofig.Bool, "", false, "", "libstorage.profiles.enabled")
	r.Key(gofig.Bool, "", false, "", "libstorage.profiles.client")
	r.Key(gofig.String, "", "local=127.0.0.1", "", "libstorage.profiles.groups")
	gofig.Register(r)
}
