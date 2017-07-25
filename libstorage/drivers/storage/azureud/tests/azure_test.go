package azureud

import (
	"testing"

	apitests "github.com/codedellemc/rexray/libstorage/api/tests"

	// load the driver packages
	"github.com/codedellemc/rexray/libstorage/drivers/storage/azureud"
	_ "github.com/codedellemc/rexray/libstorage/drivers/storage/azureud/storage"
)

func TestSuite(t *testing.T) {
	apitests.RunSuite(t, azureud.Name)
}
