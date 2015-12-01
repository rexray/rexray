package errors

import (
	"github.com/akutz/goof"
)

// RexRayErrCode is a REX-Ray error code value.
type RexRayErrCode int

const (
	// ErrCodeUnknown is the error code for an unknown error.
	ErrCodeUnknown RexRayErrCode = iota

	// ErrCodeNoOSDetected is the error code for when no OS is detected.
	ErrCodeNoOSDetected

	// ErrCodeNoVolumesDetected is the error code for when no volumes are
	// detected.
	ErrCodeNoVolumesDetected

	// ErrCodeNoStorageDetected is the error code for when no storage is
	// detected.
	ErrCodeNoStorageDetected

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

	// ErrCodeRunAsyncFromVolume is the error code for when an asynchronous
	// create volume is received.
	ErrCodeRunAsyncFromVolume
)

var (
	// ErrNoOSDetected is the error for when no OS is detected.
	ErrNoOSDetected = ErrRexRay(ErrCodeNoOSDetected)

	// ErrNoVolumesDetected is the error for when no volumes are detected.
	ErrNoVolumesDetected = ErrRexRay(ErrCodeNoVolumesDetected)

	// ErrNoStorageDetected is the error for when no storage is detected.
	ErrNoStorageDetected = ErrRexRay(ErrCodeNoStorageDetected)

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

	// ErrRunAsyncFromVolume is the error for when an asynchronous
	// create volume is received.
	ErrRunAsyncFromVolume = ErrRexRay(ErrCodeRunAsyncFromVolume)
)

// ErrRexRay creates a new instance of a RexRayErr with a given error code.
func ErrRexRay(code RexRayErrCode) error {
	return goof.New(errCodeToString(code))
}

// Error returns the string version of the error code.
func errCodeToString(code RexRayErrCode) string {
	switch code {
	case ErrCodeNoOSDetected,
		ErrCodeNoVolumesDetected,
		ErrCodeNoStorageDetected,
		ErrCodeNoOSDrivers,
		ErrCodeNoVolumeDrivers,
		ErrCodeNoStorageDrivers,
		ErrCodeNoVolumesReturned:
		return errCodeNoToString(code)
	case ErrCodeDriverBlockDeviceDiscovery,
		ErrCodeDriverInstanceDiscovery,
		ErrCodeDriverVolumeDiscovery,
		ErrCodeDriverSnapshotDiscovery:
		return errCodeFailedToString(code)
	case ErrCodeMultipleDriversDetected,
		ErrCodeMultipleVolumesReturned:
		return errCodeMultiToString(code)
	case ErrCodeUnknownOS,
		ErrCodeUnknownFileSystem:
		return errCodeUnknownToString(code)
	case ErrCodeMissingVolumeID:
		return "missing volume ID"
	case ErrCodeLocalVolumeMaps:
		return "getting local volume mounts"
	case ErrCodeRunAsyncFromVolume:
		return "cannot create volume from volume and run asynchronously"
	case ErrCodeNotImplemented:
		return "not implemented"
	default:
		return "unknown error"
	}
}

func errCodeFailedToString(code RexRayErrCode) string {
	switch code {
	case ErrCodeDriverBlockDeviceDiscovery:
		return "driver block device discovery failed"
	case ErrCodeDriverInstanceDiscovery:
		return "driver instance discovery failed"
	case ErrCodeDriverVolumeDiscovery:
		return "driver volume discovery failed"
	case ErrCodeDriverSnapshotDiscovery:
		return "driver snapshot discovery failed"
	default:
		return "unknown error"
	}
}

func errCodeNoToString(code RexRayErrCode) string {
	switch code {
	case ErrCodeNoOSDetected:
		return "no OS detected"
	case ErrCodeNoVolumesDetected:
		return "no volumes detected"
	case ErrCodeNoStorageDetected:
		return "no storage detected"
	case ErrCodeNoOSDrivers:
		return "no OS drivers initialized"
	case ErrCodeNoVolumeDrivers:
		return "no volume drivers initialized"
	case ErrCodeNoStorageDrivers:
		return "no storage drivers initialized"
	case ErrCodeNoVolumesReturned:
		return "no Volumes returned"
	default:
		return "unknown error"
	}
}

func errCodeMultiToString(code RexRayErrCode) string {
	switch code {
	case ErrCodeMultipleDriversDetected:
		return "multiple drivers detected"
	case ErrCodeMultipleVolumesReturned:
		return "multiple volumes returned"
	default:
		return "unknown error"
	}
}

func errCodeUnknownToString(code RexRayErrCode) string {
	switch code {
	case ErrCodeUnknownOS:
		return "unknown OS"
	case ErrCodeUnknownFileSystem:
		return "unknown file system"
	default:
		return "unknown error"
	}
}
