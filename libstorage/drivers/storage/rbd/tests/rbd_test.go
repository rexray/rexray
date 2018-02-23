package rbd

import (
	"testing"

	apitests "github.com/rexray/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/rexray/rexray/libstorage/drivers/storage/rbd"
	_ "github.com/rexray/rexray/libstorage/drivers/storage/rbd/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, rbd.Name)
}
