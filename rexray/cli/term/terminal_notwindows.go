// Based on ssh/terminal:
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux darwin freebsd openbsd netbsd dragonfly

package term

import (
	"syscall"
	"unsafe"
)

// IsTerminal returns true if the given file descriptor is a terminal.
func IsTerminal() bool {
	stdout := syscall.Stdout
	var stdoutTermios Termios
	_, _, errStdout := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(stdout),
		ioctlReadTermios, uintptr(unsafe.Pointer(&stdoutTermios)), 0, 0, 0)
	if errStdout == 0 {
		return true
	}

	stderr := syscall.Stderr
	var stderrTermios Termios
	_, _, errStderr := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(stderr),
		ioctlReadTermios, uintptr(unsafe.Pointer(&stderrTermios)), 0, 0, 0)

	return errStderr == 0
}
