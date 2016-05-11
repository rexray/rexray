package context

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/emccode/libstorage/api/types"
)

const (
	keyFieldOffset = 1000
)

func (ctx *lsc) ctxFields() map[string]interface{} {
	f := map[string]interface{}{}
	for k := Key(keyLoggable - 1); k > keyEOF; k-- {
		val := ctx.Value(k)
		switch tv := val.(type) {
		case types.Client:
			if tv.OS() != nil {
				f["osDriver"] = tv.OS().Name()
			}
			if tv.Storage() != nil {
				f["storageDriver"] = tv.Storage().Name()
			}
			if tv.Integration() != nil {
				f["integrationDriver"] = tv.Integration().Name()
			}
		case types.ContextLoggerFieldAware:
			k, v := tv.ContextLoggerField()
			f[k] = v
		case types.ContextLoggerFieldsAware:
			for k, v := range tv.ContextLoggerFields() {
				f[k] = v
			}
		case string:
			f[k.String()] = tv
		case hasName:
			f[k.String()] = tv.Name()
		case hasID:
			f[k.String()] = tv.ID()
		case fmt.Stringer:
			f[k.String()] = tv.String()
		case bool, uint, uint32, uint64, int, int32, int64, float32, float64:
			f[k.String()] = fmt.Sprintf("%v", tv)
		}
	}
	return f
}

func (ctx *lsc) addkeyFieldOffsetFieldsToEntry(entry *log.Entry) {
	fields := ctx.ctxFields()
	for k, v := range fields {
		entry.Data[k] = v
	}
}

func (ctx *lsc) logger() log.FieldLogger {
	return ctx.Value(LoggerKey).(log.FieldLogger)
}

func (ctx *lsc) WithField(key string, value interface{}) *log.Entry {
	entry := ctx.logger().WithField(key, value)
	ctx.addkeyFieldOffsetFieldsToEntry(entry)
	return entry
}
func (ctx *lsc) WithFields(fields log.Fields) *log.Entry {
	entry := ctx.logger().WithFields(fields)
	ctx.addkeyFieldOffsetFieldsToEntry(entry)
	return entry
}
func (ctx *lsc) WithError(err error) *log.Entry {
	entry := ctx.logger().WithError(err)
	ctx.addkeyFieldOffsetFieldsToEntry(entry)
	return entry
}

func (ctx *lsc) Debugf(format string, args ...interface{}) {
	ctx.Debug(fmt.Sprintf(format, args...))
}

func (ctx *lsc) Infof(format string, args ...interface{}) {
	ctx.Info(fmt.Sprintf(format, args...))
}

func (ctx *lsc) Printf(format string, args ...interface{}) {
	ctx.Print(fmt.Sprintf(format, args...))
}

func (ctx *lsc) Warnf(format string, args ...interface{}) {
	ctx.Warn(fmt.Sprintf(format, args...))
}

func (ctx *lsc) Warningf(format string, args ...interface{}) {
	ctx.Warning(fmt.Sprintf(format, args...))
}

func (ctx *lsc) Errorf(format string, args ...interface{}) {
	ctx.Error(fmt.Sprintf(format, args...))
}

func (ctx *lsc) Fatalf(format string, args ...interface{}) {
	ctx.Fatal(fmt.Sprintf(format, args...))
}

func (ctx *lsc) Panicf(format string, args ...interface{}) {
	ctx.Panic(fmt.Sprintf(format, args...))
}

func (ctx *lsc) Debug(args ...interface{}) {
	ctx.logger().WithFields(ctx.ctxFields()).Debug(args...)
}

func (ctx *lsc) Info(args ...interface{}) {
	ctx.logger().WithFields(ctx.ctxFields()).Info(args...)
}

func (ctx *lsc) Print(args ...interface{}) {
	ctx.logger().WithFields(ctx.ctxFields()).Print(args...)
}

func (ctx *lsc) Warn(args ...interface{}) {
	ctx.logger().WithFields(ctx.ctxFields()).Warn(args...)
}

func (ctx *lsc) Warning(args ...interface{}) {
	ctx.logger().WithFields(ctx.ctxFields()).Warning(args...)
}

func (ctx *lsc) Error(args ...interface{}) {
	ctx.logger().WithFields(ctx.ctxFields()).Error(args...)
}

func (ctx *lsc) Fatal(args ...interface{}) {
	ctx.logger().WithFields(ctx.ctxFields()).Fatal(args...)
}

func (ctx *lsc) Panic(args ...interface{}) {
	ctx.logger().WithFields(ctx.ctxFields()).Panic(args...)
}

func (ctx *lsc) Debugln(args ...interface{}) {
	ctx.logger().Debug(args...)
}

func (ctx *lsc) Infoln(args ...interface{}) {
	ctx.logger().Info(args...)
}

func (ctx *lsc) Println(args ...interface{}) {
	ctx.logger().Print(args...)
}

func (ctx *lsc) Warnln(args ...interface{}) {
	ctx.logger().Warn(args...)
}

func (ctx *lsc) Warningln(args ...interface{}) {
	ctx.logger().Warning(args...)
}

func (ctx *lsc) Errorln(args ...interface{}) {
	ctx.logger().Error(args...)
}

func (ctx *lsc) Fatalln(args ...interface{}) {
	ctx.logger().Fatal(args...)
}

func (ctx *lsc) Panicln(args ...interface{}) {
	ctx.logger().Panic(args...)
}
