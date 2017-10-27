package gocsi

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"

	"github.com/thecodeteam/gocsi/csi"
)

// Version is a type that responds with Major, Minor, and Patch
// information typical of a semantic version.
type Version interface {
	GetMajor() uint32
	GetMinor() uint32
	GetPatch() uint32
}

const maxuint32 = 4294967295

// ParseVersion parses any string that matches \d+\.\d+\.\d+ and
// returns a Version.
func ParseVersion(s string) (Version, error) {
	var major uint32
	var minor uint32
	var patch uint32
	n, err := fmt.Sscanf(s, "%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return nil, err
	}
	if n != 3 {
		return nil, fmt.Errorf("error: parsed %d vals", n)
	}
	return &csi.Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// SprintfVersion formats a Version as a string.
func SprintfVersion(v Version) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d", v.GetMajor(), v.GetMinor(), v.GetPatch())
}

// CompareVersions compares two versions and returns:
//
//   -1 if a > b
//    0 if a = b
//    1 if a < b
func CompareVersions(a, b Version) int8 {
	if a == nil && b == nil {
		return 0
	}
	if a != nil && b == nil {
		return -1
	}
	if a == nil && b != nil {
		return 1
	}
	if a.GetMajor() > b.GetMajor() {
		return -1
	}
	if a.GetMajor() < b.GetMajor() {
		return 1
	}
	if a.GetMinor() > b.GetMinor() {
		return -1
	}
	if a.GetMinor() < b.GetMinor() {
		return 1
	}
	if a.GetPatch() > b.GetPatch() {
		return -1
	}
	if a.GetPatch() < b.GetPatch() {
		return 1
	}
	return 0
}

// GetCSIEndpoint returns the network address specified by the
// environment variable CSI_ENDPOINT.
func GetCSIEndpoint() (network, addr string, err error) {
	protoAddr := os.Getenv(CSIEndpoint)
	if emptyRX.MatchString(protoAddr) {
		return "", "", ErrMissingCSIEndpoint
	}
	return ParseProtoAddr(protoAddr)
}

// GetCSIEndpointListener returns the net.Listener for the endpoint
// specified by the environment variable CSI_ENDPOINT.
func GetCSIEndpointListener() (net.Listener, error) {
	proto, addr, err := GetCSIEndpoint()
	if err != nil {
		return nil, err
	}
	return net.Listen(proto, addr)
}

const (
	protoAddrGuessPatt = `(?i)^(?:tcp|udp|ip|unix)[^:]*://`

	protoAddrExactPatt = `(?i)^((?:(?:tcp|udp|ip)[46]?)|` +
		`(?:unix(?:gram|packet)?))://(.+)$`
)

var (
	emptyRX          = regexp.MustCompile(`^\s*$`)
	protoAddrGuessRX = regexp.MustCompile(protoAddrGuessPatt)
	protoAddrExactRX = regexp.MustCompile(protoAddrExactPatt)
)

// ErrParseProtoAddrRequired occurs when an empty string is provided
// to ParseProtoAddr.
var ErrParseProtoAddrRequired = errors.New(
	"non-empty network address is required")

// ParseProtoAddr parses a Golang network address.
func ParseProtoAddr(protoAddr string) (proto string, addr string, err error) {

	if emptyRX.MatchString(protoAddr) {
		return "", "", ErrParseProtoAddrRequired
	}

	// If the provided network address does not begin with one
	// of the valid network protocols then treat the string as a
	// file path.
	//
	// First check to see if the file exists at the specified path.
	// If it does then assume it's a valid file path and return it.
	//
	// Otherwise attempt to create the file. If the file can be created
	// without error then remove the file and return the result a UNIX
	// socket file path.
	if !protoAddrGuessRX.MatchString(protoAddr) {

		// If the file already exists then assume it's a valid sock
		// file and return it.
		if _, err := os.Stat(protoAddr); !os.IsNotExist(err) {
			return "unix", protoAddr, nil
		}

		f, err := os.Create(protoAddr)
		if err != nil {
			return "", "", fmt.Errorf(
				"invalid implied sock file: %s: %v", protoAddr, err)
		}
		if err := f.Close(); err != nil {
			return "", "", fmt.Errorf(
				"failed to verify network address as sock file: %s", protoAddr)
		}
		if err := os.RemoveAll(protoAddr); err != nil {
			return "", "", fmt.Errorf(
				"failed to remove verified sock file: %s", protoAddr)
		}
		return "unix", protoAddr, nil
	}

	// Parse the provided network address into the protocol and address parts.
	m := protoAddrExactRX.FindStringSubmatch(protoAddr)
	if m == nil {
		return "", "", fmt.Errorf("invalid network address: %s", protoAddr)
	}
	return m[1], m[2], nil
}
