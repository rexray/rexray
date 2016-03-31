package main

import (
	// load the driver
	"github.com/emccode/libstorage/drivers/storage/mock"

	"github.com/emccode/libstorage/cli/servers"
)

func main() {
	servers.Run(mock.Name1, "", false)
}
