// Package nfs provides utilities for mounting/unmounting NFS exported
// directories
package nfs

import (
	"fmt"
	"os/exec"
)

// Supported queries the underlying system to check if the required system
// executables are present
// If not, it returns an error
func Supported() error {
	if _, err := exec.Command("/bin/ls", "/sbin/mount.nfs").CombinedOutput(); err != nil {
		return fmt.Errorf("Required binary /sbin/mount.nfs is missing")
	}
	if _, err := exec.Command("/bin/ls", "/sbin/mount.nfs4").CombinedOutput(); err != nil {
		return fmt.Errorf("Required binary /sbin/mount.nfs4 is missing")
	}
	return nil
}
