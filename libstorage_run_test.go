// +build run

package libstorage

import (
	"os"
	"strconv"
	"testing"
)

func TestRun(t *testing.T) {

	host := os.Getenv("LIBSTORAGE_TESTRUN_HOST")
	tls, _ := strconv.ParseBool(os.Getenv("LIBSTORAGE_TESTRUN_TLS"))

	_, _, errs := getServer(host, tls, t)
	err := <-errs
	if err != nil {
		t.Error(err)
	}
}
