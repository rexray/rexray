// +build !windows

package executor

import "os/exec"

const (
	newline = 10
)

func getHostName() (string, error) {
	buf, err := exec.Command("hostname").Output()
	if err != nil {
		return "", err
	}
	if buf[len(buf)-1] == 10 {
		buf = buf[:len(buf)-1]
	}
	return string(buf), nil
}
