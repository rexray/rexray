// +build go1.7

package gournal

import (
	"context"
)

// A Context carries a deadline, a cancelation signal, and other values across
// API boundaries.
//
// Context's methods may be called by multiple goroutines simultaneously.
type Context interface {
	context.Context
}

// Background returns a non-nil, empty Context. It is never canceled, has no
// values, and has no deadline.  It is typically used by the main function,
// initialization, and tests, and as the top-level Context for incoming
// requests.
func Background() context.Context {
	return context.Background()
}

// WithValue returns a copy of parent in which the value associated with key is
// val.
//
// Use context Values only for request-scoped data that transits processes and
// APIs, not for passing optional parameters to functions.
func WithValue(parent context.Context, key, val interface{}) context.Context {
	return context.WithValue(parent, key, val)
}
