package azureud

import (
	"testing"

	apitests "github.com/rexray/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/rexray/rexray/libstorage/drivers/storage/azureud"
	_ "github.com/rexray/rexray/libstorage/drivers/storage/azureud/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, azureud.Name)
}
