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

	t.Logf("lsx.lock=%s", Run.Join("lsx.lock"))

	if Home.get() == "" {
		td, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(td)
		libstorageHome = td
	}

	// gometalinter (go vet) complains here because of 'k' being an unrecognized
	// printf verb. it's a custom verb, as is supported by the fmt pkg. The
	// vet tool just doesn't handle validating custom verbs.
	t.Logf("%5[1]k  %[1]s", Home)
	t.Logf("%5[1]k  %[1]s", Etc)
	t.Logf("%5[1]k  %[1]s", Lib)
	t.Logf("%5[1]k  %[1]s", Log)
	t.Logf("%5[1]k  %[1]s", Run)
	t.Logf("%5[1]k  %[1]s", LSX)
}
