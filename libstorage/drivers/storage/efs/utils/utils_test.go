package utils

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
)

func skipTest(t *testing.T) {
	if ok, _ := strconv.ParseBool(os.Getenv("EFS_UTILS_TEST")); !ok {
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
