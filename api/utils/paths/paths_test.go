package paths

import (
	"io/ioutil"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	os.Exit(m.Run())
}

func TestPaths(t *testing.T) {
	if Home.get() == "" {
		td, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(td)
		libstorageHome = td
	}

	t.Logf("%5[1]k  %[1]s", Home)
	t.Logf("%5[1]k  %[1]s", Etc)
	t.Logf("%5[1]k  %[1]s", Lib)
	t.Logf("%5[1]k  %[1]s", Log)
	t.Logf("%5[1]k  %[1]s", Run)
	t.Logf("%5[1]k  %[1]s", LSX)
}
