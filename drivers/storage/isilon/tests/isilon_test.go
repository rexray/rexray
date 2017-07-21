package isilon

import (
	"testing"

	apitests "github.com/codedellemc/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/libstorage/drivers/storage/isilon"
	_ "github.com/codedellemc/libstorage/drivers/storage/isilon/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, isilon.Name)
}
