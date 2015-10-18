package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/emccode/rexray/core/version"
)

var r10 string
var tmpHomeDirs []string

func TestMain(m *testing.M) {
	r10 = RandomString(10)

	exitCode := m.Run()
	for _, d := range tmpHomeDirs {
		os.RemoveAll(d)
	}
	os.Exit(exitCode)
}

func resetPaths() {
	prefix = ""
	homeDir = ""
	binDirPath = ""
	binFilePath = ""
	logDirPath = ""
	libDirPath = ""
	runDirPath = ""
	etcDirPath = ""
	pidFilePath = ""
}

func newHomeDir(testName string, t *testing.T) string {
	resetPaths()

	tmpDir, err := ioutil.TempDir(
		"", fmt.Sprintf("rexray-util_test-%s", testName))
	if err != nil {
		t.Fatal(err)
	}

	Prefix(tmpDir)
	os.MkdirAll(tmpDir, 0755)
	tmpHomeDirs = append(tmpHomeDirs, tmpDir)
	return tmpDir
}

func TestHomeDir(t *testing.T) {
	if HomeDir() == "" {
		t.Fatal("HomeDir() == \"\"")
	}
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

	tmpDir := newHomeDir("TestHomeDir", t)
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
	tmpDir := newHomeDir("TestPrefixAndDirs", t)

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

	binDirPath := BinDirPath()
	expBinDirPath := fmt.Sprintf("%s/usr/bin", tmpDir)
	if binDirPath != expBinDirPath {
		t.Fatalf("BinDirPath() == %s, != %s", binDirPath, expBinDirPath)
	}

	binDirFilePath := BinFilePath()
	expBinFilePath := fmt.Sprintf("%s/%s", binDirPath, "rexray")
	if expBinFilePath != binDirFilePath {
		t.Fatalf("BinFilePath(\"rexray\") == %s, != %s",
			binDirFilePath, expBinFilePath)
	}

	pidFilePath := PidFilePath()
	expPidFilePath := fmt.Sprintf("%s/var/run/rexray/rexray.pid", tmpDir)
	if expPidFilePath != pidFilePath {
		t.Fatalf("PidFilePath() == %s, != %s", pidFilePath, expPidFilePath)
	}
}

