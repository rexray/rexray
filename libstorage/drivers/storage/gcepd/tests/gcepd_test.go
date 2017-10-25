package gcepd

import (
	"testing"

	apitests "github.com/thecodeteam/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/thecodeteam/rexray/libstorage/drivers/storage/gcepd"
	_ "github.com/thecodeteam/rexray/libstorage/drivers/storage/gcepd/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, gcepd.Name)
}
