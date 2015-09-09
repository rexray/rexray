// +build darwin freebsd openbsd netbsd dragonfly

package terminal

import "syscall"

const ioctlReadTermios = syscall.TIOCGETA

type Termios syscall.Termios
