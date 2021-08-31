package cinder

import (
	"testing"

	apitests "github.com/AVENTER-UG/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/cinder"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/cinder/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, cinder.Name)
}
