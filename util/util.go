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

	"github.com/kardianos/osext"

	"github.com/emccode/rexray/core/errors"
	"github.com/emccode/rexray/core/version"
)

const (
	logDirPathSuffix = "/var/log/rexray"
	etcDirPathSuffix = "/etc/rexray"
	binDirPathSuffix = "/usr/bin"
	runDirPathSuffix = "/var/run/rexray"
	libDirPathSuffix = "/var/lib/rexray"

	trimPattern          = `(?s)^\s*(.*?)\s*$`
	networkAdressPattern = `(?i)^((?:(?:tcp|udp|ip)[46]?)|(?:unix(?:gram|packet)?))://(.+)$`

	letterBytes     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIndexBits = 6
	letterIndexMask = 1<<letterIndexBits - 1
	letterIndexMax  = 63 / letterIndexBits

	// UnitFilePath is the path to the SystemD service's unit file.
	UnitFilePath = "/etc/systemd/system/rexray.service"

	// InitFilePath is the path to the SystemV Service's init script.
	InitFilePath = "/etc/init.d/rexray"

	// EnvFileName is the name of the environment file used by the SystemD
	// service.
	EnvFileName = "rexray.env"
)

var (
	trimRx    *regexp.Regexp
	netAddrRx *regexp.Regexp

	thisExeDir     string
	thisExeName    string
	thisExeAbsPath string

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

	trimRx = regexp.MustCompile(trimPattern)
	netAddrRx = regexp.MustCompile(networkAdressPattern)

	thisExeDir, thisExeName, thisExeAbsPath = GetThisPathParts()
}

// GetPrefix gets the root path to the REX-Ray data.
func GetPrefix() string {
	return prefix
}

// Prefix sets the root path to the REX-Ray data.
func Prefix(p string) {
	if p == "" || p == "/" {
		return
	}

	prefix = p
}

// IsPrefixed returns a flag indicating whether or not a prefix value is set.
func IsPrefixed() bool {
	return !(prefix == "" || prefix == "/")
}

// HomeDir returns the home directory of the user that owns the current process.
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

// Install executes the system install command.
func Install(args ...string) {
	exec.Command("install", args...).Run()
}

// InstallChownRoot executes the system install command and chowns the target
// to the root user and group.
func InstallChownRoot(args ...string) {
	a := []string{"-o", "0", "-g", "0"}
	for _, i := range args {
		a = append(a, i)
	}
	exec.Command("install", a...).Run()
}

// InstallDirChownRoot executes the system install command with a -d flag and
// chowns the target to the root user and group.
func InstallDirChownRoot(dirPath string) {
	InstallChownRoot("-d", dirPath)
}

// EtcDirPath returns the path to the REX-Ray etc directory.
func EtcDirPath() string {
	if etcDirPath == "" {
		etcDirPath = fmt.Sprintf("%s%s", prefix, etcDirPathSuffix)
		os.MkdirAll(etcDirPath, 0755)
	}
	return etcDirPath
}

// RunDirPath returns the path to the REX-Ray run directory.
func RunDirPath() string {
	if runDirPath == "" {
		runDirPath = fmt.Sprintf("%s%s", prefix, runDirPathSuffix)
		os.MkdirAll(runDirPath, 0755)
	}
	return runDirPath
}

// LogDirPath returns the path to the REX-Ray log directory.
func LogDirPath() string {
	if logDirPath == "" {
		logDirPath = fmt.Sprintf("%s%s", prefix, logDirPathSuffix)
		os.MkdirAll(logDirPath, 0755)
	}
	return logDirPath
}

// LibDirPath returns the path to the REX-Ray bin directory.
func LibDirPath() string {
	if libDirPath == "" {
		libDirPath = fmt.Sprintf("%s%s", prefix, libDirPathSuffix)
		os.MkdirAll(libDirPath, 0755)
	}
	return libDirPath
}

// LibFilePath returns the path to a file inside the REX-Ray lib directory
// with the provided file name.
func LibFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", LibDirPath(), fileName)
}

// BinDirPath returns the path to the REX-Ray bin directory.
func BinDirPath() string {
	if binDirPath == "" {
		binDirPath = fmt.Sprintf("%s%s", prefix, binDirPathSuffix)
		os.MkdirAll(binDirPath, 0755)
	}
	return binDirPath
}

// PidFilePath returns the path to the REX-Ray PID file.
func PidFilePath() string {
	if pidFilePath == "" {
		pidFilePath = fmt.Sprintf("%s/rexray.pid", RunDirPath())
	}
	return pidFilePath
}

// BinFilePath returns the path to the REX-Ray executable.
func BinFilePath() string {
	if binFilePath == "" {
		binFilePath = fmt.Sprintf("%s/rexray", BinDirPath())
	}
	return binFilePath
}

// EtcFilePath returns the path to a file inside the REX-Ray etc directory
// with the provided file name.
func EtcFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", EtcDirPath(), fileName)
}

// LogFilePath returns the path to a file inside the REX-Ray log directory
// with the provided file name.
func LogFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", LogDirPath(), fileName)
}

// LogFile returns a writer to a file inside the REX-Ray log directory
// with the provided file name.
func LogFile(fileName string) (io.Writer, error) {
	return os.OpenFile(
		LogFilePath(fileName), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
}

// StdOutAndLogFile returns a mutltiplexed writer for the current process's
// stdout descriptor and a REX-Ray log file with the provided name.
func StdOutAndLogFile(fileName string) (io.Writer, error) {
	lf, lfErr := LogFile(fileName)
	if lfErr != nil {
		return nil, lfErr
	}
	return io.MultiWriter(os.Stdout, lf), nil
}

// WriteStringToFile writes the string to the file at the provided path.
func WriteStringToFile(text, path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	if err != nil {
		return err
	}

	f.WriteString(text)
	return nil
}

// ReadFileToString reads the file at the provided path to a string.
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

// WritePidFile writes the current process ID to the REX-Ray PID file.
func WritePidFile(pid int) error {

	if pid < 0 {
		pid = os.Getpid()
	}

	return WriteStringToFile(fmt.Sprintf("%d", pid), PidFilePath())
}

// ReadPidFile reads the REX-Ray PID from the PID file.
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

// IsDirEmpty returns a flag indicating whether or not a directory has any
// child objects such as files or directories in it.
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

// LineReader returns a channel that reads the contents of a file line-by-line.
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

// GetPathParts returns the absolute directory path, the file name, and the
// absolute path of the provided path string.
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

// GetThisPathParts returns the same information as GetPathParts for the
// current executable.
func GetThisPathParts() (dirPath, fileName, absPath string) {
	exeFile, err := osext.Executable()
	if err != nil {
		panic(err)
	}
	return GetPathParts(exeFile)
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

// PrintVersion prints the current version information to the provided writer.
func PrintVersion(out io.Writer) {
	fmt.Fprintf(out, "Binary: %s\n", thisExeAbsPath)
	fmt.Fprintf(out, "SemVer: %s\n", version.SemVer)
	fmt.Fprintf(out, "OsArch: %s\n", version.Arch)
	fmt.Fprintf(out, "Branch: %s\n", version.Branch)
	fmt.Fprintf(out, "Commit: %s\n", version.ShaLong)
	fmt.Fprintf(out, "Formed: %s\n", version.EpochToRfc1123())
}
