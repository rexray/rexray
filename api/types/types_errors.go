package types

import (
	"github.com/akutz/goof"
)

// ErrNotImplemented is the error that Driver implementations should return if
// a function is not implemented.
var ErrNotImplemented = goof.New("not implemented")

// ErrTimedOut is the error that is used to indicate an operation timed out.
var ErrTimedOut = goof.New("timed out")

// ErrUnsupportedForClientType is the error that occurs when an operation is
// invoked that is unsupported for the current client type.
type ErrUnsupportedForClientType struct{ goof.Goof }

// ErrBadAdminToken occurs when a bad admin token is provided.
type ErrBadAdminToken struct{ goof.Goof }

// ErrNotFound occurs when a Driver inspects or sends an operation to a
// resource that cannot be found.
type ErrNotFound struct{ goof.Goof }

// ErrMissingLocalDevices occurs when an operation requires local devices
// and they're missing.
type ErrMissingLocalDevices struct{ goof.Goof }

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

// ErrMissingStorageService occurs when the storage service is expected in
// the provided context but is not there.
var ErrMissingStorageService = goof.New("missing storage service")

// ErrSecTokInvalid occurs when a security token is invalid.
type ErrSecTokInvalid struct {
	// InvalidToken is a flag that indicates whether or not the token was able
	// to be parsed at all.
	InvalidToken bool `json:"invalidToken"`

	// InvalidSig is a flag that indicates whether or not the security token
	// has a valid signature.
	InvalidSig bool `json:"invalidSig"`

	// MissingClaim is empty if all claims are missing or set to the name
	// of the first, detected, missing claim.
	MissingClaim string `json:"claim"`

	// Denied is a flag that indicates whether or not the security token
	// was denied access.
	Denied bool

	// InnerError is the inner error that caused this one.
	InnerError error `json:"innerError,omitempty"`
}

// Error returns the error string.
func (e *ErrSecTokInvalid) Error() string {
	return "invalid security token"
}
