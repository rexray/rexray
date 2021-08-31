package main

import "C"

import (
	"os"
	"unsafe"

	log "github.com/sirupsen/logrus"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/server"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	apicfg "github.com/AVENTER-UG/rexray/libstorage/api/utils/config"
)

// closeOnAbort is a helper function that can be called by programs, such as
// tests or a command line or service application.
//
//export closeOnAbort
func closeOnAbort() {
	server.CloseOnAbort()
}

// serve starts the libStorage server, blocking the current thread until the
// server is shutdown.
//
//   char*  host    the golang network-style address. if no value is provided
//                  then a random TCP port bound to localhost is used
//
//   short  argc    the length of the argv array
//
//   void*  argv    a pointer to an array of char* strings that represent the
//                  drivers and corresponding service names to load. if the
//                  array is odd-numbered then the service for the trailing
//                  driver takes the name of the driver
//
// char* serve(char* host, int argc, void* argv);
//export serve
func serve(
	host *C.char, argc C.int, argv unsafe.Pointer) *C.char {

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

	log.WithFields(log.Fields{
		"host": szHost,
		"args": args,
	}).Info("serving")

	ctx := context.Background()
	ctx = ctx.WithValue(context.PathConfigKey, utils.NewPathConfig(ctx, "", ""))
	registry.ProcessRegisteredConfigs(ctx)

	config, err := apicfg.NewConfig(ctx)
	if err != nil {
		return C.CString(err.Error())
	}

	if len(szHost) > 0 {
		os.Setenv("LIBSTORAGE_HOST", szHost)
	}

	_, errs, err := server.Serve(ctx, config)
	if err != nil {
		return C.CString(err.Error())
	}

	<-errs
	return nil
}
