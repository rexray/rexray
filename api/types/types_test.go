package types

import (
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	os.Exit(m.Run())
}
