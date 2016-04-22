package mock

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/akutz/gofig"
	"github.com/stretchr/testify/assert"

	"github.com/emccode/libstorage/api/server/executors"
	apitests "github.com/emccode/libstorage/api/tests"
	apihttp "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/client"

	// load the  driver
	"github.com/emccode/libstorage/drivers/storage/mock"
	mockx "github.com/emccode/libstorage/drivers/storage/mock/executor"
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
	apitests.Run(t, mock.Name, configYAML, apitests.TestRoot)
}

func TestServices(t *testing.T) {
	tf := func(config gofig.Config, client client.Client, t *testing.T) {
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
	tf := func(config gofig.Config, client client.Client, t *testing.T) {
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
	tf := func(config gofig.Config, client client.Client, t *testing.T) {
		reply, err := client.Volumes()
		if err != nil {
			t.Fatal(err)
		}
		apitests.LogAsJSON(reply, t)
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumeCreate(t *testing.T) {

	tf := func(config gofig.Config, client client.Client, t *testing.T) {
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
	tf := func(config gofig.Config, client client.Client, t *testing.T) {
		err := client.VolumeRemove("mock", "vol-000")
		if err != nil {
			t.Fatal(err)
		}
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumeSnapshot(t *testing.T) {

	tf := func(config gofig.Config, client client.Client, t *testing.T) {
		volumeID := "vol-000"
		snapshotName := "snapshot1"
		opts := map[string]interface{}{
			"priority": 2,
		}

		request := &apihttp.VolumeSnapshotRequest{
			SnapshotName: snapshotName,
			Opts: opts,
		}

		reply, err := client.VolumeSnapshot("mock", volumeID,request)
		if err != nil {
			t.Fatal(err)
		}
		apitests.LogAsJSON(reply, t)

		assert.Equal(t, snapshotName, reply.Name)
		assert.Equal(t, volumeID, reply.VolumeID)

	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestSnapshotRemove(t *testing.T) {
	tf := func(config gofig.Config, client client.Client, t *testing.T) {
		err := client.VolumeRemove("mock", "vol-000")
		if err != nil {
			t.Fatal(err)
		}
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestExecutors(t *testing.T) {
	apitests.Run(t, mock.Name, configYAML, apitests.TestExecutors)
}

func TestExecutorHead(t *testing.T) {
	apitests.RunGroup(
		t, mock.Name, configYAML,
		apitests.TestHeadExecutorLinux,
		apitests.TestHeadExecutorDarwin,
		apitests.TestHeadExecutorWindows)
}

func TestExecutorGet(t *testing.T) {
	apitests.RunGroup(
		t, mock.Name, configYAML,
		apitests.TestGetExecutorLinux,
		apitests.TestGetExecutorDarwin,
		apitests.TestGetExecutorWindows)
}

func TestInstanceID(t *testing.T) {
	apitests.RunGroup(
		t, mock.Name, nil,
		(&apitests.InstanceIDTest{
			Driver:   mock.Name,
			Expected: mockx.GetInstanceID(),
		}).Test)
}
