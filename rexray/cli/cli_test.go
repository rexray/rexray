// +build ignore

package cli

import (
	"fmt"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
)

var defaultFlags = []string{
	fmt.Sprintf("--osDrivers=%s", mock.MockOSDriverName),
	fmt.Sprintf("--volumeDrivers=%s", mock.MockVolDriverName),
	fmt.Sprintf("--storageDrivers=%s", mock.MockStorDriverName),
}

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	mock.RegisterMockDrivers()
	mock.RegisterBadMockDrivers()
	os.Exit(m.Run())
}

func a(t *testing.T, a ...string) {
	a = append(a, defaultFlags...)
	c := NewWithArgs(a...)
	c.Execute()
	s, _ := c.r.Config.ToJSON()
	t.Logf(s)
	k := "mockprovider.docker.minvolsize"
	t.Logf("%s=%v", k, c.r.Config.Get(k))
}

func TestVolumeGetYaml(t *testing.T) {
	a(t, "volume", "get")
}

func TestVolumeGetJSON(t *testing.T) {
	a(t, "volume", "get", "-f", "json")
}

func TestAdapterGet(t *testing.T) {
	a(t, "adapter", "get")
}

func TestVerbose(t *testing.T) {
	a(t, "-v")
}

func TestEnv(t *testing.T) {
	a(t, "env")
}

func TestMockDockerMinVolSize(t *testing.T) {
	a(t, "env", "--mockProviderDockerMinVolSize=128")
}
