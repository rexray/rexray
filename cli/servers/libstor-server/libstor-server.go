package main

import (
	"fmt"
	"os"

	"github.com/emccode/libstorage/cli/servers"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: libstor-server DRIVER")
		os.Exit(1)
	}

	if err := servers.Run("", false, os.Args[1:]...); err != nil {
		fmt.Fprintf(os.Stderr, "libstor-server: error: %v\n", err)
		os.Exit(1)
	}
}
