package server

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AVENTER-UG/rexray/core"
	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/services"
)

var (
	// DisableStartupInfo is a flag that indicates whether or not to print
	// startup info. This is sometimes disabled for CI systems to reduce logs.
	DisableStartupInfo, _ = strconv.ParseBool(os.Getenv(
		"LIBSTORAGE_DISABLE_STARTUP_INFO"))
)

const (
	dateFormat = "2006/01/02 15:04:05.000"

	serverStartupLogo = `
                  _ _ _     _____ _
                 | (_) |   / ____| |
                 | |_| |__| (___ | |_ ___  _ __ __ _  __ _  ___
                 | | | '_ \\___ \| __/ _ \| '__/ _' |/ _' |/ _ \
                 | | | |_) |___) | || (_) | | | (_| | (_| |  __/
                 |_|_|_.__/_____/ \__\___/|_|  \__,_|\__, |\___|
                                                      __/ |
                                                     |___/

`
)

func (s *server) PrintServerStartupHeader(w io.Writer) {

	if DisableStartupInfo {
		return
	}

	var (
		n          int
		b          = &bytes.Buffer{}
		bar        = strings.Repeat("#", 80)
		barl       = fmt.Sprintf("##%s##", strings.Repeat(" ", 76))
		now        = time.Now().UTC().Format(dateFormat)
		vts        = core.CommitTime.Format(time.RFC1123)
		pathConfig = context.MustPathConfig(s.ctx)
	)

	if pathConfig == nil {
		panic("pathConfig is nil")
	}

	fmt.Fprint(b, serverStartupLogo)
	fmt.Fprintln(b, bar)
	fmt.Fprintln(b, barl)
	fmt.Fprint(b, "##                  ")
	fmt.Fprintf(b, "libStorage starting - %s", now)
	fmt.Fprintln(b, "             ##")
	fmt.Fprintln(b, barl)

	n, _ = fmt.Fprintf(b, "##     server:      %s", s.name)
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")

	n, _ = fmt.Fprintf(b, "##      token:      %s", s.adminToken)
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")

	fmt.Fprintln(b, barl)

	n, _ = fmt.Fprintf(b, "##     semver:      %s", core.SemVer)
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")
	n, _ = fmt.Fprintf(b, "##     osarch:      %s", core.Arch)
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")
	n, _ = fmt.Fprintf(b, "##     commit:      %s", core.CommitSha7)
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")
	n, _ = fmt.Fprintf(b, "##     formed:      %s", vts)
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")

	fmt.Fprintln(b, barl)

	n, _ = fmt.Fprintf(b, "##        etc:      %s", pathConfig.Etc)
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")
	n, _ = fmt.Fprintf(b, "##        tls:      %s", pathConfig.TLS)
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")
	n, _ = fmt.Fprintf(b, "##        lib:      %s", pathConfig.Lib)
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")
	n, _ = fmt.Fprintf(b, "##        log:      %s", pathConfig.Log)
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")
	n, _ = fmt.Fprintf(b, "##        run:      %s", pathConfig.Run)
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")

	fmt.Fprintln(b, barl)

	fmt.Fprintln(b, bar)
	fmt.Fprintln(b)
	io.Copy(w, b)
}

func (s *server) PrintServerStartupFooter(w io.Writer) {

	if DisableStartupInfo {
		return
	}

	var (
		n     int
		srvcs []string
		drvrs []string
		addrs = s.Addrs()
		b     = &bytes.Buffer{}
		bar   = strings.Repeat("#", 80)
		barl  = fmt.Sprintf("##%s##", strings.Repeat(" ", 76))
		now   = time.Now().UTC().Format(dateFormat)
	)

	fmt.Fprintln(b)
	fmt.Fprintln(b, bar)
	fmt.Fprintln(b, barl)
	fmt.Fprint(b, "##                  ")
	fmt.Fprintf(b, "libStorage started  - %s", now)
	fmt.Fprintln(b, "             ##")
	fmt.Fprintln(b, barl)

	n, _ = fmt.Fprintf(b, "##     endpoints:   %s", addrs[0])
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")

	if len(addrs) > 1 {
		for x := range addrs {
			if x == 0 {
				continue
			}
			n, _ = fmt.Fprintf(b, "##                  %s", addrs[x])
			fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
			fmt.Fprintln(b, "##")
		}
	}

	fmt.Fprintln(b, barl)

	for v := range services.StorageServices(s.ctx) {
		srvcs = append(srvcs, v.Name())
		drvrs = append(drvrs, v.Driver().Name())
	}

	n, _ = fmt.Fprintf(b, "##      services:   name=%s, driver=%s",
		srvcs[0], drvrs[0])
	fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
	fmt.Fprintln(b, "##")

	if len(srvcs) > 1 {
		for x := range srvcs {
			if x == 0 {
				continue
			}
			n, _ = fmt.Fprintf(b, "##                  name=%s, driver=%s",
				srvcs[x], drvrs[x])
			fmt.Fprint(b, strings.Repeat(" ", trunc80(n)))
			fmt.Fprintln(b, "##")
		}
	}

	fmt.Fprintln(b, barl)
	fmt.Fprintln(b, bar)
	fmt.Fprintln(b)
	io.Copy(w, b)
}

func trunc80(n int) int {
	i := 80 - (n + 2)
	if i < 0 {
		return 0
	}
	return i
}
