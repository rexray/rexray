package http

import (
	"github.com/emccode/libstorage/api/types"
)

/******************************************************************************
**                                 Root                                      **
*******************************************************************************/

// RootResponse is the response when getting root information about the service.
type RootResponse []string

/******************************************************************************
**                                Drivers                                    **
*******************************************************************************/

// DriversResponse is the response when getting one to many DriverInfos.
type DriversResponse map[string]*types.DriverInfo

// DriverInspectResponse is the response when getting a single DriverInfo.
type DriverInspectResponse *types.DriverInfo

/******************************************************************************
**                               Executors                                    **
*******************************************************************************/

// ExecutorsListResponse is the response when getting one to many ExecutorInfos.
type ExecutorsListResponse []*types.ExecutorInfo

/******************************************************************************
**                               Instances                                    **
*******************************************************************************/
/*
// InstanceRef is a wrapper for an Instance object along with its location.
type InstanceRef struct {
	Service  string    `json:"service,omitempty"`
	Instance *Instance `json:"instance"`
}

// InstancesResponse is the response when getting one to many Instances.
type InstancesResponse []*InstanceRef

// InstanceInspectResponse is the response when getting a single Instance.
type InstanceInspectResponse *Instance*/

/******************************************************************************
**                                Services                                    **
*******************************************************************************/

// ServicesResponse is the response when getting one to many ServiceInfos.
type ServicesResponse map[string]*types.ServiceInfo

// ServiceInspectResponse is the response when getting a single ServiceInfo.
type ServiceInspectResponse *types.ServiceInfo

/******************************************************************************
**                               Snapshots                                    **
*******************************************************************************/

// ServiceSnapshotMap is the response for listing snapshots for multiple
// services.
type ServiceSnapshotMap map[string]SnapshotMap

// SnapshotMap is the response for listing snapshots for a single service.
type SnapshotMap map[string]*types.Snapshot

// SnapshotInspectResponse is the response when getting a single Snapshot.
type SnapshotInspectResponse *types.Snapshot

// SnapshotCreateResponse is the response when creating a Snapshot.
type SnapshotCreateResponse *types.Snapshot

// SnapshotCopyResponse is the response when copying a Snapshot.
type SnapshotCopyResponse *types.Snapshot

// SnapshotRemoveResponse is the response when removing a Snapshot.
type SnapshotRemoveResponse struct {
}

/******************************************************************************
**                                Volumes                                     **
*******************************************************************************/

// ServiceVolumeMap is the response for listing volumes for multiple services.
type ServiceVolumeMap map[string]VolumeMap

// VolumeMap is the response for listing volumes for a single service.
type VolumeMap map[string]*types.Volume

// VolumeInspectResponse is the response when getting a single Volume.
type VolumeInspectResponse *types.Volume

// VolumeCreateResponse is the response when creating a Volume.
type VolumeCreateResponse *types.Volume

// VolumeCopyResponse is the response when copying a Volume.
type VolumeCopyResponse *types.Volume

// VolumeCreateFromSnapshotResponse is the response when creating a Volume
// from an existing snapshot.
type VolumeCreateFromSnapshotResponse *types.Volume

// VolumeRemoveResponse is the response when removing a Volume.
type VolumeRemoveResponse struct {
}

// VolumeAttachResponse is the response when attaching a Volume.
type VolumeAttachResponse struct {
}

// VolumeDetachResponse is the response when detching a single Volume.
type VolumeDetachResponse struct {
}

// VolumeDetachMultipleResponse is the response when detching multiple Volumes.
type VolumeDetachMultipleResponse ServiceVolumeMap
