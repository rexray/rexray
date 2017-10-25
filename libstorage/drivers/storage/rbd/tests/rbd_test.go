package rbd

import (
	"testing"

	apitests "github.com/thecodeteam/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/thecodeteam/rexray/libstorage/drivers/storage/rbd"
	_ "github.com/thecodeteam/rexray/libstorage/drivers/storage/rbd/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, rbd.Name)
}
