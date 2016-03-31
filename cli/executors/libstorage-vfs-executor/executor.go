package main

import (
	"github.com/emccode/libstorage/cli/executors"

	// load the executor
	_ "github.com/emccode/libstorage/drivers/storage/vfs/executor"
)

func main() {
	executors.Run()
}
