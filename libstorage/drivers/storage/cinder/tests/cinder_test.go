package cinder

import (
	"testing"

	apitests "github.com/rexray/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/rexray/rexray/libstorage/drivers/storage/cinder"
	_ "github.com/rexray/rexray/libstorage/drivers/storage/cinder/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, cinder.Name)
}
