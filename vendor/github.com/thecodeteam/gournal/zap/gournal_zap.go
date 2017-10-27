// Package zap provides a Zap logger that implements the Gournal Appender
// interface.
package zap

import (
	"context"
	"fmt"
	"time"

	"github.com/uber-go/zap"

	"github.com/thecodeteam/gournal"
)

type appender struct {
	logger zap.Logger
}

// New returns a logrus logger that implements the Gournal Appender interface.
func New() gournal.Appender {
	return &appender{zap.New(zap.NewJSONEncoder())}
}

// NewWithOptions returns a zap logger that implements the Gournal Appender
// interface.
func NewWithOptions(enc zap.Encoder, opts ...zap.Option) gournal.Appender {
	return &appender{zap.New(enc, opts...)}
}

func (a *appender) Append(
	ctx context.Context,
	lvl gournal.Level,
	fields map[string]interface{},
	msg string) {

	zapLvl := lvlTranslator[lvl]

	if len(fields) == 0 {
		a.logger.Log(zapLvl, msg)
		return
	}

	zapFields := make([]zap.Field, len(fields))

	i := 0
	for k, v := range fields {
		switch tv := v.(type) {
		case zap.LogMarshaler:
			zapFields[i] = zap.Marshaler(k, tv)
		case bool:
			zapFields[i] = zap.Bool(k, tv)
		case []byte:
			zapFields[i] = zap.Base64(k, tv)
		case float64:
			zapFields[i] = zap.Float64(k, tv)
		case int:
			zapFields[i] = zap.Int(k, tv)
		case int64:
			zapFields[i] = zap.Int64(k, tv)
		case uint:
			zapFields[i] = zap.Uint(k, tv)
		case uint64:
			zapFields[i] = zap.Uint64(k, tv)
		case string:
			zapFields[i] = zap.String(k, tv)
		case fmt.Stringer:
			zapFields[i] = zap.String(k, tv.String())
		case time.Time:
			zapFields[i] = zap.Time(k, tv)
		case error:
			zapFields[i] = zap.Error(tv)
		case time.Duration:
			zapFields[i] = zap.Duration(k, tv)
		default:
			zapFields[i] = zap.Object(k, tv)
		}

		i++
	}

	a.logger.Log(zapLvl, msg, zapFields...)
}

var lvlTranslator = map[gournal.Level]zap.Level{
	gournal.DebugLevel: zap.DebugLevel,
	gournal.InfoLevel:  zap.InfoLevel,
	gournal.WarnLevel:  zap.WarnLevel,
	gournal.ErrorLevel: zap.ErrorLevel,
	gournal.FatalLevel: zap.FatalLevel,
	gournal.PanicLevel: zap.PanicLevel,
}
