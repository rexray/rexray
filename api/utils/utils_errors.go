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
