package scaleio

import (
	"testing"

	apitests "github.com/rexray/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/rexray/rexray/libstorage/drivers/storage/scaleio"
	_ "github.com/rexray/rexray/libstorage/drivers/storage/scaleio/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, scaleio.Name)
}
