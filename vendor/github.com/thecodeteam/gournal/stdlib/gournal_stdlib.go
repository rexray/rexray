// Package stdlib provides a StdLib logger that implements the Gournal Appender
// interface.
package stdlib

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/thecodeteam/gournal"
)

// New returns a stdlib logger that implements the Gournal Appender interface.
func New() gournal.Appender {
	return &appender{
		log.New(os.Stdout, "",
			log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile),
	}
}

// NewWithOptions returns a stdlib logger that implements the Gournal Appender
// interface.
func NewWithOptions(out io.Writer, prefix string, flags int) gournal.Appender {
	return &appender{log.New(out, prefix, flags)}
}

type appender struct {
	logger *log.Logger
}

func (a *appender) Append(
	ctx context.Context,
	lvl gournal.Level,
	fields map[string]interface{},
	msg string) {

	var logf func(string, ...interface{})

	switch lvl {
	case gournal.PanicLevel:
		logf = a.logger.Panicf
	case gournal.FatalLevel:
		logf = a.logger.Fatalf
	default:
		logf = a.logger.Printf
	}

	if len(fields) == 0 {
		logf(msg)
		return
	}

	logf(msg, fields)
}
