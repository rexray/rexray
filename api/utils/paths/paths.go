package paths

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"runtime"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gotil"
)

type fileKey int

const (

	// Home is the application home directory.
	Home fileKey = iota

	minDirKey

	// Etc is the application etc directory.
	Etc

	// Lib is the application lib directory.
	Lib

	// Log is the application log directory.
	Log

	// Run is the application run directory.
	Run

	maxDirKey

	minFileKey fileKey = maxDirKey + 1

	// LSX is the path to the libStorage executor.
	LSX
)

// Exists returns a flag indicating whether or not the file/directory exists.
func (k fileKey) Exists() bool {
	return gotil.FileExists(k.String())
}

// Format may call Sprint(f) or Fprint(f) etc. to generate its output.
func (k fileKey) Format(f fmt.State, c rune) {
	fs := &bytes.Buffer{}
	fs.WriteRune('%')
	if f.Flag('+') {
		fs.WriteRune('+')
	}
	if f.Flag('-') {
		fs.WriteRune('-')
	}
	if f.Flag('#') {
		fs.WriteRune('#')
	}
	if f.Flag(' ') {
		fs.WriteRune(' ')
	}
	if f.Flag('0') {
		fs.WriteRune('0')
	}
	if w, ok := f.Width(); ok {
		fs.WriteString(fmt.Sprintf("%d", w))
	}
	if p, ok := f.Precision(); ok {
		fs.WriteString(fmt.Sprintf("%d", p))
	}
	var (
		s  string
		cc = c
	)
	if c == 'k' {
		s = k.key()
		cc = 's'
	} else {
		s = k.String()
	}
	fs.WriteRune(cc)
	fmt.Fprintf(f, fs.String(), s)
}

func (k fileKey) parent() fileKey {
	if k < maxDirKey {
		return Home
	}
	return Lib
}

func (k fileKey) perms() os.FileMode {
	if k <= LSX {
		return 0755
	}
	return 0644
}

func (k fileKey) key() string {
	switch k {
	case Home:
		return "home"
	case Etc:
		return "etc"
	case Lib:
		return "lib"
	case Log:
		return "log"
	case Run:
		return "run"
	case LSX:
		return "lsx"
	}
	return ""
}

func (k fileKey) defaultVal() string {
	switch k {
	case Home:
		return libstorageHome
	case Etc:
		return "/etc/libstorage"
	case Lib:
		return "/var/lib/libstorage"
	case Log:
		return "/var/log/libstorage"
	case Run:
		return "/var/run/libstorage"
	case LSX:
		return fmt.Sprintf("lsx-%s", runtime.GOOS)
	}
	return ""
}

func (k fileKey) get() string {
	if v, ok := keyCache[k]; ok {
		return v
	}
	if k == Home {
		return libstorageHome
	}

	if p, ok := keyCache[k.parent()]; ok {
		return Join(p, k.defaultVal())
	}
	return k.defaultVal()
}

func (k fileKey) Join(elem ...string) string {
	log.WithField("elem", elem).Debug("enter join")

	var elems []string
	if _, ok := keyCache[k]; !ok {
		elems = []string{Home.String()}
	}
	if k != Home {
		elems = append(elems, k.String())
	}
	elems = append(elems, elem...)
	log.WithField("elem", elems).Debug("exit join")
	return path.Join(elems...)
}

func (k fileKey) String() string {
	if k == Home {
		homeLock.Lock()
		defer homeLock.Unlock()
	}

	if v, ok := keyCache[k]; ok {
		return v
	}

	log.WithFields(log.Fields{
		"key": k.key(),
	}).Debug("must init path")

	k.init()
	k.cache()

	return keyCache[k]
}

func (k fileKey) cache() {
	keyCache[k] = k.get()
	log.WithFields(log.Fields{
		"key":  k.key(),
		"path": k.get(),
	}).Debug("cached key")
}

func (k fileKey) init() {

	if k == Home {
		if !checkPerms(k, false) {
			failedPath := k.get()
			libstorageHome = Join(gotil.HomeDir(), ".libstorage")
			log.WithFields(log.Fields{
				"failedPath": failedPath,
				"newPath":    k.get(),
			}).Debug("first make homedir failed, trying again")
			checkPerms(k, true)
		}
		return
	}

	checkPerms(k, true)
}

var (
	libstorageHome = os.Getenv("LIBSTORAGE_HOME")
	keyCache       = map[fileKey]string{}
	homeLock       = &sync.Mutex{}
	thisExeDir     string
	thisExeName    string
	thisExeAbsPath string
	slashRX        = regexp.MustCompile(`^((?:/)|(?:[a-zA-Z]\:\\?))?$`)
)

// TODO Fix this logic
func init() {
	if libstorageHome == "" {
		libstorageHome = "/"
	}

	// if not root and home is /, change home to user's home dir
	if os.Geteuid() != 0 && libstorageHome == "/" {
		libstorageHome = Join(gotil.HomeDir(), ".libstorage")
	}

	thisExeDir, thisExeName, thisExeAbsPath = gotil.GetThisPathParts()
}

// Init is a way to manually initialize the package.
func Init() {
	Home.init()
}

// Join joins one or more paths.
func Join(elem ...string) string {
	return path.Join(elem...)
}

func checkPerms(k fileKey, mustPerm bool) bool {
	if k > maxDirKey {
		return true
	}

	p := k.get()

	fields := log.Fields{
		"path":     p,
		"perms":    k.perms(),
		"mustPerm": mustPerm,
	}

	if gotil.FileExists(p) {
		if log.GetLevel() == log.DebugLevel {
			log.WithField("path", p).Debug("file exists")
		}
	} else {
		log.WithFields(fields).Info("making libStorage directory")
		noPermMkdirErr := fmt.Sprintf("mkdir %s: permission denied", p)
		if err := os.MkdirAll(p, k.perms()); err != nil {
			if err.Error() == noPermMkdirErr {
				if mustPerm {
					log.WithFields(fields).Panic(noPermMkdirErr)
				}
				return false
			}
		}
	}

	touchFilePath := Join(p, fmt.Sprintf(".touch-%v", time.Now().Unix()))
	defer os.RemoveAll(touchFilePath)

	noPermTouchErr := fmt.Sprintf("open %s: permission denied", touchFilePath)

	if _, err := os.Create(touchFilePath); err != nil {
		if err.Error() == noPermTouchErr {
			if mustPerm {
				log.WithFields(fields).Panic(noPermTouchErr)
			}
			return false
		}
	}

	return true
}

// LogFile returns a writer to a file inside the log directory with the
// provided file name.
func LogFile(fileName string) (io.Writer, error) {
	return os.OpenFile(
		Log.Join(fileName), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
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
