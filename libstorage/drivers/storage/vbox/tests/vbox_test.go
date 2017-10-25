package vbox

import (
	"testing"

	apitests "github.com/thecodeteam/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/thecodeteam/rexray/libstorage/drivers/storage/vbox"
	_ "github.com/thecodeteam/rexray/libstorage/drivers/storage/vbox/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, vbox.Name)
}
