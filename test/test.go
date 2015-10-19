// Package test is a package that exists purely to provide coverage for the
// following packages:
//
//   - github.com/emccode/rexray
//   - github.com/emccode/rexray/core
//
// Because of the way drivers are loaded, it's not possible for the core
// package to share mock drivers with any other package. Thus in order to
// prevent duplicate mock drivers, a single test pacakge to provide coverage
// of packages requiring mock drivers has been created.
package test

import (
	// loads the packages to alleviate test warnings about no depdencies
	_ "github.com/emccode/rexray"
	_ "github.com/emccode/rexray/core"
)
