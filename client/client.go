package client

import (
	apiclient "github.com/emccode/libstorage/api/client"
	"github.com/emccode/libstorage/api/types"
	apihttp "github.com/emccode/libstorage/api/types/http"

	// load the local drivers
	_ "github.com/emccode/libstorage/imports/local"
)

var (
	// EnableInstanceIDHeaders is a flag indicating whether or not the
	// client will automatically send the instance ID header(s) along with
	// storage-related API requests. The default is enabled.
	EnableInstanceIDHeaders = true

	// EnableLocalDevicesHeaders is a flag indicating whether or not the
	// client will automatically send the local devices header(s) along with
	// storage-related API requests. The default is enabled.
	EnableLocalDevicesHeaders = true
)

// Client is the libStorage client.
type Client interface {

	// EnableInstanceIDHeaders is a flag indicating whether or not the
	// client will automatically send the instance ID header(s) along with
	// storage-related API requests. The default is enabled.
	EnableInstanceIDHeaders(enabled bool)

	// EnableLocalDevicesHeaders is a flag indicating whether or not the
	// client will automatically send the local devices header(s) along with
	// storage-related API requests. The default is enabled.
	EnableLocalDevicesHeaders(enabled bool)

	// ServerName returns the name of the server to which the client is
	// connected. This is not the same as the host name, rather it's the
	// randomly generated name the server creates for unique identification
	// when the server starts for the first time.
	ServerName() string

	// API returns the underlying API client.
	API() *apiclient.Client

	// InstanceID gets the client's instance ID.
	InstanceID(service string) (*types.InstanceID, error)

	// LocalDevices gets the client's local devices map.
	LocalDevices(service string) (map[string]string, error)

	// NextDevice gets the next available device ID.
	NextDevice(service string) (string, error)

	// Services returns a map of the configured Services.
	Services() (apihttp.ServicesMap, error)

	// ServiceInspect returns information about a service.
	ServiceInspect(name string) (*types.ServiceInfo, error)

	// Volumes returns a list of all Volumes for all Services.
	Volumes(attachments bool) (apihttp.ServiceVolumeMap, error)

	// VolumesByService returns a list of all Volumes for a service.
	VolumesByService(
		service string, attachments bool) (apihttp.VolumeMap, error)

	// VolumeInspect gets information about a single volume.
	VolumeInspect(
		service, volumeID string, attachments bool) (*types.Volume, error)

	// VolumeCopy copies a single volume.
	VolumeCopy(
		service, volumeID string,
		request *apihttp.VolumeCopyRequest) (*types.Volume, error)

	// VolumeCreate creates a single volume.
	VolumeCreate(
		service string,
		request *apihttp.VolumeCreateRequest) (*types.Volume, error)

	// SnapshotCreate creates a single volume from a snapshot.
	VolumeCreateFromSnapshot(
		service, snapshotID string,
		request *apihttp.VolumeCreateRequest) (*types.Volume, error)

	// VolumeRemove removes a single volume.
	VolumeRemove(service, volumeID string) error

	// VolumeAttach attaches a single volume.
	VolumeAttach(
		service string,
		volumeID string,
		request *apihttp.VolumeAttachRequest) (*types.Volume, error)

	// VolumeAttach attaches a single volume.
	VolumeDetach(
		service string,
		volumeID string,
		request *apihttp.VolumeDetachRequest) (*types.Volume, error)

	// VolumeDetachByService detaches all volumes in a service
	VolumeDetachAllForService(
		service string,
		request *apihttp.VolumeDetachRequest) (apihttp.VolumeMap, error)

	// VolumeDetachAll detaches all volumes from all services
	VolumeDetachAll(
		request *apihttp.VolumeDetachRequest) (apihttp.ServiceVolumeMap, error)

	// VolumeSnapshot creates a single snapshot.
	VolumeSnapshot(
		service string,
		volumeID string,
		request *apihttp.VolumeSnapshotRequest) (*types.Snapshot, error)

	// Snapshots returns a list of all Snapshots for all services.
	Snapshots() (apihttp.ServiceSnapshotMap, error)

	// SnapshotsByService returns a list of all Snapshots for a single service.
	SnapshotsByService(service string) (apihttp.SnapshotMap, error)

	// SnapshotInspect gets information about a single snapshot.
	SnapshotInspect(service, snapshotID string) (*types.Snapshot, error)

	// SnapshotRemove removes a single snapshot.
	SnapshotRemove(service, snapshotID string) error

	// SnapshotCopy copies a snapshot to a new snapshot.
	SnapshotCopy(
		service, snapshotID string,
		request *apihttp.SnapshotCopyRequest) (*types.Snapshot, error)
}
