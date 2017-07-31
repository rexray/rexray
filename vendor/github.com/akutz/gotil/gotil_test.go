package gotil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	r10           string
	tmpPrefixDirs []string
)

func newTestDir(testName string, t *testing.T) string {
	tmpDir, err := ioutil.TempDir(
		"", fmt.Sprintf("gotil_test-%s", testName))
	if err != nil {
		t.Fatal(err)
	}

	os.MkdirAll(tmpDir, 0755)
	tmpPrefixDirs = append(tmpPrefixDirs, tmpDir)
	return tmpDir
}

func TestMain(m *testing.M) {
	r10 = RandomString(10)
	exitCode := m.Run()
	for _, d := range tmpPrefixDirs {
		os.RemoveAll(d)
	}
	os.Exit(exitCode)
}

func TestWriteAndReadStringToFile(t *testing.T) {
	tmpDir := newTestDir("TestWriteAndReadStringToFile", t)

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

func TestIsDirEmpty(t *testing.T) {
	if _, err := IsDirEmpty(r10); err == nil {
		t.Fatal("expected error for invalid path")
	}

	tmpDir := newTestDir("TestIsDirEmpty", t)

	var err error
	var isEmpty bool

	if isEmpty, err = IsDirEmpty(tmpDir); err != nil {
		t.Fatal(err)
	}
	if !isEmpty {
		t.Fatalf("%s expected to be empty", tmpDir)
	}

	WriteStringToFile(r10, fmt.Sprintf("%s/temp.log", tmpDir))

	if isEmpty, err = IsDirEmpty(tmpDir); err != nil {
		t.Fatal(err)
	}
	if isEmpty {
		t.Fatalf("%s expected to not be empty", tmpDir)
	}
}

func TestLineReader(t *testing.T) {
	r := bytes.NewReader([]byte(`

30
`))

	c, err := LineReader(r)
	if err != nil {
		t.Fatal(err)
	}

	var lines []string
	for s := range c {
		lines = append(lines, s)
	}

	assert.Equal(t, 3, len(lines))
	assert.Equal(t, "30", lines[2])
}

func TestLineReaderFrom(t *testing.T) {
	c, err := LineReaderFrom(r10)
	if err != nil {
		t.Fatal("expected nil channel for invalid path")
	}

	tmpDir := newTestDir("TestLineReaderFrom", t)

	path := fmt.Sprintf("%s/temp.log", tmpDir)
	if err = WriteStringToFile(r10, path); err != nil {
		t.Fatal(err)
	}

	if c, err = LineReaderFrom(path); err != nil {
		t.Fatal(err)
	}

	var lines []string
	for s := range c {
		lines = append(lines, s)
	}

	assert.Equal(t, 1, len(lines))
	assert.Equal(t, r10, lines[0])

	os.Chmod(path, 0000)
	c, err = LineReaderFrom(r10)
	if err != nil {
		t.Fatal("expected nil channel for access denied")
	}
}

func TestGetLocalIP(t *testing.T) {
	var ip string
	if ip = GetLocalIP(); ip == "" {
		t.Fatal("ip == ''")
	}
	t.Logf("ip=%s", ip)
}

func TestParseIpAddress(t *testing.T) {
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
}

func TestParseUdpAddress(t *testing.T) {
	var err error
	var addr, proto, path string

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
}

func TestParseUnixAddress(t *testing.T) {
	var err error
	var addr, proto, path string

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

func TestStringInSliceCS(t *testing.T) {
	var r bool

	r = StringInSliceCS("hi", []string{"hello", "world"})
	if r {
		t.Fatal("hi there!")
	}

	r = StringInSliceCS("hi", []string{"hi", "world"})
	if !r {
		t.Fatal("hi where?")
	}

	r = StringInSliceCS("hi", []string{"Hi", "world"})
	if r {
		t.Fatal("Hi where?")
	}

	a := []string{"a", "B", "c"}
	if !StringInSliceCS("B", a) {
		t.Fatal("b not in 'a, B, c'")
	}
	if StringInSliceCS("A", a) {
		t.Fatal("A not in 'a, B, c'")
	}
	if StringInSliceCS("d", a) {
		t.Fatal("d in 'a, b, c'")
	}
}

func TestWriteIndented(t *testing.T) {
	s := []byte("Hello,\nworld.")
	s2 := "  Hello,\n  world."
	s4 := "    Hello,\n    world."

	w2 := &bytes.Buffer{}
	if err := WriteIndentedN(w2, s, 2); err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, s2, w2.String())

	w1 := &bytes.Buffer{}
	if err := WriteIndented(w1, s); err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, s4, w1.String())
}

func TestHomeDir(t *testing.T) {
	assert.NotEqual(t, "", HomeDir())
}

func TestRandomTCPPort(t *testing.T) {
	p := RandomTCPPort()
	addr := fmt.Sprintf("127.0.0.1:%d", p)
	t.Logf("listening on addr %s", addr)
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIsTCPPortAvailable(t *testing.T) {
	p := RandomTCPPort()
	addr := fmt.Sprintf("127.0.0.1:%d", p)
	t.Logf("listening on addr %s", addr)
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, err2 := net.Listen("tcp", addr)
	if err2 == nil {
		t.Fatalf("addr should be in use %s", addr)
	}
}
