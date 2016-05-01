package client

import (
	"io"
	"net/http"

	"github.com/emccode/libstorage/api/types"
)

// Client is the libStorage API client.
type Client struct {

	// HTTP is the underlying HTTP client.
	Client *http.Client

	// Host is the host[:port] of the remote libStorage API.
	Host string

	// LogRequests is a flag indicating whether or not to log HTTP requests.
	LogRequests bool

	// LogResponses is a flag indicating whether or not to log HTTP responses.
	LogResponses bool

	// Headers are headers to send with each HTTP request.
	Headers http.Header

	// ServerName returns the name of the server to which the client is
	// connected. This is not the same as the host name, rather it's the
	// randomly generated name the server creates for unique identification
	// when the server starts for the first time. This value is updated
	// by every request to the server that returns the server name header
	// as part of its response.
	ServerName string
}

// APIClient is the libStorage API client used for communicating with a remote
// libStorage endpoint.
type APIClient interface {

	// Root returns a list of root resources.
	Root(ctx types.Context) ([]string, error)

	// Instances returns a list of instances.
	Instances(ctx types.Context) ([]*types.Instance, error)

	// InstanceInspect inspects an instance.
	InstanceInspect(ctx types.Context, service string) (*types.Instance, error)

	// Services returns a map of the configured Services.
	Services(ctx types.Context) (map[string]*types.ServiceInfo, error)

	// ServiceInspect returns information about a service.
	ServiceInspect(ctx types.Context, name string) (*types.ServiceInfo, error)

	// Volumes returns a list of all Volumes for all Services.
	Volumes(
		ctx types.Context,
		attachments bool) (types.ServiceVolumeMap, error)

	// VolumesByService returns a list of all Volumes for a service.
	VolumesByService(
		ctx types.Context,
		service string,
		attachments bool) (types.VolumeMap, error)

	// VolumeInspect gets information about a single volume.
	VolumeInspect(
		ctx types.Context,
		service, volumeID string,
		attachments bool) (*types.Volume, error)

	// VolumeCreate creates a single volume.
	VolumeCreate(
		ctx types.Context,
		service string,
		request *types.VolumeCreateRequest) (*types.Volume, error)

	// VolumeCreateFromSnapshot creates a single volume from a snapshot.
	VolumeCreateFromSnapshot(
		ctx types.Context,
		service, snapshotID string,
		request *types.VolumeCreateRequest) (*types.Volume, error)

	// VolumeCopy copies a single volume.
	VolumeCopy(
		ctx types.Context,
		service, volumeID string,
		request *types.VolumeCopyRequest) (*types.Volume, error)

	// VolumeRemove removes a single volume.
	VolumeRemove(
		ctx types.Context,
		service, volumeID string) error

	// VolumeAttach attaches a single volume.
	VolumeAttach(
		ctx types.Context,
		service string,
		volumeID string,
		request *types.VolumeAttachRequest) (*types.Volume, error)

	// VolumeDetach attaches a single volume.
	VolumeDetach(
		ctx types.Context,
		service string,
		volumeID string,
		request *types.VolumeDetachRequest) (*types.Volume, error)

	// VolumeDetachAll attaches all volumes from all types.
	VolumeDetachAll(
		ctx types.Context,
		request *types.VolumeDetachRequest) (types.ServiceVolumeMap, error)

	// VolumeDetachAllForService detaches all volumes from a service.
	VolumeDetachAllForService(
		ctx types.Context,
		service string,
		request *types.VolumeDetachRequest) (types.VolumeMap, error)

	// VolumeSnapshot creates a single snapshot.
	VolumeSnapshot(
		ctx types.Context,
		service string,
		volumeID string,
		request *types.VolumeSnapshotRequest) (*types.Snapshot, error)

	// Snapshots returns a list of all Snapshots for all types.
	Snapshots(ctx types.Context) (types.ServiceSnapshotMap, error)

	// SnapshotsByService returns a list of all Snapshots for a single service.
	SnapshotsByService(
		ctx types.Context, service string) (types.SnapshotMap, error)

	// SnapshotInspect gets information about a single snapshot.
	SnapshotInspect(
		ctx types.Context,
		service, snapshotID string) (*types.Snapshot, error)

	// SnapshotRemove removes a single snapshot.
	SnapshotRemove(
		ctx types.Context,
		service, snapshotID string) error

	// SnapshotCopy copies a snapshot to a new snapshot.
	SnapshotCopy(
		ctx types.Context,
		service, snapshotID string,
		request *types.SnapshotCopyRequest) (*types.Snapshot, error)

	// Executors returns information about the executors.
	Executors(
		ctx types.Context) (map[string]*types.ExecutorInfo, error)

	// ExecutorHead returns information about an executor.
	ExecutorHead(
		ctx types.Context,
		name string) (*types.ExecutorInfo, error)

	// ExecutorGet downloads an executor.
	ExecutorGet(
		ctx types.Context, name string) (io.ReadCloser, error)
}
