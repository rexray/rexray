package s3fs

import (
	"testing"

	apitests "github.com/codedellemc/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/libstorage/drivers/storage/s3fs"
	_ "github.com/codedellemc/libstorage/drivers/storage/s3fs/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, s3fs.Name)
}
