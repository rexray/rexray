package types

import (
	"io/ioutil"
	"os"
	"testing"
)

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
	t.Logf("%5[1]s  %[2]s", Home.key(), Home)
	t.Logf("%5[1]s  %[2]s", Etc.key(), Etc)
	t.Logf("%5[1]s  %[2]s", TLS.key(), TLS)
	t.Logf("%5[1]s  %[2]s", DefaultTLSCertFile.key(), DefaultTLSCertFile)
	t.Logf("%5[1]s  %[2]s", DefaultTLSKeyFile.key(), DefaultTLSKeyFile)
	t.Logf("%5[1]s  %[2]s",
		DefaultTLSTrustedRootsFile.key(), DefaultTLSTrustedRootsFile)
	t.Logf("%5[1]s  %[2]s", DefaultTLSKnownHosts.key(), DefaultTLSKnownHosts)
	t.Logf("%5[1]s  %[2]s", Lib.key(), Lib)
	t.Logf("%5[1]s  %[2]s", Log.key(), Log)
	t.Logf("%5[1]s  %[2]s", Run.key(), Run)
	t.Logf("%5[1]s  %[2]s", LSX.key(), LSX)
}
