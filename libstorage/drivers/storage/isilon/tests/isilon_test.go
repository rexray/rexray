package isilon

import (
	"testing"

	apitests "github.com/AVENTER-UG/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/isilon"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/isilon/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, isilon.Name)
}
