package util

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/akutz/goof"
)

const (
	letterBytes          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIndexBits      = 6
	letterIndexMask      = 1<<letterIndexBits - 1
	letterIndexMax       = 63 / letterIndexBits
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

// ParseAddress parses a standard golang network address and returns the
// protocol and path.
func ParseAddress(addr string) (proto string, path string, err error) {
	m := netAddrRx.FindStringSubmatch(addr)
	if m == nil {
		return "", "", goof.WithField("address", addr, "invalid address")
	}
	return m[1], m[2], nil
}

// WriteIndented indents all lines four spaces.
func WriteIndented(w io.Writer, b []byte) error {
	s := bufio.NewScanner(bytes.NewReader(b))
	for s.Scan() {
		if _, err := fmt.Fprintf(w, "    %s\n", s.Text()); err != nil {
			return err
		}
	}
	return nil
}

// FileExists returns a flag indicating whether a provided file path exists.
func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return true
	}
	return false
}

// FileExistsInPath returns a flag indicating whether the provided file exists
// in the current path.
func FileExistsInPath(fileName string) bool {
	_, err := exec.LookPath(fileName)
	return err == nil
}

// RandomString generates a random set of characters with the given lenght.
func RandomString(length int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, length)
	for i, cache, remain := length-1, src.Int63(), letterIndexMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIndexMax
		}
		if idx := int(cache & letterIndexMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIndexBits
		remain--
	}

	return string(b)
}
