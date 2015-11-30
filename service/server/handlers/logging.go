package handlers

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/util"
)

const (
	lowerhex = "0123456789abcdef"
)

// LoggingHandler is an HTTP logging handler for the libStorage service
// endpoint.
type LoggingHandler struct {
	Enabled      bool
	LogRequests  bool
	LogResponses bool
	StdOut       io.WriteCloser
	StdErr       io.WriteCloser

	handler http.Handler
	config  gofig.Config
}

// NewLoggingHandler instantiates a new instance of the LoggingHandler type.
func NewLoggingHandler(
	handler http.Handler,
	config gofig.Config) *LoggingHandler {

	h := &LoggingHandler{
		handler: handler,
		config:  config,
	}

	h.Enabled = config.GetBool("libstorage.service.http.logging.enabled")
	if h.Enabled {
		h.StdOut = GetLogIO(
			"libstorage.service.http.logging.out", config)
		h.LogRequests = config.GetBool(
			"libstorage.service.http.logging.logrequest")
		h.LogResponses = config.GetBool(
			"libstorage.service.http.logging.logresponse")
	}
	return h
}

// GetLogIO gets an io.WriteCloser using a given property name from a
// configuration instance.
func GetLogIO(
	propName string,
	config gofig.Config) io.WriteCloser {

	if path := config.GetString(propName); path != "" {
		logio, err := os.OpenFile(
			path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Error(err)
		}
		log.WithFields(log.Fields{
			"logType": propName,
			"logPath": path,
		}).Debug("using log file")
		return logio
	}
	return log.StandardLogger().Writer()
}

// ServeHTTP serves the HTTP request and writes the response.
func (h *LoggingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !h.Enabled {
		h.handler.ServeHTTP(w, req)
		return
	}

	var err error
	var reqDump []byte
	if h.LogRequests {
		if reqDump, err = httputil.DumpRequest(req, true); err != nil {
			log.Error(err)
		}
	}

	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)

	logRequest(h.LogRequests, h.StdOut, rec, req, reqDump)
	if h.LogResponses {
		fmt.Fprintln(h.StdOut, "")
		logResponse(h.StdOut, rec, req)
		fmt.Fprintln(h.StdOut, "")
	}

	w.WriteHeader(rec.Code)
	for k, v := range rec.HeaderMap {
		w.Header()[k] = v
	}
	w.Write(rec.Body.Bytes())
}

// Close closes the resources associated with the handler.
func (h *LoggingHandler) Close() error {
	if h.StdOut != nil {
		if err := h.StdOut.Close(); err != nil {
			return err
		}
	}
	if h.StdErr != nil {
		if err := h.StdErr.Close(); err != nil {
			return err
		}
	}
	return nil
}

func logRequest(
	l bool,
	w io.Writer,
	rec *httptest.ResponseRecorder,
	req *http.Request,
	reqDump []byte) {

	cll := buildCommonLogLine(
		req, *req.URL, time.Now(), rec.Code, rec.Body.Len())
	fmt.Fprintln(w, string(cll))

	if !l || len(reqDump) == 0 {
		return
	}

	fmt.Fprintln(w, "")
	fmt.Fprint(w, "    -------------------------- ")
	fmt.Fprint(w, "HTTP REQUEST (SERVER)")
	fmt.Fprintln(w, " --------------------------")
	util.WriteIndented(w, reqDump)
}

func logResponse(
	w io.Writer,
	rec *httptest.ResponseRecorder,
	req *http.Request) {

	fmt.Fprint(w, "    -------------------------- ")
	fmt.Fprint(w, "HTTP RESPONSE (SERVER)")
	fmt.Fprintln(w, " -------------------------")

	for k, v := range rec.HeaderMap {
		fmt.Fprintf(w, "    %s=%s\n", k, strings.Join(v, ","))
	}
	fmt.Fprintln(w, "")
	util.WriteIndented(w, rec.Body.Bytes())
}

// buildCommonLogLine builds a log entry for req in Apache Common Log Format.
// ts is the timestamp with which the entry should be logged.
// status and size are used to provide the response HTTP status and size.
//
// This function was taken from the Gorilla toolkit's handlers.go file.
func buildCommonLogLine(
	req *http.Request,
	url url.URL,
	ts time.Time,
	status int,
	size int) []byte {

	username := "-"
	if url.User != nil {
		if name := url.User.Username(); name != "" {
			username = name
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)

	if err != nil {
		host = req.RemoteAddr
	}

	uri := req.RequestURI

	// Requests using the CONNECT method over HTTP/2.0 must use
	// the authority field (aka r.Host) to identify the target.
	// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
	if req.ProtoMajor == 2 && req.Method == "CONNECT" {
		uri = req.Host
	}
	if uri == "" {
		uri = url.RequestURI()
	}

	buf := make([]byte, 0, 3*(len(host)+len(username)+
		len(req.Method)+len(uri)+len(req.Proto)+50)/2)
	buf = append(buf, host...)
	buf = append(buf, " - "...)
	buf = append(buf, username...)
	buf = append(buf, " ["...)
	buf = append(buf, ts.Format("02/Jan/2006:15:04:05 -0700")...)
	buf = append(buf, `] "`...)
	buf = append(buf, req.Method...)
	buf = append(buf, " "...)
	buf = appendQuoted(buf, uri)
	buf = append(buf, " "...)
	buf = append(buf, req.Proto...)
	buf = append(buf, `" `...)
	buf = append(buf, strconv.Itoa(status)...)
	buf = append(buf, " "...)
	buf = append(buf, strconv.Itoa(size)...)
	return buf
}

func appendQuoted(buf []byte, s string) []byte {
	var runeTmp [utf8.UTFMax]byte
	for width := 0; len(s) > 0; s = s[width:] {
		r := rune(s[0])
		width = 1
		if r >= utf8.RuneSelf {
			r, width = utf8.DecodeRuneInString(s)
		}
		if width == 1 && r == utf8.RuneError {
			buf = append(buf, `\x`...)
			buf = append(buf, lowerhex[s[0]>>4])
			buf = append(buf, lowerhex[s[0]&0xF])
			continue
		}
		if r == rune('"') || r == '\\' { // always backslashed
			buf = append(buf, '\\')
			buf = append(buf, byte(r))
			continue
		}
		if strconv.IsPrint(r) {
			n := utf8.EncodeRune(runeTmp[:], r)
			buf = append(buf, runeTmp[:n]...)
			continue
		}
		switch r {
		case '\a':
			buf = append(buf, `\a`...)
		case '\b':
			buf = append(buf, `\b`...)
		case '\f':
			buf = append(buf, `\f`...)
		case '\n':
			buf = append(buf, `\n`...)
		case '\r':
			buf = append(buf, `\r`...)
		case '\t':
			buf = append(buf, `\t`...)
		case '\v':
			buf = append(buf, `\v`...)
		default:
			switch {
			case r < ' ':
				buf = append(buf, `\x`...)
				buf = append(buf, lowerhex[s[0]>>4])
				buf = append(buf, lowerhex[s[0]&0xF])
			case r > utf8.MaxRune:
				r = 0xFFFD
				fallthrough
			case r < 0x10000:
				buf = append(buf, `\u`...)
				for s := 12; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			default:
				buf = append(buf, `\U`...)
				for s := 28; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			}
		}
	}
	return buf

}
