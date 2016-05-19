package types

import "io"

// Server is the interface for a libStorage server.
type Server interface {
	io.Closer

	// Name returns the name of the server.
	Name() string

	// Addrs returns the server's configured endpoint addresses.
	Addrs() []string
}
