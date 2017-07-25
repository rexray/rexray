package ebs

import (
	"testing"

	apitests "github.com/codedellemc/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/rexray/libstorage/drivers/storage/ebs"
	_ "github.com/codedellemc/rexray/libstorage/drivers/storage/ebs/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, ebs.Name)
}
