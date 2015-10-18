package version

import (
	"strconv"
	"time"
)

var (
	// SemVer is the semantic version string
	SemVer string

	// ShaLong is the commit hash from which this package was built
	ShaLong string

	// Epoch is the epoch value at the time this package was built
	Epoch string

	// Branch is the branch name from which this package was built
	Branch string

	// Arch is the OS-Arch string of the system on which this package is
	// supported.
	Arch string

	epoch     int64
	buildDate time.Time
)

// EpochToRfc1123 returns the Epoch value as an RFC-1123 formatted date/time
// string.
func EpochToRfc1123() string {
	if !buildDate.IsZero() {
		return buildDate.Format(time.RFC1123)
	}
	if Epoch == "" {
		return ""
	}
	var err error
	epoch, err = strconv.ParseInt(Epoch, 10, 64)
	if err != nil {
		panic(err)
	}
	buildDate = time.Unix(epoch, 0)
	return buildDate.Format(time.RFC1123)
}
