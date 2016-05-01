package tests

import (
	"encoding/json"
	"testing"

	"github.com/akutz/gofig"
	"github.com/stretchr/testify/assert"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
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

	iid, err := client.API().InstanceID(nil, tt.Driver)
	assert.NoError(t, err)

	iidBuf, err := json.Marshal(iid)
	assert.NoError(t, err)
	iidJSON := string(iidBuf)

	assert.Equal(t, expectedJSON, iidJSON)
}

// InstanceTest is the test harness for testing instance inspection.
type InstanceTest struct {

	// Driver is the name of the driver/executor for which to inspect the
	// instance.
	Driver string

	// Expected is the expected instance.
	Expected *types.Instance
}

// Test is the APITestFunc for the InstanceTest.
func (tt *InstanceTest) Test(
	config gofig.Config,
	client client.Client,
	t *testing.T) {

	expectedBuf, err := json.Marshal(tt.Expected)
	assert.NoError(t, err)
	expectedJSON := string(expectedBuf)

	ctx := context.Background().WithServiceName(tt.Driver)
	iid, err := client.Storage().InstanceInspect(ctx, utils.NewStore())
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
	val, err := client.API().NextDevice(nil, tt.Driver)
	assert.NoError(t, err)
	assert.Equal(t, tt.Expected, val)
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

	val, err := client.API().LocalDevices(nil, tt.Driver)
	assert.NoError(t, err)

	buf, err := json.Marshal(val)
	assert.NoError(t, err)
	actualJSON := string(buf)

	assert.Equal(t, expectedJSON, actualJSON)
}
