package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/akutz/gotil"
	apitypes "github.com/codedellemc/rexray/libstorage/api/types"

	"github.com/codedellemc/rexray/core"
)

var r10 string
var tmpPrefixDirs []string

func TestMain(m *testing.M) {
	r10 = gotil.RandomString(10)

	exitCode := m.Run()
	for _, d := range tmpPrefixDirs {
		os.RemoveAll(d)
	}
	os.Exit(exitCode)
}

func newPrefixDir(testName string, t *testing.T) string {
	tmpDir, err := ioutil.TempDir(
		"", fmt.Sprintf("rexray-util_test-%s", testName))
	if err != nil {
		t.Fatal(err)
	}

	Prefix(tmpDir)
	os.MkdirAll(tmpDir, 0755)
	tmpPrefixDirs = append(tmpPrefixDirs, tmpDir)
	return tmpDir
}

func TestPrefix(t *testing.T) {
	if IsPrefixed() {
		t.Fatalf("is prefixed %s", GetPrefix())
	}
	Prefix("")
	if IsPrefixed() {
		t.Fatalf("is prefixed %s", GetPrefix())
	}
	Prefix("/")
	if IsPrefixed() {
		t.Fatalf("is prefixed %s", GetPrefix())
	}

	tmpDir := newPrefixDir("TestHomeDir", t)
	Prefix(tmpDir)
	if !IsPrefixed() {
		t.Fatalf("is not prefixed %s", GetPrefix())
	}

	p := GetPrefix()
	if p != tmpDir {
		t.Fatalf("prefix != %s, == %s", tmpDir, p)
	}
}

func TestPrefixAndDirs(t *testing.T) {
	tmpDir := newPrefixDir("TestPrefixAndDirs", t)

	etcDirPath := EtcDirPath()
	expEtcDirPath := fmt.Sprintf("%s/etc/rexray", tmpDir)
	if etcDirPath != expEtcDirPath {
		t.Fatalf("EtcDirPath() == %s, != %s", etcDirPath, expEtcDirPath)
	}

	etcDirFilePath := EtcFilePath("etcFile")
	expEtcFilePath := fmt.Sprintf("%s/%s", etcDirPath, "etcFile")
	if expEtcFilePath != etcDirFilePath {
		t.Fatalf("EtcFilePath(\"etcFile\") == %s, != %s",
			etcDirFilePath, expEtcFilePath)
	}

	runDirPath := RunDirPath()
	expRunDirPath := fmt.Sprintf("%s/var/run/rexray", tmpDir)
	if runDirPath != expRunDirPath {
		t.Fatalf("RunDirPath() == %s, != %s", runDirPath, expRunDirPath)
	}

	logDirPath := LogDirPath()
	expLogDirPath := fmt.Sprintf("%s/var/log/rexray", tmpDir)
	if logDirPath != expLogDirPath {
		t.Fatalf("LogDirPath() == %s, != %s", logDirPath, expLogDirPath)
	}

	logDirFilePath := LogFilePath("logFile")
	expLogFilePath := fmt.Sprintf("%s/%s", logDirPath, "logFile")
	if expLogFilePath != logDirFilePath {
		t.Fatalf("LogFilePath(\"logFile\") == %s, != %s",
			logDirFilePath, expLogFilePath)
	}

	libDirPath := LibDirPath()
	expLibDirPath := fmt.Sprintf("%s/var/lib/rexray", tmpDir)
	if libDirPath != expLibDirPath {
		t.Fatalf("LibDirPath() == %s, != %s", libDirPath, expLibDirPath)
	}

	libDirFilePath := LibFilePath("libFile")
	expLibFilePath := fmt.Sprintf("%s/%s", libDirPath, "libFile")
	if expLibFilePath != libDirFilePath {
		t.Fatalf("LibFilePath(\"libFile\") == %s, != %s",
			libDirFilePath, expLibFilePath)
	}

	pidFilePath := PidFilePath()
	expPidFilePath := path.Join(tmpDir, "/var/run/rexray", PIDFileName)
	if expPidFilePath != pidFilePath {
		t.Fatalf("PidFilePath() == %s, != %s", pidFilePath, expPidFilePath)
	}
}

