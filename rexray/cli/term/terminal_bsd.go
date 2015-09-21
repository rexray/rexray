// +build darwin freebsd openbsd netbsd dragonfly

package term

import "syscall"

const ioctlReadTermios = syscall.TIOCGETA

type Termios syscall.Termios
