package libstorage

import (
	"errors"
	"fmt"

	"github.com/thecodeteam/gocsi/csi"

	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
)

var errMissingServiceName = errors.New("missing service name")

func gib2b(i int64) uint64 {
	return uint64(i * 1024 * 1024 * 1024)
}
func b2gib(i uint64) int64 {
	if i == 0 {
		return 0
	}
	return int64(i / 1024 / 1024 / 1024)
}
func addrOfInt64(i int64) *int64 {
	return &i
}

func toNodeID(iid *apitypes.InstanceID) *csi.NodeID {
	return &csi.NodeID{
		Values: map[string]string{
			"id":      iid.ID,
			"driver":  iid.Driver,
			"service": iid.Service,
		},
	}
}

var (
	errNodeIDNil            = errors.New("nodeID is nil")
	errNodeIDMissingID      = errors.New("nodeID missing id key")
	errNodeIDMissingDriver  = errors.New("nodeID missing driver key")
	errNodeIDMissingService = errors.New("nodeID missing service key")
)

func toInstanceID(nodeID *csi.NodeID) (*apitypes.InstanceID, error) {
	if nodeID == nil {
		return nil, errNodeIDNil
	}

	i, idOK := nodeID.Values["id"]
	if !idOK {
		return nil, errNodeIDMissingID
	}

	d, driverOK := nodeID.Values["driver"]
	if !driverOK {
		return nil, errNodeIDMissingDriver
	}

	s, serviceOK := nodeID.Values["service"]
	if !serviceOK {
		return nil, errNodeIDMissingService
	}

	return &apitypes.InstanceID{
		ID:      i,
		Driver:  d,
		Service: s,
	}, nil
}

func toVolumeInfo(v *apitypes.Volume) *csi.VolumeInfo {

	mdv := map[string]string{}
	if v.AttachmentState > 0 {
		mdv["attachmentstate"] = fmt.Sprintf("%v", v.AttachmentState)
	}
	if v.AvailabilityZone != "" {
		mdv["availabilityzone"] = v.AvailabilityZone
	}
	if v.Encrypted {
		mdv["encrypted"] = "true"
	}
	if v.IOPS > 0 {
		mdv["iops"] = fmt.Sprintf("%d", v.IOPS)
	}
	if v.Name != "" {
		mdv["name"] = v.Name
	}
	if v.NetworkName != "" {
		mdv["networkname"] = v.NetworkName
	}
	if v.Status != "" {
		mdv["status"] = v.Status
	}
	if v.Type != "" {
		mdv["type"] = v.Type
	}
	for k, v := range v.Fields {
		if _, ok := mdv[k]; !ok {
			mdv[k] = v
		}
	}

	return &csi.VolumeInfo{
		Id:            &csi.VolumeID{Values: map[string]string{"id": v.ID}},
		Metadata:      &csi.VolumeMetadata{Values: mdv},
		CapacityBytes: gib2b(v.Size),
	}
}

func isVolCapSupported(
	storType apitypes.StorageType, caps ...*csi.VolumeCapability) bool {

	if len(caps) == 0 {
		return true
	}

	reqCapsIncludeRaw := false
	for _, vc := range caps {
		if vc.GetBlock() != nil {
			reqCapsIncludeRaw = true
			break
		}
	}

	// The bridge can only support volumes if the requested
	// capabilities do not include raw support AND the storage
	// type is something other than Block.
	return !(reqCapsIncludeRaw && storType != apitypes.Block)
}
