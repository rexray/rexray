package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"
)

func main() {
	var (
		tpl    *template.Template
		format string
		output string
		export bool
	)

	flag.StringVar(
		&format,
		"f",
		"ver",
		"The output format: env, go, json, mk, rpm, ver")
	flag.StringVar(
		&output,
		"o",
		"",
		"The output file")
	flag.BoolVar(
		&export,
		"x",
		false,
		"Export env vars. Used with -f env")
	flag.Parse()

	if strings.EqualFold("env", format) {
		format = "env"
	} else if strings.EqualFold("go", format) {
		format = "go"
	} else if strings.EqualFold("json", format) {
		format = "json"
	} else if strings.EqualFold("mk", format) {
		format = "mk"
	} else if strings.EqualFold("rpm", format) {
		format = "rpm"
	} else if strings.EqualFold("ver", format) {
		format = "ver"
	} else {
		if fileExists(format) {
			buf, err := ioutil.ReadFile(format)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: read tpl failed: %v\n", err)
				os.Exit(1)
			}
			format = string(buf)
		}
		tpl = template.Must(template.New("tpl").Parse(format))
		format = "tpl"
	}

	var w io.Writer = os.Stdout
	if len(output) > 0 {
		fout, err := os.Create(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		w = fout
		defer fout.Close()
	}

	gitdesc := chkErr(doExec("git", "describe", "--long", "--dirty"))
	rx := regexp.MustCompile(
		`^[^\d]*(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z].+?))?(?:-(\d+)-g(.+?)(?:-(dirty))?)?\s*$`)
	m := rx.FindStringSubmatch(gitdesc)
	if len(m) == 0 {
		fmt.Fprintf(os.Stderr, "error: match git describe failed: %s\n", gitdesc)
		os.Exit(1)
	}

	goos := os.Getenv("XGOOS")
	if goos == "" {
		goos = runtime.GOOS
	}
	goarch := os.Getenv("XGOARCH")
	if goarch == "" {
		goarch = runtime.GOARCH
	}

	ver := &semver{
		GOOS:   goos,
		GOARCH: goarch,
		OS:     goosToUname[goos],
		Arch:   goarchToUname[goarch],
		Major:  toInt(m[1]),
		Minor:  toInt(m[2]),
		Patch:  toInt(m[3]),
		Notes:  m[4],
		Build:  toInt(m[5]),
		Sha7:   m[6],
		Sha32:  chkErr(doExec("git", "log", "-n1", `--format=%H`)),
		Dirty:  m[7] != "",
		Epoch:  toInt64(chkErr(doExec("git", "log", "-n1", `--format=%ct`))),
	}
	ver.SemVer = ver.String()
	ver.SemVerRPM = ver.RPM()
	ver.BuildDate = ver.Timestamp().Format("Mon, 02 Jan 2006 15:04:05 MST")
	ver.ReleaseDate = ver.Timestamp().Format("06-01-02")

	switch format {
	case "env":
		for _, v := range ver.EnvVars() {
			if export {
				fmt.Fprint(w, "export ")
			}
			fmt.Fprintln(w, v)
		}
	case "go":
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(ver); err != nil {
			fmt.Fprintf(os.Stderr, "error: encode to json failed: %v\n", err)
			os.Exit(1)
		}
	case "mk":
		for _, v := range ver.EnvVars() {
			p := strings.SplitN(v, "=", 2)
			key := p[0]
			fmt.Fprintf(w, "%s ?=", key)
			if len(p) == 1 {
				fmt.Fprintln(w)
				continue
			}
			val := p[1]
			if strings.HasPrefix(val, `"`) &&
				strings.HasSuffix(val, `"`) {
				val = val[1 : len(val)-1]
			}
			val = strings.Replace(val, "$", "$$", -1)
			fmt.Fprintf(w, " %s\n", val)
		}
	case "rpm":
		fmt.Fprintln(w, ver.RPM())
	case "tpl":
		if err := tpl.Execute(w, ver); err != nil {
			fmt.Fprintf(os.Stderr, "error: template failed: %v\n", err)
			os.Exit(1)
		}
	case "ver":
		fmt.Fprintln(w, ver.String())
	}
}

