package vbox

import (
	"testing"

	apitests "github.com/rexray/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/rexray/rexray/libstorage/drivers/storage/vbox"
	_ "github.com/rexray/rexray/libstorage/drivers/storage/vbox/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, vbox.Name)
}
