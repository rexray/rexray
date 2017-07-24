package vbox

import (
	"testing"

	apitests "github.com/codedellemc/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/libstorage/drivers/storage/vbox"
	_ "github.com/codedellemc/libstorage/drivers/storage/vbox/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, vbox.Name)
}
