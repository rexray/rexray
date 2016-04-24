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
func NewStoreKeyErr(storeKey string) error {
	return &types.ErrStoreKey{
		Goof: goof.WithField("storeKey", storeKey, "missing store key"),
	}
}

// NewContextKeyErr returns a new ErrContextKey error.
func NewContextKeyErr(contextKey string) error {
	return &types.ErrContextKey{
		Goof: goof.WithField("contextKey", contextKey, "missing context key"),
	}
}

// NewContextTypeErr returns a new ErrContextType error.
func NewContextTypeErr(
	contextKey, expectedType, actualType string) error {

	return &types.ErrContextType{
		Goof: goof.WithFields(
			goof.Fields{
				"contextKey":   contextKey,
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
