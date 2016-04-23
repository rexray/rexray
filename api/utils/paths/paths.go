package paths

import (
	"fmt"
	"io"
	"os"

	"github.com/akutz/gotil"
)

var (
	etcDirPathSuffix = "/etc/libstorage"
	libDirPathSuffix = "/var/lib/libstorage"
	logDirPathSuffix = "/var/log/libstorage"
	runDirPathSuffix = "/var/run/libstorage"
)

var (
	thisExeDir     string
	thisExeName    string
	thisExeAbsPath string

	prefix string

	etcDirPath string
	libDirPath string
	logDirPath string
	runDirPath string
	usrDirPath string
)

func init() {
	prefix = os.Getenv("LIBSTORAGE_HOME")
	thisExeDir, thisExeName, thisExeAbsPath = gotil.GetThisPathParts()
}

// GetPrefix gets the root path to the libStorage data.
func GetPrefix() string {
	return prefix
}

// Prefix sets the root path to the libStorage data.
func Prefix(p string) {
	if p == "" || p == "/" {
		return
	}

	logDirPath = ""
	etcDirPath = ""

	prefix = p
}

// IsPrefixed returns a flag indicating whether or not a prefix value is set.
func IsPrefixed() bool {
	return !(prefix == "" || prefix == "/")
}

// EtcDirPath returns the path to the REX-Ray etc directory.
func EtcDirPath() string {
	if etcDirPath == "" {
		etcDirPath = fmt.Sprintf("%s%s", prefix, etcDirPathSuffix)
		os.MkdirAll(etcDirPath, 0755)
	}
	return etcDirPath
}

// LibDirPath returns the path to the lib directory.
func LibDirPath() string {
	if libDirPath == "" {
		libDirPath = fmt.Sprintf("%s%s", prefix, libDirPathSuffix)
		os.MkdirAll(libDirPath, 0755)
	}
	return libDirPath
}

// LogDirPath returns the path to the log directory.
func LogDirPath() string {
	if logDirPath == "" {
		logDirPath = fmt.Sprintf("%s%s", prefix, logDirPathSuffix)
		os.MkdirAll(logDirPath, 0755)
	}
	return logDirPath
}

// RunDirPath returns the path to the run directory.
func RunDirPath() string {
	if runDirPath == "" {
		runDirPath = fmt.Sprintf("%s%s", prefix, runDirPathSuffix)
		os.MkdirAll(runDirPath, 0755)
	}
	return runDirPath
}

// UsrDirPath is the path to the user libStorage config directory.
func UsrDirPath() string {
	if usrDirPath == "" {
		usrDirPath = fmt.Sprintf("%s/.libstorage", gotil.HomeDir())
		os.MkdirAll(usrDirPath, 0755)
	}

	return usrDirPath
}

// EtcFilePath returns the path to a file inside the etc directory with the
// provided file name.
func EtcFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", EtcDirPath(), fileName)
}

// LibFilePath returns the path to a file inside the lib directory with the
// the provided file name.
func LibFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", LibDirPath(), fileName)
}

// LogFilePath returns the path to a file inside the log directory with the
// provided file name.
func LogFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", LogDirPath(), fileName)
}

// RunFilePath returns the path to a file inside the run directory with the
// provided file name.
func RunFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", RunDirPath(), fileName)
}

// UsrFilePath returns the path to a file inside the usr directory with the
// provided file name.
func UsrFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", UsrDirPath(), fileName)
}

// LogFile returns a writer to a file inside the log directory with the
// provided file name.
func LogFile(fileName string) (io.Writer, error) {
	return os.OpenFile(
		LogFilePath(fileName), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
}

// StdOutAndLogFile returns a mutltiplexed writer for the current process's
// stdout descriptor and alog file with the provided name.
func StdOutAndLogFile(fileName string) (io.Writer, error) {
	lf, lfErr := LogFile(fileName)
	if lfErr != nil {
		return nil, lfErr
	}
	return io.MultiWriter(os.Stdout, lf), nil
}
