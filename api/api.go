package api

import (
	"github.com/blang/semver"
)

var (
	// Version of the current REST API
	Version = semver.MustParse("1.0.0")

	// MinVersion represents Minimun REST API version supported
	MinVersion = semver.MustParse("1.0.0")
)
