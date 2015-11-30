package util

import (
	"os"
	"testing"
)

var r10 string

func TestMain(m *testing.M) {
	r10 = RandomString(10)
	os.Exit(m.Run())
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