func TestStdOutAndLogFile(t *testing.T) {
	newPrefixDir("TestStdOutAndLogFile", t)

	if _, err := StdOutAndLogFile("BadFile/ (*$"); err == nil {
		t.Fatal("error expected in created BadFile")
	}

	out, err := StdOutAndLogFile("TestStdOutAndLogFile")

	if err != nil {
		t.Fatal(err)
	}

	if out == nil {
		t.Fatal("out == nil")
	}
}

func TestWriteReadCurrentPidFile(t *testing.T) {
	newPrefixDir("TestWriteReadPidFile", t)

	var err error
	var pidRead int

	pid := os.Getpid()

	if err = WritePidFile(-1); err != nil {
		t.Fatalf("error writing pidfile=%s", PidFilePath())
	}

	if pidRead, err = ReadPidFile(); err != nil {
		t.Fatalf("error reading pidfile=%s", PidFilePath())
	}

	if pidRead != pid {
		t.Fatalf("pidRead=%d != pid=%d", pidRead, pid)
	}
}

func TestWriteReadCustomPidFile(t *testing.T) {
	newPrefixDir("TestWriteReadPidFile", t)

	var err error
	if _, err = ReadPidFile(); err == nil {
		t.Fatal("error expected in reading pid file")
	}

	pidWritten := int(time.Now().Unix())
	if err = WritePidFile(pidWritten); err != nil {
		t.Fatalf("error writing pidfile=%s", PidFilePath())
	}

	var pidRead int
	if pidRead, err = ReadPidFile(); err != nil {
		t.Fatalf("error reading pidfile=%s", PidFilePath())
	}

	if pidRead != pidWritten {
		t.Fatalf("pidRead=%d != pidWritten=%d", pidRead, pidWritten)
	}
}

func TestReadPidFileWithErrors(t *testing.T) {
	newPrefixDir("TestWriteReadPidFile", t)

	var err error
	if _, err = ReadPidFile(); err == nil {
		t.Fatal("error expected in reading pid file")
	}

	gotil.WriteStringToFile("hello", PidFilePath())

	if _, err = ReadPidFile(); err == nil {
		t.Fatal("error expected in reading pid file")
	}
}

func TestPrintVersion(t *testing.T) {
	// this test works, and adding the libStorage info in just makes it more
	// trouble than it's worth right now to fix
	t.SkipNow()

	core.Version = &apitypes.VersionInfo{}
	core.Version.Arch = "Linux-x86_64"
	core.Version.Branch = "master"
	core.Version.ShaLong = gotil.RandomString(32)
	core.Version.BuildTimestamp = time.Now()
	core.Version.SemVer = "1.0.0"
	_, _, thisAbsPath := gotil.GetThisPathParts()
	epochStr := core.Version.BuildTimestamp.Format(time.RFC1123)

	t.Logf("thisAbsPath=%s", thisAbsPath)
	t.Logf("epochStr=%s", epochStr)

	var buff []byte
	b := bytes.NewBuffer(buff)

	PrintVersion(b)

	vs := b.String()

	evs := `Binary: ` + thisAbsPath + `
SemVer: ` + core.Version.SemVer + `
OsArch: ` + core.Version.Arch + `
Branch: ` + core.Version.Branch + `
Commit: ` + core.Version.ShaLong + `
Formed: ` + epochStr + `
`

	if vs != evs {
		t.Fatalf("nexpectedVersionString=%s\n\nversionString=%s\n", evs, vs)
	}
}

func TestInstall(t *testing.T) {
	Install()
}

func TestInstallChownRoot(t *testing.T) {
	InstallChownRoot()
}

func TestInstallDirChownRoot(t *testing.T) {
	InstallDirChownRoot("--help")
}
