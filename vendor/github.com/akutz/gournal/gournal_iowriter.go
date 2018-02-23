package gournal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
)

// NewAppender returns an Appender that writes to os.Stdout.
func NewAppender() Appender {
	return &appender{os.Stdout}
}

// NewAppenderWithOptions returns an Appender that writes to the provided
// io.Writer object.
func NewAppenderWithOptions(w io.Writer) Appender {
	return &appender{w}
}

type appender struct {
	w io.Writer
}

func (a *appender) Append(
	ctx context.Context,
	lvl Level,
	fields map[string]interface{},
	msg string) {

	var (
		w        = a.w
		panicBuf *bytes.Buffer
	)

	if lvl == PanicLevel {
		// the base size of the buffer is:
		//
		//   8 - len("[PANIC] ")
		//   x - len(msg)
		//   1 - len("\n")
		//
		// then add 64 bytes for every possible field. this may be more than
		// is required, or it may be less. but by pre-allocating more than is
		// required the buffer will not need to double in size if additional
		// capacity is needed to write the fields object
		//panicBufLen := 8 + len(msg) + (len(fields) * 64)

		panicBuf = &bytes.Buffer{} //bytes.NewBuffer(make([]byte, panicBufLen))
		w = io.MultiWriter(a.w, panicBuf)
	}

	if len(fields) == 0 {
		fmt.Fprintf(w, "[%s] %s\n", lvl, msg)
	} else {
		fmt.Fprintf(w, "[%s] %s %v\n", lvl, msg, fields)
	}

	if lvl == FatalLevel {
		os.Exit(1)
	}

	if lvl == PanicLevel {
		panic(panicBuf.String())
	}
}
