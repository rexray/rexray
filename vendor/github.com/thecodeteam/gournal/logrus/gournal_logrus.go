// Package logrus provides a Logrus logger that implements the Gournal Appender
// interface.
package logrus

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"

	"github.com/thecodeteam/gournal"
)

type appender struct {
	logger *logrus.Logger
}

// New returns a logrus logger that implements the Gournal Appender interface.
func New() gournal.Appender {
	return &appender{logrus.New()}
}

// NewWithOptions returns a logrus logger that implements the Gournal Appender
// interface.
func NewWithOptions(
	out io.Writer,
	lvl logrus.Level,
	formatter logrus.Formatter) gournal.Appender {

	return &appender{&logrus.Logger{Out: out, Level: lvl, Formatter: formatter}}
}

func (a *appender) Append(
	ctx context.Context,
	lvl gournal.Level,
	fields map[string]interface{},
	msg string) {

	if len(fields) == 0 {
		switch lvl {
		case gournal.DebugLevel:
			a.logger.Debugf(msg)
		case gournal.InfoLevel:
			a.logger.Infof(msg)
		case gournal.WarnLevel:
			a.logger.Warnf(msg)
		case gournal.ErrorLevel:
			a.logger.Errorf(msg)
		case gournal.FatalLevel:
			a.logger.Fatalf(msg)
		case gournal.PanicLevel:
			a.logger.Panicf(msg)
		}
		return
	}

	switch lvl {
	case gournal.DebugLevel:
		a.logger.WithFields(fields).Debugf(msg)
	case gournal.InfoLevel:
		a.logger.WithFields(fields).Infof(msg)
	case gournal.WarnLevel:
		a.logger.WithFields(fields).Warnf(msg)
	case gournal.ErrorLevel:
		a.logger.WithFields(fields).Errorf(msg)
	case gournal.FatalLevel:
		a.logger.WithFields(fields).Fatalf(msg)
	case gournal.PanicLevel:
		a.logger.WithFields(fields).Panicf(msg)
	}
}
