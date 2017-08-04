package core

import "time"

func init() {
	SemVer = "{{.SemVer}}"
	CommitSha7 = "{{.Sha7}}"
	CommitSha32 = "{{.Sha32}}"
	CommitTime = time.Unix({{.Epoch}}, 0)
}
