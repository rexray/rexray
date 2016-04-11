package main

import (
	// load the driver
	"github.com/emccode/libstorage/drivers/storage/mock"

	"github.com/emccode/libstorage/cli/servers"
)

func main() {
	servers.Run("", false,
		mock.Name1, mock.Name1,
		mock.Name2, mock.Name2,
		mock.Name3, mock.Name3)
}
