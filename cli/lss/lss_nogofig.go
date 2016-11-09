// +build !gofig !pflag

package lss

import (
	"fmt"
	"os"
	"runtime"
)

// Run the server.
func Run() {
	fmt.Fprintf(os.Stderr, "lss-%s was built without gofig\n", runtime.GOOS)
	os.Exit(1)
}
