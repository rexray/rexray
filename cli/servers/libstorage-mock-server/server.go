package main

import (
	// load the driver
	"github.com/emccode/libstorage/drivers/storage/mock"

	"github.com/emccode/libstorage/cli/servers"
)

func main() {
	servers.Run("", false,
		mock.Name, "service1",
		mock.Name, "service2",
		mock.Name, "service3")
}
