package core

import (
	apitypes "github.com/emccode/libstorage/api/types"
)

var (
	// Version of REX-Ray.
	Version *apitypes.VersionInfo
)

type os string

func (o os) String() string {
	switch o {
	case "linux":
		return "Linux"
	case "darwin":
		return "Darwin"
	case "windows":
		return "Windows"
	default:
		return string(o)
	}
}

type arch string

func (a arch) String() string {
	switch a {
	case "386":
		return "i386"
	case "amd64":
		return "x86_64"
	default:
		return string(a)
	}
}
