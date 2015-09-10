package version_info

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

const SemVerPatt = `^[^\d]*(\d+)\.(\d+)\.(\d+)(?:-(.+?))?(?:-(\d+)-g(.+))?$`

var (
	GitDescribe        string
	CommitHash         string
	CommitDateEpochStr string
	BuildDateEpochStr  string
	BranchName         string
	TargetVersion      string

	version  *Version
	semVerRx *regexp.Regexp
)

type Version struct {
	SemVer          string `json:"semVer"`
	FullSemVer      string `json:"fullSemVer"`
	Major           int32  `json:"major"`
	Minor           int32  `json:"minor"`
	Patch           int32  `json:"patch"`
	MajorMinorPatch string `json:"major.minor.patch"`
	PreReleaseTag   string `json:"preReleaseTag"`
	Build           int64  `json:"build"`
	Sha             string `json:"sha"`
	BranchName      string `json:"branchName"`
	CommitDate      int64  `json:"commitDate"`
	BuildDate       int64  `json:"buildDate"`
}

func init() {
	var semVerRxErr error
	semVerRx, semVerRxErr = regexp.Compile(SemVerPatt)
	if semVerRxErr != nil {
		panic(semVerRxErr)
	}
}

func getVersion() *Version {
	if version == nil {
		versionFields := map[string]interface{}{
			"gitDescribe":     GitDescribe,
			"commitHash":      CommitHash,
			"commitDateEpoch": CommitDateEpochStr,
			"buildDateEpcoh":  BuildDateEpochStr,
			"branchName":      BranchName,
			"targetVersion":   TargetVersion,
		}
		log.WithFields(versionFields).Debug("rexray version info")
		version = parseSemVer(GitDescribe, TargetVersion)
	}

	return version
}

func ParseSemVer(semVer string) *Version {
	return parseSemVer(semVer, "")
}

func parseSemVer(semVer, targetVer string) *Version {

	commitDate, commitDateErr := strconv.ParseInt(CommitDateEpochStr, 10, 64)
	if commitDateErr != nil {
		panic(commitDateErr)
	}
	buildDate, buildDateErr := strconv.ParseInt(BuildDateEpochStr, 10, 64)
	if buildDateErr != nil {
		panic(buildDateErr)
	}

	v := &Version{
		Sha:        CommitHash,
		BranchName: BranchName,
		CommitDate: commitDate,
		BuildDate:  buildDate,
	}

	parseSemVer_(v, semVer, targetVer)
	log.WithField("version", v).Debug("parsed version")
	return v
}

func parseSemVer_(v *Version, semVer, targetVer string) {

	svm := semVerRx.FindStringSubmatch(semVer)
	if svm == nil {
		return
	}

	var tvm []string
	if targetVer != "" {
		tvm = semVerRx.FindStringSubmatch(targetVer)
	}

	if tvm == nil {
		parseSemVerParts(v, svm)
	} else {
		parseSemVerParts(v, tvm)
	}

	v.MajorMinorPatch = fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)

	var semVerBuff bytes.Buffer
	semVerBuff.WriteString(v.MajorMinorPatch)

	if v.PreReleaseTag != "" {
		semVerBuff.WriteString("-")
		semVerBuff.WriteString(v.PreReleaseTag)
		v.SemVer = semVerBuff.String()
	}

	if len(svm) >= 6 && svm[5] != "" {
		log.WithField("build", svm[5]).Debug("parsed build")
		build, buildErr := strconv.ParseInt(svm[5], 10, 64)
		if buildErr != nil {
			panic(buildErr)
		}
		v.Build = build
		semVerBuff.WriteString("+")
		semVerBuff.WriteString(svm[5])
	}
	v.FullSemVer = semVerBuff.String()
}

func parseSemVerParts(v *Version, m []string) {
	log.WithField("semVerParts", m).Debug("parsing matched semver parts")

	major, majorErr := strconv.ParseInt(m[1], 10, 32)
	if majorErr != nil {
		panic(majorErr)
	}
	minor, minorErr := strconv.ParseInt(m[2], 10, 32)
	if minorErr != nil {
		panic(minorErr)
	}
	patch, patchErr := strconv.ParseInt(m[3], 10, 32)
	if patchErr != nil {
		panic(patchErr)
	}

	var preReleaseTag string
	if len(m) >= 5 && m[4] != "" {
		preReleaseTag = m[4]
	}

	/*log.WithFields(log.Fields{
		"major":         major,
		"minor":         minor,
		"patch":         patch,
		"preReleaseTag": preReleaseTag,
	}).Debug("parsed major, minor, patch, & pre-release tag")*/

	v.Major = int32(major)
	v.Minor = int32(minor)
	v.Patch = int32(patch)
	v.PreReleaseTag = preReleaseTag
}

func SemVer() string {
	return getVersion().SemVer
}

func FullSemVer() string {
	return getVersion().FullSemVer
}

func Major() int32 {
	return getVersion().Major
}

func Minor() int32 {
	return getVersion().Minor
}

func Patch() int32 {
	return getVersion().Patch
}

func MajorMinorPatch() string {
	return getVersion().MajorMinorPatch
}

func PreReleaseTag() string {
	return getVersion().PreReleaseTag
}

func Build() int64 {
	return getVersion().Build
}

func Sha() string {
	return getVersion().Sha
}

func Branch() string {
	return getVersion().BranchName
}

func CommitDate() int64 {
	return getVersion().CommitDate
}

func BuildDate() int64 {
	return getVersion().BuildDate
}
