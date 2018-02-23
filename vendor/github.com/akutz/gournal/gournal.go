/*
Package gournal (pronounced "Journal") is a Context-aware logging framework.

Gournal introduces the Google Context type (https://blog.golang.org/context) as
a first-class parameter to all common log functions such as Info, Debug, etc.

Instead of being Yet Another Go Log library, Gournal actually takes its
inspiration from the Simple Logging Facade for Java (SLF4J). Gournal is not
attempting to replace anyone's favorite logger, rather existing logging
frameworks such as Logrus, Zap, etc. can easily participate as a Gournal
Appender.

For more information on Gournal's features or how to use it, please refer
to the project's README file or https://github.com/akutz/gournal.
*/
package gournal

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var debug, _ = strconv.ParseBool(os.Getenv("GOURNAL_DEBUG"))

var (
	// ErrorKey defines the key when adding errors using WithError.
	ErrorKey = "error"

	// DefaultLevel is used when a Level is not present in a Context.
	DefaultLevel = ErrorLevel

	// DefaultAppender is used when an Appender is not present in a Context.
	DefaultAppender = NewAppender()

	// DefaultContext is used when a log method is invoked with a nil Context.
	DefaultContext = context.Background()
)

type contextKey uint8

const (
	levelKeyC contextKey = iota
	fieldsKeyC
	appenderKeyC
)

var (
	levelKey    interface{} = levelKeyC
	fieldsKey   interface{} = fieldsKeyC
	appenderKey interface{} = appenderKeyC
)

// LevelKey returns the Context key used for storing and retrieving the log
// level.
func LevelKey() interface{} {
	return levelKey
}

// FieldsKey returns the Context key for storing and retrieving Context-specific
// data that is appended along with each log entry. Three different types of
// data are inspected for this context key:
//
//     * map[string]interface{}
//
//     * func() map[string]interface{}
//
//     * func(ctx context.Context,
//            lvl Level,
//            fields map[string]interface{},
//            msg string) map[string]interface{}
func FieldsKey() interface{} {
	return fieldsKey
}

// AppenderKey returns the Context key for storing and retrieving an Appender
// object.
func AppenderKey() interface{} {
	return appenderKey
}

// Level is a log level.
type Level uint8

// These are the different logging levels.
const (
	// UnknownLevel is an unknown level.
	UnknownLevel Level = iota

	// PanicLevel level, highest level of severity. Logs and then calls panic
	// with the message passed to Debug, Info, ...
	PanicLevel

	// FatalLevel level. Logs and then calls os.Exit(1). It will exit even
	// if the logging level is set to Panic.
	FatalLevel

	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel

	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel

	// InfoLevel level. General operational entries about what's going on
	// inside the application.
	InfoLevel

	// DebugLevel level. Usually only enabled when debugging. Very verbose
	// logging.
	DebugLevel

	levelCount
)

const (
	unknownLevelStr = "UNKNOWN"
	panicLevelStr   = "PANIC"
	fatalLevelStr   = "FATAL"
	errorLevelStr   = "ERROR"
	warnLevelStr    = "WARN"
	warningLevelStr = "WARNING"
	infoLevelStr    = "INFO"
	debugLevelStr   = "DEBUG"
)

var (
	lvlValsToStrs = [levelCount]string{
		unknownLevelStr,
		panicLevelStr,
		fatalLevelStr,
		errorLevelStr,
		warnLevelStr,
		infoLevelStr,
		debugLevelStr,
	}
)

// String returns string representation of a Level.
func (level Level) String() string {
	if level < PanicLevel || level >= levelCount {
		return lvlValsToStrs[UnknownLevel]
	}
	return lvlValsToStrs[level]
}

// ParseLevel parses a string and returns its constant.
func ParseLevel(lvl string) Level {
	switch {
	case strings.EqualFold(lvl, debugLevelStr):
		return DebugLevel
	case strings.EqualFold(lvl, infoLevelStr):
		return InfoLevel
	case strings.EqualFold(lvl, warnLevelStr),
		strings.EqualFold(lvl, warningLevelStr):
		return WarnLevel
	case strings.EqualFold(lvl, errorLevelStr):
		return ErrorLevel
	case strings.EqualFold(lvl, fatalLevelStr):
		return FatalLevel
	case strings.EqualFold(lvl, panicLevelStr):
		return PanicLevel
	}
	return UnknownLevel
}

