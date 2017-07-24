package dobs

import (
	"testing"

	apitests "github.com/codedellemc/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/libstorage/drivers/storage/dobs"
	_ "github.com/codedellemc/libstorage/drivers/storage/dobs/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, dobs.Name)
}
