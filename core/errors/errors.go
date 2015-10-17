package errors

import (
	"fmt"
)

// RexRayErrCode is a REX-Ray error code value.
type RexRayErrCode int

const (
	// ErrCodeUnknown is the error code for an unknown error.
	ErrCodeUnknown RexRayErrCode = iota

	// ErrCodeNoOSDetected is the error code for when no OS is detected.
	ErrCodeNoOSDetected

	// ErrCodeDriverBlockDeviceDiscovery is the error code for for when there
	// is an error discovering one or more block devices.
	ErrCodeDriverBlockDeviceDiscovery

	// ErrCodeDriverInstanceDiscovery is the error code for when there is an
	// error discovering an instance.
	ErrCodeDriverInstanceDiscovery

	// ErrCodeDriverVolumeDiscovery is the error code for when there is an
	// error discovering one or more volumes.
	ErrCodeDriverVolumeDiscovery

	// ErrCodeDriverSnapshotDiscovery is the error code for when there is an
	// error discovering one or more snapshots.
	ErrCodeDriverSnapshotDiscovery

	// ErrCodeMultipleDriversDetected is the error code for when there are
	// multiple drivers with the same name detected.
	ErrCodeMultipleDriversDetected

	// ErrCodeNoOSDrivers is the error code for when there are no registered
	// OS drivers.
	ErrCodeNoOSDrivers

	// ErrCodeNoVolumeDrivers is the error code for when there are no
	// registered volume drivers.
	ErrCodeNoVolumeDrivers

	// ErrCodeNoStorageDrivers is the error code for when there are no
	// registered storage drivers.
	ErrCodeNoStorageDrivers

	// ErrCodeNotImplemented is the error code for when some function is not
	// yet implemented.
	ErrCodeNotImplemented

	// ErrCodeUnknownOS is the error code for when the OS is detected but it
	// is an unknown or unsupported type.
	ErrCodeUnknownOS

	// ErrCodeUnknownFileSystem is the error code for when the filesystem is
	// detected but it is an unknown or unsupported type.
	ErrCodeUnknownFileSystem

	// ErrCodeMissingVolumeID is the error code for when the volume ID is
	// missing.
	ErrCodeMissingVolumeID

	// ErrCodeMultipleVolumesReturned is the error code for when only a single
	// volume is expected to be returned but multiple volumes are returned.
	ErrCodeMultipleVolumesReturned

	// ErrCodeNoVolumesReturned is the error code for when no volumes are
	// returned.
	ErrCodeNoVolumesReturned

	// ErrCodeLocalVolumeMaps is the error code for when there is an error
	// discovering local volume mappings.
	ErrCodeLocalVolumeMaps
)

var (
	// ErrNoOSDetected is the error for when no OS is detected.
	ErrNoOSDetected = ErrRexRay(ErrCodeNoOSDetected)

	// ErrDriverBlockDeviceDiscovery is the error for for when there
	// is an error discovering block devices.
	ErrDriverBlockDeviceDiscovery = ErrRexRay(ErrCodeDriverBlockDeviceDiscovery)

	// ErrDriverInstanceDiscovery is the error for when there is an
	// error discovering an instance.
	ErrDriverInstanceDiscovery = ErrRexRay(ErrCodeDriverInstanceDiscovery)

	// ErrDriverVolumeDiscovery is the error for when there is an
	// error discovering one or more volumes.
	ErrDriverVolumeDiscovery = ErrRexRay(ErrCodeDriverVolumeDiscovery)

	// ErrDriverSnapshotDiscovery is the error for when there is an
	// error discovering one or more snapshots.
	ErrDriverSnapshotDiscovery = ErrRexRay(ErrCodeDriverSnapshotDiscovery)

	// ErrMultipleDriversDetected is the error for when there are
	// multiple drivers with the same name detected.
	ErrMultipleDriversDetected = ErrRexRay(ErrCodeMultipleDriversDetected)

	// ErrNoOSDrivers is the error for when there are no registered
	// OS drivers.
	ErrNoOSDrivers = ErrRexRay(ErrCodeNoOSDrivers)

	// ErrNoVolumeDrivers is the error for when there are no registered
	// volume drivers.
	ErrNoVolumeDrivers = ErrRexRay(ErrCodeNoVolumeDrivers)

	// ErrNoStorageDrivers is the error for when there are no registered
	// storage drivers.
	ErrNoStorageDrivers = ErrRexRay(ErrCodeNoStorageDrivers)

	// ErrNotImplemented is the error for when some function is not
	// yet implemented.
	ErrNotImplemented = ErrRexRay(ErrCodeNotImplemented)

	// ErrUnknownOS is the error for when the OS is detected but it
	// is an unknown or unsupported type.
	ErrUnknownOS = ErrRexRay(ErrCodeUnknownOS)

	// ErrUnknownFileSystem is the error for when the filesystem is
	// detected but it is an unknown or unsupported type.
	ErrUnknownFileSystem = ErrRexRay(ErrCodeUnknownFileSystem)

	// ErrMissingVolumeID is the error for when the volume ID is missing.
	ErrMissingVolumeID = ErrRexRay(ErrCodeMissingVolumeID)

	// ErrMultipleVolumesReturned is the error for when only a single
	// volume is expected to be returned but multiple volumes are returned.
	ErrMultipleVolumesReturned = ErrRexRay(ErrCodeMultipleVolumesReturned)

	// ErrNoVolumesReturned is the error for when no volumes are returned.
	ErrNoVolumesReturned = ErrRexRay(ErrCodeNoVolumesReturned)

	// ErrLocalVolumeMaps is the error for when there is an error
	// discovering local volume mappings.
	ErrLocalVolumeMaps = ErrRexRay(ErrCodeLocalVolumeMaps)
)

