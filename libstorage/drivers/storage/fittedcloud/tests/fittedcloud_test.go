package fittedcloud

import (
	"testing"

	apitests "github.com/rexray/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/rexray/rexray/libstorage/drivers/storage/fittedcloud"
	_ "github.com/rexray/rexray/libstorage/drivers/storage/fittedcloud/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, fittedcloud.Name)
}
