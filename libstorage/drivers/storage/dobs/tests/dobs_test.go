package dobs

import (
	"testing"

	apitests "github.com/rexray/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/rexray/rexray/libstorage/drivers/storage/dobs"
	_ "github.com/rexray/rexray/libstorage/drivers/storage/dobs/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, dobs.Name)
}
