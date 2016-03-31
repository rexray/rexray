package main

import (
	"github.com/emccode/libstorage/cli/executors"

	// load the executor
	_ "github.com/emccode/libstorage/drivers/storage/mock/executor"
)

func main() {
	executors.Run()
}
