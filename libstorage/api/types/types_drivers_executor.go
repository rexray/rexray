package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DeviceScanType is a type of device scan algorithm.
type DeviceScanType int

const (

	// DeviceScanQuick performs a shallow, quick scan.
	DeviceScanQuick DeviceScanType = iota

	// DeviceScanDeep performs a deep, longer scan.
	DeviceScanDeep
)

// String returns the string representation of a DeviceScanType.
func (st DeviceScanType) String() string {
	switch st {
	case DeviceScanQuick:
		return "quick"
	case DeviceScanDeep:
		return "deep"
	}
	return ""
}

// ParseDeviceScanType parses a device scan type.
func ParseDeviceScanType(i interface{}) DeviceScanType {
	switch ti := i.(type) {
	case string:
		lti := strings.ToLower(ti)
		if lti == DeviceScanQuick.String() {
			return DeviceScanQuick
		} else if lti == DeviceScanDeep.String() {
			return DeviceScanDeep
		}
		i, err := strconv.Atoi(ti)
		if err != nil {
			return DeviceScanQuick
		}
		return ParseDeviceScanType(i)
	case int:
		st := DeviceScanType(ti)
		if st == DeviceScanQuick || st == DeviceScanDeep {
			return st
		}
		return DeviceScanQuick
	default:
		return ParseDeviceScanType(fmt.Sprintf("%v", ti))
	}
}

// LocalDevicesOpts are options when getting a list of local devices.
type LocalDevicesOpts struct {
	ScanType DeviceScanType
	Opts     Store
}

// WaitForDeviceOpts are options when waiting on specific local device to
// appear.
type WaitForDeviceOpts struct {
	LocalDevicesOpts

	// Token is the value returned by a remote VolumeAttach call that the
	// client can use to block until a specific device has appeared in the
	// local devices list.
	Token string

	// Timeout is the maximum duration for which to wait for a device to
	// appear in the local devices list.
	Timeout time.Duration
}

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
		opts *LocalDevicesOpts) (*LocalDevices, error)
}

// StorageExecutorWithSupported is an interface that executor implementations
// may use by defining the function "Supported(Context, Store) (bool, error)".
// This function indicates whether a storage platform is valid when executing
// the executor binary on a given client.
type StorageExecutorWithSupported interface {
	StorageExecutorFunctions

	// Supported returns a flag indicating whether or not the platform
	// implementing the executor is valid for the host on which the executor
	// resides.
	Supported(
		ctx Context,
		opts Store) (bool, error)
}

// StorageExecutorWithMount is an interface that executor implementations
// may use to become part of the mount workflow.
type StorageExecutorWithMount interface {

	// Mount mounts a device to a specified path.
	Mount(
		ctx Context,
		deviceName, mountPoint string,
		opts *DeviceMountOpts) error
}

// StorageExecutorWithMounts is an interface that executor implementations
// may use to become part of the mounts workflow.
type StorageExecutorWithMounts interface {

	// Mounts get a list of mount points.
	Mounts(
		ctx Context,
		opts Store) ([]*MountInfo, error)
}

// StorageExecutorWithUnmount is an interface that executor implementations
// may use to become part of unmount workflow.
type StorageExecutorWithUnmount interface {

	// Unmount unmounts the underlying device from the specified path.
	Unmount(
		ctx Context,
		mountPoint string,
		opts Store) error
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
	StorageExecutorWithMount
	StorageExecutorWithMounts
	StorageExecutorWithUnmount

	// WaitForDevice blocks until the provided attach token appears in the
	// map returned from LocalDevices or until the timeout expires, whichever
	// occurs first.
	//
	// The return value is a boolean flag indicating whether or not a match was
	// discovered as well as the result of the last LocalDevices call before a
	// match is discovered or the timeout expires.
	WaitForDevice(
		ctx Context,
		opts *WaitForDeviceOpts) (bool, *LocalDevices, error)

	// Supported returns a flag indicating whether the executor supports
	// specific functions for a storage platform on the current host.
	Supported(
		ctx Context,
		opts Store) (LSXSupportedOp, error)
}

// LSXSupportedOp is a bit for the mask returned from an executor's Supported
// function.
type LSXSupportedOp int

const (
	// LSXSOpInstanceID indicates an executor supports "InstanceID".
	// "InstanceID" operation.
	LSXSOpInstanceID LSXSupportedOp = 1 << iota // 1

	// LSXSOpNextDevice indicates an executor supports "NextDevice".
	LSXSOpNextDevice

	// LSXSOpLocalDevices indicates an executor supports "LocalDevices".
	LSXSOpLocalDevices

	// LSXSOpWaitForDevice indicates an executor supports "WaitForDevice".
	LSXSOpWaitForDevice

	// LSXSOpMount indicates an executor supports "Mount".
	LSXSOpMount

	// LSXSOpUmount indicates an executor supports "Umount".
	LSXSOpUmount

	// LSXSOpMounts indicates an executor supports "Mounts".
	LSXSOpMounts
)

const (
	// LSXSOpNone indicates the executor is not supported for the platform.
	LSXSOpNone LSXSupportedOp = 0

	// LSXOpAll indicates the executor supports all operations.
	LSXOpAll LSXSupportedOp = LSXSOpInstanceID |
		LSXSOpNextDevice |
		LSXSOpLocalDevices |
		LSXSOpWaitForDevice |
		LSXSOpMount |
		LSXSOpUmount |
		LSXSOpMounts

	// LSXOpAllNoMount indicates the executor supports all operations except
	// mount and unmount.
	LSXOpAllNoMount = LSXOpAll & ^LSXSOpMount & ^LSXSOpUmount & ^LSXSOpMounts
)

// InstanceID returns a flag that indicates whether the LSXSOpInstanceID bit
// is set.
func (v LSXSupportedOp) InstanceID() bool {
	return v.bitSet(LSXSOpInstanceID)
}

// NextDevice returns a flag that indicates whether the LSXSOpNextDevice bit
// is set.
func (v LSXSupportedOp) NextDevice() bool {
	return v.bitSet(LSXSOpNextDevice)
}

// LocalDevices returns a flag that indicates whether the LSXSOpLocalDevices
// bit is set.
func (v LSXSupportedOp) LocalDevices() bool {
	return v.bitSet(LSXSOpLocalDevices)
}

// WaitForDevice returns a flag that indicates whether the LSXSOpWaitForDevice
// bit is set.
func (v LSXSupportedOp) WaitForDevice() bool {
	return v.bitSet(LSXSOpWaitForDevice)
}

// Mount returns a flag that indicates whether the LSXSOpMount bit
// is set.
func (v LSXSupportedOp) Mount() bool {
	return v.bitSet(LSXSOpMount)
}

// Umount returns a flag that indicates whether the LSXSOpUmount bit
// is set.
func (v LSXSupportedOp) Umount() bool {
	return v.bitSet(LSXSOpUmount)
}

// Mounts returns a flag that indicates whether the LSXSOpMounts bit
// is set.
func (v LSXSupportedOp) Mounts() bool {
	return v.bitSet(LSXSOpMounts)
}

func (v LSXSupportedOp) bitSet(b LSXSupportedOp) bool {
	return v&b == b
}