func doExec(cmd string, args ...string) ([]byte, error) {
	c := exec.Command(cmd, args...)
	c.Stderr = os.Stderr
	return c.Output()
}

func chkErr(out []byte, err error) string {
	if err == nil {
		return strings.TrimSpace(string(out))
	}

	e, ok := err.(*exec.ExitError)
	if !ok {
		os.Exit(1)
	}

	st, ok := e.Sys().(syscall.WaitStatus)
	if !ok {
		os.Exit(1)
	}

	os.Exit(st.ExitStatus())
	return ""
}

type semver struct {
	GOOS        string `json:"goos"`
	GOARCH      string `json:"goarch"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	Major       int    `json:"major"`
	Minor       int    `json:"minor"`
	Patch       int    `json:"patch"`
	Build       int    `json:"build"`
	Notes       string `json:"notes"`
	Dirty       bool   `json:"dirty"`
	Sha7        string `json:"sha7"`
	Sha32       string `json:"sha32"`
	Epoch       int64  `json:"epoch"`
	SemVer      string `json:"semver"`
	SemVerRPM   string `json:"semverRPM"`
	BuildDate   string `json:"buildDate"`
	ReleaseDate string `json:"releaseDate"`
}

func (v *semver) String() string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.Notes) > 0 {
		fmt.Fprintf(buf, "-%s", v.Notes)
	}
	if v.Build > 0 {
		fmt.Fprintf(buf, "+%d", v.Build)
	}
	if v.Dirty {
		fmt.Fprint(buf, "+dirty")
	}
	return buf.String()
}

func (v *semver) RPM() string {
	return strings.Replace(v.String(), "-", "+", -1)
}

func (v *semver) EnvVars() []string {
	return []string{
		fmt.Sprintf("GOOS=%s", v.GOOS),
		fmt.Sprintf("GOARCH=%s", v.GOARCH),
		fmt.Sprintf("OS=%s", v.OS),
		fmt.Sprintf("ARCH=%s", v.Arch),
		fmt.Sprintf("MAJOR=%d", v.Major),
		fmt.Sprintf("MINOR=%d", v.Minor),
		fmt.Sprintf("PATCH=%d", v.Patch),
		fmt.Sprintf("BUILD=%d", v.Build),
		fmt.Sprintf("NOTES=\"%s\"", v.Notes),
		fmt.Sprintf("DIRTY=%v", v.Dirty),
		fmt.Sprintf("SHA7=%s", v.Sha7),
		fmt.Sprintf("SHA32=%s", v.Sha32),
		fmt.Sprintf("EPOCH=%d", v.Epoch),
		fmt.Sprintf("SEMVER=\"%s\"", v.SemVer),
		fmt.Sprintf("SEMVER_RPM=\"%s\"", v.SemVerRPM),
		fmt.Sprintf("BUILD_DATE=\"%s\"", v.BuildDate),
		fmt.Sprintf("RELEASE_DATE=\"%s\"", v.ReleaseDate),
	}
}

func (v *semver) Timestamp() time.Time {
	return time.Unix(v.Epoch, 0)
}

func toInt(sz string) int {
	i, _ := strconv.Atoi(sz)
	return i
}

func toInt64(sz string) int64 {
	i, _ := strconv.Atoi(sz)
	return int64(i)
}

var goosToUname = map[string]string{
	"android":   "Android",
	"darwin":    "Darwin",
	"dragonfly": "DragonFly",
	"freebsd":   "kFreeBSD",
	"linux":     "Linux",
	"nacl":      "NaCl",
	"netbsd":    "NetBSD",
	"openbsd":   "OpenBSD",
	"plan9":     "Plan9",
	"solaris":   "Solaris",
	"windows":   "Windows",
}

var goarchToUname = map[string]string{
	"386":      "i386",
	"amd64":    "x86_64",
	"amd64p32": "x86_64_P32",
	"arm":      "ARMv7",
	"arm64":    "ARMv8",
	"mips":     "MIPS32",
	"mips64":   "MIPS64",
	"mips64le": "MIPS64LE",
	"mipsle":   "MIPS32LE",
	"ppc64":    "PPC64",
	"ppc64le":  "PPC64LE",
	"s390x":    "S390X",
}

func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return true
	}
	return false
}
