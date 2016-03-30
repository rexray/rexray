package main

import (
	"github.com/emccode/libstorage/cli/executor"

	// load the executor
	_ "github.com/emccode/libstorage/drivers/storage/mock/executor"
)

func main() {
	executor.Run()
}
