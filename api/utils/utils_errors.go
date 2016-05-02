package utils

import (
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/types"
)

// NewNotFoundError returns a new ErrNotFound error.
func NewNotFoundError(resourceID string) error {
	return &types.ErrNotFound{
		Goof: goof.WithField("resourceID", resourceID, "resource not found"),
	}
}

// NewMissingInstanceIDError returns a new ErrMissingInstanceID error.
func NewMissingInstanceIDError(service string) error {
	return &types.ErrMissingInstanceID{
		Goof: goof.WithField("service", service, "missing instance ID"),
	}
}

// NewStoreKeyErr returns a new ErrStoreKey error.
func NewStoreKeyErr(key string) error {
	return &types.ErrStoreKey{
		Goof: goof.WithField("storeKey", key, "missing store key"),
	}
}

// NewContextKeyErr returns a new ErrContextKey error.
func NewContextKeyErr(key types.ContextKey) error {
	return &types.ErrContextKey{
		Goof: goof.WithField("contextKey", key.String(), "missing context key"),
	}
}

// NewContextErr returns a new ErrContextKey error.
func NewContextErr(key types.ContextKey) error {
	return NewContextKeyErr(key)
}

// NewContextTypeErr returns a new ErrContextType error.
func NewContextTypeErr(
	contextKey types.ContextKey, expectedType, actualType string) error {
	return &types.ErrContextType{
		Goof: goof.WithFields(
			goof.Fields{
				"contextKey":   contextKey.String(),
				"expectedType": expectedType,
				"actualType":   actualType,
			}, "invalid context type")}
}

// NewDriverTypeErr returns a new ErrDriverTypeErr error.
func NewDriverTypeErr(expectedType, actualType string) error {
	return &types.ErrDriverTypeErr{Goof: goof.WithFields(goof.Fields{
		"expectedType": expectedType,
		"actualType":   actualType,
	}, "invalid driver type")}
}

// NewBatchProcessErr returns a new ErrBatchProcess error.
func NewBatchProcessErr(completed interface{}, err error) error {
	return &types.ErrBatchProcess{Goof: goof.WithFieldE(
		"completed", completed, "batch processing error", err)}
}

// NewBadFilterErr returns a new ErrBadFilter error.
func NewBadFilterErr(filter string, err error) error {
	return &types.ErrBadFilter{Goof: goof.WithFieldE(
		"filter", filter, "bad filter", err)}
}
