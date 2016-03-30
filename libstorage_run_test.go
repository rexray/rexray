package libstorage

import (
	"os"
	"strconv"
	"testing"
)

func TestRun(t *testing.T) {
	continueIfTestRun(t)

	host := os.Getenv("LIBSTORAGE_TESTRUN_HOST")
	tls, _ := strconv.ParseBool(os.Getenv("LIBSTORAGE_TESTRUN_TLS"))

	_, _, errs := getServer(host, tls, t)
	err := <-errs
	if err != nil {
		t.Error(err)
	}
}

func skipIfTestRun(t *testing.T) {
	if isTestRun() {
		t.SkipNow()
	}
}

func continueIfTestRun(t *testing.T) {
	if !isTestRun() {
		t.SkipNow()
	}
}

func isTestRun() bool {
	b, _ := strconv.ParseBool(os.Getenv("LIBSTORAGE_TESTRUN"))
	return b
}
