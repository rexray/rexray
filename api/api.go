package api

import "github.com/emccode/libstorage/api/types"

var (
	// Version of the current REST API
	Version *types.VersionInfo
)

//go:generate make -C ../ ./api/api_generated.go
