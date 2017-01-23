// +build !libstorage_storage_driver libstorage_storage_driver_azure

package utils

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/codedellemc/libstorage/api/context"
)

func skipTest(t *testing.T) {
	if ok, _ := strconv.ParseBool(os.Getenv("AZURE_UTILS_TEST")); !ok {
		t.Skip()
	}
}

func TestInstanceID(t *testing.T) {
	skipTest(t)
	iid, err := InstanceID(context.Background())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	t.Logf("instanceID=%s", iid.String())
}
