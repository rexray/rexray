package scaleio

import (
	"testing"

	apitests "github.com/codedellemc/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/libstorage/drivers/storage/scaleio"
	_ "github.com/codedellemc/libstorage/drivers/storage/scaleio/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, scaleio.Name)
}
