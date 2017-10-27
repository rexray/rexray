package logrus

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/thecodeteam/gournal"
)

func TestLogrusAppenderNoFields(t *testing.T) {
	gournal.Info(ctx(), "Hello %s", "Bob")
}

func TestLogrusAppenderWithField(t *testing.T) {
	gournal.WithField("size", 2).Info(ctx(), "Hello %s", "Alice")
}

func TestLogrusAppenderWithFields(t *testing.T) {
	gournal.WithFields(map[string]interface{}{
		"size":     1,
		"location": "Austin",
	}).Warn(ctx(), "Hello %s", "Mary")
}

func TestLogrusAppenderPanic(t *testing.T) {

	defer func() {
		r := recover()
		assert.NotNil(t, r, "no panic")
		assert.IsType(t, &logrus.Entry{}, r)
		entry := r.(*logrus.Entry)
		assert.Equal(t, logrus.PanicLevel, entry.Level)
		assert.Equal(t, "Hello Bob", entry.Message)
	}()

	gournal.Panic(ctx(), "Hello %s", "Bob")
}

func ctx() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, gournal.LevelKey(), gournal.InfoLevel)
	ctx = context.WithValue(ctx, gournal.AppenderKey(), New())
	return ctx
}
