package main

import (
	"fmt"
	"os"

	"github.com/emccode/libstorage/api/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(
			os.Stderr,
			"usage: libstor-server DRIVER [SERVICE] [DRIVER [SERVICE]]...")
		os.Exit(1)
	}

	server.CloseOnAbort()

	err := server.Run("", false, os.Args[1:]...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "libstor-server: error: %v\n", err)
		os.Exit(1)
	}
}