func TestStdOutAndLogFile(t *testing.T) {
	newHomeDir("TestStdOutAndLogFile", t)

	if _, err := StdOutAndLogFile("BadFile/"); err == nil {
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

func TestWriteAndReadStringToFile(t *testing.T) {
	tmpDir := newHomeDir("TestWriteAndReadStringToFile", t)

	tmpFile, _ := ioutil.TempFile(tmpDir, "temp")
	WriteStringToFile("Hello, world.", tmpFile.Name())
	if s, _ := ReadFileToString(tmpFile.Name()); s != "Hello, world." {
		t.Fatalf("s != 'Hello, world.', == %s", s)
	}
}

func TestWriteStringToFileError(t *testing.T) {
	if err := WriteStringToFile("error", "/badtmpdir/badfile"); err == nil {
		t.Fatal("error expected in writing temp file")
	}
}

func TestReadtringToFileError(t *testing.T) {
	if _, err := ReadFileToString("/badtmpdir/badfile"); err == nil {
		t.Fatal("error expected in reading temp file")
	}
}

func TestWriteReadCurrentPidFile(t *testing.T) {
	newHomeDir("TestWriteReadPidFile", t)

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
	newHomeDir("TestWriteReadPidFile", t)

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
	newHomeDir("TestWriteReadPidFile", t)

	var err error
	if _, err = ReadPidFile(); err == nil {
		t.Fatal("error expected in reading pid file")
	}

	WriteStringToFile("hello", PidFilePath())

	if _, err = ReadPidFile(); err == nil {
		t.Fatal("error expected in reading pid file")
	}
}

func TestIsDirEmpty(t *testing.T) {
	if _, err := IsDirEmpty(r10); err == nil {
		t.Fatal("expected error for invalid path")
	}

	tmpDir := newHomeDir("TestWriteReadPidFile", t)

	var err error
	var isEmpty bool

	if isEmpty, err = IsDirEmpty(tmpDir); err != nil {
		t.Fatal(err)
	}
	if !isEmpty {
		t.Fatalf("%s expected to be empty", tmpDir)
	}

	WritePidFile(100)

	if isEmpty, err = IsDirEmpty(tmpDir); err != nil {
		t.Fatal(err)
	}
	if isEmpty {
		t.Fatalf("%s expected to not be empty", tmpDir)
	}
}

func TestLineReader(t *testing.T) {
	if c := LineReader(r10); c != nil {
		t.Fatal("expected nil channel for invalid path")
	}

	newHomeDir("TestLineReader", t)

	WritePidFile(100)
	c := LineReader(PidFilePath())

	var lines []string
	for s := range c {
		lines = append(lines, s)
	}

	ll := len(lines)
	if ll != 1 {
		t.Fatalf("len(lines) != 1, == %d", ll)
	}

	if lines[0] != "100" {
		t.Fatalf("lines[0] != 100, == %s", lines[0])
	}
}

func TestGetLocalIP(t *testing.T) {
	var ip string
	if ip = GetLocalIP(); ip == "" {
		t.Fatal("ip == ''")
	}
	t.Logf("ip=%s", ip)
}

func TestParseAddress(t *testing.T) {
	var err error
	var addr, proto, path string

	addr = "ipv4://127.0.0.1:80/hello"
	if proto, path, err = ParseAddress(addr); err == nil {
		t.Fatalf("expected error parsing %s", addr)
	}

	addr = "TCP://127.0.0.1:80/hello"
	if proto, path, err = ParseAddress(addr); err != nil {
		t.Fatalf("error parsing %s %v", addr, err)
	}
	if proto != "TCP" {
		t.Fatalf("proto != TCP, == %s", proto)
	}
	if path != "127.0.0.1:80/hello" {
		t.Fatalf("path != 127.0.0.1:80/hello == %s", path)
	}

	addr = "tcp://127.0.0.1:80/hello"
	if proto, path, err = ParseAddress(addr); err != nil {
		t.Fatalf("error parsing %s %v", addr, err)
	}
	if proto != "tcp" {
		t.Fatalf("proto != tcp, == %s", proto)
	}
	if path != "127.0.0.1:80/hello" {
		t.Fatalf("path != 127.0.0.1:80/hello == %s", path)
	}

	addr = "ip://127.0.0.1:443/secure"
	if proto, path, err = ParseAddress(addr); err != nil {
		t.Fatalf("error parsing %s %v", addr, err)
	}
	if proto != "ip" {
		t.Fatalf("proto != ip, == %s", proto)
	}
	if path != "127.0.0.1:443/secure" {
		t.Fatalf("path != 127.0.0.1:443/secure == %s", path)
	}

	addr = "udp://127.0.0.1:443/secure"
	if proto, path, err = ParseAddress(addr); err != nil {
		t.Fatalf("error parsing %s %v", addr, err)
	}
	if proto != "udp" {
		t.Fatalf("proto != udp, == %s", proto)
	}
	if path != "127.0.0.1:443/secure" {
		t.Fatalf("path != 127.0.0.1:443/secure == %s", path)
	}

	addr = "unix:///var/run/rexray/rexray.sock"
	if proto, path, err = ParseAddress(addr); err != nil {
		t.Fatalf("error parsing %s %v", addr, err)
	}
	if proto != "unix" {
		t.Fatalf("proto != unix, == %s", proto)
	}
	if path != "/var/run/rexray/rexray.sock" {
		t.Fatalf("path != /var/run/rexray/rexray.sock == %s", path)
	}

	addr = "unixgram:///var/run/rexray/rexray.sock"
	if proto, path, err = ParseAddress(addr); err != nil {
		t.Fatalf("error parsing %s %v", addr, err)
	}
	if proto != "unixgram" {
		t.Fatalf("proto != unixgram, == %s", proto)
	}
	if path != "/var/run/rexray/rexray.sock" {
		t.Fatalf("path != /var/run/rexray/rexray.sock == %s", path)
	}

	addr = "unixpacket:///var/run/rexray/rexray.sock"
	if proto, path, err = ParseAddress(addr); err != nil {
		t.Fatalf("error parsing %s %v", addr, err)
	}
	if proto != "unixpacket" {
		t.Fatalf("proto != unixpacket, == %s", proto)
	}
	if path != "/var/run/rexray/rexray.sock" {
		t.Fatalf("path != /var/run/rexray/rexray.sock == %s", path)
	}
}

func TestPrintVersion(t *testing.T) {
	version.Arch = "Linux-x86_64"
	version.Branch = "master"
	version.ShaLong = RandomString(32)
	version.Epoch = fmt.Sprintf("%d", time.Now().Unix())
	version.SemVer = "1.0.0"
	_, _, thisAbsPath := GetThisPathParts()
	epochStr := version.EpochToRfc1123()

	t.Logf("thisAbsPath=%s", thisAbsPath)
	t.Logf("epochStr=%s", epochStr)

	var buff []byte
	b := bytes.NewBuffer(buff)

	PrintVersion(b)

	vs := b.String()

	evs := `Binary: ` + thisAbsPath + `
SemVer: ` + version.SemVer + `
OsArch: ` + version.Arch + `
Branch: ` + version.Branch + `
Commit: ` + version.ShaLong + `
Formed: ` + epochStr + `
`

	if vs != evs {
		t.Fatalf("nexpectedVersionString=%s\n\nversionString=%s\n", evs, vs)
	}
}

func TestTrimSingleWord(t *testing.T) {

	s := Trim(`

						hi




    `)

	if s != "hi" {
		t.Fatalf("trim failed '%v'", s)
	}
}

func TestTrimMultipleWords(t *testing.T) {

	s := Trim(`

						hi

		there

		     you
    `)

	if s != `hi

		there

		     you` {
		t.Fatalf("trim failed '%v'", s)
	}
}

func TestFileExists(t *testing.T) {
	if FileExists(r10) {
		t.Fatal("file should not exist")
	}

	if !FileExists("/bin/sh") {
		t.Fail()
	}
}

func TestFileExistsInPath(t *testing.T) {
	if FileExistsInPath(r10) {
		t.Fatal("file should not exist")
	}

	if !FileExistsInPath("sh") {
		t.Fail()
	}
}

func TestGetPathParts(t *testing.T) {
	d, n, a := GetPathParts("/bin/sh")
	if d != "/bin" {
		t.Fatalf("dir != /bin, == %s", d)
	}
	if n != "sh" {
		t.Fatalf("n != sh, == %s", n)
	}
	if a != "/bin/sh" {
		t.Fatalf("name != /bin/sh, == %s", a)
	}
}

func TestGetThisPathParts(t *testing.T) {
	_, n, _ := GetThisPathParts()
	if !strings.Contains(n, ".test") {
		t.Fatalf("n !=~ .test, == %s", n)
	}
}

func TestRandomString(t *testing.T) {
	var lastR string
	for x := 0; x < 100; x++ {
		r := RandomString(10)
		if r == lastR {
			t.Fail()
		}
		lastR = r
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

func TestStringInSlice(t *testing.T) {
	var r bool

	r = StringInSlice("hi", []string{"hello", "world"})
	if r {
		t.Fatal("hi there!")
	}

	r = StringInSlice("hi", []string{"hi", "world"})
	if !r {
		t.Fatal("hi where?")
	}

	a := []string{"a", "b", "c"}
	if !StringInSlice("b", a) {
		t.Fatal("b not in 'a, b, c'")
	}
	if !StringInSlice("A", a) {
		t.Fatal("A not in 'a, b, c'")
	}
	if StringInSlice("d", a) {
		t.Fatal("d in 'a, b, c'")
	}
}
