package cinder

import (
	"testing"

	apitests "github.com/codedellemc/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/rexray/libstorage/drivers/storage/cinder"
	_ "github.com/codedellemc/rexray/libstorage/drivers/storage/cinder/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, cinder.Name)
}
