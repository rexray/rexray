package efs

import (
	"testing"

	apitests "github.com/AVENTER-UG/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/efs"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/efs/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, efs.Name)
}
