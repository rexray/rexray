package util

import (
	"bufio"
	"io"
)

type writerWrapper struct {
	f func(msg string, args ...interface{})
	w io.Writer
}

// NewWriterFor returns an io.Writer that will write data using the
// provided function.
func NewWriterFor(f func(msg string, args ...interface{})) io.Writer {
	l := &writerWrapper{f: f}
	r, w := io.Pipe()
	l.w = w
	go func() {
		scan := bufio.NewScanner(r)
		for scan.Scan() {
			f(scan.Text())
		}
	}()
	return l
}

func (w *writerWrapper) Write(data []byte) (int, error) {
	return w.w.Write(data)
}
