package stdlib

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/codedellemc/gournal"
)

func TestStdLibAppenderNoFields(t *testing.T) {
	gournal.Info(ctx(), "Hello %s\n", "Bob")
}

func TestStdLibAppenderWithField(t *testing.T) {
	gournal.WithField("size", 2).Info(ctx(), "Hello %s\n", "Alice")
}

func TestStdLibAppenderWithFields(t *testing.T) {
	gournal.WithFields(map[string]interface{}{
		"size":     1,
		"location": "Austin",
	}).Warn(ctx(), "Hello %s\n", "Mary")
}

func TestStdLibAppenderPanic(t *testing.T) {

	defer func() {
		r := recover()
		assert.NotNil(t, r, "no panic")
		assert.IsType(t, r, "")
		assert.Equal(t, "Hello Bob\n", r)
	}()

	gournal.Panic(ctx(), "Hello %s\n", "Bob")
}

func ctx() gournal.Context {
	ctx := gournal.Background()
	ctx = gournal.WithValue(ctx, gournal.LevelKey(), gournal.InfoLevel)
	ctx = gournal.WithValue(ctx, gournal.AppenderKey(), New())
	return ctx
}
