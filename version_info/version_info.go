package version_info

import (
	"strconv"
	"time"
)

var (
	SemVer  string
	ShaLong string
	Epoch   string
	Branch  string
	Arch    string

	epoch     int64
	buildDate time.Time
)

func init() {
	if Epoch == "" {
		return
	}
	var err error
	epoch, err = strconv.ParseInt(Epoch, 10, 64)
	if err != nil {
		panic(err)
	}
	buildDate = time.Unix(epoch, 0)
}

func EpochToRfc1123() string {
	if Epoch == "" {
		return ""
	}
	return buildDate.Format(time.RFC1123)
}
