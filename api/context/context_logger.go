package context

import (
	log "github.com/Sirupsen/logrus"
)

func (ctx *lsc) WithField(key string, value interface{}) *log.Entry {
	return ctx.Logger.WithField(key, value)
}
func (ctx *lsc) WithFields(fields log.Fields) *log.Entry {
	return ctx.Logger.WithFields(fields)
}
func (ctx *lsc) WithError(err error) *log.Entry {
	return ctx.Logger.WithError(err)
}

func (ctx *lsc) Debugf(format string, args ...interface{}) {
	ctx.Logger.Debugf(format, args...)
}
func (ctx *lsc) Infof(format string, args ...interface{}) {
	ctx.Logger.Infof(format, args...)
}
func (ctx *lsc) Printf(format string, args ...interface{}) {
	ctx.Logger.Printf(format, args...)
}
func (ctx *lsc) Warnf(format string, args ...interface{}) {
	ctx.Logger.Warnf(format, args...)
}
func (ctx *lsc) Warningf(format string, args ...interface{}) {
	ctx.Logger.Warningf(format, args...)
}
func (ctx *lsc) Errorf(format string, args ...interface{}) {
	ctx.Logger.Errorf(format, args...)
}
func (ctx *lsc) Fatalf(format string, args ...interface{}) {
	ctx.Logger.Fatalf(format, args...)
}
func (ctx *lsc) Panicf(format string, args ...interface{}) {
	ctx.Logger.Panicf(format, args...)
}

func (ctx *lsc) Debug(args ...interface{}) {
	ctx.Logger.Debug(args...)
}
func (ctx *lsc) Info(args ...interface{}) {
	ctx.Logger.Info(args...)
}
func (ctx *lsc) Print(args ...interface{}) {
	ctx.Logger.Print(args...)
}
func (ctx *lsc) Warn(args ...interface{}) {
	ctx.Logger.Warn(args...)
}
func (ctx *lsc) Warning(args ...interface{}) {
	ctx.Logger.Warning(args...)
}
func (ctx *lsc) Error(args ...interface{}) {
	ctx.Logger.Error(args...)
}
func (ctx *lsc) Fatal(args ...interface{}) {
	ctx.Logger.Fatal(args...)
}
func (ctx *lsc) Panic(args ...interface{}) {
	ctx.Logger.Panic(args...)
}

func (ctx *lsc) Debugln(args ...interface{}) {
	ctx.Logger.Debugln(args...)
}
func (ctx *lsc) Infoln(args ...interface{}) {
	ctx.Logger.Infoln(args...)
}
func (ctx *lsc) Println(args ...interface{}) {
	ctx.Logger.Println(args...)
}
func (ctx *lsc) Warnln(args ...interface{}) {
	ctx.Logger.Warnln(args...)
}
func (ctx *lsc) Warningln(args ...interface{}) {
	ctx.Logger.Warningln(args...)
}
func (ctx *lsc) Errorln(args ...interface{}) {
	ctx.Logger.Errorln(args...)
}
func (ctx *lsc) Fatalln(args ...interface{}) {
	ctx.Logger.Fatalln(args...)
}
func (ctx *lsc) Panicln(args ...interface{}) {
	ctx.Logger.Panicln(args...)
}
