package isilon

import (
	"testing"

	apitests "github.com/thecodeteam/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/thecodeteam/rexray/libstorage/drivers/storage/isilon"
	_ "github.com/thecodeteam/rexray/libstorage/drivers/storage/isilon/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, isilon.Name)
}
