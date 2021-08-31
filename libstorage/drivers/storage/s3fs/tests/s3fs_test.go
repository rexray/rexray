package s3fs

import (
	"testing"

	apitests "github.com/AVENTER-UG/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/s3fs"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/s3fs/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, s3fs.Name)
}
