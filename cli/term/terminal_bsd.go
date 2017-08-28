// +build darwin freebsd openbsd netbsd dragonfly

package term

import "syscall"

const ioctlReadTermios = syscall.TIOCGETA

// Termios is the Terminal Input/Output structure
type Termios syscall.Termios
