package utils

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"

	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/ebs"
)

func skipTest(t *testing.T) {
	if ok, _ := strconv.ParseBool(os.Getenv("EBS_UTILS_TEST")); !ok {
		t.Skip()
	}
}

func TestInstanceID(t *testing.T) {
	skipTest(t)
	iid, err := InstanceID(context.Background(), ebs.Name)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	t.Logf("instanceID=%s", iid.String())
}
