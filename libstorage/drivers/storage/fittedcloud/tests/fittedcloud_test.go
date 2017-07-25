package fittedcloud

import (
	"testing"

	apitests "github.com/codedellemc/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/rexray/libstorage/drivers/storage/fittedcloud"
	_ "github.com/codedellemc/rexray/libstorage/drivers/storage/fittedcloud/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, fittedcloud.Name)
}
