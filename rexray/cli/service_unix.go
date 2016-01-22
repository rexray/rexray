// +build darwin dragonfly freebsd !android,linux netbsd openbsd solaris

package cli

import (
	"os"
	"syscall"
)

func findProcess(pid int) (*os.Process, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}

	// make sure that a process with this pid is actually running and to
	// which we have access
	if err := syscall.Kill(p.Pid, syscall.Signal(0)); err != nil {
		if err.Error() == "no such process" {
			return nil, nil
		}
		return nil, err
	}

	return p, nil
}
