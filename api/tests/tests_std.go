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

	apiclient "github.com/emccode/libstorage/api/client"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/client"
)

// TestRoot tests the GET / route.
var TestRoot = func(config gofig.Config, client client.Client, t *testing.T) {
	reply, err := client.(apiclient.APIClient).Root()
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
func (idt *InstanceIDTest) Test(
	config gofig.Config,
	client client.Client,
	t *testing.T) {

	expectedBuf, err := json.Marshal(idt.Expected)
	assert.NoError(t, err)
	expectedJSON := string(expectedBuf)

	reply, err := client.(apiclient.APIClient).ExecutorGet(lsxbin)
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
		f.Name(), idt.Driver, "instanceID").CombinedOutput()
	assert.NoError(t, err)
	assert.Equal(t, expectedJSON, gotil.Trim(string(out)))
}
