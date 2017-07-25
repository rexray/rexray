package gcepd

import (
	"testing"

	apitests "github.com/codedellemc/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/rexray/libstorage/drivers/storage/gcepd"
	_ "github.com/codedellemc/rexray/libstorage/drivers/storage/gcepd/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, gcepd.Name)
}
