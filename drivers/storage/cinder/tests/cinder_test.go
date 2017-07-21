package cinder

import (
	"testing"

	apitests "github.com/codedellemc/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/libstorage/drivers/storage/cinder"
	_ "github.com/codedellemc/libstorage/drivers/storage/cinder/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, cinder.Name)
}
