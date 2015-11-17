package util

import (
	"regexp"
	"strings"

	"github.com/akutz/goof"
	"golang.org/x/net/context"

	"github.com/emccode/libstorage/model"
)

const (
	networkAdressPattern = `(?i)^((?:(?:tcp|udp|ip)[46]?)|(?:unix(?:gram|packet)?))://(.+)$`
)

var (
	netAddrRx = regexp.MustCompile(networkAdressPattern)
)

// StringInSlice returns a flag indicating whether or not a provided string
// exists in a string slice.
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToLower(b) == strings.ToLower(a) {
			return true
		}
	}
	return false
}

// WithInstanceID returns a new context with the provided InstanceID object
// accessible with the key "instanceID".
func WithInstanceID(instanceID *model.InstanceID) context.Context {
	return context.WithValue(context.Background(), "instanceID", instanceID)
}

// ParseAddress parses a standard golang network address and returns the
// protocol and path.
func ParseAddress(addr string) (proto string, path string, err error) {
	m := netAddrRx.FindStringSubmatch(addr)
	if m == nil {
		return "", "", goof.WithField("address", addr, "invalid address")
	}
	return m[1], m[2], nil
}
