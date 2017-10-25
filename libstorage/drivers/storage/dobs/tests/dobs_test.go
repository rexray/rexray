package dobs

import (
	"testing"

	apitests "github.com/thecodeteam/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/thecodeteam/rexray/libstorage/drivers/storage/dobs"
	_ "github.com/thecodeteam/rexray/libstorage/drivers/storage/dobs/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, dobs.Name)
}
