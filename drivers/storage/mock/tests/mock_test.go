package mock

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/stretchr/testify/assert"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/server"
	"github.com/emccode/libstorage/api/server/executors"
	apitests "github.com/emccode/libstorage/api/tests"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"

	// load the  driver

	"github.com/emccode/libstorage/drivers/storage/mock"
	mockx "github.com/emccode/libstorage/drivers/storage/mock/executor"
)

var (
	lsxbin string

	lsxLinuxInfo, _  = executors.ExecutorInfoInspect("lsx-linux", false)
	lsxDarwinInfo, _ = executors.ExecutorInfoInspect("lsx-darwin", false)
	//lsxWindowsInfo, _ = executors.ExecutorInfoInspect("lsx-windows.exe", false)

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
	if travis, _ := strconv.ParseBool(os.Getenv("TRAVIS")); !travis {
		// semaphore.Unlink(types.LSX)
	}
}

func TestMain(m *testing.M) {
	server.CloseOnAbort()
	ec := m.Run()
	os.Exit(ec)
}

func TestStorageDriverVolumes(t *testing.T) {
	apitests.Run(t, mock.Name, configYAML,
		func(config gofig.Config, client types.Client, t *testing.T) {

			vols, err := client.Storage().Volumes(
				context.Background().WithValue(
					context.ServiceKey, mock.Name),
				&types.VolumesOpts{Attachments: true, Opts: utils.NewStore()})
			assert.NoError(t, err)
			assert.Len(t, vols, 1)
		})
}

func TestClient(t *testing.T) {
	apitests.Run(t, mock.Name, configYAML,
		func(config gofig.Config, client types.Client, t *testing.T) {

			ctx := context.Background()

			iid, err := client.Executor().InstanceID(
				ctx.WithValue(context.ServiceKey, mock.Name),
				utils.NewStore())
			assert.NoError(t, err)
			assert.NotNil(t, iid)

			iid, err = client.Executor().InstanceID(
				ctx.WithValue(context.ServiceKey, "mock2"),
				utils.NewStore())
			assert.NoError(t, err)
			assert.NotNil(t, iid)

			iid, err = client.Executor().InstanceID(
				ctx.WithValue(context.ServiceKey, "mock3"),
				utils.NewStore())
			assert.NoError(t, err)
			assert.NotNil(t, iid)
		})
}

func TestInstanceID(t *testing.T) {
	iid := mockx.GetInstanceID()
	iid.Formatted = true
	iid.Driver = mock.Name
	apitests.Run(
		t, mock.Name, nil,
		(&apitests.InstanceIDTest{
			Driver:   mock.Name,
			Expected: iid,
		}).Test)
}

func TestInstance(t *testing.T) {
	iid := mockx.GetInstanceID()
	iid.Formatted = true
	iid.Driver = mock.Name
	apitests.Run(
		t, mock.Name, nil,
		(&apitests.InstanceTest{
			Driver: mock.Name,
			Expected: &types.Instance{
				InstanceID: iid,
				Name:       "mockInstance",
			},
		}).Test)
}

func TestRoot(t *testing.T) {
	apitests.Run(t, mock.Name, configYAML, apitests.TestRoot)
}

