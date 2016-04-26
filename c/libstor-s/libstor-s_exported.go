package main

import "C"

import (
	"unsafe"

	log "github.com/Sirupsen/logrus"

	"github.com/emccode/libstorage/api/server"
)

// serve starts the libStorage server, blocking the current thread until the
// server is shutdown.
//
//   char*  host    the golang network-style address. if no value is provided
//                  then a random TCP port bound to localhost is used
//
//   int    tls     a flag indicating whether or not to use TLS. >0 = true
//
//   short  argc    the length of the argv array
//
//   void*  argv    a pointer to an array of char* strings that represent the
//                  drivers and corresponding service names to load. if the
//                  array is odd-numbered then the service for the trailing
//                  driver takes the name of the driver
//
// char* serve(char* host, short tls, int argc, void* argv);
//export serve
func serve(
	host *C.char, tls C.short, argc C.int, argv unsafe.Pointer) *C.char {

	iargc := int(argc)
	args := make([]string, iargc)
	pargv := argv

	for x := 0; x < iargc; x++ {
		gostr := C.GoString((*C.char)(pargv))
		log.WithFields(log.Fields{
			"x":     x,
			"gostr": gostr,
		}).Info("parsed gostr")
		args[x] = gostr
		pargv = unsafe.Pointer(uintptr(pargv) + unsafe.Sizeof(gostr))
	}

	szHost := C.GoString(host)
	bTLS := tls > 0

	log.WithFields(log.Fields{
		"host": szHost,
		"tls":  bTLS,
		"args": args,
	}).Info("serving")

	if err := server.Run(szHost, bTLS, args...); err != nil {
		return C.CString(err.Error())
	}
	return nil
}
