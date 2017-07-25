package efs

import (
	"testing"

	apitests "github.com/codedellemc/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/rexray/libstorage/drivers/storage/efs"
	_ "github.com/codedellemc/rexray/libstorage/drivers/storage/efs/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, efs.Name)
}
