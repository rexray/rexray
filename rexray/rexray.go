package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	glog "github.com/akutz/golf/logrus"
	"github.com/emccode/rexray/rexray/cli"

	// This blank import loads the drivers package
	_ "github.com/emccode/rexray/drivers"
)

func main() {
	log.SetFormatter(&glog.TextFormatter{log.TextFormatter{}})

	defer func() {
		r := recover()
		if r == nil {
			os.Exit(0)
		}
		switch r := r.(type) {
		case int:
			log.Debugf("exiting with error code %d", r)
			os.Exit(r)
		case error:
			log.Panic(r)
		default:
			log.Debugf("exiting with default error code 1, r=%v", r)
			os.Exit(1)
		}
	}()

	cli.Exec()
}
