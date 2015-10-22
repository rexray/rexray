package test

import (
	"fmt"
	"testing"

	"github.com/emccode/rexray/rexray/cli"
)

var defaultFlags = []string{
	fmt.Sprintf("--osDrivers=%s", mockOSDriverName),
	fmt.Sprintf("--volumeDrivers=%s", mockVolDriverName),
	fmt.Sprintf("--storageDrivers=%s", mockStorDriverName),
}

func a(a ...string) {
	a = append(a, defaultFlags...)
	cli.ExecuteWithArgs(a...)
}

func TestVolumeGetYaml(t *testing.T) {
	a("volume", "get")
}

func TestVolumeGetJSON(t *testing.T) {
	a("volume", "get", "-f", "json")
}
