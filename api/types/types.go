package types

import (
	"fmt"
	"runtime"
)

// LSX is the default  name of the libStorage executor for the current OS.
var LSX string

func init() {
	if runtime.GOOS == "windows" {
		LSX = "lsx-windows.exe"
	} else {
		LSX = fmt.Sprintf("lsx-%s", runtime.GOOS)
	}
}
