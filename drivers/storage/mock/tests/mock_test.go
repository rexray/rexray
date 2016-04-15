package mock

import (
	"testing"

	"github.com/akutz/gofig"
	"github.com/stretchr/testify/assert"

	apiclient "github.com/emccode/libstorage/api/client"
	"github.com/emccode/libstorage/api/server/executors"
	apitests "github.com/emccode/libstorage/api/tests"
	"github.com/emccode/libstorage/api/types"

	// load the  driver
	"github.com/emccode/libstorage/drivers/storage/mock"
)

var (
	lsxLinuxInfo, _   = executors.ExecutorInfoInspect("lsx-linux", false)
	lsxDarwinInfo, _  = executors.ExecutorInfoInspect("lsx-darwin", false)
	lsxWindowsInfo, _ = executors.ExecutorInfoInspect("lsx-windows.exe", false)

	configYAML = []byte(`
libstorage:
  driver: mock
  server:
    services:
      mock2:
      mock3:
`)
)

func TestRoot(t *testing.T) {

	tf := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.Root()
		if err != nil {
			t.Fatal(err)
		}
		apitests.LogAsJSON(reply, t)
	}

	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumes(t *testing.T) {

	tf := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.Volumes()
		if err != nil {
			t.Fatal(err)
		}
		apitests.LogAsJSON(reply, t)
	}

	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestExecutors(t *testing.T) {

	apitests.Run(t, mock.Name, configYAML,
		func(config gofig.Config, client apiclient.Client, t *testing.T) {
			reply, err := client.Executors()
			if err != nil {
				t.Fatal(err)
			}
			assertLSXWindows(t, reply["lsx-windows.exe"])
			assertLSXLinux(t, reply["lsx-linux"])
			assertLSXDarwin(t, reply["lsx-darwin"])
		})
}

func TestExecutorHead(t *testing.T) {

	tf1 := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.ExecutorHead("lsx-windows.exe")
		if err != nil {
			t.Fatal(err)
		}
		assertLSXWindows(t, reply)
	}

	tf2 := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.ExecutorHead("lsx-linux")
		if err != nil {
			t.Fatal(err)
		}
		assertLSXLinux(t, reply)
	}

	tf3 := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.ExecutorHead("lsx-darwin")
		if err != nil {
			t.Fatal(err)
		}
		assertLSXDarwin(t, reply)
	}

	apitests.RunGroup(t, mock.Name, configYAML, tf1, tf2, tf3)
}

func assertLSXWindows(t *testing.T, i *types.ExecutorInfo) {
	assert.Equal(t, lsxWindowsInfo.Name, i.Name)
	assert.EqualValues(t, lsxWindowsInfo.Size, i.Size)
	assert.Equal(t, lsxWindowsInfo.MD5Checksum, i.MD5Checksum)
}

func assertLSXLinux(t *testing.T, i *types.ExecutorInfo) {
	assert.Equal(t, lsxLinuxInfo.Name, i.Name)
	assert.EqualValues(t, lsxLinuxInfo.Size, i.Size)
	assert.Equal(t, lsxLinuxInfo.MD5Checksum, i.MD5Checksum)
}

func assertLSXDarwin(t *testing.T, i *types.ExecutorInfo) {
	assert.Equal(t, lsxDarwinInfo.Name, i.Name)
	assert.EqualValues(t, lsxDarwinInfo.Size, i.Size)
	assert.Equal(t, lsxDarwinInfo.MD5Checksum, i.MD5Checksum)
}
