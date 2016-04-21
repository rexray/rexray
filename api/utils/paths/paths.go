package paths

import (
	"fmt"
	"io"
	"os"

	"github.com/akutz/gotil"
)

var (
	logDirPathSuffix = "/var/log/libstor"
	etcDirPathSuffix = "/etc/libstor"

	// UsrDirPath is the path to the user libStorage config directory.
	UsrDirPath = fmt.Sprintf("%s/.libstor", gotil.HomeDir())
)

var (
	thisExeDir     string
	thisExeName    string
	thisExeAbsPath string

	prefix string

	logDirPath string
	etcDirPath string
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

// LogDirPath returns the path to the log directory.
func LogDirPath() string {
	if logDirPath == "" {
		logDirPath = fmt.Sprintf("%s%s", prefix, logDirPathSuffix)
		os.MkdirAll(logDirPath, 0755)
	}
	return logDirPath
}

// EtcFilePath returns the path to a file inside the etc directory with the
// provided file name.
func EtcFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", EtcDirPath(), fileName)
}

// LogFilePath returns the path to a file inside the log directory with the
// provided file name.
func LogFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", LogDirPath(), fileName)
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
