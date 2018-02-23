package api

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	log "github.com/akutz/gournal"
)

func isBinOctetBody(h http.Header) bool {
	return h.Get(headerKeyContentType) == headerValContentTypeBinaryOctetStream
}

func logRequest(ctx context.Context, w io.Writer, req *http.Request) {
	fmt.Fprintln(w, "")
	fmt.Fprint(w, "    -------------------------- ")
	fmt.Fprint(w, "GOISILON HTTP REQUEST")
	fmt.Fprintln(w, " -------------------------")
	buf, err := httputil.DumpRequest(req, !isBinOctetBody(req.Header))
	if err != nil {
		return
	}
	WriteIndented(w, buf)
	fmt.Fprintln(w)
}

func logResponse(ctx context.Context, res *http.Response) {
	w := &bytes.Buffer{}

	fmt.Fprintln(w)
	fmt.Fprint(w, "    -------------------------- ")
	fmt.Fprint(w, "GOISILON HTTP RESPONSE")
	fmt.Fprintln(w, " -------------------------")

	buf, err := httputil.DumpResponse(res, !isBinOctetBody(res.Header))
	if err != nil {
		return
	}

	bw := &bytes.Buffer{}
	WriteIndented(bw, buf)

	scanner := bufio.NewScanner(bw)
	for {
		if !scanner.Scan() {
			break
		}
		fmt.Fprintln(w, scanner.Text())
	}

	log.Debug(ctx, w.String())
}

// WriteIndentedN indents all lines n spaces.
func WriteIndentedN(w io.Writer, b []byte, n int) error {
	s := bufio.NewScanner(bytes.NewReader(b))
	if !s.Scan() {
		return nil
	}
	l := s.Text()
	for {
		for x := 0; x < n; x++ {
			if _, err := fmt.Fprint(w, " "); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(w, l); err != nil {
			return err
		}
		if !s.Scan() {
			break
		}
		l = s.Text()
		if _, err := fmt.Fprint(w, "\n"); err != nil {
			return err
		}
	}
	return nil
}

// WriteIndented indents all lines four spaces.
func WriteIndented(w io.Writer, b []byte) error {
	return WriteIndentedN(w, b, 4)
}
