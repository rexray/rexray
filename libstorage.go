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
	"github.com/akutz/gofig"
	"golang.org/x/net/context"

	"github.com/emccode/libstorage/api/server"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/client"
)

// New starts an embedded libStorage server and returns both the server
// instnace as well as a client connected to said instnace.
//
// While a new server may be launched, it's still up to the caller to provide
// a config instance with the correct properties to specify service
// information for a libStorage server.
func New(
	ctx context.Context,
	config gofig.Config) (types.Client, types.Server, <-chan error, error) {

	s, errs, err := server.Serve(ctx, config)
	if err != nil {
		return nil, nil, nil, err
	}

	if h := config.GetString(types.ConfigHost); h == "" {
		config.Set(types.ConfigHost, s.Addrs()[0])
	}

	c, err := client.New(ctx, config)
	if err != nil {
		return nil, nil, nil, err
	}

	return c, s, errs, nil
}
