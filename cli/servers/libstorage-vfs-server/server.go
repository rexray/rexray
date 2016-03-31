package main

import (
	// load the driver
	"github.com/emccode/libstorage/drivers/storage/vfs"

	"github.com/emccode/libstorage/cli/servers"
)

func main() {
	servers.Run(vfs.Name, "", false)
}
