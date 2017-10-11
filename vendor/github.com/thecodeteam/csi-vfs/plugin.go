// +build linux,plugin

package main

import "C"

import (
	"github.com/thecodeteam/csi-vfs/provider"
)

////////////////////////////////////////////////////////////////////////////////
//                              Go Plug-in                                    //
////////////////////////////////////////////////////////////////////////////////

// ServiceProviders is an exported symbol that provides a host program
// with a map of the service provider names and constructors.
var ServiceProviders = map[string]func() interface{}{
	name: func() interface{} { return provider.New() },
}
