package main

//#cgo CFLAGS: -I${SRCDIR}
//#include "libstor-c_types.h"
import "C"

import (
	"sync"
	"unsafe"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

var (
	clients    = map[C.h]types.Client{}
	clientsRWL = sync.RWMutex{}
)

//export new_client
func new_client(configPath *C.char) C.result {
	clientsRWL.Lock()
	defer clientsRWL.Unlock()

	result := C.result{}

	client_id, err := newClientID()
	if err != nil {
		result.err = C.CString(err.Error())
		return result
	}

	c, err := newWithConfig(C.GoString(configPath))
	if err != nil {
		result.err = C.CString(err.Error())
		return result
	}

	clients[*client_id] = c
	result.val = unsafe.Pointer(client_id)
	return result
}

//export close
func close(clientID C.h) {
	clientsRWL.Lock()
	defer clientsRWL.Unlock()
	delete(clients, clientID)
}

//export volumes
func volumes(clientID C.h, attachments C.short) C.result {
	result := C.result{}

	c, err := getClient(clientID)
	if err != nil {
		result.err = C.CString(err.Error())
		return result
	}

	svcToVolMap, err := c.API().Volumes(
		nil, types.VolumeAttachmentsTypes(attachments))
	if err != nil {
		result.err = C.CString(err.Error())
		return result
	}

	service_names := []*C.char{}
	volume_maps := []*C.volume_map{}

	for service, volMap := range svcToVolMap {

		service_names = append(service_names, C.CString(service))

		volumes := []*C.volume{}
		volume_ids := []*C.char{}

		for volumeID, volume := range volMap {
			volume_ids = append(volume_ids, C.CString(volumeID))
			cVol, err := toCVolume(volume)
			if err != nil {
				result.err = C.CString(err.Error())
				return result
			}
			volumes = append(volumes, cVol)
		}

		lcVolMap := len(volMap)
		cVolMap := C.new_volume_map()
		cVolMap.volumes_c = C.int(lcVolMap)
		if lcVolMap > 0 {
			cVolMap.volume_ids = &volume_ids[0]
			cVolMap.volumes = &volumes[0]
		}
		volume_maps = append(volume_maps, cVolMap)
	}

	lcSvcVolMap := len(svcToVolMap)
	cSvcVolMap := C.new_service_volume_map()
	cSvcVolMap.services_c = C.int(lcSvcVolMap)
	if lcSvcVolMap > 0 {
		cSvcVolMap.service_names = &service_names[0]
		cSvcVolMap.volumes = &volume_maps[0]
	}

	result.val = unsafe.Pointer(cSvcVolMap)
	return result
}

//export mount
func mount(clientID C.h, service, volumeID *C.char) C.result {
	return C.result{}
}

//export mount_with_name
func mount_with_name(clientID C.h, service, volumeName *C.char) C.result {
	return C.result{}
}

//export unmount
func unmount(clientID C.h, service, volumeID *C.char) C.result {
	return C.result{}
}

//export unmount_with_name
func unmount_with_name(clientID C.h, service, volumeName *C.char) C.result {
	return C.result{}
}
