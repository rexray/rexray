package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/emccode/rexray/rexray/commands"
	"os"
)

func main() {
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

	commands.Exec()
}