// Logger provides backwards-compatibility for code that does not yet use
// context-aware logging.
type Logger interface {

	// Debug emits a log entry at the DEBUG level.
	Debug(msg string, args ...interface{})

	// Info emits a log entry at the INFO level.
	Info(msg string, args ...interface{})

	// Print emits a log entry at the INFO level.
	Print(msg string, args ...interface{})

	// Warn emits a log entry at the WARN level.
	Warn(msg string, args ...interface{})

	// Error emits a log entry at the ERROR level.
	Error(msg string, args ...interface{})

	// Fatal emits a log entry at the FATAL level.
	Fatal(msg string, args ...interface{})

	// Panic emits a log entry at the PANIC level.
	Panic(msg string, args ...interface{})
}

// New returns a Logger for the provided context.
func New(ctx context.Context) Logger {
	return &logger{ctx}
}

type logger struct {
	ctx context.Context
}

func (l *logger) Debug(msg string, args ...interface{}) {
	Debug(l.ctx, msg, args...)
}

func (l *logger) Info(msg string, args ...interface{}) {
	Info(l.ctx, msg, args...)
}

func (l *logger) Print(msg string, args ...interface{}) {
	Print(l.ctx, msg, args...)
}

func (l *logger) Warn(msg string, args ...interface{}) {
	Warn(l.ctx, msg, args...)
}

func (l *logger) Error(msg string, args ...interface{}) {
	Error(l.ctx, msg, args...)
}

func (l *logger) Fatal(msg string, args ...interface{}) {
	Fatal(l.ctx, msg, args...)
}

func (l *logger) Panic(msg string, args ...interface{}) {
	Panic(l.ctx, msg, args...)
}

// Entry is the interface for types that contain information to be emmitted
// to a log appender.
type Entry interface {

	// WithField adds a single field to the Entry. The provided key will
	// override an existing, equivalent key in the Entry.
	WithField(key string, value interface{}) Entry

	// WithFields adds a map to the Entry. Keys in the provided map will
	// override existing, equivalent keys in the Entry.
	WithFields(fields map[string]interface{}) Entry

	// WithError adds the provided error to the Entry using the ErrorKey value
	// as the key.
	WithError(err error) Entry

	// Debug emits a log entry at the DEBUG level.
	Debug(ctx context.Context, msg string, args ...interface{})

	// Info emits a log entry at the INFO level.
	Info(ctx context.Context, msg string, args ...interface{})

	// Print emits a log entry at the INFO level.
	Print(ctx context.Context, msg string, args ...interface{})

	// Warn emits a log entry at the WARN level.
	Warn(ctx context.Context, msg string, args ...interface{})

	// Error emits a log entry at the ERROR level.
	Error(ctx context.Context, msg string, args ...interface{})

	// Fatal emits a log entry at the FATAL level.
	Fatal(ctx context.Context, msg string, args ...interface{})

	// Panic emits a log entry at the PANIC level.
	Panic(ctx context.Context, msg string, args ...interface{})
}

// Appender is the interface that must be implemented by the logging frameworks
// which are members of the Gournal facade.
type Appender interface {

	// Append is implemented by logging frameworks to accept the log entry
	// at the provided level, its message, and its associated field data.
	Append(
		ctx context.Context,
		lvl Level,
		fields map[string]interface{},
		msg string)
}

// WithField adds a single field to the Entry. The provided key will override
// an existing, equivalent key in the Entry.
func WithField(key string, value interface{}) Entry {
	return &entry{map[string]interface{}{key: value}}
}

// WithFields adds a map to the Entry. Keys in the provided map will override
// existing, equivalent keys in the Entry.
func WithFields(fields map[string]interface{}) Entry {
	return &entry{fields}
}

// WithError adds the provided error to the Entry using the ErrorKey value
// as the key.
func WithError(err error) Entry {
	return &entry{map[string]interface{}{ErrorKey: err.Error()}}
}

// Debug emits a log entry at the DEBUG level.
func Debug(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, DebugLevel, nil, msg, args...)
}

