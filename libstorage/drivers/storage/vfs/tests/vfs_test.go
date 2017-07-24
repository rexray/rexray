package vfs

import (
	"testing"

	apitests "github.com/codedellemc/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/libstorage/drivers/storage/vfs"

	// load the vfs driver packages
	_ "github.com/codedellemc/libstorage/drivers/storage/vfs/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, vfs.Name)
}
