package efs

import (
	"testing"

	apitests "github.com/thecodeteam/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/thecodeteam/rexray/libstorage/drivers/storage/efs"
	_ "github.com/thecodeteam/rexray/libstorage/drivers/storage/efs/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, efs.Name)
}
