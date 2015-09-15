package util

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const LOGFILE = "/var/log/rexray.log"
const PIDFILE = "/var/run/rexray.pid"
const PIDDIR = "/var/run/rexray"
const LOGDIR = "/var/log/rexray"
const CLILOG = "/var/log/rexray.log"

var (
	homeDir string
)

func init() {
	homeDir = HomeDir()
}

func HomeDir() string {
	if homeDir != "" {
		return homeDir
	}

	hd := "$HOME"
	curUser, curUserErr := user.Current()
	if curUserErr == nil {
		hd = curUser.HomeDir
	}

	return hd
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToLower(b) == strings.ToLower(a) {
			return true
		}
	}
	return false
}

func PidFile() string {
	return PIDFILE
}

func LogFile() string {
	return LOGFILE
}

func WritePidFile(pid int) error {

	if pid < 0 {
		pid = os.Getpid()
	}

	f, err := os.OpenFile(PidFile(), os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	if err != nil {
		return err
	}

	f.WriteString(fmt.Sprintf("%d", pid))
	return nil
}

func ReadPidFile() (int, error) {

	f, err := os.Open(PidFile())
	if err != nil {
		return -1, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return -1, errors.New("Error reading PID file")
	}

	pid, atoiErr := strconv.Atoi(scanner.Text())
	if atoiErr != nil {
		return -1, atoiErr
	}

	return pid, nil
}

func LineReader(filePath string) <-chan string {
	if !FileExists(filePath) {
		return nil
	}

	c := make(chan string)
	go func() {
		f, err := os.Open(filePath)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		s := bufio.NewScanner(f)
		for s.Scan() {
			c <- s.Text()
		}
		close(c)
	}()
	return c
}

func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return true
	}
	return false
}

func FileExistsInPath(fileName string) bool {
	_, err := exec.LookPath(fileName)
	return err == nil
}

func GetPathParts(path string) (dirPath, fileName, absPath string) {
	lookup, lookupErr := exec.LookPath(path)
	if lookupErr == nil {
		path = lookup
	}
	absPath, _ = filepath.Abs(path)

	dirPath = filepath.Dir(absPath)
	fileName = filepath.Base(absPath)
	return
}

func GetThisPathParts() (dirPath, fileName, absPath string) {
	return GetPathParts(os.Args[0])
}

const (
	LETTER_BYTES      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LETTER_INDEX_BITS = 6
	LETTER_INDEX_MASK = 1<<LETTER_INDEX_BITS - 1
	LETTER_INDEX_MAX  = 63 / LETTER_INDEX_BITS
)

func RandomString(length int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, length)
	for i, cache, remain := length-1, src.Int63(), LETTER_INDEX_MAX; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), LETTER_INDEX_MAX
		}
		if idx := int(cache & LETTER_INDEX_MASK); idx < len(LETTER_BYTES) {
			b[i] = LETTER_BYTES[idx]
			i--
		}
		cache >>= LETTER_INDEX_BITS
		remain--
	}

	return string(b)
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

const ADDR_PATT = "(?i)^((?:(?:tcp|udp|ip)[46]?)|(?:unix(?:gram|packet)?))://(.+)$"

func ParseAddress(addr string) (proto string, path string, err error) {
	rx, rxErr := regexp.Compile(ADDR_PATT)
	if rxErr != nil {
		return "", "", nil
	}
	if !rx.MatchString(addr) {
		return "", "", errors.New(fmt.Sprintf("Invalid address '%s'", addr))
	}
	m := rx.FindStringSubmatch(addr)
	return m[1], m[2], nil
}
