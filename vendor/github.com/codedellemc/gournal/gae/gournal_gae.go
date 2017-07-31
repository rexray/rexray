// Package gae provides a Google App Engine logger that implements the Gournal
// Appender interface.
package gae

import (
	"fmt"

	gae "google.golang.org/appengine/log"

	"github.com/codedellemc/gournal"
)

type appender int

var theAppender = appender(0)

// New returns a Google App Engine logger that implements the Gournal Appender
// interface.
func New() gournal.Appender {
	return theAppender
}

func (a appender) Append(
	ctx gournal.Context,
	lvl gournal.Level,
	fields map[string]interface{},
	msg string) {

	if len(fields) > 0 {
		msg = fmt.Sprintf("%s %v", msg, fields)
	}

	switch lvl {
	case gournal.DebugLevel:
		gae.Debugf(ctx, msg)
	case gournal.InfoLevel:
		gae.Infof(ctx, msg)
	case gournal.WarnLevel:
		gae.Warningf(ctx, msg)
	case gournal.ErrorLevel:
		gae.Errorf(ctx, msg)
	case gournal.FatalLevel, gournal.PanicLevel:
		gae.Criticalf(ctx, msg)
	}
}
