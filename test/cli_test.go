package test

import (
	"fmt"
	"testing"

	"github.com/emccode/rexray/drivers/mock"
	"github.com/emccode/rexray/rexray/cli"
)

var defaultFlags = []string{
	fmt.Sprintf("--osDrivers=%s", mock.MockOSDriverName),
	fmt.Sprintf("--volumeDrivers=%s", mock.MockVolDriverName),
	fmt.Sprintf("--storageDrivers=%s", mock.MockStorDriverName),
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