// RexRayErr is the default error type for REX-Ray errors.
type RexRayErr struct {
	Code RexRayErrCode
}

// ErrRexRay creates a new instance of a RexRayErr with a given error code.
func ErrRexRay(code RexRayErrCode) *RexRayErr {
	return &RexRayErr{code}
}

// Error returns the string version of the error code.
func (e *RexRayErr) Error() string {
	switch e.Code {
	case ErrCodeNoOSDetected:
		return "no OS detected"
	case ErrCodeDriverBlockDeviceDiscovery:
		return "driver block device discovery failed"
	case ErrCodeDriverInstanceDiscovery:
		return "driver instance discovery failed"
	case ErrCodeDriverVolumeDiscovery:
		return "driver volume discovery failed"
	case ErrCodeDriverSnapshotDiscovery:
		return "driver snapshot discovery failed"
	case ErrCodeMultipleDriversDetected:
		return "multiple drivers detected"
	case ErrCodeNoOSDrivers:
		return "no OS drivers initialized"
	case ErrCodeNoVolumeDrivers:
		return "no volume drivers initialized"
	case ErrCodeNoStorageDrivers:
		return "no storage drivers initialized"
	case ErrCodeNotImplemented:
		return "not implemented"
	case ErrCodeUnknownOS:
		return "unknown OS"
	case ErrCodeUnknownFileSystem:
		return "unknown file system"
	case ErrCodeMissingVolumeID:
		return "missing volume ID"
	case ErrCodeMultipleVolumesReturned:
		return "multiple volumes returned"
	case ErrCodeNoVolumesReturned:
		return "no Volumes returned"
	case ErrCodeLocalVolumeMaps:
		return "getting local volume mounts"
	default:
		return "unknown error"
	}
}

// Error is a structure that implements the Go Error interface as well as the
// Golf interface for extended log information capabilities.
type Error struct {
	fields map[string]interface{}
}

// Fields is a type alias for a map of interfaces.
type Fields map[string]interface{}

// Error returns the error message.
func (e *Error) Error() string {
	return e.fields["message"].(string)
}

// PlayGolf lets the logrus framework know that Error supports the Golf
// framework.
func (e *Error) PlayGolf() bool {
	return true
}

// GolfExportedFields returns the fields to use when playing golf.
func (e *Error) GolfExportedFields() map[string]interface{} {
	return e.fields
}

// New returns a new error object initialized with the provided message.
func New(message string) error {
	return &Error{Fields{"message": message}}
}

// Newf returns a new error object initialized with the messages created by
// formatting the format string with the provided arguments.
func Newf(format string, a ...interface{}) error {
	return &Error{Fields{"message": fmt.Sprintf(format, a)}}
}

// WithError returns a new error object initialized with the provided message
// and inner error.
func WithError(message string, inner error) error {
	return WithFieldsE(nil, message, inner)
}

// WithField returns a new error object initialized with the provided field
// name, value, and error message.
func WithField(key string, val interface{}, message string) error {
	return WithFields(Fields{key: val}, message)
}

// WithFieldE returns a new error object initialized with the provided field
// name, value, error message, and inner error.
func WithFieldE(key string, val interface{}, message string, inner error) error {
	return WithFieldsE(Fields{key: val}, message, inner)
}

// WithFields returns a new error object initialized with the provided fields
// and error message.
func WithFields(fields map[string]interface{}, message string) error {
	return WithFieldsE(fields, message, nil)
}

// WithFieldsE returns a new error object initialized with the provided fields,
// error message, and inner error.
func WithFieldsE(fields map[string]interface{}, message string, inner error) error {

	if fields == nil {
		fields = Fields{}
	}

	if inner != nil {
		fields["inner"] = inner
	}

	fields["message"] = message

	return &Error{fields}
}
