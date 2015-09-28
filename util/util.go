package util

import (
	"bufio"
	"fmt"
	"io"
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

	"github.com/emccode/rexray/errors"
	"github.com/kardianos/osext"
)

const (
	LogDirPathSuffix = "/var/log/rexray"
	EtcDirPathSuffix = "/etc/rexray"
	BinDirPathSuffix = "/usr/bin"
	RunDirPathSuffix = "/var/run/rexray"
	LibDirPathSuffix = "/var/lib/rexray"
	UnitFilePath     = "/etc/systemd/system/rexray.service"
	InitFilePath     = "/etc/init.d/rexray"
	EnvFileName      = "rexray.env"

	TrimPattern          = `(?s)^\s*(.*?)\s*$`
	NetworkAdressPattern = `(?i)^((?:(?:tcp|udp|ip)[46]?)|(?:unix(?:gram|packet)?))://(.+)$`

	LetterBytes     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LetterIndexBits = 6
	LetterIndexMask = 1<<LetterIndexBits - 1
	LetterIndexMax  = 63 / LetterIndexBits
)

var (
	trimRx    *regexp.Regexp
	netAddrRx *regexp.Regexp

	prefix string

	homeDir     string
	binDirPath  string
	binFilePath string
	logDirPath  string
	libDirPath  string
	runDirPath  string
	etcDirPath  string
	pidFilePath string
)

func init() {
	homeDir = HomeDir()
	prefix = os.Getenv("REXRAY_HOME")

	var err error
	trimRx, err = regexp.Compile(TrimPattern)
	if err != nil {
		panic(err)
	}

	netAddrRx, err = regexp.Compile(NetworkAdressPattern)
	if err != nil {
		panic(err)
	}
}

func GetPrefix() string {
	return prefix
}

func Prefix(p string) {
	if p == "" || p == "/" {
		return
	}

	prefix = p
}

func IsPrefixed() bool {
	return !(prefix == "" || prefix == "/")
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

func Install(args ...string) {
	exec.Command("install", args...).Run()
}

func InstallChownRoot(args ...string) {
	a := []string{"-o", "0", "-g", "0"}
	for _, i := range args {
		a = append(a, i)
	}
	exec.Command("install", a...).Run()
}

func InstallDirChownRoot(dirPath string) {
	InstallChownRoot("-d", dirPath)
}

func EtcDirPath() string {
	if etcDirPath == "" {
		etcDirPath = fmt.Sprintf("%s%s", prefix, EtcDirPathSuffix)
		os.MkdirAll(etcDirPath, 0755)
	}
	return etcDirPath
}

func RunDirPath() string {
	if runDirPath == "" {
		runDirPath = fmt.Sprintf("%s%s", prefix, RunDirPathSuffix)
		os.MkdirAll(runDirPath, 0755)
	}
	return runDirPath
}

func LogDirPath() string {
	if logDirPath == "" {
		logDirPath = fmt.Sprintf("%s%s", prefix, LogDirPathSuffix)
		os.MkdirAll(logDirPath, 0755)
	}
	return logDirPath
}

func LibDirPath() string {
	if libDirPath == "" {
		libDirPath = fmt.Sprintf("%s%s", prefix, LibDirPathSuffix)
		os.MkdirAll(libDirPath, 0755)
	}
	return libDirPath
}

func LibFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", LibDirPath(), fileName)
}

func BinDirPath() string {
	if binDirPath == "" {
		binDirPath = fmt.Sprintf("%s%s", prefix, BinDirPathSuffix)
		os.MkdirAll(binDirPath, 0755)
	}
	return binDirPath
}

func PidFilePath() string {
	if pidFilePath == "" {
		pidFilePath = fmt.Sprintf("%s/rexray.pid", RunDirPath())
	}
	return pidFilePath
}

func BinFilePath() string {
	if binFilePath == "" {
		binFilePath = fmt.Sprintf("%s/rexray", BinDirPath())
	}
	return binFilePath
}

func EtcFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", EtcDirPath(), fileName)
}

func LogFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", LogDirPath(), fileName)
}

func LogFile(fileName string) (io.Writer, error) {
	return os.OpenFile(
		LogFilePath(fileName), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
}

func StdOutAndLogFile(fileName string) (io.Writer, error) {
	lf, lfErr := LogFile(fileName)
	if lfErr != nil {
		return nil, lfErr
	}
	return io.MultiWriter(os.Stdout, lf), nil
}

func WriteStringToFile(text, path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	if err != nil {
		return err
	}

	f.WriteString(text)
	return nil
}

func ReadFileToString(path string) (string, error) {

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return "", errors.WithField("path", path, "error reading file")
	}

	return scanner.Text(), nil
}

func WritePidFile(pid int) error {

	if pid < 0 {
		pid = os.Getpid()
	}

	return WriteStringToFile(fmt.Sprintf("%d", pid), PidFilePath())
}

func ReadPidFile() (int, error) {

	pidStr, pidStrErr := ReadFileToString(PidFilePath())
	if pidStrErr != nil {
		return -1, pidStrErr
	}

	pid, atoiErr := strconv.Atoi(pidStr)
	if atoiErr != nil {
		return -1, atoiErr
	}

	return pid, nil
}

func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
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
	exeFile, err := osext.Executable()
	if err != nil {
		panic(err)
	}
	return GetPathParts(exeFile)
}

func RandomString(length int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, length)
	for i, cache, remain := length-1, src.Int63(), LetterIndexMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), LetterIndexMax
		}
		if idx := int(cache & LetterIndexMask); idx < len(LetterBytes) {
			b[i] = LetterBytes[idx]
			i--
		}
		cache >>= LetterIndexBits
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

// ParseAddress parses a standard golang network address and returns the
// protocol and path.
func ParseAddress(addr string) (proto string, path string, err error) {
	m := netAddrRx.FindStringSubmatch(addr)
	if m == nil {
		return "", "", errors.WithField("address", addr, "invalid address")
	}
	return m[1], m[2], nil
}

// Trim removes all leading and trailing whitespace, including tab, newline,
// and carriage return characters.
func Trim(text string) string {
	m := trimRx.FindStringSubmatch(text)
	if m == nil {
		return text
	}
	return m[1]
}
