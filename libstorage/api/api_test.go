package api

import "testing"

func TestVersion(t *testing.T) {
	t.Logf("%s\n", Version.String())
}
