package rbd

import (
	"testing"

	apitests "github.com/codedellemc/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/libstorage/drivers/storage/rbd"
	_ "github.com/codedellemc/libstorage/drivers/storage/rbd/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, rbd.Name)
}
