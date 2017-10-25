package fittedcloud

import (
	"testing"

	apitests "github.com/thecodeteam/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/thecodeteam/rexray/libstorage/drivers/storage/fittedcloud"
	_ "github.com/thecodeteam/rexray/libstorage/drivers/storage/fittedcloud/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, fittedcloud.Name)
}
