package zap

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/codedellemc/gournal"
)

func TestZapAppenderNoFields(t *testing.T) {
	gournal.Info(ctx(), "Hello %s", "Bob")
}

func TestZapAppenderWithField(t *testing.T) {
	gournal.WithField("size", 2).Info(ctx(), "Hello %s", "Alice")
}

func TestZapAppenderWithFields(t *testing.T) {
	gournal.WithFields(map[string]interface{}{
		"size":     1,
		"location": "Austin",
	}).Warn(ctx(), "Hello %s", "Mary")
}

func TestZapAppenderPanic(t *testing.T) {

	defer func() {
		r := recover()
		assert.NotNil(t, r, "no panic")
		assert.IsType(t, "", r)
		assert.Equal(t, "Hello Bob", r)
	}()

	gournal.Panic(ctx(), "Hello %s", "Bob")
}

func ctx() gournal.Context {
	ctx := gournal.Background()
	ctx = gournal.WithValue(ctx, gournal.LevelKey(), gournal.InfoLevel)
	ctx = gournal.WithValue(ctx, gournal.AppenderKey(), New())
	return ctx
}
