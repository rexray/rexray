// +build windows

package executor

func getHostName() (string, error) {
	return "windows", nil
}
