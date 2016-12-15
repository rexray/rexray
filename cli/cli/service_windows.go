// +build !rexray_build_type_client

package cli

import (
	"os"
)

func findProcess(pid int) (*os.Process, error) {
	return os.FindProcess(pid)
}
