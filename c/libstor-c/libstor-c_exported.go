package main

//#cgo CFLAGS: -I${SRCDIR}/..
//#include "types.h"
import "C"

import (
	"sync"

	"github.com/emccode/libstorage/client"
)

var (
	clients    = map[C.hc]client.Client{}
	clientsRWL = sync.RWMutex{}
)

//export new
func new(configPath *C.char) (C.hc, *C.error) {
	clientsRWL.Lock()
	defer clientsRWL.Unlock()

	ch, err := newClientID()
	if err != nil {
		return 0, toCError(err)
	}

	c, err := newWithConfig(C.GoString(configPath))
	if err != nil {
		return 0, toCError(err)
	}

	clients[ch] = c
	return ch, nil
}

//export close
func close(clientID C.hc) {
	clientsRWL.Lock()
	defer clientsRWL.Unlock()
	delete(clients, clientID)
}

//export mount
func mount(clientID C.hc, service, volumeID *C.char) *C.error {
	return nil
}

//export mount_with_name
func mount_with_name(clientID C.hc, service, volumeName *C.char) *C.error {
	return nil
}

//export unmount
func unmount(clientID C.hc, service, volumeID *C.char) *C.error {
	return nil
}

//export unmount_with_name
func unmount_with_name(clientID C.hc, service, volumeName *C.char) *C.error {
	return nil
}
