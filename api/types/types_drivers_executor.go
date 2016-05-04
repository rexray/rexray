package types

import (
	"time"
)

// NewStorageExecutor is a function that constructs a new StorageExecutors.
type NewStorageExecutor func() StorageExecutor

// StorageExecutor is the part of a storage driver that is downloaded at
// runtime by the libStorage client.
type StorageExecutor interface {
	Driver
	StorageExecutorFunctions
}

// StorageExecutorFunctions is the collection of functions that are required of
// a StorageExecutor.
type StorageExecutorFunctions interface {
	// InstanceID returns the local system's InstanceID.
	InstanceID(
		ctx Context,
		opts Store) (*InstanceID, error)

	// NextDevice returns the next available device.
	NextDevice(
		ctx Context,
		opts Store) (string, error)

	// LocalDevices returns a map of the system's local devices.
	LocalDevices(
		ctx Context,
		opts Store) (map[string]string, error)
}

// ProvidesStorageExecutorCLI is a type that provides the StorageExecutorCLI.
type ProvidesStorageExecutorCLI interface {
	// XCLI returns the StorageExecutorCLI.
	XCLI() StorageExecutorCLI
}

// StorageExecutorCLI provides a way to interact with the CLI tool built with
// the driver implementations of the StorageExecutor interface.
type StorageExecutorCLI interface {
	StorageExecutorFunctions

	// WaitForDevice blocks until the provided attach token appears in the
	// map returned from LocalDevices or until the timeout expires, whichever
	// occurs first.
	//
	// The return value is a boolean flag indicating whether or not a match was
	// discovered as well as the result of the last LocalDevices call before a
	// match is discovered or the timeout expires.
	WaitForDevice(
		ctx Context,
		attachToken string,
		timeout time.Duration,
		opts Store) (bool, map[string]string, error)
}