func TestServices(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().Services(nil)
		assert.NoError(t, err)
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
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().ServiceInspect(nil, "mock2")
		assert.NoError(t, err)
		assert.Equal(t, "mock2", reply.Name)
		assert.Equal(t, mock.Name, reply.Driver.Name)
		assert.True(t, reply.Driver.NextDevice.Ignore)
		assert.Equal(t, "xvd", reply.Driver.NextDevice.Prefix)
		assert.Equal(t, `\w`, reply.Driver.NextDevice.Pattern)
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestSnapshotsByService(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().SnapshotsByService(nil, mock.Name)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)
		_, ok := reply["snap-000"]
		assert.Equal(t, ok, true)
		assert.Equal(t, reply["snap-000"].Name, "Snapshot 0")
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumes(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().Volumes(nil, false)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)
		assert.Len(t, reply, 3)
		assert.Len(t, reply[mock.Name], 3)
		assert.Len(t, reply[mock.Name]["vol-000"].Attachments, 0)
		assert.Len(t, reply["mock2"]["vol-000"].Attachments, 0)
		assert.Len(t, reply["mock3"]["vol-000"].Attachments, 0)
	}

	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumesWithAttachments(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().Volumes(nil, true)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)
		assert.Len(t, reply, 3)
		assert.Len(t, reply[mock.Name], 1)
		assert.Len(t, reply[mock.Name]["vol-000"].Attachments, 3)
		assert.Equal(
			t, "/var/log", reply[mock.Name]["vol-000"].Attachments[0].MountPoint)
		assert.Equal(
			t, "/home", reply[mock.Name]["vol-000"].Attachments[1].MountPoint)
		assert.Equal(
			t, "/net/share", reply[mock.Name]["vol-000"].Attachments[2].MountPoint)
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumesWithAttachmentsNoLocalDevices(t *testing.T) {
	t.SkipNow()
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().Volumes(nil, true)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)
		assert.Len(t, reply, 3)
		assert.Len(t, reply[mock.Name], 1)
		assert.Len(t, reply[mock.Name]["vol-000"].Attachments, 3)
		assert.NotEqual(
			t, "/var/log", reply[mock.Name]["vol-000"].Attachments[0].MountPoint)
		assert.NotEqual(
			t, "/home", reply[mock.Name]["vol-000"].Attachments[1].MountPoint)
		assert.NotEqual(
			t, "/net/share", reply[mock.Name]["vol-000"].Attachments[2].MountPoint)
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumesByService(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().VolumesByService(nil, mock.Name, false)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)
		_, ok := reply["vol-000"]
		assert.Equal(t, ok, true)
		assert.Equal(t, reply["vol-000"].AvailabilityZone, "zone-000")
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumeCreate(t *testing.T) {

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		volumeName := "Volume 001"
		availabilityZone := "US"
		iops := int64(1000)
		size := int64(10240)
		volType := "myType"

		opts := map[string]interface{}{
			"priority": 2,
			"owner":    "root@example.com",
		}

		volumeCreateRequest := &types.VolumeCreateRequest{
			Name:             volumeName,
			AvailabilityZone: &availabilityZone,
			IOPS:             &iops,
			Size:             &size,
			Type:             &volType,
			Opts:             opts,
		}

		reply, err := client.API().VolumeCreate(
			nil, mock.Name, volumeCreateRequest)
		assert.NoError(t, err)
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

func TestVolumeCreateWithError(t *testing.T) {

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		volumeName := "Volume 010"
		availabilityZone := "US"
		iops := int64(1000)
		size := int64(10240)
		volType := "myType"

		opts := map[string]interface{}{
			"priority": 2,
			"owner":    "root@example.com",
		}

		volumeCreateRequest := &types.VolumeCreateRequest{
			Name:             volumeName,
			AvailabilityZone: &availabilityZone,
			IOPS:             &iops,
			Size:             &size,
			Type:             &volType,
			Opts:             opts,
		}

		expectedError := goof.NewHTTPError(goof.WithFieldE(
			"iops", volumeCreateRequest.IOPS,
			"iops required",
			goof.WithFieldE(
				"size", volumeCreateRequest.Size,
				"size required",
				goof.New("bzzzzT BROKEN"),
			),
		), 500)

		_, err := client.API().VolumeCreate(
			nil, mock.Name, volumeCreateRequest)

		assert.Error(t, err)

		expBuf, _ := json.Marshal(expectedError)
		actBuf, _ := json.Marshal(err)
		assert.EqualValues(t, expBuf, actBuf)

		apitests.LogAsJSON(err, t)
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumeRemove(t *testing.T) {

	tf1 := func(config gofig.Config, client types.Client, t *testing.T) {
		err := client.API().VolumeRemove(nil, mock.Name, "vol-000")
		assert.NoError(t, err)
	}

	apitests.Run(t, mock.Name, configYAML, tf1, tf1)

	tf2 := func(config gofig.Config, client types.Client, t *testing.T) {
		err := client.API().VolumeRemove(nil, mock.Name, "vol-000")
		assert.Error(t, err)
		httpErr := err.(goof.HTTPError)
		assert.Equal(t, "resource not found", httpErr.Error())
		assert.Equal(t, 404, httpErr.Status())
	}

	apitests.RunGroup(t, mock.Name, configYAML, tf1, tf2)
}

func TestVolumeSnapshot(t *testing.T) {

	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		volumeID := "vol-000"
		snapshotName := "snapshot1"
		opts := map[string]interface{}{
			"priority": 2,
		}

		request := &types.VolumeSnapshotRequest{
			SnapshotName: snapshotName,
			Opts:         opts,
		}

		reply, err := client.API().VolumeSnapshot(
			nil, mock.Name, volumeID, request)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)

		assert.Equal(t, snapshotName, reply.Name)
		assert.Equal(t, volumeID, reply.VolumeID)

	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestSnapshots(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().Snapshots(nil)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)
		_, ok := reply[mock.Name]
		assert.Equal(t, true, ok)
		assert.Equal(t, 3, len(reply[mock.Name]))
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestSnapshotInspect(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		reply, err := client.API().SnapshotInspect(nil, mock.Name, "snap-000")
		assert.NoError(t, err)
		assert.Equal(t, "Snapshot 0", reply.Name)
		assert.Equal(t, "snap-000", reply.ID)
		assert.Equal(t, "vol-000", reply.VolumeID)
		assert.Equal(t, int64(100), reply.VolumeSize)
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumeCreateFromSnapshot(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		volumeName := "Volume from snap-000"

		availabilityZone := "US"
		iops := int64(1000)
		size := int64(10240)
		volType := "myType"

		opts := map[string]interface{}{
			"priority": 2,
			"owner":    "root@example.com",
		}

		snapshotCreateRequest := &types.VolumeCreateRequest{
			Name:             volumeName,
			AvailabilityZone: &availabilityZone,
			IOPS:             &iops,
			Size:             &size,
			Type:             &volType,
			Opts:             opts,
		}

		reply, err := client.API().VolumeCreateFromSnapshot(nil,
			mock.Name, "snap-000", snapshotCreateRequest)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)

		assert.Equal(t, volumeName, reply.Name)
		assert.Equal(t, opts["priority"], 2)
		assert.Equal(t, opts["owner"], "root@example.com")

	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestSnapshotRemove(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		assert.NoError(t, client.API().SnapshotRemove(
			nil, mock.Name, "snap-000"))
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestSnapshotCopy(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		snapshotName := "Snapshot from snap-000"

		opts := map[string]interface{}{
			"priority": 2,
			"owner":    "root@example.com",
		}

		request := &types.SnapshotCopyRequest{
			SnapshotName: snapshotName,
			Opts:         opts,
		}

		reply, err := client.API().SnapshotCopy(nil, mock.Name, "snap-000", request)
		assert.NoError(t, err)
		apitests.LogAsJSON(reply, t)

		assert.Equal(t, snapshotName, reply.Name)
		assert.Equal(t, opts["priority"], 2)
		assert.Equal(t, opts["owner"], "root@example.com")

	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumeAttach(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {

		opts := map[string]interface{}{
			"priority": 2,
			"owner":    "root@example.com",
		}

		nd, _ := client.Executor().NextDevice(
			context.Background().WithValue(context.ServiceKey, mock.Name),
			utils.NewStore())
		request := &types.VolumeAttachRequest{
			NextDeviceName: &nd,
			Opts:           opts,
		}

		reply, attTokn, err := client.API().VolumeAttach(
			nil, mock.Name, "vol-001", request)
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}
		assert.Equal(t, "1234", attTokn)
		assert.Equal(
			t, "vol-001", reply.ID)
		assert.Equal(
			t, "/dev/xvde", reply.Attachments[0].DeviceName)

		reply, attTokn, err = client.API().VolumeAttach(
			nil, mock.Name, "vol-002", request)
		assert.NoError(t, err)
		assert.Equal(t, "1234", attTokn)
		assert.Equal(
			t, "vol-002", reply.ID)
		assert.Equal(
			t, "/dev/xvde", reply.Attachments[0].DeviceName)

	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumeDetach(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {
		request := &types.VolumeDetachRequest{}
		vol, err := client.API().VolumeDetach(nil, mock.Name, "vol-000", request)
		assert.NoError(t, err)
		assert.Equal(t, "vol-000", vol.ID)
		assert.Len(t, vol.Attachments, 0)
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumeDetachAllForService(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {

		opts := map[string]interface{}{
			"priority": 2,
			"owner":    "root@example.com",
		}

		nd, err := client.Executor().NextDevice(
			context.Background().WithValue(context.ServiceKey, mock.Name),
			utils.NewStore())
		assert.NoError(t, err)
		request := &types.VolumeAttachRequest{
			NextDeviceName: &nd,
			Opts:           opts,
		}

		_, _, err = client.API().VolumeAttach(nil, "mock2", "vol-000", request)
		assert.NoError(t, err)
		var vol *types.Volume
		vol, err = client.API().VolumeInspect(nil, "mock2", "vol-000", true)
		assert.NoError(t, err)
		assert.Len(
			t, vol.Attachments, 1)

		requestD := &types.VolumeDetachRequest{
			Opts: opts,
		}
		_, err = client.API().VolumeDetachAllForService(nil, mock.Name, requestD)
		assert.NoError(t, err)
	}
	apitests.Run(t, mock.Name, configYAML, tf)
}

func TestVolumeDetachAll(t *testing.T) {
	tf := func(config gofig.Config, client types.Client, t *testing.T) {

		opts := map[string]interface{}{
			"priority": 2,
			"owner":    "root@example.com",
		}

		request := &types.VolumeDetachRequest{
			Opts: opts,
		}

		_, err := client.API().VolumeDetachAll(nil, request)
		assert.NoError(t, err)
		vol, err := client.API().VolumeInspect(nil, mock.Name, "vol-000", true)
		assert.NoError(t, err)
		assert.Len(
			t, vol.Attachments, 0)
		vol, err = client.API().VolumeInspect(nil, mock.Name, "vol-001", true)
		assert.NoError(t, err)
		assert.Len(
			t, vol.Attachments, 0)
		vol, err = client.API().VolumeInspect(nil, mock.Name, "vol-002", true)
		assert.NoError(t, err)
		assert.Len(
			t, vol.Attachments, 0)
		vol, err = client.API().VolumeInspect(nil, "mock2", "vol-000", true)
		assert.NoError(t, err)
		assert.Len(
			t, vol.Attachments, 0)

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
		apitests.TestHeadExecutorDarwin)
	//apitests.TestHeadExecutorWindows)
}

func TestExecutorGet(t *testing.T) {
	apitests.RunGroup(
		t, mock.Name, configYAML,
		apitests.TestGetExecutorLinux,
		apitests.TestGetExecutorDarwin)
	//apitests.TestGetExecutorWindows)
}
