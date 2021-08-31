package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/akutz/gotil"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

const (
	lowerhex = "0123456789abcdef"
)

// loggingHandler is an HTTP logging handler for the libStorage service
// endpoint.
type loggingHandler struct {
	handler      types.APIFunc
	logRequests  bool
	logResponses bool
	writer       io.Writer
}

// NewLoggingHandler instantiates a new instance of the loggingHandler type.
func NewLoggingHandler(
	w io.Writer,
	logHTTPRequests, logHTTPResponses bool) types.Middleware {

	return &loggingHandler{
		writer:       w,
		logRequests:  logHTTPRequests,
		logResponses: logHTTPResponses,
	}
}

func (h *loggingHandler) Name() string {
	return "logging-handler"
}

func (h *loggingHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&loggingHandler{m, h.logRequests, h.logResponses, h.writer}).Handle
}

// Handle is the type's Handler function.
func (h *loggingHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	bw := &bytes.Buffer{}
	defer func(w io.Writer) {
		h.writer.Write(bw.Bytes())
	}(bw)

	var err error
	var reqDump []byte
	if h.logRequests {
		if reqDump, err = httputil.DumpRequest(req, true); err != nil {
			return err
		}
	}

	rec := httptest.NewRecorder()
	reqErr := h.handler(ctx, rec, req, store)

	logRequest(h.logRequests, bw, rec, req, reqDump)

	if reqErr != nil {
		return reqErr
	}

	if h.logResponses {
		fmt.Fprintln(bw, "")
		logResponse(bw, rec, req)
		fmt.Fprintln(bw, "")
	}

	for k, v := range rec.HeaderMap {
		w.Header()[k] = v
	}
	w.WriteHeader(rec.Code)
	if req.Method != http.MethodHead {
		w.Write(rec.Body.Bytes())
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
	gotil.WriteIndented(w, reqDump)
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
	if !isBinaryContent(rec.HeaderMap) {
		gotil.WriteIndented(w, rec.Body.Bytes())
	}
}

func isBinaryContent(headers http.Header) bool {
	v, ok := headers["Content-Type"]
	if !ok || len(v) == 0 {
		return false
	}
	return v[0] == "application/octet-stream"
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
		buf = appendRune(buf, s, r)
	}
	return buf
}

func appendRune(buf []byte, s string, r rune) []byte {
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
	return buf
}
