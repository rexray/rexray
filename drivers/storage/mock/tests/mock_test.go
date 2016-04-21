package mock

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/akutz/gofig"
	"github.com/stretchr/testify/assert"

	apiclient "github.com/emccode/libstorage/api/client"
	"github.com/emccode/libstorage/api/server/executors"
	apitests "github.com/emccode/libstorage/api/tests"
	"github.com/emccode/libstorage/api/types"
	apihttp "github.com/emccode/libstorage/api/types/http"

	// load the  driver
	"github.com/emccode/libstorage/drivers/storage/mock"
)

var (
	lsxbin string

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

func init() {
	if runtime.GOOS == "windows" {
		lsxbin = "lsx-windows.exe"
	} else {
		lsxbin = fmt.Sprintf("lsx-%s", runtime.GOOS)
	}
}

func TestRoot(t *testing.T) {
	tf := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.Root()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(reply), 5)
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestServices(t *testing.T) {
	tf := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.Services()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(reply), 3)

		_, ok := reply[mock.Name]
		assert.True(t, ok)

		_, ok = reply["mock2"]
		assert.True(t, ok)

		_, ok = reply["mock3"]
		assert.True(t, ok)
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestServiceInpspect(t *testing.T) {
	tf := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.ServiceInspect("mock2")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "mock2", reply.Name)
		assert.Equal(t, "mock", reply.Driver.Name)
		assert.False(t, reply.Driver.NextDevice.Ignore)
		assert.Equal(t, "xvd", reply.Driver.NextDevice.Prefix)
		assert.Equal(t, `\w`, reply.Driver.NextDevice.Pattern)
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

func TestVolumeCreate(t *testing.T) {

	tf := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		volumeName := "Volume 001"
		availabilityZone := "US"
		iops := int64(1000)
		size := int64(10240)
		volType := "myType"

		opts := map[string]interface{}{
			"priority": 2,
			"owner":    "root@example.com",
		}

		volumeCreateRequest := &apihttp.VolumeCreateRequest{
			Name:             volumeName,
			AvailabilityZone: &availabilityZone,
			IOPS:             &iops,
			Size:             &size,
			Type:             &volType,
			Opts:             opts,
		}

		reply, err := client.VolumeCreate("mock", volumeCreateRequest)
		if err != nil {
			t.Fatal(err)
		}
		apitests.LogAsJSON(reply, t)

		assert.Equal(t, availabilityZone, reply.AvailabilityZone)
		assert.Equal(t, iops, reply.IOPS)
		assert.Equal(t, volumeName, reply.Name)
		assert.Equal(t, size, reply.Size)
		assert.Equal(t, volType, reply.Type)
		assert.Equal(t, opts["priority"], 2)
		assert.Equal(t, opts["owner"], "root@example.com")

	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumeRemove(t *testing.T) {
	tf := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		err := client.VolumeRemove("mock", "vol-000")
		if err != nil {
			t.Fatal(err)
		}
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

func TestExecutorGet(t *testing.T) {

	tf1 := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.ExecutorGet("lsx-windows.exe")
		if err != nil {
			t.Fatal(err)
		}
		defer reply.Close()
		buf, err := ioutil.ReadAll(reply)
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(t, lsxWindowsInfo.Size, len(buf))
	}

	tf2 := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.ExecutorGet("lsx-linux")
		if err != nil {
			t.Fatal(err)
		}
		defer reply.Close()
		buf, err := ioutil.ReadAll(reply)
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(t, lsxLinuxInfo.Size, len(buf))
	}

	tf3 := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.ExecutorGet("lsx-darwin")
		if err != nil {
			t.Fatal(err)
		}
		defer reply.Close()
		buf, err := ioutil.ReadAll(reply)
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(t, lsxDarwinInfo.Size, len(buf))
	}

	apitests.RunGroup(t, mock.Name, configYAML, tf1, tf2, tf3)
}

const (
	mockInstanceIDJSON = `{"id":"12345","metadata":{"max":10,"min":0,"rad":"cool","totally":"tubular"}}
`
)

func TestInstanceID(t *testing.T) {

	tf1 := func(config gofig.Config, client apiclient.Client, t *testing.T) {
		reply, err := client.ExecutorGet(lsxbin)
		assert.NoError(t, err)
		defer reply.Close()

		f, err := ioutil.TempFile("", "")
		assert.NoError(t, err)

		_, err = io.Copy(f, reply)
		assert.NoError(t, err)

		err = f.Close()
		assert.NoError(t, err)

		os.Chmod(f.Name(), 0755)

		out, err := exec.Command(
			f.Name(), mock.Name, "instanceID").CombinedOutput()
		assert.NoError(t, err)
		assert.Equal(t, mockInstanceIDJSON, string(out))
	}

	apitests.RunGroup(t, mock.Name, nil, tf1)
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