// Info emits a log entry at the INFO level.
func Info(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, InfoLevel, nil, msg, args...)
}

// Print emits a log entry at the INFO level.
func Print(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, InfoLevel, nil, msg, args...)
}

// Warn emits a log entry at the WARN level.
func Warn(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, WarnLevel, nil, msg, args...)
}

// Error emits a log entry at the ERROR level.
func Error(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, ErrorLevel, nil, msg, args...)
}

// Fatal emits a log entry at the FATAL level.
func Fatal(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, FatalLevel, nil, msg, args...)
}

// Panic emits a log entry at the PANIC level.
func Panic(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, PanicLevel, nil, msg, args...)
}

func sendToAppender(
	ctx context.Context,
	lvl Level,
	fields map[string]interface{},
	msg string,
	args ...interface{}) {

	if ctx == nil {
		ctx = DefaultContext
	}

	// do not append if the provided log level is less than that of the
	// provided context's log level
	if getLevel(ctx) < lvl {
		return
	}

	// do not proceed without an appender
	a := getAppender(ctx)

	// format the message with args if any
	if len(msg) == 0 && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if len(msg) > 0 && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	// grab any of the context fields to append alongside each new log entry
	inspectCustomCtxFields(ctx, lvl, &fields, msg)

	if debug {
		if len(fields) == 0 {
			fmt.Fprintf(os.Stderr,
				"GOURNAL: append: a=%T, lvl=%s, msg=%s\n",
				a, lvl, msg)
		} else {
			fmt.Fprintf(os.Stderr,
				"GOURNAL: append: a=%T, lvl=%s, msg=%s, fields=%v\n",
				a, lvl, msg, fields)
		}
	}

	a.Append(ctx, lvl, fields, msg)
}

func getLevel(ctx context.Context) Level {
	if ctx == nil {
		return DefaultLevel
	}

	if v, ok := ctx.Value(levelKey).(Level); ok {
		return v
	}
	return DefaultLevel
}

func getAppender(ctx context.Context) Appender {
	if ctx == nil {
		return DefaultAppender
	}

	if ctx == DefaultContext {
		return DefaultAppender
	}

	if v, ok := ctx.Value(appenderKey).(Appender); ok {
		return v
	}

	return DefaultAppender
}

func inspectCustomCtxFields(
	ctx context.Context,
	lvl Level,
	fields *map[string]interface{},
	msg string) {

	switch tv := ctx.Value(fieldsKey).(type) {
	case map[string]interface{}:
		swapFields(fields, &tv)
	case func() map[string]interface{}:
		ctxFields := tv()
		swapFields(fields, &ctxFields)
	case func(
		ctx context.Context,
		lvl Level,
		fields map[string]interface{},
		msg string) map[string]interface{}:

		ctxFields := tv(ctx, lvl, *fields, msg)
		swapFields(fields, &ctxFields)
	}
}

func swapFields(appendFields, ctxFields *map[string]interface{}) {
	if len(*ctxFields) == 0 {
		return
	}

	if len(*appendFields) == 0 {
		*appendFields = *ctxFields
		return
	}

	for k, v := range *ctxFields {
		(*appendFields)[k] = v
	}
}

type entry struct {
	fields map[string]interface{}
}

func (e *entry) WithField(key string, value interface{}) Entry {
	e.fields[key] = value
	return e
}
func (e *entry) WithFields(fields map[string]interface{}) Entry {
	for k, v := range fields {
		e.fields[k] = v
	}
	return e
}
func (e *entry) WithError(err error) Entry {
	e.fields[ErrorKey] = err.Error()
	return e
}

func (e *entry) Debug(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, DebugLevel, e.fields, msg, args...)
}

func (e *entry) Info(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, InfoLevel, e.fields, msg, args...)
}

func (e *entry) Print(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, InfoLevel, e.fields, msg, args...)
}

func (e *entry) Warn(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, WarnLevel, e.fields, msg, args...)
}

func (e *entry) Error(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, ErrorLevel, e.fields, msg, args...)
}

func (e *entry) Fatal(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, FatalLevel, e.fields, msg, args...)
}

func (e *entry) Panic(ctx context.Context, msg string, args ...interface{}) {
	sendToAppender(ctx, PanicLevel, e.fields, msg, args...)
}
