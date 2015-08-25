// +build !exclude_module

package daemon

import (
	"log"
	"os"

	"github.com/emccode/rexray/daemon/module"
	_ "github.com/emccode/rexray/drivers/storage"
)

func Start(host string, init chan error, stop <-chan os.Signal) {

	isErr := false

	initModErr := module.InitializeDefaultModules()

	if initModErr != nil {
		init <- initModErr
		isErr = true
		log.Println("Default module(s) failed to initialize")

	} else {

		startModErr := module.StartDefaultModules()
		if startModErr != nil {
			init <- startModErr
			isErr = true
			log.Println("Default module(s) failed to start")
		}

	}

	log.Println("Service sent registered modules start signals")

	// if there is a channel receiving initialization errors go ahead and
	// close it so that callers reading this channel will know that the
	// initialization of the daemon is complete
	if init != nil {
		close(init)
	}

	// if there were initialization errors go ahead and return instead of
	// waiting for a stop signal
	if isErr {
		log.Println("Service initialized failed")
		return
	}

	log.Println("Service successfully initialized, waiting on stop signal")

	// if a channel to receive a stop signal is provided then block until
	// a stop signal is received
	if stop != nil {
		<-stop
		log.Println("Service received stop signal")
	}
}
