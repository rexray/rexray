package main

import (
	// This blank import loads the drivers package
	_ "github.com/emccode/rexray/drivers"

	"github.com/emccode/rexray/rexray/cli"
)

func main() {
	cli.Execute()
}
