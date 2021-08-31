package main

//#cgo CFLAGS: -I${SRCDIR}
//#include "libstor-c_types.h"
import "C"

import (
	"encoding/binary"
	"math/rand"
	"time"

	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	"github.com/AVENTER-UG/rexray/libstorage/client"
)

func toCVolume(v *types.Volume) (*C.volume, error) {

	lcva, cva, err := toCVolumeAttachments(v)
	if err != nil {
		return nil, err
	}

	cv := C.volume{
		id:                C.CString(v.ID),
		name:              C.CString(v.Name),
		size:              C.int64_t(v.Size),
		iops:              C.int64_t(v.IOPS),
		status:            C.CString(v.Status),
		volume_type:       C.CString(v.Type),
		availability_zone: C.CString(v.AvailabilityZone),
		network_name:      C.CString(v.NetworkName),
	}
	cv.attachments_c = lcva
	if lcva > 0 {
		cv.attachments = &cva[0]
	}

	return &cv, nil
}

func toCVolumeAttachments(
	v *types.Volume) (C.int, []*C.volume_attachment, error) {

	la := C.int(len(v.Attachments))
	if la == 0 {
		return 0, nil, nil
	}

	ca := make([]*C.volume_attachment, la)
	for x, a := range v.Attachments {
		cva, err := toCVolumeAttachment(a)
		if err != nil {
			return 0, nil, err
		}
		ca[x] = cva
	}

	return la, ca, nil
}

func toCVolumeAttachment(
	va *types.VolumeAttachment) (*C.volume_attachment, error) {

	ciid, err := toCInstanceID(va.InstanceID)
	if err != nil {
		return nil, err
	}

	cva := C.volume_attachment{
		volume_id:   C.CString(va.VolumeID),
		device_name: C.CString(va.DeviceName),
		mount_point: C.CString(va.MountPoint),
		status:      C.CString(va.Status),
	}
	cva.instance_id = ciid

	return &cva, nil
}

func toCInstanceID(i *types.InstanceID) (*C.instance_id, error) {
	return &C.instance_id{
		id: C.CString(i.ID),
	}, nil
}

func newWithConfig(configPath string) (types.Client, error) {
	config := gofig.New()
	if err := config.ReadConfigFile(configPath); err != nil {
		return nil, err
	}
	ctx := context.Background()
	if _, ok := context.PathConfig(ctx); !ok {
		pathConfig := utils.NewPathConfig(ctx, "", "")
		ctx = ctx.WithValue(context.PathConfigKey, pathConfig)
		registry.ProcessRegisteredConfigs(ctx)
	}
	return client.New(ctx, config)
}

func getClient(clientID C.h) (types.Client, error) {
	clientsRWL.RLock()
	defer clientsRWL.RUnlock()
	c, ok := clients[clientID]
	if !ok {
		return nil, goof.Newf("invalid client ID %d", int64(clientID))
	}
	return c, nil
}

var (
	rng = rand.New(rand.NewSource(time.Now().Unix()))
)

func newClientID() (*C.h, error) {
	buf := make([]byte, 16)
	if _, err := rng.Read(buf[:]); err != nil {
		return nil, err
	}
	buf[8] = (buf[8] | 0x40) & 0x7F
	buf[6] = (buf[6] & 0xF) | (4 << 4)
	u, _ := binary.Uvarint(buf)
	return C.new_client_id(C.ulonglong(u)), nil
}
