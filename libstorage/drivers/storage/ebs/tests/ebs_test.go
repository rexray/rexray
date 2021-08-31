package ebs

import (
	"testing"

	apitests "github.com/AVENTER-UG/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/ebs"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/ebs/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, ebs.Name)
}
