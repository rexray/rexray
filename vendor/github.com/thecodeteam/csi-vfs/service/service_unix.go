// +build linux darwin

package service

import "syscall"

func getAvailableBytes(dir string) uint64 {
	var stat syscall.Statfs_t
	syscall.Statfs(dir, &stat)
	return stat.Bavail * uint64(stat.Bsize)
}
