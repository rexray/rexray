package types

import (
	"github.com/akutz/goof"
)

// JSONError is the base type for errors returned from the REST API.
type JSONError struct {
	Message    string      `json:"message"`
	Status     int         `json:"status"`
	InnerError interface{} `json:"error,omitempty"`
}

// Error returns the error message.
func (je *JSONError) Error() string {
	return je.Message
}

// ErrNotImplemented is the error that Driver implementations should return if
// a function is not implemented.
var ErrNotImplemented = goof.New("not implemented")

// ErrNotFound occurs when a Driver inspects or sends an operation to a
// resource that cannot be found.
type ErrNotFound struct{ goof.Goof }

// ErrMissingInstanceID occurs when an operation requires the instance ID for
// the configured service to be avaialble.
type ErrMissingInstanceID struct{ goof.Goof }

// ErrStoreKey occurs when no value exists for a specified store key.
type ErrStoreKey struct{ goof.Goof }

// ErrContextKey occurs when no value exists for a specified context key.
type ErrContextKey struct{ goof.Goof }

// ErrContextType occurs when a value exists in the context but is not the
// expected typed.
type ErrContextType struct{ goof.Goof }

// ErrDriverTypeErr occurs when a Driver is constructed with an invalid type.
type ErrDriverTypeErr struct{ goof.Goof }

// ErrBatchProcess occurs when a batch process is interrupted by an error
// before the process is complete. This error will contain information about
// the objects for which the process did complete.
type ErrBatchProcess struct{ goof.Goof }

// ErrBadFilter occurs when a bad filter is supplied via the filter query
// string.
type ErrBadFilter struct{ goof.Goof }
