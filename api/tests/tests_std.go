package tests

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"
	"github.com/stretchr/testify/assert"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/client"
)

// TestRoot tests the GET / route.
var TestRoot = func(config gofig.Config, client client.Client, t *testing.T) {
	reply, err := client.API().Root(nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(reply), 5)
}

// InstanceIDTest is the test harness for testing the instance ID.
type InstanceIDTest struct {

	// Driver is the name of the driver/executor for which to get the instance
	// ID.
	Driver string

	// Expected is the expected instance ID value.
	Expected *types.InstanceID
}

// Test is the APITestFunc for the InstanceIDTest.
func (tt *InstanceIDTest) Test(
	config gofig.Config,
	client client.Client,
	t *testing.T) {

	expectedBuf, err := json.Marshal(tt.Expected)
	assert.NoError(t, err)
	expectedJSON := string(expectedBuf)

	iid, err := client.InstanceID(tt.Driver)
	assert.NoError(t, err)

	iidBuf, err := json.Marshal(iid)
	assert.NoError(t, err)
	iidJSON := string(iidBuf)

	assert.Equal(t, expectedJSON, iidJSON)
}

// NextDeviceTest is the test harness for testing getting the next device.
type NextDeviceTest struct {

	// Driver is the name of the driver/executor for which to get the next
	// device.
	Driver string

	// Expected is the expected next device.
	Expected string
}

// Test is the APITestFunc for the NextDeviceTest.
func (tt *NextDeviceTest) Test(
	config gofig.Config,
	client client.Client,
	t *testing.T) {

	reply, err := client.API().ExecutorGet(nil, lsxbin)
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
		f.Name(), tt.Driver, "nextDevice").CombinedOutput()
	assert.NoError(t, err)
	assert.Equal(t, tt.Expected, gotil.Trim(string(out)))
}

// LocalDevicesTest is the test harness for testing getting the local devices.
type LocalDevicesTest struct {

	// Driver is the name of the driver/executor for which to get the local
	// devices.
	Driver string

	// Expected is the expected local devices.
	Expected map[string]string
}

// Test is the APITestFunc for the LocalDevicesTest.
func (tt *LocalDevicesTest) Test(
	config gofig.Config,
	client client.Client,
	t *testing.T) {

	expectedBuf, err := json.Marshal(tt.Expected)
	assert.NoError(t, err)
	expectedJSON := string(expectedBuf)

	reply, err := client.API().ExecutorGet(nil, lsxbin)
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
		f.Name(), tt.Driver, "localDevices").CombinedOutput()
	assert.NoError(t, err)

	assert.Equal(t, expectedJSON, gotil.Trim(string(out)))
}
